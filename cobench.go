package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
	log "github.com/sirupsen/logrus"
)

func main() {
	commandFile := parseArgs()

	commandStrings, err := commands.Read(*commandFile)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"file": *commandFile,
		}).Fatalln("Could not read command file")
	}
	if len(commandStrings) < 1 || (len(commandStrings) < 2 && !*noCoSched) {
		log.Fatalln("You must provide more commands")
	}

	hostname, _ := os.Hostname()
	log.WithField("host", hostname).Infoln("Benchmark started")

	storeConfig(commandStrings)

	defer cleanup()

	indvCommands := commands.GenerateIndv(commandStrings)

	if len(indvCommands) != len(commandStrings) {
		log.Infof("Remove %v duplicates from commands for individual runs.\n", len(commandStrings)-len(indvCommands))
	}

	// run apps individually
	individualRuns(commandStrings)

	if *noCoSched {
		return
	}

	commandPairs := commands.GeneratePairs(commandStrings)
	coSchedRuns(commandPairs)
}

func cleanup() {
	log.Infoln("Benchmark runs complete")

	err := stats.StoreToFile(*resultFilename)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"file": *resultFilename,
		}).Errorln("Error store measurements")

		j, _ := stats.CreateJSON()
		log.Infoln(string(j))
	}

	absfilename, err := filepath.Abs(*resultFilename)
	if err == nil {
		*resultFilename = absfilename
	}
	hostname, err := os.Hostname()
	if err == nil {
		*resultFilename = hostname + ":" + *resultFilename
	}

	log.WithField("file", *resultFilename).Infoln("Result file written")
}

func individualRuns(commands []string) {

	log.Infoln("Running apps individually")

	// run app individually without CAT (if CAT was requested)
	for _, c := range commands {
		catConfig := [2]uint64{stats.NoCATMask, stats.NoCATMask}

		log.WithFields(log.Fields{
			"app": c,
		}).Infoln("Running app")
		r, err := runSingle(c, catConfig)
		if err != nil {
			log.WithError(err).Fatalln("Error running app")
		}

		stat := stats.AddReferenceRuntime(c, r) // TODO stat also contain old runs
		printStats(c, stat, catConfig[0])       // TODO and old catConfigs
	}

	if !*cat || *noIndvSched {
		return
	}

	minBits, numBits, err := setupCAT()
	if err != nil {
		log.WithError(err).Fatalln("Error setting up CAT")
	}
	defer resetCAT()

	catPairs := generateCatConfigs(minBits, numBits)

	for _, c := range commands {

		log.WithFields(log.Fields{
			"app": c,
		}).Infoln("Running app with CAT")

		for _, catConfig := range catPairs {
			runtime, err := runSingle(c, catConfig)
			if err != nil {
				log.WithError(err).Fatalln("Error running app")
			}

			stat := stats.AddCATRuntime(c, catConfig[0], runtime) // TODO see above
			printStats(c, stat, catConfig[0])
		}
	}

	log.Infoln("Individual runs done")
}

func coSchedRuns(commandPairs [][2]string) {
	log.Infoln("Executing the following command pairs")
	for i, c := range commandPairs {
		log.WithFields(log.Fields{
			"app0": c[0],
			"app1": c[1],
		}).Infof("%v", i)
	}

	// run co-scheduling *without* cat
	for i, c := range commandPairs {
		log.WithFields(log.Fields{
			"app0": c[0],
			"app1": c[1],
		}).Infof("Running pair %v", i)

		catConfig := [2]uint64{stats.NoCATMask, stats.NoCATMask}
		runtimes, err := runPair(c, catConfig)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"app0": c[0],
				"app1": c[1],
			}).Fatalln("Error running pair")
		}

		err = processRuntime(i, c, catConfig, runtimes)
		if err != nil {
			log.WithError(err).Fatalln("Error processing runtime")
		}
	}

	if !*cat {
		return
	}

	minBits, numBits, err := setupCAT()
	if err != nil {
		log.WithError(err).Fatalln("Error setting up CAT")
	}
	defer resetCAT()

	catPairs := generateCatConfigs(minBits, numBits)

	for i, c := range commandPairs {
		log.WithFields(log.Fields{
			"app0": c[0],
			"app1": c[1],
		}).Infof("Running pair %v", i)

		for _, catConfig := range catPairs {
			runtimes, err := runPair(c, catConfig)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"app0": c[0],
					"app1": c[1],
				}).Fatalln("Error running pair")
			}

			err = processRuntime(i, c, catConfig, runtimes)
			if err != nil {
				log.WithError(err).Fatalln("Error processing runtime")
			}
		}
	}
}

func processRuntime(id int, cPair [2]string, catMasks [2]uint64, runtimes [][]stats.DataPerRun) error {

	for i, runtime := range runtimes {
		var stat stats.RuntimeT

		if catMasks[0] != 0 && catMasks[1] != 0 {
			stat = stats.AddCoSchedCATRuntime(cPair[i], cPair[(i+1)%2], catMasks[i], runtime)
		} else {
			stat = stats.AddCoSchedRuntime(cPair[i], cPair[(i+1)%2], runtime)
		}

		printStats(cPair[i], stat, catMasks[i]) // TODO see above
	}

	return nil
}

func printStats(c string, stat stats.RuntimeT, catMask uint64) {
	ref := stats.GetReferenceRuntime(c)
	slowdown := math.NaN()
	if ref != nil {
		slowdown = stat.Mean / ref.Mean
	}

	log.WithFields(log.Fields{
		"Ø":        fmt.Sprintf("%9.2f", stat.Mean),
		"σ":        fmt.Sprintf("%1.6f", stat.Stddev),
		"σ²":       fmt.Sprintf("%1.6f", stat.Vari),
		"runs":     fmt.Sprintf("%3d", stat.Runs),
		"CAT":      fmt.Sprintf("%6x", catMask),
		"slowdown": fmt.Sprintf("%1.6f", slowdown),
	}).Infof("%v", commands.Pretty(c))
}
