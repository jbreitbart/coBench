package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
)

func runSingle(c string, id int) ([]time.Duration, error) {
	env := os.Environ()

	// TODO CAT?

	var cmd *exec.Cmd
	if *hermitcore {
		cmd = exec.Command("numactl", "--physcpubind", cpus[0], "/bin/sh", "-c", c)
		cmd.Env = append(env, "HERMIT_CPUS="+*threads, "HERMIT_MEM=4G", "HERMIT_ISLE=uhyve")
	} else {
		cmd = exec.Command("/bin/sh", "-c", c)
		cmd.Env = append(env, "GOMP_CPU_AFFINITY="+cpus[0], "OMP_NUM_THREADS="+*threads)
	}

	filename := fmt.Sprintf("%v", id)
	filename += ".log"
	outfile, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("Error while creating file: %v", err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	// used to count how many apps have reached their min limit
	done := make(chan int, 1)
	done <- 1

	// used to return an error from the go-routines
	errs := make(chan error, 1)

	// used to return the app runtimes
	var runtimes []time.Duration

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

// TODO remove id
func runPair(cPair [2]string, id int, catConfig []uint64) ([][]time.Duration, error) {
	env := os.Environ()

	if *cat {
		if err := writeCATConfig(catConfig); err != nil {
			return nil, fmt.Errorf("Error while writting CAT config: %v", err)
		}
	}

	var cmds [len(cpus)]*exec.Cmd
	// setup commands
	for i, _ := range cmds {
		if *hermitcore {
			cmds[i] = exec.Command("numactl", "--physcpubind", cpus[i], "/bin/sh", "-c", cPair[i])
			cmds[i].Env = append(env, "HERMIT_CPUS="+*threads, "HERMIT_MEM=4G", "HERMIT_ISLE=uhyve")
		} else {
			cmds[i] = exec.Command("/bin/sh", "-c", cPair[i])
			cmds[i].Env = append(env, "GOMP_CPU_AFFINITY="+cpus[i], "OMP_NUM_THREADS="+*threads)
		}

		filename := fmt.Sprintf("%v-%v", id, i)
		if *cat {
			filename += fmt.Sprintf("-%x", catConfig[i])
		}
		filename += ".log"
		outfile, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("Error while creating file: %v", err)
		}
		defer outfile.Close()
		cmds[i].Stdout = outfile
	}

	// used to count how many apps have reached their min limit
	done := make(chan int, 1)
	done <- 0

	// used to return an error from the go-routines
	errs := make(chan error, len(cmds))

	// used to return the app runtimes
	runtimes := make([][]time.Duration, len(cmds))

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

func runCmdMinTimes(cmd *exec.Cmd, min int, wg *sync.WaitGroup, runtime *[]time.Duration, done chan int, errs chan error) {
	defer wg.Done()

	oldVariance := 0.0
	var runtimeInSeconds []float64
	completed := false

	for i := 1; ; i++ {
		// create a copy of the command
		cmd := *cmd

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

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
			*runtime = append(*runtime, elapsed)
			runtimeInSeconds = append(runtimeInSeconds, elapsed.Seconds())
		}

		// did we run min times?
		if !completed && i >= min {
			vari, _ := stats.Variance(runtimeInSeconds)
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
