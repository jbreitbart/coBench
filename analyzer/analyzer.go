package main

import (
	"flag"
	"io/ioutil"
	"log"
	"math/bits"
	"sort"
	"strconv"

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

	CATDatFiles := createIndvCATDatFiles(apps)
	writeGNUPlotCATIndvFile(apps, CATDatFiles)
}

// Returns the filenames
func createIndvCATDatFiles(apps []string) []string {
	log.Println("Creating dat files for individual CAT runs.")

	ret := make([]string, 0)

	for i, app := range apps {
		catRuntime := stats.GetAllCATRuntimes(app)
		if catRuntime == nil {
			continue
		}

		filename := indvCATDatFilename(i)
		out := "# " + app + "\n"
		out += "# Bits Runtime Std.Dev.\n"

		var sortedKeys []uint64
		for k, _ := range *catRuntime {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Slice(sortedKeys[:], func(i, j int) bool {
			return bits.OnesCount64(sortedKeys[i]) < bits.OnesCount64(sortedKeys[j])
		})

		for _, k := range sortedKeys {
			v, exist := (*catRuntime)[k]
			if !exist {
				log.Fatalln("Could not find key. Should never happen.")
			}
			out += strconv.FormatFloat(1.5*float64(bits.OnesCount64(k)), 'E', -1, 64) + " " + strconv.FormatFloat(v.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(v.Stddev, 'E', -1, 64) + "\n"
		}

		ref := stats.GetReferenceRuntime(app)
		// TODO fix hardcoded 20
		out += strconv.FormatFloat(1.5*20, 'E', -1, 64) + " " + strconv.FormatFloat(ref.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref.Stddev, 'E', -1, 64) + "\n"

		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.Fatalf("Error while write file %v: %v", filename, err)
		}

		ret = append(ret, filename)
	}

	return ret
}

func indvCATDatFilename(i int) string {
	return strconv.FormatInt(int64(i), 10) + "-cat.dat"
}
