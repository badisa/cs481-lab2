package main

import (
    "os"
    "os/exec"
    "fmt"
    "bytes"
    "strconv"
    "encoding/json"
    "flag"
    "sync"
    "time"
)


type StatsData struct {
    lock sync.RWMutex
    data map[string][]map[string]string
}

func main() {
    maxTime := flag.Int("max-time", 5, "Maximum amount of time to run each proc")
    stepSize := flag.Int("step", 1, "How much to increase time step by until max is reached, also the starting time. In Seconds")
    procs := flag.Int("procs", 10, "How many procs to run in each iteration")
    flag.Parse()
    fmt.Println("Building CPU Intensive binary")
    cmd := exec.Command("go", "build", "-ldflags", "-s -w -d", "cmd/cpu_intensive.go")
    var output bytes.Buffer
    cmd.Stdout = &output
    cmd.Stderr = &output
    err := cmd.Run()
    if err != nil {
        fmt.Printf("Failed to build IO intensive binary: %s\n", output.String())
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
    stat := StatsData{data:make(map[string][]map[string]string, 10)}
    start := time.Now()
    var wg sync.WaitGroup
    for runTime := *stepSize; runTime <= *maxTime; runTime += *stepSize {
        runStart := time.Now()
        fmt.Printf("Running IO only Procs for %ds\n", runTime)
        key := fmt.Sprintf("io-only-%d", runTime)
        for i := 0; i < *procs; i++ {
            wg.Add(1)
            go stat.RunIOProccess(key, runTime, &wg)
        }
        wg.Wait()
        fmt.Printf("Running CPU only Procs for %ds\n", runTime)
        key = fmt.Sprintf("cpu-only-%d", runTime)
        for i := 0; i < *procs; i++ {
            wg.Add(1)
            go stat.RunCPUProccess(key, runTime, &wg)
        }
        wg.Wait()
        fmt.Printf("Running Mixed Procs for %ds\n", runTime)
        key = fmt.Sprintf("mixed-%d", runTime)
        for i := 0; i < *procs; i++ {
            wg.Add(1)
            if i % 2 == 0 {
                go stat.RunCPUProccess(key, runTime, &wg)
            } else {
                go stat.RunIOProccess(key, runTime, &wg)
            }
        }
        wg.Wait()
        fmt.Printf("Finished runs with runTime of %ds, took %s\n", runTime, time.Since(runStart))
    }
    fmt.Printf("Finished running processes, took %s\n", time.Since(start))
    path := fmt.Sprintf("lab-part-1-max-%d-step-%d-proc-%d.json", *maxTime, *stepSize, *procs)
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
    if _, ok := d.data[key]; !ok{
        d.data[key] = []map[string]string{data}
    } else {
        d.data[key] = append(d.data[key], data)
    }
}

func (d *StatsData) RunIOProccess(key string, time int, wg *sync.WaitGroup) {
    strTime := strconv.Itoa(time)
    cmd := exec.Command("./io_intensive", "-time", strTime)
    var output bytes.Buffer
    cmd.Stdout = &output
    cmd.Stderr = &output
    err := cmd.Run()
    wg.Done()
    if err != nil {
        fmt.Printf("Failed to run IO Process for key %s: %s\n", key, output.String())
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

func (d *StatsData) RunCPUProccess(key string, time int, wg *sync.WaitGroup) {
    strTime := strconv.Itoa(time)
    cmd := exec.Command("./cpu_intensive", "-time", strTime)
    var output bytes.Buffer
    cmd.Stdout = &output
    cmd.Stderr = &output
    err := cmd.Run()
    wg.Done()
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
