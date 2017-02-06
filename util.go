package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/djavorszky/ddn/inet"
	"github.com/djavorszky/ddn/model"
	"github.com/djavorszky/notif"
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

func textsOccur(file *os.File, t ...[]string) (map[int]bool, error) {
	found := make(map[int]bool)

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
				found[lines[0]] = true
			}

			file.Seek(0, 0)
		}
	}

	return found, nil
}

func removeLinesFromFile(file *os.File, lines map[int]bool) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "ddnc")
	if err != nil {
		return nil, fmt.Errorf("could not create tempfile: %s", err.Error())
	}

	writer := bufio.NewWriter(tmpFile)

	curLine := 1
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		if _, ok := lines[curLine]; ok {
			continue
		}

		writer.WriteString(line)
		curLine++
	}

	err = writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("could not flush writer: %s", err.Error())
	}

	tmpFilePath, _ := filepath.Abs(tmpFile.Name())
	newFilePath, _ := filepath.Abs(file.Name())

	os.Rename(tmpFilePath, newFilePath)

	return os.Open(newFilePath)
}

func registerConnector() error {
	endpoint := fmt.Sprintf("%s/%s", conf.MasterAddress, "alive")

	if !inet.AddrExists(endpoint) {
		return fmt.Errorf("Master server does not exist at given endpoint")
	}

	ddnc := model.RegisterRequest{
		ID:        connector.ID,
		ShortName: conf.ShortName,
		LongName:  fmt.Sprintf("%s %s", conf.Vendor, conf.Version),
		Version:   version,
	}

	register := fmt.Sprintf("%s/%s", conf.MasterAddress, "register")

	resp, err := notif.SndLoc(ddnc, register)
	if err != nil {
		return fmt.Errorf("Could not register with the master server: %s", err.Error())
	}

	var c model.Connector

	err = json.NewDecoder(bytes.NewBufferString(resp)).Decode(&c)
	if err != nil {
		log.Fatalf("Could not decode server response: %s", err.Error())
	}

	connector = c

	log.Printf("Master server registration complete. Got assigned ID '%d'", connector.ID)

	return nil
}

func unregisterConnector() {
	connector.Up = false

	unregister := fmt.Sprintf("%s/%s", conf.MasterAddress, "unregister")
	_, err := notif.SndLoc(connector, unregister)
	if err != nil {
		log.Fatalf("Could not register with the master server: %s", err.Error())
	}

	log.Fatalf("Successfully unregistered the connector.")
}
