package stats

import (
	"log"
	"math/bits"
)

func checkIfReferenceExists(application string) {
	if _, ok := runtimeStats.Runtimes[application]; !ok {
		log.Fatalln("Error while inserting CAT runtime. Application key does not exist. Call AddReferenceRuntime() first.")
	}
}

// AddReferenceRuntime adds the individual runtime without CAT
func AddReferenceRuntime(application string, data []DataPerRun) RuntimeT {
	if runtimeStats.Runtimes == nil {
		runtimeStats.Runtimes = make(map[string]*RuntimePerAppT, 1)
	}

	var old RuntimeT

	if runtimeStats.Runtimes[application] != nil {
		old = runtimeStats.Runtimes[application].ReferenceRuntimes
	}
	old.update(NoCATMask, data)

	var temp RuntimePerAppT
	temp.ReferenceRuntimes = old
	runtimeStats.Runtimes[application] = &temp

	return old
}

// AddCATRuntime adds the individual runtime with CAT
func AddCATRuntime(application string, CATMask uint64, data []DataPerRun) RuntimeT {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CATRuntimes == nil {
		temp := make(map[int]RuntimeT, 1)
		runtimeStats.Runtimes[application].CATRuntimes = &temp
	}

	key := bits.OnesCount64(CATMask)
	old := (*runtimeStats.Runtimes[application].CATRuntimes)[key]
	old.update(CATMask, data)

	(*runtimeStats.Runtimes[application].CATRuntimes)[key] = old

	return old
}

// AddCoSchedRuntime adds the co-scheduling runtime of 'application' co-scheduled with coSchedApplication without CAT
func AddCoSchedRuntime(application string, coSchedApplication string, data []DataPerRun) RuntimeT {
	checkIfReferenceExists(application)

	if runtimeStats.Runtimes[application].CoSchedRuntimes == nil {
		temp := make(map[string]RuntimeT, 1)
		runtimeStats.Runtimes[application].CoSchedRuntimes = &temp
	}

	old := (*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication]
	old.update(NoCATMask, data)

	(*runtimeStats.Runtimes[application].CoSchedRuntimes)[coSchedApplication] = old

	return old
}

// AddCoSchedCATRuntime adds the co-scheduling runtime of 'application' co-scheduled with coSchedApplication with CAT
func AddCoSchedCATRuntime(application string, coSchedApplication string, CATMask uint64, data []DataPerRun) RuntimeT {
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
	old.update(CATMask, data)

	(*runtimeStats.Runtimes[application].CoSchedCATRuntimes)[coSchedApplication][key] = old

	return old
}
