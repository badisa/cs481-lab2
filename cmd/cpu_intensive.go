package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/cs481-lab2/logic"
	"os"
	"os/exec"
	"time"
)

func main() {
	secondsToCompletion := flag.Int("time", 10, "How long to run computation in seconds")
	flag.Parse()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*time.Duration(*secondsToCompletion)))
	defer cancel()
	logic.CPUIntensive(ctx)
	cmd := exec.Command("cat", "/proc/self/sched")
	var output bytes.Buffer
	cmd.Stdout = &output
	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed get proc information")
		os.Exit(1)
	}
	fmt.Printf(output.String())
}
