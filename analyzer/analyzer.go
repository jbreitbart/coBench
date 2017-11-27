package main

import (
	"flag"
	"log"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
)

func main() {

	inputFile := flag.String("input", "", "Input result file")
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("No input file provided.\n")
	}

	err := stats.ReadFromFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	apps := stats.GetAllApplications()

	log.Println("Found data for the following applications:")
	for _, app := range apps {
		log.Printf("\t%v\n", app)
	}
	log.Println("")

	indvApps := commands.GenerateIndv(apps)
	CATDatFiles := createIndvCATDatFiles(indvApps)
	writeGNUPlotCATIndvFile(indvApps, CATDatFiles)

	pairs := commands.GeneratePairs(apps)
	CATCoSchedDatFiles := createCoSchedCATDatFiles(pairs)
	writeGNUPlotCATCoSchedFile(pairs, CATCoSchedDatFiles)
}
