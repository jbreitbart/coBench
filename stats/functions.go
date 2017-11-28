package stats

import (
	"log"
	"math/bits"
	"time"

	"github.com/montanaflynn/stats"
)

// TODO comment functions! :)

// GetAllApplications returns a string slice containing all applications that are currently stored
func GetAllApplications() []string {
	return runtimeStats.Commandline.Commands
}

// GetCoSchedCATRuntimes returns the runtime of application when running in parallel to cosched with CAT
func GetCoSchedCATRuntimes(application string, cosched string) *map[int]RuntimeT {
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
func GetIndvCATRuntimes(application string) *map[int]RuntimeT {
	_, exists := runtimeStats.Runtimes[application]
	if exists {
		return runtimeStats.Runtimes[application].CATRuntimes
	}
	return nil
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
		log.Fatalln("Error while inserting CAT runtime. Application key does not exist. Call AddReferenceRuntime() first.")
	}
}

func AddReferenceRuntime(application string, runtime []time.Duration) {
	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*RuntimePerAppT, 1)
	}

	var old RuntimeT

	if runtimeStats.Runtimes[application] != nil {
		old = runtimeStats.Runtimes[application].ReferenceRuntimes
	}

	var temp RuntimePerAppT
	temp.ReferenceRuntimes = ComputeRuntimeStats(runtime, NoCATMask, old)
	runtimeStats.Runtimes[application] = &temp
}

func AddCATRuntime(application string, CAT uint64, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[int]RuntimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}

	key := bits.OnesCount64(CAT)
	old := (*runtimeStats.Runtimes[application].CATRuntimes)[key]

	(*runtimeStats.Runtimes[application].CATRuntimes)[key] = ComputeRuntimeStats(runtime, CAT, old)
}

func AddCoSchedRuntime(application string, coSchedApplication string, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}

	old := (*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication]

	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = ComputeRuntimeStats(runtime, NoCATMask, old)
}

func AddCoSchedCATRuntime(application string, coSchedApplication string, CAT uint64, runtime []time.Duration) {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedCATRuntimes == nil {
		temp := make(map[string]map[int]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedCATRuntimes = &temp
	}
	if (*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] == nil {
		temp := make(map[int]RuntimeT, 1)
		(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] = temp
	}

	key := bits.OnesCount64(CAT)
	old := (*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][key]

	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][key] = ComputeRuntimeStats(runtime, CAT, old)
}

// ComputeRuntimeStats creates a RuntimeT object based on the runtime
func ComputeRuntimeStats(runtime []time.Duration, CATMask uint64, old RuntimeT) RuntimeT {
	// old will be updated and returned

	if old.RawRuntimesByMask == nil {
		temp := make(map[uint64][]time.Duration, 1)
		old.RawRuntimesByMask = &temp
	}
	if _, exists := (*old.RawRuntimesByMask)[CATMask]; exists {
		(*old.RawRuntimesByMask)[CATMask] = append((*old.RawRuntimesByMask)[CATMask], runtime...)
	} else {
		(*old.RawRuntimesByMask)[CATMask] = runtime
	}

	var runtimeSeconds []float64
	for _, v := range *old.RawRuntimesByMask {
		for _, r := range v {
			runtimeSeconds = append(runtimeSeconds, r.Seconds())
		}
	}

	// TODO handle error?
	old.Mean, _ = stats.Mean(runtimeSeconds)
	old.Stddev, _ = stats.StandardDeviation(runtimeSeconds)
	old.Vari, _ = stats.Variance(runtimeSeconds)
	old.RuntimeSum, _ = stats.Sum(runtimeSeconds)

	old.Runs = len(runtimeSeconds)

	return old
}

// SetCommandline stores the command line options in the config struct
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
