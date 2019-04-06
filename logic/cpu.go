package logic

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
)

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
		tmp.Sync()
		_, err = tmp.Seek(0, io.SeekStart)
		if err != nil {
			panic(fmt.Sprintf("Failed to sync: %s", err))
		}
		err = binary.Read(tmp, binary.BigEndian, &pi)
		if err != nil {
			panic(fmt.Sprintf("Failed to read: %s", err))
		}
		if IsCanceled(ctx) {
			break
		}
	}
}
