package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// main vars
	bbs_server_url string
	bbs_username   string
	bbs_password   string
	no_ssl_verify  = false
	description    = "GitHub CLI extension to analyze BitBucket Server for migration statistics"
	log_file       *os.File
	output_file    string
	page_limit     = 100
	// Create some colors and a spinner
	red  = color.New(color.FgRed).SprintFunc()
	cyan = color.New(color.FgCyan).SprintFunc()
	sp   = spinner.New(spinner.CharSets[2], 100*time.Millisecond)
	// Create the root cobra command
	rootCmd = &cobra.Command{
		Use:           "gh bbs-analyzer",
		Short:         description,
		Long:          description,
		Version:       "0.1.0",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          Process,
	}
)

// custom structs for JSON conversion
type BitBucketProjectResponse struct {
	Size          int                `json:"size"`
	Limit         int                `json:"limit"`
	IsLastPage    bool               `json:"isLastPage"`
	Values        []BitBucketProject `json:"values"`
	Start         int                `json:"start"`
	Filter        string             `json:"filter"`
	NextPageStart int                `json:"nextPageStart"`
}
type BitBucketRepositoryResponse struct {
	Size          int                   `json:"size"`
	Limit         int                   `json:"limit"`
	IsLastPage    bool                  `json:"isLastPage"`
	Values        []BitBucketRepository `json:"values"`
	Start         int                   `json:"start"`
	Filter        string                `json:"filter"`
	NextPageStart int                   `json:"nextPageStart"`
}
type BitBucketProject struct {
	Key    string `json:"key"`
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
	Type   string `json:"type"`
}
type BitBucketRepository struct {
	Slug          string           `json:"slug"`
	ID            int              `json:"id"`
	Name          string           `json:"name"`
	HierarchyId   string           `json:"hierarchyId"`
	ScmId         string           `json:"scmId"`
	State         string           `json:"state"`
	StatusMessage string           `json:"statusMessage"`
	Forkable      bool             `json:"forkable"`
	Public        bool             `json:"public"`
	Archived      bool             `json:"archived"`
	Project       BitBucketProject `json:"project"`
	Size          BitBucketRepositorySize
}
type BitBucketRepositorySize struct {
	Repository  int `json:"repository"`
	Attachments int `json:"attachments"`
}

func init() {

	// add flags here
	rootCmd.PersistentFlags().StringVarP(
		&bbs_server_url,
		"bbs-server-url",
		"s",
		"",
		"The full URL of the Bitbucket Server/Data Center to migrate from. E.g. http://bitbucket.contoso.com:7990",
	)
	rootCmd.PersistentFlags().StringVarP(
		&bbs_username,
		"bbs-username",
		"u",
		"",
		"The Bitbucket username of a user with site admin privileges. If not set will be read from BBS_USERNAME environment variable.",
	)
	rootCmd.PersistentFlags().StringVarP(
		&bbs_password,
		"bbs-password",
		"p",
		"",
		"The Bitbucket password of the user specified by --bbs-username. If not set will be read from BBS_PASSWORD environment variable.",
	)
	rootCmd.PersistentFlags().BoolVar(
		&no_ssl_verify,
		"no-ssl-verify",
		false,
		"Disables SSL verification when communicating with your Bitbucket Server/Data Center instance. All other migration steps will continue to verify SSL. If your Bitbucket instance has a self-signed SSL certificate then setting this flag will allow the migration archive to be exported.",
	)
	rootCmd.PersistentFlags().StringVarP(
		&output_file,
		"output-file",
		"o",
		"results.csv",
		"The file to output the results to.",
	)

	// add args here
	rootCmd.Args = cobra.MaximumNArgs(0)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		ExitOnError(err)
	}
}

func ExitOnError(err error) {
	if err != nil {
		rootCmd.PrintErrln(err.Error())
		os.Exit(1)
	}
}

func OutputFlags(key string, value string) {
	sep := ": "
	fmt.Println(fmt.Sprint(cyan(key), sep, value))
	Log(fmt.Sprint(key, sep, value))
}

func OutputNotice(message string) {
	Output(message, "default", false, false)
}

func OutputError(message string, exit bool) {
	sp.Stop()
	Output(message, "red", true, exit)
}

func Output(message string, color string, isErr bool, exit bool) {

	if isErr {
		message = fmt.Sprint("[ERROR] ", message)
	}
	Log(message)

	switch {
	case color == "red":
		message = red(message)
	}
	fmt.Println(message)
	if exit {
		fmt.Println("")
		os.Exit(1)
	}
}

func DebugAndStatus(message string) string {
	sp.Suffix = fmt.Sprint(
		" ",
		message,
	)
	return Debug(message)
}

func Debug(message string) string {
	Log(message)
	return message
}

func Log(message string) {
	if message != "" {
		message = fmt.Sprint(
			"[",
			time.Now().Format("2006-01-02 15:04:05"),
			"] ",
			message,
		)
	}
	_, err := log_file.WriteString(
		fmt.Sprintln(message),
	)
	if err != nil {
		fmt.Println(red("Unable to write to log file."))
		fmt.Println(red(err))
		os.Exit(1)
	}
}

func LF() {
	Output("", "default", false, false)
}

func LogLF() {
	Log("")
}

func Truncate(str string, limit int) string {
	lastSpaceIx := -1
	len := 0
	for i, r := range str {
		if unicode.IsSpace(r) {
			lastSpaceIx = i
		}
		len++
		if len >= limit {
			if lastSpaceIx != -1 {
				return fmt.Sprint(str[:lastSpaceIx], "...")
			} else {
				return fmt.Sprint(str[:limit], "...")
			}
		}
	}
	return str
}

func Process(cmd *cobra.Command, args []string) (err error) {

	// Create log file
	log_file, err = os.Create(fmt.Sprint(time.Now().Format("20060102_150401"), ".log"))
	if err != nil {
		return err
	}
	defer log_file.Close()

	LF()
	Debug("---- VALIDATING FLAGS & ENV VARS ----")

	// validate flags
	r, _ := regexp.Compile("^http(s|):(//|)")
	if !r.MatchString(bbs_server_url) {
		OutputError("BitBucket server url should contain http(s) prefix and does not.", true)
	}
	if bbs_server_url == "" {
		OutputError("A BitBucket server URL must be provided.", true)
	}

	// user environment variables if the flags aren't provided
	if bbs_username == "" {
		bbs_username_env, is_set := os.LookupEnv("BBS_USERNAME")
		if is_set {
			bbs_username = bbs_username_env
			Debug("BitBucket username set from Environment Variable BBS_USERNAME")
		} else {
			OutputError("A BitBucket username was not provided via --bbs-username or environment variable BBS_USERNAME", true)
		}
	}
	if bbs_password == "" {
		bbs_password_env, is_set := os.LookupEnv("BBS_PASSWORD")
		if is_set {
			bbs_password = bbs_password_env
			Debug("BitBucket password set from Environment Variable BBS_PASSWORD")
		} else {
			OutputError("A BitBucket password was not provided via --bbs-password or environment variable BBS_PASSWORD", true)
		}
	}

	// output flags for reference
	OutputFlags("BitBucket Server URL", bbs_server_url)
	OutputFlags("BitBucket Username", bbs_username)
	OutputFlags("BitBucket Password", "**********")
	OutputFlags("SSL Verification Disabled", strconv.FormatBool(no_ssl_verify))
	LF()

	// Start the spinner in the CLI
	sp.Start()

	// Get all projects
	projects, err := GetProjects([]BitBucketProject{}, 0)
	if err != nil {
		OutputError(fmt.Sprintf("Error looking up projects: %s", err), true)
	}

	// get all repos
	var repositories []BitBucketRepository
	for _, project := range projects {
		repositories, err = GetRepositories(project.Key, repositories, 0)
		if err != nil {
			OutputError(fmt.Sprintf("Error looking up repositories: %s", err), true)
		}
	}

	// get repo sizes
	total_size := 0
	for i, repository := range repositories {
		size, err := GetRepositorySize(repository)
		if err != nil {
			OutputError(fmt.Sprintf("Error looking up repository size: %s", err), true)
		}
		// set the size object on the repo
		repositories[i].Size = size
		total_size += size.Repository
	}

	// do some quick math
	display_size := fmt.Sprintf("%d B", total_size)
	if total_size >= 1000000000 {
		display_size = fmt.Sprintf("%d GB", total_size/1000000000)
	} else if total_size >= 1000000 {
		display_size = fmt.Sprintf("%d MB", total_size/1000000)
	} else if total_size >= 1000 {
		display_size = fmt.Sprintf("%d KB", total_size/1000)
	}

	// Stop the spinner in the CLI
	sp.Stop()

	OutputNotice(fmt.Sprintf("Projects found: %d", len(projects)))
	OutputNotice(fmt.Sprintf("Repositories found: %d", len(repositories)))
	OutputNotice(fmt.Sprintf("Total Size: %s", display_size))

	// Create output file
	out_file, err := os.Create(output_file)
	if err != nil {
		return err
	}
	defer out_file.Close()

	// write header
	_, err = out_file.WriteString(
		fmt.Sprintln("project,repository,size"),
	)
	if err != nil {
		OutputError("Error writing to output file.", true)
	}
	// write body
	for _, repository := range repositories {
		_, err = out_file.WriteString(
			fmt.Sprintln(fmt.Sprintf("%s,%s,%d", repository.Project.Key, repository.Slug, repository.Size.Repository)),
		)
		if err != nil {
			OutputError("Error writing to output file.", true)
		}
	}

	// always return
	return err
}

// pagination method for projects
func GetProjects(projects []BitBucketProject, start int) ([]BitBucketProject, error) {

	// get all projects
	endpoint := fmt.Sprintf("/projects?limit=%d&start=%d", page_limit, start)
	DebugAndStatus(fmt.Sprintf("Making HTTP request to %s", endpoint))
	data, err := BBSAPIRequest(endpoint, "GET")
	if err != nil {
		return projects, err
	}

	// convert response
	var response BitBucketProjectResponse
	Debug(fmt.Sprintf("Attempting to unmarshal response %s", data))
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return projects, err
	}

	// if first iteration executing, set
	// otherwise merge values
	if len(projects) == 0 {
		projects = response.Values
	} else {
		projects = append(response.Values, projects...)
	}

	// check for next page recursively
	if !response.IsLastPage {
		Debug("Not the last page. Recursively looking up next page.")
		projects, err = GetProjects(projects, response.NextPageStart)
	}

	return projects, err
}

// getting repo sizes
func GetRepositorySize(repository BitBucketRepository) (size BitBucketRepositorySize, err error) {

	// get repo size
	endpoint := fmt.Sprintf("/projects/%s/repos/%s/sizes", repository.Project.Key, repository.Slug)
	DebugAndStatus(fmt.Sprintf("Making HTTP request to: %s", endpoint))
	data, err := BBSRequest(endpoint, "GET")
	if err != nil {
		return size, err
	}

	// convert response
	Debug(fmt.Sprintf("Attempting to unmarshal response %s", data))
	err = json.Unmarshal([]byte(data), &size)
	if err != nil {
		return size, err
	}

	return size, err
}

// pagination method for repos
func GetRepositories(project string, repositories []BitBucketRepository, start int) ([]BitBucketRepository, error) {

	// get all projects
	endpoint := fmt.Sprintf("/projects/%s/repos?limit=%d&start=%d", project, page_limit, start)
	DebugAndStatus(fmt.Sprintf("Making HTTP request to /%s", endpoint))
	data, err := BBSAPIRequest(endpoint, "GET")
	if err != nil {
		return repositories, err
	}

	// convert response
	var response BitBucketRepositoryResponse
	Debug(fmt.Sprintf("Attempting to unmarshal response %s", data))
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return repositories, err
	}

	// if first iteration executing, set
	// otherwise merge values
	if len(repositories) == 0 {
		repositories = response.Values
	} else {
		repositories = append(response.Values, repositories...)
	}

	// check for next page recursively
	if !response.IsLastPage {
		Debug("Not the last page. Recursively looking up next page.")
		repositories, err = GetRepositories(project, repositories, response.NextPageStart)
	}

	return repositories, err
}

func BBSRequest(endpoint string, method string) (data string, err error) {
	return BBSHTTPRequest("", endpoint, method)
}

func BBSAPIRequest(endpoint string, method string) (data string, err error) {
	return BBSHTTPRequest("/rest/api/1.0", endpoint, method)
}

func BBSHTTPRequest(path string, endpoint string, method string) (data string, err error) {

	// set up endpoint
	url := fmt.Sprintf("%s%s%s", bbs_server_url, path, endpoint)
	Debug(fmt.Sprintf("Request URL: %s", url))

	// set SSL verification using inverse bool from flag
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: no_ssl_verify},
	}

	// create client
	client := &http.Client{Transport: tr}

	// create request and add authentication
	req, err := http.NewRequest(method, url, nil)
	req.SetBasicAuth(bbs_username, bbs_password)

	// perform request
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// defer body close so we can read it
	defer res.Body.Close()

	// error out if non-200 code received
	if res.StatusCode != http.StatusOK {
		return "", err
	}

	// read the body
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	// return the response and nil error
	return string(bodyBytes), err
}
