package main

import (
	"flag"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
	log "github.com/sirupsen/logrus"
)

func main() {
	//log.SetLevel(log.DebugLevel)

	inputFile := flag.String("input", "", "Input result file")
	flag.Parse()

	if *inputFile == "" {
		log.Fatalln("No input file provided. Use -input <file>")
	}

	err := stats.ReadFromFile(*inputFile)
	if err != nil {
		log.WithError(err).Fatalln("Cannot read input file")
	}

	apps := stats.GetAllApplications()

	log.Infoln("Found data for the following applications:")
	for i, app := range apps {
		log.WithFields(log.Fields{
			"app": app,
		}).Infof("%v", i)
	}

	indvApps := commands.GenerateIndv(apps)
	CATDatFiles, perfNames := createIndvCATDatFiles(indvApps)
	writeGNUPlotCATIndvFile(indvApps, CATDatFiles, perfNames)

	pairs := commands.GeneratePairs(apps)
	CATCoSchedDatFiles, perfNames := createCoSchedCATDatFiles(pairs, false)
	writeGNUPlotCATCoSchedFile(pairs, CATCoSchedDatFiles, perfNames, false)

	CATCoSchedDatFiles, perfNames = createCoSchedCATDatFiles(pairs, true)
	writeGNUPlotCATCoSchedFile(pairs, CATCoSchedDatFiles, perfNames, true)
}
