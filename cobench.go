package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

var referenceRuntimes map[string]runtimeT

func main() {
	commandFile := parseArgs()

	commands, err := readCommands(*commandFile)
	if err != nil {
		log.Fatalf("Error reading command file %v: %v", *commandFile, err)
	}
	if len(commands) < 2 {
		log.Fatal("You must provide at least 2 commands")
	}

	if err := storeConfig(commands); err != nil {
		log.Fatalf("Could not save config: %v\n", err)
	}

	commandPairs := generateCommandPairs(commands)

	referenceRuntimes = make(map[string]runtimeT, len(commandPairs))

	// run apps indiviually
	fmt.Println("Running apps individually:")
	for i, cmd := range commands {
		fmt.Printf("Running %v\n", cmd)
		r, err := runSingle(cmd, i)
		if err != nil {
			log.Fatalf("Error running application individually: %v\n", err)
		}
		stat := computeRuntimeStats(r)
		referenceRuntimes[cmd] = stat
		printStats(cmd, stat, 0)
	}
	fmt.Println("Individual runs done. \n")

	// run co-scheduling without cat
	if *cat {
		*cat = false
		for i, c := range commandPairs {
			fmt.Printf("Running pair %v\n", i)
			fmt.Println(c)

			catConfig := []uint64{0, 0}
			runtimes, err := runPair(c, i, catConfig)
			if err != nil {
				log.Fatalf("Error while running pair %v (%v): %v", i, c, err)
			}

			err = processRuntime(i, c, catConfig, runtimes)
			if err != nil {
				log.Fatalf("Error processing runtime: %v", err)
			}
		}
		*cat = true
	}

	minBits := uint64(0)
	numBits := uint64(0)

	if *cat {
		var err error
		minBits, numBits, err = setupCAT()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		defer resetCAT()
	}

	catPairs := generateCatConfigs(minBits, numBits)

	fmt.Println("Executing the following command pairs:")
	for _, c := range commandPairs {
		fmt.Println(c)
	}

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

func generateCommandPairs(commands []string) [][2]string {
	var pairs [][2]string
	for i, c0 := range commands {
		for j, c1 := range commands {
			if i >= j {
				continue
			}
			pairs = append(pairs, [2]string{c0, c1})
		}
	}
	return pairs
}

func readCommands(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("Error opening file " + filename + ": " + err.Error())
	}
	defer file.Close()

	var commands []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		temp := scanner.Text()
		temp = strings.TrimSpace(temp)
		if len(temp) > 0 && temp[0] != '#' {
			commands = append(commands, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("Error scanning commands: " + err.Error())
	}

	return commands, nil
}
