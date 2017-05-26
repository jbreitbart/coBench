package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func runApp() error {
	return nil
}

// global command line parameters
var runs *int
var cpus0 *string
var cpus1 *string
var threads *string

func main() {
	runs = flag.Int("arun", 2, "Number of times the applications are executed")
	commandFile := flag.String("cmd", "cmd.txt", "Text file containing the commands to execute")
	cpus0 = flag.String("cpus0", "0-4", "List of CPUs to be used for the 1st command")
	cpus1 = flag.String("cpus1", "5-9", "List of CPUs to be used for the 2nd command")
	threads = flag.String("threads", "5", "Number of threads to be used")
	//resctrlPath := flag.String("resctrl", "/sys/fs/resctrl/", "Root path of the resctrl file system")
	flag.Parse()

	commands, err := readCommands(*commandFile)
	if err != nil {
		log.Fatalf("Error reading command file %v: %v", *commandFile, err)
	}
	if len(commands) < 2 {
		log.Fatal("You must provide at least 2 commands")
	}

	commandPairs := generateCommandPairs(commands)

	fmt.Println("Executing the following command pairs:")
	for _, c := range commandPairs {
		fmt.Println(c)
	}

	for i, c := range commandPairs {
		fmt.Printf("Running pair %v\n", i)
		fmt.Println(c)
		// TODO run for every combination of CAT setup
		err := runPair(c, i)
		if err != nil {
			log.Fatalf("Error while running pair %v (%v): %v", i, c, err)
		}
	}
}

func runCmdMinTimes(cmd *exec.Cmd, min int, measurement *string, done chan int, errs chan error) {

	i := 0

	for {
		// create a copy of the command
		cmd := *cmd

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)
		*measurement += strconv.FormatInt(elapsed.Nanoseconds(), 10)
		*measurement += "\n"

		if err != nil {
			errs <- fmt.Errorf("Error running %v: %v", cmd.Args, err)
			return
		}

		i++

		// did the other cmd result in an error?
		if len(errs) != 0 {
			return
		}

		d := <-done
		// did we run min times?
		if i == min {
			d++
		}
		done <- d

		// both applications are done
		if d == 2 {
			return
		}
	}
}

func runPair(cPair [2]string, id int) error {
	cmd0 := exec.Command("/bin/sh", "-c", cPair[0])
	cmd1 := exec.Command("/bin/sh", "-c", cPair[1])

	env := os.Environ()
	cmd0.Env = append(env, "GOMP_CPU_AFFINITY="+*cpus0, "OMP_NUM_THREADS="+*threads)
	cmd1.Env = append(env, "GOMP_CPU_AFFINITY="+*cpus1, "OMP_NUM_THREADS="+*threads)

	outfile0, err := os.Create(fmt.Sprintf("%v-0.log", id))
	if err != nil {
		return fmt.Errorf("Error while creating file: %v", err)
	}
	defer outfile0.Close()
	outfile1, err := os.Create(fmt.Sprintf("%v-1.log", id))
	if err != nil {
		return fmt.Errorf("Error while creating file: %v", err)
	}
	defer outfile1.Close()

	cmd0.Stdout = outfile0
	cmd1.Stdout = outfile1

	var measurements [2]string
	// used to count how many apps have reached there min limit
	done := make(chan int, 1)
	done <- 0

	// used to return an error from the go-routines
	errs := make(chan error, 2)

	go runCmdMinTimes(cmd0, *runs, &measurements[0], done, errs)
	runCmdMinTimes(cmd1, *runs, &measurements[1], done, errs)

	if len(errs) != 0 {
		return <-errs
	}

	for i, s := range measurements {
		measurementsFile, err := os.Create(fmt.Sprintf("%v-%v.time", id, i))
		if err != nil {
			return fmt.Errorf("Error while creating file: %v", err)
		}
		defer measurementsFile.Close()

		_, err = measurementsFile.WriteString("#nanoseconds\n" + s)
		if err != nil {
			return fmt.Errorf("Error while writing measurements file: %v", err)
		}
	}

	return nil
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
		commands = append(commands, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("Error scanning commands: " + err.Error())
	}

	return commands, nil
}
