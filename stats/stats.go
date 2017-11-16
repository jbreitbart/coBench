package stats

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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

type StatsT struct {
	// Application command line as a key
	Runtimes map[string]*RuntimePerAppT

	// Command line options passed to coBench
	Commandline CommandlineT
}

type CommandlineT struct {
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

type RuntimePerAppT struct {
	// individual run
	ReferenceRuntimes RuntimeT

	// individual runtime with CAT config used as key
	CATRuntimes *map[uint64]RuntimeT

	// runtime coScheduling without CAT
	CoSchedRuntimes *map[string]RuntimeT

	// runtime coScheduling with CAT
	CoSchedCATRuntimes *map[coSchedCATKey]RuntimeT
}

type RuntimeT struct {
	Mean       float64
	Stddev     float64
	Vari       float64
	RuntimeSum float64
	Runs       int
	RawRuntime []time.Duration
}

// global variable that keeps track of everything
var runtimeStats StatsT

func checkIfReferenceExists(application string) {
	if _, ok := runtimeStats.Runtimes[application]; !ok {
		log.Fatalln("Error while inserting CAT runtime. Application key does not exist.")
	}
}

func AddReferenceTime(application string, referenceTime RuntimeT) {
	var temp RuntimePerAppT
	temp.ReferenceRuntimes = referenceTime

	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*RuntimePerAppT, 1)
	}
	runtimeStats.Runtimes[application] = &temp
}

func AddCATRuntime(application string, CAT uint64, runtime RuntimeT) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[uint64]RuntimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CATRuntimes)[CAT] = runtime
}

func AddCoSchedRuntime(application string, coSchedApplication string, runtime RuntimeT) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = runtime
}

func AddCoSchedCATRuntime(application string, coSchedApplication string, CAT uint64, runtime RuntimeT) {
	checkIfReferenceExists(application)

	key := coSchedCATKey{coSchedApplication, CAT}

	if runtimeStats.Runtimes[application].CoSchedCATRuntimes == nil {
		temp := make(map[coSchedCATKey]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedCATRuntimes = &temp
	}
	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[key] = runtime
}

func StoreToFile(filename string) error {
	json, err := json.Marshal(runtimeStats)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, json, 0644)
	return err
}

func ReadFromFile(filename string) error {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &runtimeStats)
	return err
}

func ComputeRuntimeStats(runtime []time.Duration) RuntimeT {
	var stat RuntimeT
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

func ProcessRuntime(id int, cPair [2]string, catMasks [2]uint64, runtimes [][]time.Duration) error {

	for i, runtime := range runtimes {
		stat := ComputeRuntimeStats(runtime)

		printStats(cPair[i], stat, catMasks[i])

		if catMasks[0] != 0 && catMasks[1] != 0 {
			AddCoSchedCATRuntime(cPair[i], cPair[(i+1)%2], catMasks[i], stat)
		} else {
			AddCoSchedRuntime(cPair[i], cPair[(i+1)%2], stat)
		}
	}

	return nil
}

func printStats(c string, stat RuntimeT, catMask uint64) {
	s := fmt.Sprintf("%v \t %9.2fs avg. runtime \t %1.6f std. dev. \t %1.6f variance \t %3d runs", c, stat.Mean, stat.Stddev, stat.Vari, stat.Runs)
	if catMask != 0 {
		s += fmt.Sprintf("\t %6x CAT", catMask)
	} else {
		s += "\t           "
	}

	/*ref, ok := referenceRuntimes[c]
	if ok {
		s += fmt.Sprintf("\t %1.6f co-slowdown", stat.Mean/ref.Mean)
	} else {
		s += "\t ref missing"
	}*/

	fmt.Println(s)
}

func SetCommandline(cat bool, catBitChunk uint64, catDirs []string, cpus [2]string, commands []string, hermitcore bool, resctrlPath string, runs int, threads string, varianceDiff float64) {
	runtimeStats.Commandline.CAT = cat
	runtimeStats.Commandline.CATChunk = catBitChunk
	runtimeStats.Commandline.CATDirs = catDirs
	runtimeStats.Commandline.CPUs = cpus
	runtimeStats.Commandline.Commands = commands
	runtimeStats.Commandline.HermitCore = hermitcore
	runtimeStats.Commandline.ResctrlPath = resctrlPath
	runtimeStats.Commandline.Runs = runs
	runtimeStats.Commandline.Threads = threads
	runtimeStats.Commandline.VarianceDiff = varianceDiff
}
