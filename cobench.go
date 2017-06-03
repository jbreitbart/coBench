package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/jbreitbart/coBench/bit"
	"github.com/montanaflynn/stats"
)

// global command line parameters
var runs *int
var cpus [2]string
var threads *string
var hermitcore *bool
var cat *bool
var resctrlPath *string

func swap(x, y uint64) (uint64, uint64) {
	return y, x
}

func createDirsCAT(dirs []string) error {
	for _, dir := range dirs {
		err := os.Mkdir(dir, 0777)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("CAT: %v", err)
		}
	}

	return nil
}

func setupCAT() error {
	// TODO hardcoded length
	dirs := []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	numbers := regexp.MustCompile("[0-9]+")

	if err := createDirsCAT(dirs); err != nil {
		return err
	}

	for i, cpu := range cpus {
		cpuIDs := numbers.FindAllString(cpu, -1)
		if len(cpuIDs)%2 != 0 {
			return fmt.Errorf("Unsupported CPU list: %v", cpu)
		}

		var bitset int64

		// loop over every pair
		for i := 0; i < len(cpuIDs); i += 2 {
			var start, end uint64
			start, err := strconv.ParseUint(cpuIDs[i], 10, 64)
			if err != nil {
				return fmt.Errorf("Parse number: %v", start)
			}
			end, err = strconv.ParseUint(cpuIDs[i+1], 10, 64)
			if err != nil {
				return fmt.Errorf("Parse number: %v", end)
			}
			if end < start {
				start, end = swap(start, end)
			}
			for c := start; c <= end; c++ {
				// TODO validate by max cpu size
				bitset = bit.Set(bitset, c)
			}
		}

		file, err := os.OpenFile(dirs[i]+"/cpus", os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			return fmt.Errorf("CAT could not open cpus file: %v", err)
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("%v", bitset))
		if err != nil {
			return fmt.Errorf("CAT could write to cpus file: %v", err)
		}
	}

	return nil
}

func writeCATConfig(configs []int64) error {
	// TODO duplicated line
	dirs := []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	if len(dirs) != len(configs) {
		return fmt.Errorf("Internal error")
	}

	for i, dir := range dirs {
		file, err := os.OpenFile(dir+"/schemata", os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			return fmt.Errorf("CAT could not open cpus file: %v", err)
		}
		defer file.Close()

		// TODO hardcoded string
		_, err = file.WriteString(fmt.Sprintf("L3:0=%x;1=%x", (uint)(configs[i]), (uint)(configs[i])))
		if err != nil {
			return fmt.Errorf("CAT could write to schemata file: %v", err)
		}
	}
	return nil
}

func main() {
	commandFile := parseArgs()

	commands, err := readCommands(*commandFile)
	if err != nil {
		log.Fatalf("Error reading command file %v: %v", *commandFile, err)
	}
	if len(commands) < 2 {
		log.Fatal("You must provide at least 2 commands")
	}

	var minBits int
	var numBits int

	if *cat {
		err := setupCAT()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
		// TODO read from resctrlPath
		minBits = 2
		numBits = 20
	} else {
		minBits = 0
		numBits = 0
	}

	commandPairs := generateCommandPairs(commands)

	fmt.Println("Executing the following command pairs:")
	for _, c := range commandPairs {
		fmt.Println(c)
	}

	for i, c := range commandPairs {
		fmt.Printf("Running pair %v\n", i)
		fmt.Println(c)
		// TODO max chunk size configurable
		for bits := minBits; bits <= numBits; bits += 2 {
			bitsets := []int64{0, 0}
			bitsets[0] = bit.SetFirstN(bitsets[0], bits)
			bitsets[1] = bit.SetLastN(bitsets[1], bits, numBits)

			err := runPair(c, i, bitsets)
			if err != nil {
				log.Fatalf("Error while running pair %v (%v): %v", i, c, err)
			}
		}
	}
}

func runCmdMinTimes(cmd *exec.Cmd, min int, wg *sync.WaitGroup, measurement *string, done chan int, errs chan error) {
	var runtime []float64

	defer wg.Done()

	defer func() {
		mean, _ := stats.Mean(runtime)
		stddev, _ := stats.StandardDeviation(runtime)
		vari, _ := stats.Variance(runtime)

		fmt.Printf("%v \t %9.2fs avg. runtime \t %1.6f std. dev. \t %1.6f variance \t %v runs\n", cmd.Args, mean, stddev, vari, len(runtime))
	}()

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
		if d == len(cpus) {
			// no
			*measurement += "# "
		} else {
			runtime = append(runtime, elapsed.Seconds())
		}
		*measurement += strconv.FormatInt(elapsed.Nanoseconds(), 10)
		*measurement += "\n"

		// did we run min times?
		if i == min {
			d++
		}
		done <- d

		// both applications are done
		if d == len(cpus) {
			return
		}
	}
}

func runPair(cPair [2]string, id int, catConfig []int64) error {
	env := os.Environ()

	if *cat {
		if err := writeCATConfig(catConfig); err != nil {
			return fmt.Errorf("Error while writting CAT config: %v", err)
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

		outfile, err := os.Create(fmt.Sprintf("%v-%v-%x.log", id, i, catConfig[i]))
		if err != nil {
			return fmt.Errorf("Error while creating file: %v", err)
		}
		defer outfile.Close()
		cmds[i].Stdout = outfile
	}

	var measurements [len(cmds)]string
	// used to count how many apps have reached their min limit
	done := make(chan int, 1)
	done <- 0

	// used to return an error from the go-routines
	errs := make(chan error, len(cmds))

	// used to wait for the following 2 goroutines
	var wg sync.WaitGroup
	wg.Add(len(cmds))

	for i, c := range cmds {
		go runCmdMinTimes(c, *runs, &wg, &measurements[i], done, errs)
	}

	wg.Wait()

	for i, s := range measurements {
		measurementsFile, err := os.Create(fmt.Sprintf("%v-%v.time", id, i))
		if err != nil {
			return fmt.Errorf("Error while creating file: %v", err)
		}
		defer measurementsFile.Close()

		out := "# runtime in nanoseconds of \"" + cPair[i] + "\" on CPUs " + cpus[i] + "while \"" + cPair[(i+1)%2] + "\" was running on cores " + cpus[(i+1)%len(cpus)]
		if *cat {
			out += fmt.Sprintf(" with CAT %x ", catConfig[i])
		}
		out += "\n"
		out += s

		_, err = measurementsFile.WriteString(out)
		if err != nil {
			return fmt.Errorf("Error while writing measurements file: %v", err)
		}
	}

	if len(errs) != 0 {
		return <-errs
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

func parseArgs() *string {
	runs = flag.Int("arun", 2, "Number of times the applications are executed")
	commandFile := flag.String("cmd", "cmd.txt", "Text file containing the commands to execute")
	cpus0 := flag.String("cpus0", "0-4", "List of CPUs to be used for the 1st command")
	cpus1 := flag.String("cpus1", "5-9", "List of CPUs to be used for the 2nd command")
	threads = flag.String("threads", "5", "Number of threads to be used")
	hermitcore = flag.Bool("hermitcore", false, "Use if you are executing hermitcore binaries")
	cat = flag.Bool("cat", false, "Measure with all CAT settings")
	resctrlPath = flag.String("resctrl", "/sys/fs/resctrl/", "Root path of the resctrl file system")
	flag.Parse()

	cpus[0] = *cpus0
	cpus[1] = *cpus1

	return commandFile
}
