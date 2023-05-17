package main

import (
	"crypto/tls"
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
	logFile        *os.File
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

func init() {

	// add flags here
	rootCmd.PersistentFlags().StringVar(
		&bbs_server_url,
		"bbs-server-url",
		"",
		"The full URL of the Bitbucket Server/Data Center to migrate from. E.g. http://bitbucket.contoso.com:7990",
	)
	rootCmd.PersistentFlags().StringVar(
		&bbs_username,
		"bbs-username",
		"",
		"The Bitbucket username of a user with site admin privileges. If not set will be read from BBS_USERNAME environment variable.",
	)
	rootCmd.PersistentFlags().StringVar(
		&bbs_password,
		"bbs-password",
		"",
		"The Bitbucket password of the user specified by --bbs-username. If not set will be read from BBS_PASSWORD environment variable.",
	)
	rootCmd.PersistentFlags().BoolVar(
		&no_ssl_verify,
		"no-ssl-verify",
		false,
		"Disables SSL verification when communicating with your Bitbucket Server/Data Center instance. All other migration steps will continue to verify SSL. If your Bitbucket instance has a self-signed SSL certificate then setting this flag will allow the migration archive to be exported.",
	)

	bbs_username_env, is_set := os.LookupEnv("BBS_USERNAME")
	if is_set {
		bbs_username = bbs_username_env
	}

	bbs_password_env, is_set := os.LookupEnv("BBS_PASSWORD")
	if is_set {
		bbs_password = bbs_password_env
	}

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
	_, err := logFile.WriteString(
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
	logFile, err = os.Create(fmt.Sprint(time.Now().Format("20060102_150401"), ".log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	// validate flags
	r, _ := regexp.Compile("^http(s|):(//|)")
	if !r.MatchString(bbs_server_url) {
		OutputError("BitBucket server url should contain http(s) prefix and does not.", true)
	}
	if bbs_server_url == "" {
		OutputError("A BitBucket server URL must be provided.", true)
	}
	if bbs_username == "" {
		OutputError("A BitBucket username must be provided or environment variable set.", true)
	}
	if bbs_password == "" {
		OutputError("A BitBucket password must be provided or environment variable set.", true)
	}

	// output flags for reference
	OutputFlags("BitBucket Server URL", bbs_server_url)
	OutputFlags("BitBucket Username", bbs_username)
	OutputFlags("BitBucket Password", "**********")
	OutputFlags("SSL Verification Disabled", strconv.FormatBool(no_ssl_verify))

	// do logic here
	data, err := BBSRequest("projects")
	if err != nil {
		OutputError(fmt.Sprintf("Error making request: %s", err), true)
	}

	OutputNotice("client: got response!")
	OutputNotice(fmt.Sprintf("client: response: %s\n", data))

	// always return
	return err
}

func BBSRequest(endpoint string) (data string, err error) {

	// set up endpoint
	url := fmt.Sprintf("%s/rest/api/1.0/%s", bbs_server_url, endpoint)
	Debug(fmt.Sprintf("Request URL: %s", url))

	// set SSL verification using inverse bool from flag
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !no_ssl_verify},
	}

	// create client
	client := &http.Client{Transport: tr}

	// create request and add authentication
	req, err := http.NewRequest("GET", url, nil)
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
