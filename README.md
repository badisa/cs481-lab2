Lab 2 - UNM CS481
=================
---

Summary
-------
A lab from UNM's CS481 to evaluate how the Completely Fair Scheduler (CFS) handles procceses that are IO versus CPU intensive. A second portion of the lab is to evaluate the difference in how memory managed, but as of writing this it is not yet implemented in the repo.

This repo consistents of a Golang componenet which performs all of the computation/IO that is being evaluated as well as a script that can be used to generate json output files that summarize each run. The summaries are made up of the fields contained in `/proc/<pid>/sched`. The IO and CPU methods attempt to be as similar as possible to improve the accuracy of the comparisons, however the IO one is contrived to trigger blocking on IO.

Components
----------

### `Golang` - Tested using 1.11
  - Scripts for generating results related to the CFS scheduler

### `Python` - Requires 3.4+
  - Jupyter Notebook for data analysis and graph generation


Installation
------------

### Golang

Refer to the golang [docs](https://golang.org/doc/install). 

### Python

#### TODO

Running
-------

### Golang

```
// To run processes for 8 to 64 seconds (incrementing by 8) with 100
// proccesses run concurrently in each phase (IO only, CPU only, Half and Half)
$ go run cmd/stats_collector.go -time 64 -step 8 -procs 100
```

### Python
The python component is a Ipython notebook, allowing review of how the analysis was performed.
```
// Start the notebook server, then navigate to the analysis.ipynb
$ jupyter notebook
```

