package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jbreitbart/coBench/stats"
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

func storeConfig(commands []string) {

	stats.SetCommandline(*cat, *catBitChunk, catDirs, cpus, commands, *hermitcore, *resctrlPath, *runs, *threads, *varianceDiff)

}
