package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

const (
	PHASE_ALL   = "all"
	PHASE_CPU   = "cpu"
	PHASE_IO    = "io"
	PHASE_MIXED = "mixed"
    IO_CPU      = "io-cpu"
    PAGE_TABLE  = "page-table"
)

type StatsData struct {
	lock sync.RWMutex
	data map[string][]map[string]string
}

type InputData struct {
    phase     string
    maxTime   int
    stepSize  int
    procs     int
    batchSize int
}

func main() {
    mode := flag.String("mode", "io-cpu", "Whether to run the 'io-cpu' experiment or the 'page-table' experiment")
	phase := flag.String("phases", PHASE_ALL, "Type of phases to run. Options are all, cpu, io, mixed")
	maxTime := flag.Int("max-time", 5, "Maximum amount of time to run each proc")
	stepSize := flag.Int("step", 1, "How much to increase time step by until max is reached, also the starting time. In Seconds")
	procs := flag.Int("procs", 10, "How many procs to run in each iteration")
	batchSize := flag.Int("batch", 10, "Maximum number of processes to run simultaneously")
    flag.Parse()
    if *mode != IO_CPU && *mode != PAGE_TABLE {
        fmt.Println(fmt.Sprintf("Invalid mode: %s", *mode))
        os.Exit(1)
    }
    inputData := InputData{
        phase: *phase,
        maxTime: *maxTime,
        stepSize: *stepSize,
        procs: *procs,
        batchSize: *batchSize,
    }
    if *mode == IO_CPU {
        compareIOCPUProcesses(&inputData)
    } else {
        comparePageTableProcesses(&inputData)
    }
}

func comparePageTableProcesses (input *InputData) {
    fmt.Println("Building Sane Page table binary")
    cmd := exec.Command("go", "build", "-ldflags", "-s -w -d", "cmd/sane_page_table.go")
    var output bytes.Buffer
    cmd.Stdout = &output
    cmd.Stderr = &output
    err := cmd.Run()
    if err != nil {
        fmt.Printf("Failed to build sane page table binary: %s\n", output.String())
        os.Exit(1)
    }
    defer os.Remove("sane_page_table")
    output.Reset()
    fmt.Println("Building Insane page table binary")
    cmd = exec.Command("go", "build", "-ldflags", "-s -w -d", "cmd/insane_page_table.go")
    cmd.Stdout = &output
    cmd.Stderr = &output
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Failed to build insane page table binary: %s\n", output.String())
        os.Exit(1)
    }
    defer os.Remove("insane_page_table")
    output.Reset()
	stat := StatsData{data: make(map[string][]map[string]string, 10)}
	start := time.Now()
	var wg sync.WaitGroup
	var key string
    for runTime := input.stepSize; runTime <= input.maxTime; runTime += input.stepSize {
		runStart := time.Now()
		fmt.Printf("Running Mixed Page Table processes for %ds in batches of %d\n", runTime, input.batchSize)
		key = fmt.Sprintf("mixed-page-table-%d", runTime)
		for i := 0; i < input.procs; i++ {
			if i%input.batchSize == 0 {
				wg.Wait()
			}
			wg.Add(1)
            if i % 2 == 0 {
                go stat.RunInsanePageTableProcess(key, runTime, &wg)
            } else {
			    go stat.RunSanePageTableProcess(key, runTime, &wg)
            }
		}
		wg.Wait()
		fmt.Printf("Finished runs with runTime of %ds, took %s\n", runTime, time.Since(runStart))
    }
	fmt.Printf("Finished running processes, took %s\n", time.Since(start))
	path := fmt.Sprintf("lab-part-2-max-%d-step-%d-proc-%d-batches-%d.json", input.maxTime, input.stepSize, input.procs, input.batchSize)
	err = stat.Dump(path)
	if err != nil {
		fmt.Printf("Failed to dump: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote results to %s\n", path)
}

func compareIOCPUProcesses (input *InputData) {
    fmt.Println("Building CPU Intensive binary")
    cmd := exec.Command("go", "build", "-ldflags", "-s -w -d", "cmd/cpu_intensive.go")
    var output bytes.Buffer
    cmd.Stdout = &output
    cmd.Stderr = &output
    err := cmd.Run()
    if err != nil {
        fmt.Printf("Failed to build CPU intensive binary: %s\n", output.String())
        os.Exit(1)
    }
    defer os.Remove("cpu_intensive")
    output.Reset()
    fmt.Println("Building IO Intensive binary")
    cmd = exec.Command("go", "build", "-ldflags", "-s -w -d", "cmd/io_intensive.go")
    cmd.Stdout = &output
    cmd.Stderr = &output
    err = cmd.Run()
    if err != nil {
        fmt.Printf("Failed to build IO intensive binary: %s\n", output.String())
        os.Exit(1)
    }
    defer os.Remove("io_intensive")
    output.Reset()
	stat := StatsData{data: make(map[string][]map[string]string, 10)}
	start := time.Now()
	var wg sync.WaitGroup
	var key string
	for runTime := input.stepSize; runTime <= input.maxTime; runTime += input.stepSize {
		runStart := time.Now()
		if input.phase == PHASE_ALL || input.phase == PHASE_IO {
			fmt.Printf("Running IO only Procs for %ds in batches of %d\n", runTime, input.batchSize)
			key = fmt.Sprintf("io-only-%d", runTime)
			for i := 0; i < input.procs; i++ {
				if i%input.batchSize == 0 {
					wg.Wait()
				}
				wg.Add(1)
				go stat.RunIOProcess(key, runTime, &wg)
			}
			wg.Wait()
		}
		if input.phase == PHASE_ALL || input.phase == PHASE_CPU {
			fmt.Printf("Running CPU only Procs for %ds in batches of %d\n", runTime, input.batchSize)
			key = fmt.Sprintf("cpu-only-%d", runTime)
			for i := 0; i < input.procs; i++ {
				if i%input.batchSize == 0 {
					wg.Wait()
				}
				wg.Add(1)
				go stat.RunCPUProcess(key, runTime, &wg)
			}
			wg.Wait()
		}
		if input.phase == PHASE_ALL || input.phase == PHASE_MIXED {
			fmt.Printf("Running Mixed Procs for %ds in batches of %d\n", runTime, input.batchSize)
			key = fmt.Sprintf("mixed-%d", runTime)
			for i := 0; i < input.procs; i++ {
				if i%input.batchSize == 0 {
					wg.Wait()
				}
				wg.Add(1)
				if i%2 == 0 {
					go stat.RunCPUProcess(key, runTime, &wg)
				} else {
					go stat.RunIOProcess(key, runTime, &wg)
				}
			}
			wg.Wait()
		}
		fmt.Printf("Finished runs with runTime of %ds, took %s\n", runTime, time.Since(runStart))
	}
	fmt.Printf("Finished running processes, took %s\n", time.Since(start))
	path := fmt.Sprintf("lab-part-1-max-%d-step-%d-proc-%d-batches-%d.json", input.maxTime, input.stepSize, input.procs, input.batchSize)
	err = stat.Dump(path)
	if err != nil {
		fmt.Printf("Failed to dump: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote results to %s\n", path)
}

func (d StatsData) Dump(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create output file:%s", err)
	}
	data, err := json.MarshalIndent(d.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to dump marshal data: %s", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %s", err)
	}
	if err = file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %s", err)
	}
	return nil
}

func (d *StatsData) WriteRun(key string, data map[string]string) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, ok := d.data[key]; !ok {
		d.data[key] = []map[string]string{data}
	} else {
		d.data[key] = append(d.data[key], data)
	}
}

func (d *StatsData) RunInsanePageTableProcess(key string, time int, wg *sync.WaitGroup) {
	strTime := strconv.Itoa(time)
	cmd := exec.Command("./insane_page_table", "-time", strTime)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	defer wg.Done()
	if err != nil {
		fmt.Printf("Failed to run insane page table process for key %s: error:%s, output:%s\n", key, err, output.String())
		return
	}
	schedStats := make(map[string]string)
	err = json.Unmarshal(output.Bytes(), &schedStats)
	if err != nil {
		fmt.Printf("Failed to parse IO proc stats: %s\n", output.String())
		return
	}
	d.WriteRun(key, schedStats)
}


func (d *StatsData) RunSanePageTableProcess(key string, time int, wg *sync.WaitGroup) {
	strTime := strconv.Itoa(time)
	cmd := exec.Command("./sane_page_table", "-time", strTime)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	defer wg.Done()
	if err != nil {
		fmt.Printf("Failed to run insane page table process for key %s: error:%s, output:%s\n", key, err, output.String())
		return
	}
	schedStats := make(map[string]string)
	err = json.Unmarshal(output.Bytes(), &schedStats)
	if err != nil {
		fmt.Printf("Failed to parse IO proc stats: %s\n", output.String())
		return
	}
	d.WriteRun(key, schedStats)
}

func (d *StatsData) RunIOProcess(key string, time int, wg *sync.WaitGroup) {
	strTime := strconv.Itoa(time)
	cmd := exec.Command("./io_intensive", "-time", strTime)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	defer wg.Done()
	if err != nil {
		fmt.Printf("Failed to run IO Process for key %s: error:%s, output:%s\n", key, err, output.String())
		return
	}
	schedStats := make(map[string]string)
	err = json.Unmarshal(output.Bytes(), &schedStats)
	if err != nil {
		fmt.Printf("Failed to parse IO proc stats: %s\n", output.String())
		return
	}
	d.WriteRun(key, schedStats)
}

func (d *StatsData) RunCPUProcess(key string, time int, wg *sync.WaitGroup) {
	strTime := strconv.Itoa(time)
	cmd := exec.Command("./cpu_intensive", "-time", strTime)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	defer wg.Done()
	if err != nil {
		fmt.Printf("Failed to run CPU Process for key %s: %s\n", key, output.String())
		return
	}
	schedStats := make(map[string]string)
	err = json.Unmarshal(output.Bytes(), &schedStats)
	if err != nil {
		fmt.Printf("Failed to parse CPU proc stats: %s\n", err)
		return
	}
	d.WriteRun(key, schedStats)
}
