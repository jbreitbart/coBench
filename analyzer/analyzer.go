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

	for i, app := range apps {
		catRuntime := stats.GetAllCATRuntimes(app)
		if catRuntime == nil {
			continue
		}

		filename := indvCATRuntime(i)
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
			out += strconv.FormatInt(int64(bits.OnesCount64(k)), 10) + " " + strconv.FormatFloat(v.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(v.Stddev, 'E', -1, 64) + "\n"
		}

		ref := stats.GetReferenceRuntime(app)
		out += "20 " + strconv.FormatFloat(ref.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref.Stddev, 'E', -1, 64) + "\n"

		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.Fatalf("Error while write file %v: %v", filename, err)
		}
	}

	err = ioutil.WriteFile("indv-cat.plot", []byte(gnuplotStringCAT(apps)), 0644)
	if err != nil {
		log.Fatalf("Error while write file indv-cat.plot: %v", err)
	}
}

func indvCATRuntime(i int) string {
	return strconv.FormatInt(int64(i), 10) + "-cat.dat"
}

func gnuplotStringCAT(apps []string) string {
	var ret string
	ret += "set terminal pdf\n"
	ret += "set output 'indv-cat.pdf'\n"

	ret += "set yrange [0:*]\n"
	ret += "set key right bottom\n"
	ret += "unset x2tics\n"
	ret += "unset y2tics\n"
	ret += "set border 3\n"

	ret += "set xlabel 'CAT bits'\n"
	ret += "set ylabel 'Runtime (s)'\n"

	ret += "Shadecolor = '#80E0A080'\n"

	for i, app := range apps {
		ret += "plot '" + indvCATRuntime(i) + "' using 1:($2+$3):($2-$3) with filledcurve fc rgb Shadecolor title 'Std. dev.', '' using 1:2 smooth mcspline lw 2 title 'Mean runtime (" + app + ")'\n"
	}

	return ret
}
