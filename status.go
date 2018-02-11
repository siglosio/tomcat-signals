package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// File format is 8 lines for 3 modes:
//  Tomcat PID
//	Request Count last run time (Rate mode)
//	Request Count last counter value (Rate mode)
//	Error Count last run time (Error mode)
//	Error Count last counter value (Error mode)
//  Processing Time last run time (Latency mode)
//	Processing Time last time counter value (Latency mode)
//	Processing Time last request counter value (Latency mode)

// Read last run info from saved file
func getLastRunInfo(statusFileName string) ([8]int, error) {

	var (
		f       io.Reader
		scanner *bufio.Scanner
		err     error
		r       int
		last    [8]int
	)

	if statusFileName == "" {
		statusFileName = "status.file"
	}

	f, err = os.Open(statusFileName)

	// If not exist or error, just ignore, hope we can create when done
	if err == nil {
		scanner = bufio.NewScanner(f)

		// Read values in loop
		for i := 0; i < 8; i++ {
			if scanner.Scan() { // Only process if we got something (not blank / empty line or file)
				r, err = strconv.Atoi(strings.TrimSpace(scanner.Text()))
				checkErr(err)
				last[i] = r
			}
		}
	} else {
		if flagVerbose {
			fmt.Printf("Error opening status file: %s, ignoring.\n\n", statusFileName)
		}
		// Set all to zero
		for i := 0; i < 8; i++ {
			last[i] = 0
		}
	}
	return last, nil
}

// Save last run info to saved file
func saveLastRunInfo(statusFile string, last [8]int) (error) {
	var (
		fw  io.Writer
		err error
		s   string
	)

	if statusFile == "" {
		statusFile = "status.file"
	}
	fw, err = os.Create(statusFile)
	checkErr(err)

	// Loop writing
	for i := 0; i < 8; i++ {
		s = fmt.Sprintf("%d\n", last[i])
		_, err = io.WriteString(fw, s)
		checkErr(err)
	}

	return nil
}
