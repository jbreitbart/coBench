package main

import (
	"flag"
	"fmt"
	"os"
)

// global command line parameters
var runs *int
var cpus [2]string
var threads *string
var hermitcore *bool
var noCoSched *bool

var resctrlPath *string
var cat *bool
var catBitChunk *uint64
var catDirs []string

var varianceDiff *float64

func parseArgs() *string {
	runs = flag.Int("runs", 2, "Number of times the applications are executed")
	commandFile := flag.String("cmd", "cmd.txt", "Text file containing the commands to execute")

	cpus0 := flag.String("cpus0", "0-4", "List of CPUs to be used for the 1st command")
	cpus1 := flag.String("cpus1", "5-9", "List of CPUs to be used for the 2nd command")
	threads = flag.String("threads", "5", "Number of threads to be used")

	cat = flag.Bool("cat", false, "Measure with all CAT settings")
	catBitChunk = flag.Uint64("catChunk", 2, "Bits changed from one run to the next")
	resctrlPath = flag.String("resctrl", "/sys/fs/resctrl/", "Root path of the resctrl file system")

	hermitcore = flag.Bool("hermitcore", false, "Use if you are executing hermitcore binaries")

	varianceDiff = flag.Float64("variance", 0.001, "Minimum differences in variance required between runs")

	noCoSched = flag.Bool("no-cosched", false, "Disable co-scheduling")

	flag.Parse()
	catDirs = []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	if *runs < 1 {
		fmt.Println("runs must be > 0")
		os.Exit(0)
	}
	if *catBitChunk < 1 {
		fmt.Println("catChunk must be > 0")
		os.Exit(0)
	}

	cpus[0] = *cpus0
	cpus[1] = *cpus1

	return commandFile
}

func storeConfig(commands []string) error {

	runtimeStats.Commandline.CAT = *cat
	runtimeStats.Commandline.CATChunk = *catBitChunk
	runtimeStats.Commandline.CATDirs = catDirs
	runtimeStats.Commandline.CPUs = cpus
	runtimeStats.Commandline.Commands = commands
	runtimeStats.Commandline.HermitCore = *hermitcore
	runtimeStats.Commandline.ResctrlPath = *resctrlPath
	runtimeStats.Commandline.Runs = *runs
	runtimeStats.Commandline.Threads = *threads
	runtimeStats.Commandline.VarianceDiff = *varianceDiff

	file, err := os.Create("coBench.config")
	if err != nil {
		return fmt.Errorf("Error while creating file: %v", err)
	}

	out := fmt.Sprintf("Runs: %v \n", *runs)
	out += fmt.Sprintf("Variance diff: %v \n", *varianceDiff)

	out += fmt.Sprintf("CPUs: %v \n", cpus)
	out += fmt.Sprintf("Threads: %v \n", *threads)

	out += fmt.Sprintf("HermitCore: %v \n", *hermitcore)

	out += fmt.Sprintf("CAT: %v \n", *cat)
	out += fmt.Sprintf("CAT chunk: %v \n", *catBitChunk)
	out += fmt.Sprintf("CAT dirs: %v \n", catDirs)
	out += fmt.Sprintf("resctrlPath: %v \n", *resctrlPath)

	out += fmt.Sprintf("Commands: %v \n", commands)

	_, err = file.WriteString(out)
	if err != nil {
		return err
	}

	return nil
}
