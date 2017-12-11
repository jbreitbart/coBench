package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
	mstats "github.com/montanaflynn/stats"
)

func setupCmd(c string, cpuID int, logFilename string) (*exec.Cmd, *os.File, error) {
	var commandName string
	commandStr := make([]string, 0)

	env := os.Environ()

	if *hermitcore {
		commandName = "numactl"
		commandStr = append(commandStr, "numactl", "--physcpubind", cpus[cpuID], "/bin/sh")
		env = append(env, "HERMIT_CPUS="+*threads, "HERMIT_MEM=4G", "HERMIT_ISLE=uhyve")
	} else {
		commandName = "/bin/sh"
		env = append(env, "GOMP_CPU_AFFINITY="+cpus[cpuID], "OMP_NUM_THREADS="+*threads)
	}
	commandStr = append(commandStr, "-c")
	if *perfStat != "" {
		c = "perf stat -e " + *perfStat + " " + c
	}
	commandStr = append(commandStr, c)
	cmd := exec.Command(commandName, commandStr...)
	cmd.Env = env

	// try to avoid duplicate filenames; TODO not perfect
	if _, err := os.Stat(logFilename + ".log"); err == nil {
		for i := 0; i < 100; i++ {
			if _, err := os.Stat(logFilename + "--" + strconv.Itoa(i) + ".log"); err != nil {
				logFilename = logFilename + "-" + strconv.Itoa(i)
				break
			}
		}
	}
	logFilename += ".log"

	outfile, err := os.Create(logFilename)
	if err != nil {
		return nil, nil, fmt.Errorf("Error while creating file: %v", err)
	}
	cmd.Stdout = outfile
	cmd.Stderr = outfile

	return cmd, outfile, nil
}

func runSingle(c string, catConfig [2]uint64) ([]stats.DataPerRun, error) {

	if catConfig[0] != 0 && catConfig[1] != 0 {
		if err := writeCATConfig(catConfig); err != nil {
			return nil, fmt.Errorf("Error while writting CAT config: %v", err)
		}
	}

	filename := commands.Pretty(c)
	if catConfig[0] != 0 && catConfig[1] != 0 {
		filename += fmt.Sprintf("-%x", catConfig[0])
	}
	cmd, outFile, err := setupCmd(c, 0, filename)
	if err != nil {
		return nil, err
	}
	defer outFile.Close()

	// used to count how many apps have reached their min limit
	done := make(chan int, 1)
	done <- 1

	// used to return an error from the go-routines
	errs := make(chan error, 1)

	// used to return the app runtimes
	var runtimes []stats.DataPerRun

	// used to wait for the following 2 goroutines
	var wg sync.WaitGroup
	wg.Add(1)

	go runCmdMinTimes(cmd, *runs, &wg, &runtimes, done, errs)

	wg.Wait()

	if len(errs) != 0 {
		return nil, <-errs
	}

	return runtimes, nil
}

func runPair(cPair [2]string, catConfig [2]uint64) ([][]stats.DataPerRun, error) {

	if catConfig[0] != 0 && catConfig[1] != 0 {
		if err := writeCATConfig(catConfig); err != nil {
			return nil, fmt.Errorf("Error while writting CAT config: %v", err)
		}
	}

	var cmds [len(cpus)]*exec.Cmd
	// setup commands
	for i := range cmds {
		filename := fmt.Sprintf("%v-%v", commands.Pretty(cPair[i]), commands.Pretty(cPair[(i+1)%2]))
		if catConfig[0] != 0 && catConfig[1] != 0 {
			filename += fmt.Sprintf("-%x", catConfig[0])
		}

		var outFile *os.File
		var err error
		cmds[i], outFile, err = setupCmd(cPair[i], i, filename)
		if err != nil {
			return nil, err
		}
		defer outFile.Close()
	}

	// used to count how many apps have reached their min limit
	done := make(chan int, 1)
	done <- 0

	// used to return an error from the go-routines
	errs := make(chan error, len(cmds))

	// used to return the app runtimes
	runtimes := make([][]stats.DataPerRun, len(cmds))

	// used to wait for the following 2 goroutines
	var wg sync.WaitGroup
	wg.Add(len(cmds))

	for i, c := range cmds {
		go runCmdMinTimes(c, *runs, &wg, &runtimes[i], done, errs)
	}

	wg.Wait()

	if len(errs) != 0 {
		return nil, <-errs
	}

	return runtimes, nil
}

func runCmdMinTimes(cmd *exec.Cmd, min int, wg *sync.WaitGroup, runtime *[]stats.DataPerRun, done chan int, errs chan error) {
	defer wg.Done()

	oldVariance := 0.0
	var runtimeInSeconds []float64
	completed := false

	for i := 1; ; i++ {
		// create a copy of the command
		cmd := *cmd

		var buf bytes.Buffer
		cmd.Stdout = io.MultiWriter(cmd.Stdout, &buf)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, &buf)

		start := time.Now()
		err := cmd.Run()

		var data stats.DataPerRun
		data.Runtime = time.Since(start)
		data.Output = buf.String()

		if err != nil {
			errs <- fmt.Errorf("Error running %v: %v", cmd.Args, err)
			return
		}

		// did the other cmd result in an error?
		if len(errs) != 0 {
			return
		}

		d := <-done

		// check if the other application was running the whole time
		if d != len(cpus) {
			// yes
			*runtime = append(*runtime, data)
			runtimeInSeconds = append(runtimeInSeconds, data.Runtime.Seconds())
		}

		// did we run min times?
		if !completed && i >= min {
			vari, _ := mstats.Variance(runtimeInSeconds)
			if math.Abs(vari-oldVariance) <= *varianceDiff {
				d++
				completed = true
			}
			oldVariance = vari
		}
		done <- d

		// both applications are done
		if d == len(cpus) {
			return
		}
	}
}
