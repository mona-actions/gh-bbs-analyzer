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
	"sync"
	"time"
	"unicode"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	// main vars
	bbs_server_url string
	bbs_username   string
	bbs_password   string
	bbs_project    string
	no_ssl_verify  = false
	description    = "GitHub CLI extension to analyze BitBucket Server for migration statistics"
	log_file       *os.File
	output_file    string
	page_limit     = 100
	repositories   []BitBucketRepository
	threads        int
	total_size     = 0
	total_pr       = 0
	total_comments = 0
	waitGroup      sync.WaitGroup

	// Create some colors and a spinner
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
	sp     = spinner.New(spinner.CharSets[2], 100*time.Millisecond)
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
type BitBucketPullRequestResponse struct {
	Size          int                    `json:"size"`
	Limit         int                    `json:"limit"`
	IsLastPage    bool                   `json:"isLastPage"`
	Values        []BitBucketPullRequest `json:"values"`
	Start         int                    `json:"start"`
	Filter        string                 `json:"filter"`
	NextPageStart int                    `json:"nextPageStart"`
}
type BitBucketPullRequest struct {
	ID         int                            `json:"id"`
	Properties BitBucketPullRequestProperties `json:"properties"`
}
type BitBucketPullRequestProperties struct {
	CommentCount int `json:"commentCount"`
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
	PullRequests  []BitBucketPullRequest
	CommentCount  int
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
	rootCmd.PersistentFlags().StringVar(
		&bbs_project,
		"bbs-project",
		"",
		"A specific Bitbucket project instead of analyzing all proejcts.",
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
	rootCmd.PersistentFlags().IntVarP(
		&threads,
		"threads",
		"t",
		3,
		"Number of threads to process concurrently.",
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

func OutputWarning(message string) {
	Output(fmt.Sprint("[WARNING] ", message), "yellow", false, false)
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
	case color == "yellow":
		message = yellow(message)
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

	if threads > 10 {
		OutputError("Number of concurrent threads cannot be higher than 10.", true)
	} else if threads > 3 {
		OutputWarning("Number of concurrent threads is higher than 3. This could result in extreme load on your server.")
	}

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
	OutputFlags("Threads", fmt.Sprintf("%d", threads))
	Debug("---- PROCESSING API REQUESTS ----")

	// Start the spinner in the CLI
	sp.Start()

	// Get all projects
	var projects []BitBucketProject
	if bbs_project == "" {
		projects, err = GetProjects([]BitBucketProject{}, 0)
		if err != nil {
			OutputError(fmt.Sprintf("Error looking up projects: %s", err), true)
		}
	} else {
		project, err := GetProject(bbs_project)
		if err != nil {
			OutputError(fmt.Sprintf("Error looking up project: %s", err), true)
		}
		projects = append(projects, project)
	}

	// make sure there are projects to lookup
	if len(projects) == 0 {
		OutputError("No projects were found to look up repositories for.", true)
	}

	// get all repos
	for _, project := range projects {
		repositories, err = GetRepositories(project.Key, repositories, 0)
		if err != nil {
			OutputError(fmt.Sprintf("Error looking up repositories: %s", err), true)
		}
	}

	// set a temp var that we can batch through without effecting the original
	repositoriesToProcess := repositories
	batchThreads := threads
	batchNum := 1

	// do while we still have repos left to process
	for len(repositoriesToProcess) > 0 {

		// adjust number of threads in case our batch has less than num of threads
		repositoriesLeft := len(repositoriesToProcess)
		if repositoriesLeft < threads {
			batchThreads = repositoriesLeft
			Debug(
				fmt.Sprintf(
					"Setting number of threads to %d because there are only %d repositories left.",
					repositoriesLeft,
					repositoriesLeft,
				),
			)
		}

		DebugAndStatus(
			fmt.Sprintf(
				"Running repository analysis batch #%d (%d threads)...",
				batchNum,
				batchThreads,
			),
		)

		// get the next batch into new array and remove from processing array
		batch := repositoriesToProcess[:batchThreads]
		repositoriesToProcess = repositoriesToProcess[len(batch):]

		// add the number of wait groups needed
		waitGroup.Add(len(batch))

		// process threads
		for i := 0; i < len(batch); i++ {
			Debug(
				fmt.Sprintf(
					"Running thread %d of %d on repository '%s'",
					i+1,
					len(batch),
					batch[i].Name,
				),
			)
			go GetRepositoryStatistics(batch[i])
		}

		// wait for threads to finish
		waitGroup.Wait()
		batchNum++

	}

	// Stop the spinner in the CLI
	sp.Stop()

	// do some quick math
	display_size := fmt.Sprintf("%d B", total_size)
	b := 1024
	mb := b * b
	gb := mb * b
	if total_size >= gb {
		display_size = fmt.Sprintf("%d GB", total_size/gb)
	} else if total_size >= mb {
		display_size = fmt.Sprintf("%d MB", total_size/mb)
	} else if total_size >= b {
		display_size = fmt.Sprintf("%d KB", total_size/b)
	}

	LF()
	Debug("---- OUTPUT INFO ----")
	OutputFlags("Totals", "")
	OutputNotice(fmt.Sprintf("Projects: %d", len(projects)))
	OutputNotice(fmt.Sprintf("Repositories: %d", len(repositories)))
	OutputNotice(fmt.Sprintf("Pull Requests: %d", total_pr))
	OutputNotice(fmt.Sprintf("Comments: %d", total_comments))
	OutputNotice(fmt.Sprintf("Total Disk Size: %s", display_size))
	OutputNotice(fmt.Sprintf("Results File: ./%s", output_file))
	LF()
	Debug("---- WRITING TO CSV ----")

	// Create output file
	out_file, err := os.Create(output_file)
	if err != nil {
		return err
	}
	defer out_file.Close()

	// write header
	_, err = out_file.WriteString(
		fmt.Sprintln("project,repository,size,pull_requests,comments,archived,public"),
	)
	if err != nil {
		OutputError("Error writing to output file.", true)
	}
	// write body
	for _, repository := range repositories {
		_, err = out_file.WriteString(
			fmt.Sprintln(
				fmt.Sprintf(
					"%s,%s,%d,%d,%d,%s,%s",
					repository.Project.Key,
					repository.Slug,
					repository.Size.Repository,
					len(repository.PullRequests),
					repository.CommentCount,
					strconv.FormatBool(repository.Archived),
					strconv.FormatBool(repository.Public),
				),
			),
		)
		if err != nil {
			OutputError("Error writing to output file.", true)
		}
	}

	// always return
	return err
}

func GetRepositoryStatistics(repository BitBucketRepository) {
	// get repo size
	size, err := GetRepositorySize(repository)
	if err != nil {
		OutputError(fmt.Sprintf("Error looking up repository size: %s", err), false)
	}

	// get pull requests
	var pull_requests []BitBucketPullRequest
	pull_requests, err = GetPullRequests(repository, pull_requests, 0)
	if err != nil {
		OutputError(fmt.Sprintf("Error looking up repository pull requests: %s", err), false)
	}

	// get all pull request comments
	repository_comment_count := 0
	for _, pull_request := range pull_requests {
		repository_comment_count += pull_request.Properties.CommentCount
	}

	// add to repo information
	repository.Size = size
	repository.CommentCount = repository_comment_count
	repository.PullRequests = pull_requests

	// find index of repo in original list and overwite it
	idx := slices.IndexFunc(repositories, func(r BitBucketRepository) bool { return r.ID == repository.ID })
	if idx < 0 {
		OutputError(fmt.Sprintf("Error finding batch repository in original list: %s", repository.Name), false)
	} else {
		repositories[idx] = repository
	}

	// add to totals for quick analysis
	total_size += size.Repository
	total_comments += repository_comment_count
	total_pr += len(pull_requests)

	// finish this group
	waitGroup.Done()
}

func GetProject(project_key string) (BitBucketProject, error) {

	// get all projects
	var project BitBucketProject
	endpoint := fmt.Sprintf("/projects/%s", project_key)
	DebugAndStatus(fmt.Sprintf("Making HTTP request to %s", endpoint))
	data, err := BBSAPIRequest(endpoint, "GET")
	if err != nil {
		return project, err
	} else if data == "" {
		return project, fmt.Errorf("No data was returned from the project endpoint.")
	}

	// convert response
	var response BitBucketProject
	Debug(fmt.Sprintf("Attempting to unmarshal response data: %s", data))
	err = json.Unmarshal([]byte(data), &response)

	return response, err
}

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

func GetRepositorySize(repository BitBucketRepository) (size BitBucketRepositorySize, err error) {

	// get repo size
	endpoint := fmt.Sprintf("/projects/%s/repos/%s/sizes", repository.Project.Key, repository.Slug)
	data, err := BBSRequest(endpoint, "GET")
	if err != nil {
		return size, err
	}

	// convert response
	err = json.Unmarshal([]byte(data), &size)
	if err != nil {
		return size, err
	}

	return size, err
}

func GetPullRequests(repository BitBucketRepository, pull_requests []BitBucketPullRequest, start int) ([]BitBucketPullRequest, error) {
	// get all projects
	endpoint := fmt.Sprintf(
		"/projects/%s/repos/%s/pull-requests?state=all&limit=%d&start=%d",
		repository.Project.Key,
		repository.Slug,
		page_limit,
		start,
	)
	data, err := BBSAPIRequest(endpoint, "GET")
	if err != nil {
		return pull_requests, err
	}

	// convert response
	var response BitBucketPullRequestResponse
	err = json.Unmarshal([]byte(data), &response)
	if err != nil {
		return pull_requests, err
	}

	// if first iteration executing, set
	// otherwise merge values
	if len(pull_requests) == 0 {
		pull_requests = response.Values
	} else {
		pull_requests = append(response.Values, pull_requests...)
	}

	// check for next page recursively
	if !response.IsLastPage {
		pull_requests, err = GetPullRequests(repository, pull_requests, response.NextPageStart)
	}

	return pull_requests, err
}

func GetRepositories(project string, repositories []BitBucketRepository, start int) ([]BitBucketRepository, error) {

	// get all projects
	endpoint := fmt.Sprintf("/projects/%s/repos?limit=%d&start=%d", project, page_limit, start)
	data, err := BBSAPIRequest(endpoint, "GET")
	if err != nil {
		return repositories, err
	}

	// convert response
	var response BitBucketRepositoryResponse
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
	Debug(fmt.Sprintf("Requesting URI: %s", url))

	// set SSL verification
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

	// log response to debug
	Debug(fmt.Sprintf("Response: %s", string(bodyBytes)))

	// return the response and nil error
	return string(bodyBytes), err
}
