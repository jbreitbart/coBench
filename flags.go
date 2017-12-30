package main

import (
	"flag"
	"math"
	"time"

	"github.com/jbreitbart/coBench/stats"
	"github.com/multiplay/go-slack/chat"
	"github.com/multiplay/go-slack/lrhook"
	log "github.com/sirupsen/logrus"
)

// global command line parameters
var runs *int
var cpus [2]string
var threads *string
var hermitcore *bool
var noCoSched *bool
var noIndvSched *bool

var resctrlPath *string
var cat *bool
var inverseCat *bool
var catBitChunk *uint64
var catDirs []string

var varianceDiff *float64

var perfStat *string

var resultFilename *string

var slackChannel *string
var slackWebhook *string
var slackLevel *int

func parseArgs() *string {
	runs = flag.Int("runs", 2, "Number of times the applications are executed")
	commandFile := flag.String("cmd", "cmd.txt", "Text file containing the commands to execute")

	cpus0 := flag.String("cpus0", "0-4", "List of CPUs to be used for the 1st command")
	cpus1 := flag.String("cpus1", "5-9", "List of CPUs to be used for the 2nd command")
	threads = flag.String("threads", "5", "Number of threads to be used")

	cat = flag.Bool("cat", false, "Measure with all CAT settings")
	inverseCat = flag.Bool("cat-inverse", false, "Inverse the CAT masks")
	catBitChunk = flag.Uint64("catChunk", 2, "Bits changed from one run to the next")
	resctrlPath = flag.String("resctrl", "/sys/fs/resctrl/", "Root path of the resctrl file system")

	hermitcore = flag.Bool("hermitcore", false, "Use if you are executing hermitcore binaries")

	varianceDiff = flag.Float64("variance", math.NaN(), "Minimum differences in variance required between runs")

	noCoSched = flag.Bool("no-cosched", false, "Disable co-scheduling")
	noIndvSched = flag.Bool("no-indv", false, "Disable the individual runs")

	perfStat = flag.String("pstat", "", "If set commands are with perf stat -e <param>. Param could be intel_cqm/llc_occupancy/,LLC-load-misses")

	resultFilename = flag.String("output", time.Now().Format("06-01-02-15-04-05.result.json"), "Name of the result json file")

	slackChannel = flag.String("slack-channel", "#cobench", "The channel coBench will use for logging")
	slackWebhook = flag.String("slack-webhook", "", "The webhook of your slack application")
	slackLevel = flag.Int("slack-level", 1, "Select the lowest log level forwarded to slack. 0: Debug; 1: Info; 2: Warn")

	flag.Parse()
	catDirs = []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	if *slackWebhook != "" {
		cfg := lrhook.Config{
			MinLevel: log.InfoLevel,
			Message: chat.Message{
				Channel:   *slackChannel,
				IconEmoji: ":ghost:",
			},
		}
		if *slackLevel == 0 {
			cfg.MinLevel = log.DebugLevel
		}
		if *slackLevel == 2 {
			cfg.MinLevel = log.WarnLevel
		}
		h := lrhook.New(cfg, *slackWebhook)
		log.AddHook(h)
	}

	if *runs < 1 {
		log.Fatalln("runs must be > 0")
	}
	if *catBitChunk < 1 {
		log.Fatalln("catChunk must be > 0")
	}

	cpus[0] = *cpus0
	cpus[1] = *cpus1

	return commandFile
}

func storeConfig(commands []string) {
	stats.SetCommandline(*cat, *catBitChunk, catDirs, cpus, commands, *hermitcore, *resctrlPath, *runs, *threads, *varianceDiff)
}
