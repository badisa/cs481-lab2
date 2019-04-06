package logic

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strings"
)

func PrintSchedulerStats(procType string, format string) {
	cmd := exec.Command("cat", "/proc/self/sched")
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed get proc information")
		os.Exit(1)
	}
	result := output.String()
	// Split the header from the body
	sections := strings.Split(result, "-\n")
	body := strings.Join(sections[1:], "")
	var subSections []string
	for _, section := range strings.Split(body, ":") {
		for _, line := range strings.Split(section, "\n") {
			subSections = append(subSections, strings.Trim(line, " "))
		}
	}
	hasKey := false
	values := make(map[string]string, 20)
	// Set type of run in data for easier parsing
	values["type"] = procType
	var key string
	for _, section := range subSections {
		// Skips nonsense at the end
		if strings.Contains(section, "=") {
			continue
		}
		if format == "print" {
			fmt.Println(section)
		} else if format == "json" {
			if hasKey {
				values[key] = section
				hasKey = false
			} else {
				key = section
				hasKey = true
			}
		} else {
			fmt.Printf("Unknown format: %s\n", format)
			os.Exit(1)
		}
	}
	if format == "json" {
		result, err := json.MarshalIndent(values, "", "  ")
		if err != nil {
			fmt.Printf("Unable to marshall data: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	}
}

// Inidicate that the context is canceled
func IsCanceled(ctx context.Context) bool {
	for {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}
}

// Stole the logic from https://golang.org/doc/play/pi.go
func CPUIntensive(ctx context.Context) {
	var pi float64
	var counter uint64
	for {
		counter++
		pi += 4 * math.Pow(-1, float64(counter)) / float64((2*counter)+1)
		if IsCanceled(ctx) {
			break
		}
	}
}

// Basically the same as cpuIntensive, but writing the value
// of pi to a file and flushing it each time
func IOIntensive(ctx context.Context) {
	var pi float64
	var counter uint64
	tmp, err := ioutil.TempFile("", "io_intensive")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp file: %s", err))
	}
	defer os.Remove(tmp.Name())
	for {
		counter++
		pi += 4 * math.Pow(-1, float64(counter)) / float64((2*counter)+1)
		// Read then write from flushed file
		err = binary.Write(tmp, binary.BigEndian, &pi)
		if err != nil {
			panic(fmt.Sprintf("Failed to write: %s", err))
		}
		err = tmp.Sync()
		if err != nil {
			panic(fmt.Sprintf("Failed to sync: %s", err))
		}
		_, err = tmp.Seek(int64(-binary.Size(pi)), io.SeekEnd)
		if err != nil {
			panic(fmt.Sprintf("Failed to seek: %s", err))
		}
		err = binary.Read(tmp, binary.BigEndian, &pi)
		if err != nil {
			panic(fmt.Sprintf("Failed to read: %s", err))
		}
		_, err := tmp.Seek(4096*1024, io.SeekEnd)
		if err != nil {
			panic(fmt.Sprintf("Failed to do tail seek: %s", err))
		}
		if IsCanceled(ctx) {
			break
		}
	}
}
