package main

import (
    "io"
    "os"
    "fmt"
    "bufio"
    "strings"
)


// Read lines from a file.  Lines will start being sent to the callback function between first
// and max
func LinesFromFile(filename string, firstResult int, maxResults int, callback func(line string) bool) error {
    var err error
    var file *os.File = nil

    if (filename == "-") {
        file = os.Stdin
    } else {
        file, err = os.Open(filename)
        if err != nil {
            return err
        }
        defer file.Close()
    }

    bufr := bufio.NewReader(file)
    resultCount := 0

    for line, err := bufr.ReadString('\n') ; err == nil ; line, err = bufr.ReadString('\n') {
        if (resultCount >= firstResult) {
            line = strings.TrimSpace(line)
            if (! callback(line)) {
                return nil
            }
        }
        resultCount++
        if ((resultCount >= firstResult + maxResults) && (maxResults != -1)) {
            fmt.Fprintf(os.Stderr, "Maximum number of lines encountered (%d).  Use -c to change.\n", maxResults)
            return nil
        }
    }

    if err != io.EOF {
        return err
    } else {
        return nil
    }
}
