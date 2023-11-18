package main

import (
	"errors"
	"fmt"
	"mvdan.cc/garble/mobile/shared"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

/***
This binary will be called by gomobile. gomobile calls "go build" internally.
When "garble mobile" is called, garble will modify the PATH env var so that calls to "go"
will call this binary instead.

The call flow looks like this:
"garble mobile" -> "gomobile" -> (this binary) -> "garble build" -> "go build" (garble calls the real go binary)
*/

func main() { os.Exit(run()) }

func run() int {
	garbleBinDir, err := shared.GarbleBinDir()
	if err != nil {
		return 1
	}

	parentDir := filepath.Dir(garbleBinDir)

	logFilePath := filepath.Join(parentDir, "redirection.log")

	// since gomobile calls this binary, we can't get the output directly in garble,
	// we can only write any errors to a log file.
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 1
	}
	defer f.Close()

	args := os.Args[1:]

	var command = "garble"

	if args[0] != "build" {
		// "go build" was not called so we don't want to call garble. Fallback to the real go command.
		// make sure to specify the full path to the original go binary.
		command = os.Getenv("GARBLE_OG_GO") // path to real go binary
	}

	argsAsString := strings.Join(args, " ")

	f.Write([]byte("running command: " + command + " " + argsAsString + "\n"))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		f.Write([]byte(fmt.Sprintf("err starting command: %s with args: %s err: %s\n", command, argsAsString, err)))
		return 1
	}

	err = cmd.Wait()
	if err != nil {
		f.Write([]byte(fmt.Sprintf("err waiting command: %s with args: %s err: %s\n", command, argsAsString, err)))

		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				// exit with same code returned from cmd
				return status.ExitStatus()
			}
		}
		return 1
	}

	return 0
}
