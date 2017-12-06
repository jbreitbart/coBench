package stats

import (
	"encoding/json"
	"math/bits"
	"time"

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

// GetCATRuntimes returns all cat individual runtimes with CAT
func GetCATRuntimes(application string) *map[int]RuntimeT {
	_, exists := runtimeStats.Runtimes[application]
	if exists {
		return runtimeStats.Runtimes[application].CATRuntimes
	}
	return nil
}

// GetReferenceRuntime returns the individual runtime without CAT
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

// AddReferenceRuntime adds the individual runtime without CAT
func AddReferenceRuntime(application string, runtime []time.Duration) RuntimeT {
	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*RuntimePerAppT, 1)
	}

	var old RuntimeT

	if runtimeStats.Runtimes[application] != nil {
		old = runtimeStats.Runtimes[application].ReferenceRuntimes
	}
	old.update(NoCATMask, runtime)

	var temp RuntimePerAppT
	temp.ReferenceRuntimes = old
	runtimeStats.Runtimes[application] = &temp

	return old
}

// AddCATRuntime adds the individual runtime with CAT
func AddCATRuntime(application string, CATMask uint64, runtime []time.Duration) RuntimeT {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[int]RuntimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}

	key := bits.OnesCount64(CATMask)
	old := (*runtimeStats.Runtimes[application].CATRuntimes)[key]
	old.update(CATMask, runtime)

	(*runtimeStats.Runtimes[application].CATRuntimes)[key] = old

	return old
}

// AddCoSchedRuntime adds the co-scheduling runtime of 'application' co-scheduled with coSchedApplication without CAT
func AddCoSchedRuntime(application string, coSchedApplication string, runtime []time.Duration) RuntimeT {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}

	old := (*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication]
	old.update(NoCATMask, runtime)

	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = old

	return old
}

// AddCoSchedCATRuntime adds the co-scheduling runtime of 'application' co-scheduled with coSchedApplication with CAT
func AddCoSchedCATRuntime(application string, coSchedApplication string, CATMask uint64, runtime []time.Duration) RuntimeT {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedCATRuntimes == nil {
		temp := make(map[string]map[int]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedCATRuntimes = &temp
	}
	if (*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] == nil {
		temp := make(map[int]RuntimeT, 1)
		(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication] = temp
	}

	key := bits.OnesCount64(CATMask)
	old := (*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][key]
	old.update(CATMask, runtime)

	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][key] = old

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
