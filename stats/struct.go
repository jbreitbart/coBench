package stats

import (
	"time"
)

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

// Used as a key for maps
type coSchedCATKey struct {
	Application string
	CAT         uint64
}

type RuntimeT struct {
	Mean       float64
	Stddev     float64
	Vari       float64
	RuntimeSum float64
	Runs       int
	RawRuntime []time.Duration
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

type StatsT struct {
	// Application command line as a key
	Runtimes map[string]*RuntimePerAppT

	// Command line options passed to coBench
	Commandline CommandlineT
}

// global variable that keeps track of everything
var runtimeStats StatsT