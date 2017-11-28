package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
)

func main() {
	commandFile := parseArgs()

	commandStrings, err := commands.Read(*commandFile)
	if err != nil {
		log.Fatalf("Error reading command file %v: %v", *commandFile, err)
	}
	if len(commandStrings) < 2 {
		log.Fatal("You must provide at least 2 commands")
	}

	storeConfig(commandStrings)

	defer cleanup()

	indvCommands := commands.GenerateIndv(commandStrings)

	if len(indvCommands) != len(commandStrings) {
		log.Printf("Remove %v duplicates from commands for individual runs.\n", len(commandStrings)-len(indvCommands))
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
	err := stats.StoreToFile(time.Now().Format("06-01-02-15-04-05.result.json"))
	if err != nil {
		log.Fatalf("Error while writing measurements to file: %v\n", err)
	}
}

func individualRuns(commands []string) {

	fmt.Println("Running apps individually.")

	// run app individually without CAT (if CAT was requested)
	for i, c := range commands {
		catConfig := [2]uint64{stats.NoCATMask, stats.NoCATMask}

		fmt.Printf("Running %v\n", c)
		r, err := runSingle(c, i, catConfig)
		if err != nil {
			log.Fatalf("Error running application individually: %v\n", err)
		}
		stat := stats.ComputeRuntimeStats(r, catConfig[0], stats.RuntimeT{}) // TODO remove this call
		printStats(c, stat, catConfig[0])

		stats.AddReferenceRuntime(c, r)
	}

	if !*cat {
		return
	}

	minBits, numBits, err := setupCAT()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer resetCAT()

	catPairs := generateCatConfigs(minBits, numBits)

	for i, c := range commands {

		fmt.Printf("Running %v\n", c)

		for _, catConfig := range catPairs {
			runtime, err := runSingle(c, i, catConfig)
			if err != nil {
				log.Fatalf("Error running application individually: %v\n", err)
			}
			stat := stats.ComputeRuntimeStats(runtime, catConfig[0], stats.RuntimeT{}) // TODO remove this call

			printStats(c, stat, catConfig[0])

			stats.AddCATRuntime(c, catConfig[0], runtime)
		}
	}

	fmt.Println("Individual runs done.")
}

func coSchedRuns(commandPairs [][2]string) {
	fmt.Println("Executing the following command pairs:")
	for _, c := range commandPairs {
		fmt.Println(c)
	}

	// run co-scheduling *without* cat
	for i, c := range commandPairs {
		fmt.Printf("Running pair %v\n", i)
		fmt.Println(c)

		catConfig := [2]uint64{stats.NoCATMask, stats.NoCATMask}
		runtimes, err := runPair(c, i, catConfig)
		if err != nil {
			log.Fatalf("Error while running pair %v (%v): %v", i, c, err)
		}

		err = processRuntime(i, c, catConfig, runtimes)
		if err != nil {
			log.Fatalf("Error processing runtime: %v", err)
		}
	}

	if !*cat {
		return
	}

	minBits, numBits, err := setupCAT()
	if err != nil {
		log.Fatalf("%v\n", err)
	}
	defer resetCAT()

	catPairs := generateCatConfigs(minBits, numBits)

	for i, c := range commandPairs {
		fmt.Printf("Running pair %v\n", i)
		fmt.Println(c)

		for _, catConfig := range catPairs {
			runtimes, err := runPair(c, i, catConfig)
			if err != nil {
				log.Fatalf("Error while running pair %v (%v): %v", i, c, err)
			}

			err = processRuntime(i, c, catConfig, runtimes)
			if err != nil {
				log.Fatalf("Error processing runtime: %v", err)
			}
		}
	}
}

func processRuntime(id int, cPair [2]string, catMasks [2]uint64, runtimes [][]time.Duration) error {

	for i, runtime := range runtimes {
		stat := stats.ComputeRuntimeStats(runtime, catMasks[i], stats.RuntimeT{}) // TODO remove this call

		printStats(cPair[i], stat, catMasks[i])

		if catMasks[0] != 0 && catMasks[1] != 0 {
			stats.AddCoSchedCATRuntime(cPair[i], cPair[(i+1)%2], catMasks[i], runtime)
		} else {
			stats.AddCoSchedRuntime(cPair[i], cPair[(i+1)%2], runtime)
		}
	}

	return nil
}

func printStats(c string, stat stats.RuntimeT, catMask uint64) {
	s := fmt.Sprintf("%v \t %9.2fs avg. runtime \t %1.6f std. dev. \t %1.6f variance \t %3d runs", c, stat.Mean, stat.Stddev, stat.Vari, stat.Runs)
	if catMask != 0 {
		s += fmt.Sprintf("\t %6x CAT", catMask)
	} else {
		s += "\t           "
	}

	ref := stats.GetReferenceRuntime(c)
	if ref != nil {
		s += fmt.Sprintf("\t %1.6f co-slowdown", stat.Mean/ref.Mean)
	} else {
		s += "\t ref missing"
	}

	fmt.Println(s)
}
