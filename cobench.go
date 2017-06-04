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

func main() {
	commandFile := parseArgs()

	commands, err := readCommands(*commandFile)
	if err != nil {
		log.Fatalf("Error reading command file %v: %v", *commandFile, err)
	}
	if len(commands) < 2 {
		log.Fatal("You must provide at least 2 commands")
	}

	minBits := uint64(0)
	numBits := uint64(0)

	if *cat {
		var err error
		minBits, numBits, err = setupCAT()
		if err != nil {
			log.Fatalf("%v\n", err)
		}
	}

	commandPairs := generateCommandPairs(commands)
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

func processRuntime(id int, cPair [2]string, catMasks []uint64, runtimes [][]time.Duration) error {

	statsFile, err := os.Create("stats")
	if err == nil {
		defer statsFile.Close()

		// write header
		statsFile.WriteString("cmd \t avg. runtime (s) \t std. dev. \t variance \t runs")
		if *cat {
			statsFile.WriteString("\t CAT")
		}
		statsFile.WriteString("\t nom. perf\n")
	} else {
		if os.IsExist(err) {
			statsFile, err = os.OpenFile("stats", os.O_WRONLY|os.O_APPEND, 0777)
			if err != nil {
				return fmt.Errorf("Error while opening file: %v", err)
			}
			defer statsFile.Close()
		} else {
			return fmt.Errorf("Error while creating file: %v", err)
		}
	}

	for i, runtime := range runtimes {
		var runtimeSeconds []float64
		for _, r := range runtime {
			runtimeSeconds = append(runtimeSeconds, r.Seconds())
		}

		mean, _ := stats.Mean(runtimeSeconds)
		stddev, _ := stats.StandardDeviation(runtimeSeconds)
		vari, _ := stats.Variance(runtimeSeconds)
		runtimeSum, _ := stats.Sum(runtimeSeconds)

		s := fmt.Sprintf("%v \t %9.2fs avg. runtime \t %1.6f std. dev. \t %1.6f variance \t %3d runs", cPair[i], mean, stddev, vari, len(runtime))
		if *cat {
			s += fmt.Sprintf("\t %6x CAT", catMasks[i])
		} else {
			s += "\t           "
		}
		s += fmt.Sprintf("\t %1.6f nom. perf", (float64)(len(runtime))/runtimeSum)
		fmt.Println(s)

		statsFile.WriteString(fmt.Sprintf("%v \t %v \t %v \t %v \t %v", cPair[i], mean, stddev, vari, len(runtime)))
		if *cat {
			statsFile.WriteString(fmt.Sprintf("\t %6x", catMasks[i]))
		} else {
			statsFile.WriteString("\t       ")
		}
		statsFile.WriteString(fmt.Sprintf("\t %v\n", (float64)(len(runtime))/runtimeSum))
	}
	fmt.Print("\n")

	for i, runtime := range runtimes {
		filename := fmt.Sprintf("%v-%v", id, i)
		if *cat {
			filename += fmt.Sprintf("-%x", catMasks[i])
		}
		filename += ".time"
		measurementsFile, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("Error while creating file: %v", err)
		}
		defer measurementsFile.Close()

		out := "# runtime in nanoseconds of \"" + cPair[i] + "\" on CPUs " + cpus[i] + "while \"" + cPair[(i+1)%2] + "\" was running on cores " + cpus[(i+1)%len(cpus)]
		if *cat {
			out += fmt.Sprintf(" with CAT %6x ", catMasks[i])
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
