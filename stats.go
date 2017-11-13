package main

// TODO move to own package

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
)

// used as a key
type coSchedCATKey struct {
	Application string
	CAT         uint64
}

func (k coSchedCATKey) MarshalText() ([]byte, error) {
	c := make([]byte, 8)
	binary.LittleEndian.PutUint64(c, k.CAT)

	a := []byte(k.Application)

	return append(c, a...), nil
}

func (k coSchedCATKey) UnmarshalText(text []byte) error {
	c := text[0:8]
	k.CAT = binary.LittleEndian.Uint64(c)

	k.Application = string(text[8:])

	return nil
}

type statsT struct {
	// Application command line as a key
	Runtimes map[string]*runtimePerAppT

	// Command line options passed to coBench
	Commandline commandlineT
}

type commandlineT struct {
	Runs         int
	VarianceDiff float64
	CPUs         [2]string
	Threads      string
	HermitCore   bool
	CAT          bool
	CATChunk     uint64
	CATDirs      []string
	ResctrlPath  string
	Commands     []string
}

type runtimePerAppT struct {
	// individual run
	ReferenceRuntimes runtimeT

	// individual runtime with CAT config used as key
	CATRuntimes *map[uint64]runtimeT

	// runtime coScheduling without CAT
	CoSchedRuntimes *map[string]runtimeT

	// runtime coScheduling with CAT
	CoSchedCATRuntimes *map[coSchedCATKey]runtimeT
}

type runtimeT struct {
	Mean       float64
	Stddev     float64
	Vari       float64
	RuntimeSum float64
	Runs       int
	RawRuntime []time.Duration
}

// global variable that keeps track of everything
var runtimeStats statsT

func checkIfReferenceExists(application string) {
	if _, ok := runtimeStats.Runtimes[application]; !ok {
		log.Fatalln("Error while inserting CAT runtime. Application key does not exist.")
	}
}

func addReferenceTime(application string, referenceTime runtimeT) {
	var temp runtimePerAppT
	temp.ReferenceRuntimes = referenceTime

	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*runtimePerAppT, 1)
	}
	runtimeStats.Runtimes[application] = &temp
}

func addCATRuntime(application string, CAT uint64, runtime runtimeT) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[uint64]runtimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CATRuntimes)[CAT] = runtime
}

func addCoSchedRuntime(application string, coSchedApplication string, runtime runtimeT) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]runtimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = runtime
}

func addCoSchedCATRuntime(application string, coSchedApplication string, CAT uint64, runtime runtimeT) {
	checkIfReferenceExists(application)

	key := coSchedCATKey{coSchedApplication, CAT}

	if runtimeStats.Runtimes[application].CoSchedCATRuntimes == nil {
		temp := make(map[coSchedCATKey]runtimeT, 1)
		runtimeStats.Runtimes[application].CoSchedCATRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[key] = runtime
}

func storeToFile(filename string) error {
	json, err := json.Marshal(runtimeStats)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, json, 0644)
	return err
}

func readFromFile(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &runtimeStats)
	return err
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
	stat.RawRuntime = runtime

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
		statsFile.WriteString("cmd \t avg. runtime (s) \t std. dev. \t variance \t runs \t CAT \t co-slowdown\n")
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
	s += fmt.Sprintf("\t %1.6f", stat.Mean/ref.Mean)

	s += "\n"

	_, err := statsFile.WriteString(s)
	if err != nil {
		return err
	}

	return nil
}

func processRuntime(id int, cPair [2]string, catMasks [2]uint64, runtimes [][]time.Duration) error {

	statsFile, err := openStatsFile()
	if err != nil {
		return err
	}
	defer statsFile.Close()

	for i, runtime := range runtimes {
		stat := computeRuntimeStats(runtime)

		printStats(cPair[i], stat, catMasks[i])
		writeToStatsFile(statsFile, cPair[i], stat, catMasks[i])

		if catMasks[0] != 0 && catMasks[1] != 0 {
			addCoSchedCATRuntime(cPair[i], cPair[(i+1)%2], catMasks[i], stat)
		} else {
			addCoSchedRuntime(cPair[i], cPair[(i+1)%2], stat)
		}
	}

	fmt.Print("\n")

	for i, runtime := range runtimes {

		filename := fmt.Sprintf("%v-%v", id, i)
		if *cat {
			filename += fmt.Sprintf("-%x", catMasks[i])
		}
		filename += ".time"

		header := "# runtime in nanoseconds of \"" + cPair[i] + "\" on CPUs " + cpus[i] + " while \"" + cPair[(i+1)%2] + "\" was running on cores " + cpus[(i+1)%len(cpus)]
		if *cat {
			header += fmt.Sprintf(" with CAT %6x ", catMasks[i])
		}
		header += "\n"

		file, err := createStatsFile(filename, header)
		if err != nil {
			return err
		}
		defer file.Close()

		err = writeRuntimeFile(file, runtime)
		if err != nil {
			return err
		}
	}

	return nil
}

func createStatsFile(filename string, header string) (*os.File, error) {
	measurementsFile, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("Error while creating file: %v", err)
	}
	_, err = measurementsFile.WriteString(header)

	return measurementsFile, err
}

func writeRuntimeFile(file *os.File, runtime []time.Duration) error {
	var out string
	for _, r := range runtime {
		out += strconv.FormatInt(r.Nanoseconds(), 10)
		out += "\n"
	}

	_, err := file.WriteString(out)
	if err != nil {
		return fmt.Errorf("Error while writing measurements file: %v", err)
	}
	return nil
}
