package stats

import (
	"log"
	"time"

	"github.com/montanaflynn/stats"
)

// TODO comment functions! :)

// GetAllApplications returns a string slice containing all applications that are currently stored
func GetAllApplications() []string {
	return runtimeStats.Commandline.Commands
}

// GetCoSchedCATRuntimes returns the runtime of application when running in parallel to cosched with CAT
func GetCoSchedCATRuntimes(application string, cosched string) *map[uint64]RuntimeT {
	temp, exists := runtimeStats.Runtimes[application]
	if !exists {
		return nil
	}

	ret, exists := (*(*temp).CoSchedCATRuntimes)[cosched]
	if exists {
		return &ret
	}

	return nil
}

// GetCoSchedRuntimes returns the runtime of application when running in parallel to cosched without CAT
func GetCoSchedRuntimes(application string, cosched string) *RuntimeT {
	temp, exists := runtimeStats.Runtimes[application]
	if !exists {
		return nil
	}

	ret, exists := (*temp.CoSchedRuntimes)[cosched]
	if !exists {
		return nil
	}

	return &ret
}

// GetIndvCATRuntimes
func GetIndvCATRuntimes(application string) *map[uint64]RuntimeT {
	temp, exists := runtimeStats.Runtimes[application]
	if !exists {
		return nil
	}

	return temp.CATRuntimes
}

func GetReferenceRuntime(application string) *RuntimeT {
	_, exists := runtimeStats.Runtimes[application]
	if exists {
		return &runtimeStats.Runtimes[application].ReferenceRuntimes
	}
	return nil
}

func checkIfReferenceExists(application string) {
	if _, ok := runtimeStats.Runtimes[application]; !ok {
		log.Fatalln("Error while inserting CAT runtime. Application key does not exist.")
	}
}

func AddReferenceRuntime(application string, runtime []time.Duration) {
	var temp RuntimePerAppT
	temp.ReferenceRuntimes = ComputeRuntimeStats(runtime)

	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*RuntimePerAppT, 1)
	}
	// TODO check if already available and sum up?
	runtimeStats.Runtimes[application] = &temp
}

func AddCATRuntime(application string, CAT uint64, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[uint64]RuntimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}
	// TODO check if already available and sum up?
	(*runtimeStats.Runtimes[application].CATRuntimes)[CAT] = ComputeRuntimeStats(runtime)
}

func AddCoSchedRuntime(application string, coSchedApplication string, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}
	// TODO check if already available and sum up?
	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = ComputeRuntimeStats(runtime)
}

func AddCoSchedCATRuntime(application string, coSchedApplication string, CAT uint64, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedCATRuntimes == nil {
		temp := make(map[string]map[uint64]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedCATRuntimes = &temp
	}
	if (*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] == nil {
		temp := make(map[uint64]RuntimeT, 1)
		(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] = temp
	}

	// TODO check if already available and sum up?
	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][CAT] = ComputeRuntimeStats(runtime)
}

// ComputeRuntimeStats creates a RuntimeT object based on the runtime
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
