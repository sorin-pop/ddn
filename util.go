package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/djavorszky/sutils"
)

const defaultFailedCode = 1

// RunCommand executes a command with specified arguments and returns its exitcode, stdout
// and stderr as well.
func RunCommand(name string, args ...string) CommandResult {
	var (
		outbuf, errbuf bytes.Buffer
		exitCode       int
	)

	log.Println("Running command: ", name, args)

	cmd := exec.Command(name, args...)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Run()
	stdout := outbuf.String()
	stderr := errbuf.String()

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			// This will happen (in OSX) if `name` is not available in $PATH,
			// in this situation, exit code could not be get, and stderr will be
			// empty string very likely, so we use the default fail code, and format err
			// to string and set to stderr
			log.Printf("Could not get exit code for failed program: %v, %v", name, args)

			exitCode = defaultFailedCode

			if stderr == "" {
				stderr = err.Error()
			}
		}
	} else {
		// success, exitCode should be 0 if go is ok
		ws := cmd.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	stdout = strings.TrimSuffix(stdout, "\n")
	stderr = strings.TrimSuffix(stderr, "\n")

	return CommandResult{stdout, stderr, exitCode}
}

// CommandResult is a struct that contains the stdout, stderr and exitcode
// of a command that was executed.
type CommandResult struct {
	stdout, stderr string
	exitCode       int
}

func textsOccur(file *os.File, t ...[]string) ([]int, error) {
	var found []int

	for _, strslice := range t {
		for _, str := range strslice {
			lines, err := sutils.FindCaseSensitive(file, str)
			if err != nil {
				return nil, fmt.Errorf("searching for %q failed: %s", str, err.Error())
			}
			if len(lines) > 1 {
				return nil, fmt.Errorf("more than one %q statements found in dump", str)
			}

			if len(lines) == 1 {
				found = append(found, lines[0])
			}

			file.Seek(0, 0)
		}
	}

	return found, nil
}
