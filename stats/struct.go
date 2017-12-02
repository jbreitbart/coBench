package stats

import (
	"time"
)

// NoCATMask is used as a special value when CAT is not used
const NoCATMask = 0

// CommandlineT is used to store all command line parameters
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

// RuntimeT contains a set of runtimes and statistic values
type RuntimeT struct {
	Mean              float64
	Stddev            float64
	Vari              float64
	RuntimeSum        float64
	Runs              int
	RawRuntimesByMask *map[uint64][]time.Duration
}

// RuntimePerAppT store runtime values with different combinations for one application
type RuntimePerAppT struct {
	// individual run
	ReferenceRuntimes RuntimeT

	// individual runtime with number of bits set in CAT mask
	CATRuntimes *map[int]RuntimeT

	// runtime coScheduling without CAT
	CoSchedRuntimes *map[string]RuntimeT

	// runtime coScheduling with CAT
	CoSchedCATRuntimes *map[string]map[int]RuntimeT
}

// StatsT contains every information of a benchmark run
type StatsT struct {
	// Application command line as a key
	Runtimes map[string]*RuntimePerAppT

	// Command line options passed to coBench
	Commandline CommandlineT

	// TODO add hardware info

	// TODO version info which struct version is used
}

// global variable that keeps track of everything
var runtimeStats StatsT // TODO we should not keep a global state in the package
