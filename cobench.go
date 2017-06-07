package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
)

// global command line parameters
var runs *int
var cpus [2]string
var threads *string
var hermitcore *bool

var resctrlPath *string
var cat *bool
var catBitChunk *uint64

type runtimeT struct {
	Mean       float64
	Stddev     float64
	Vari       float64
	RuntimeSum float64
	Runs       int
}

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

func computeRuntimeStats(runtime []time.Duration) runtimeT {
	var stat runtimeT
	var runtimeSeconds []float64
	for _, r := range runtime {
		runtimeSeconds = append(runtimeSeconds, r.Seconds())
	}

	// TODO handle error?
	stat.Mean, _ = stats.Mean(runtimeSeconds)
	stat.Stddev, _ = stats.StandardDeviation(runtimeSeconds)
	stat.Vari, _ = stats.Variance(runtimeSeconds)
	stat.RuntimeSum, _ = stats.Sum(runtimeSeconds)

	stat.Runs = len(runtime)

	return stat
}

func openStatsFile() (*os.File, error) {
	var statsFile *os.File
	if _, err := os.Stat("stats"); os.IsNotExist(err) {
		// stats does not exist
		statsFile, err = os.Create("stats")
		if err != nil {
			return nil, fmt.Errorf("Error while creating file: %v", err)
		}

		// write header
		statsFile.WriteString("cmd \t avg. runtime (s) \t std. dev. \t variance \t runs")
		if *cat {
			statsFile.WriteString("\t CAT")
		}
		statsFile.WriteString("\t co-slowdown\n")
	} else {
		statsFile, err = os.OpenFile("stats", os.O_WRONLY|os.O_APPEND, 0777)
		if err != nil {
			return nil, fmt.Errorf("Error while opening file: %v", err)
		}
	}
	return statsFile, nil
}

func printStats(c string, stat runtimeT, catMask uint64) {
	s := fmt.Sprintf("%v \t %9.2fs avg. runtime \t %1.6f std. dev. \t %1.6f variance \t %3d runs", c, stat.Mean, stat.Stddev, stat.Vari, stat.Runs)
	if *cat {
		s += fmt.Sprintf("\t %6x CAT", catMask)
	} else {
		s += "\t           "
	}

	ref, ok := referenceRuntimes[c]
	if ok {
		s += fmt.Sprintf("\t %1.6f co-slowdown", stat.Mean/ref.Mean)
	} else {
		s += "\t ref missing"
	}

	fmt.Println(s)
}

func writeToStatsFile(statsFile *os.File, c string, stat runtimeT, catMask uint64) error {
	s := fmt.Sprintf("%v \t %v \t %v \t %v \t %v", c, stat.Mean, stat.Stddev, stat.Vari, stat.Runs)
	if *cat {
		s += fmt.Sprintf("\t %6x", catMask)
	} else {
		s += "\t       "
	}

	ref := referenceRuntimes[c]
	s += fmt.Sprintf("\t %1.6f co-slowdown", stat.Mean/ref.Mean)

	s += "\n"

	_, err := statsFile.WriteString(s)
	if err != nil {
		return err
	}

	return nil
}

func writeRuntimeFile(c string, cCo string, id int, i int, catMask uint64, runtime []time.Duration) error {
	filename := fmt.Sprintf("%v-%v", id, i)
	if *cat {
		filename += fmt.Sprintf("-%x", catMask)
	}
	filename += ".time"
	measurementsFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error while creating file: %v", err)
	}
	defer measurementsFile.Close()

	out := "# runtime in nanoseconds of \"" + c + "\" on CPUs " + cpus[i] + " while \"" + cCo + "\" was running on cores " + cpus[(i+1)%len(cpus)]
	if *cat {
		out += fmt.Sprintf(" with CAT %6x ", catMask)
	}
	out += "\n"
	for _, r := range runtime {
		out += strconv.FormatInt(r.Nanoseconds(), 10)
		out += "\n"
	}

	_, err = measurementsFile.WriteString(out)
	if err != nil {
		return fmt.Errorf("Error while writing measurements file: %v", err)
	}
	return nil
}

func processRuntime(id int, cPair [2]string, catMasks []uint64, runtimes [][]time.Duration) error {

	statsFile, err := openStatsFile()
	if err != nil {
		return err
	}
	defer statsFile.Close()

	for i, runtime := range runtimes {
		stat := computeRuntimeStats(runtime)

		printStats(cPair[i], stat, catMasks[i])
		writeToStatsFile(statsFile, cPair[i], stat, catMasks[i])
	}

	fmt.Print("\n")

	for i, runtime := range runtimes {

		err := writeRuntimeFile(cPair[i], cPair[(i+1)%2], id, i, catMasks[i], runtime)
		if err != nil {
			return err
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

	flag.Parse()

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
