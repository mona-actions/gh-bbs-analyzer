package main

import (
	"fmt"
	"os"
	"time"
	"unicode"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// main vars
	description = "GitHub CLI extension to analyze BitBucket Server for migration statistics"
	logFile     *os.File
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

	// do logic here
	OutputNotice("---- RUNNING PROCESS ----")

	// always return
	return err
}
