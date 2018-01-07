package stats

import (
	"encoding/json"
	"math"

	"github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
)

// CreateJSON creates a JSON representation of the current state
func CreateJSON() ([]byte, error) {
	json, err := json.Marshal(runtimeStats)
	if err != nil {
		return nil, err
	}
	return json, err
}

// StoreJSON parses the JSON and stores it in the state
func StoreJSON(raw []byte) error {
	return json.Unmarshal(raw, &runtimeStats)
}

func newRuntimeT(CATMask uint64, rawRuntimes []DataPerRun) RuntimeT {
	var ret RuntimeT
	ret.update(CATMask, rawRuntimes)
	return ret
}

func (run *RuntimeT) update(CATMask uint64, data []DataPerRun) {
	if run.RawRuntimesByMask == nil {
		temp := make(map[uint64][]DataPerRun, 1)
		run.RawRuntimesByMask = &temp
	}

	if _, exists := (*run.RawRuntimesByMask)[CATMask]; exists {
		(*run.RawRuntimesByMask)[CATMask] = append((*run.RawRuntimesByMask)[CATMask], data...)
	} else {
		(*run.RawRuntimesByMask)[CATMask] = data
	}

	var runtimeSeconds []float64
	for _, v := range *run.RawRuntimesByMask {
		for _, r := range v {
			runtimeSeconds = append(runtimeSeconds, r.Runtime.Seconds())
		}
	}

	var err error
	run.Mean, err = stats.Mean(runtimeSeconds)
	if err != nil {
		log.WithError(err).Errorln("Error while computing mean")
	}
	run.Stddev, err = stats.StandardDeviation(runtimeSeconds)
	if err != nil {
		log.WithError(err).Errorln("Error while computing stddev")
	}
	run.Vari, err = stats.Variance(runtimeSeconds)
	if err != nil {
		log.WithError(err).Errorln("Error while computing variance")
	}
	run.RuntimeSum, err = stats.Sum(runtimeSeconds)
	if err != nil {
		log.WithError(err).Errorln("Error while computing sum")
	}

	run.Runs = len(runtimeSeconds)
}

// GetAllApplications returns a string slice containing all applications that are currently stored
func GetAllApplications() []string {
	return runtimeStats.Commandline.Commands
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
	if math.IsNaN(varianceDiff) {
		runtimeStats.Commandline.VarianceDiff = -1.0
	}
}
