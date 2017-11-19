package main

import (
	"io/ioutil"
	"log"
	"math/bits"
	"sort"
	"strconv"

	"github.com/jbreitbart/coBench/stats"
)

// TODO copied from coBench. fix somehow...
func generateCommandPairs(commands []string) [][2]string {
	var pairs [][2]string
	for i, c0 := range commands {
		for j, c1 := range commands {
			if i >= j {
				continue
			}
			pairs = append(pairs, [2]string{c0, c1})
		}
	}
	return pairs
}

func createCoSchedCATDatFiles(apps []string) []string {
	log.Println("Creating dat files for co-scheduling CAT runs.")

	ret := make([]string, 0)

	pairs := generateCommandPairs(apps)

	for _, pair := range pairs {
		r0 := stats.GetCoSchedCATRuntimes(pair[0], pair[1])
		r1 := stats.GetCoSchedCATRuntimes(pair[1], pair[0])

		if r0 == nil && r1 == nil {
			continue
		}

		if r0 == nil || r1 == nil {
			log.Fatalln("CAT co-scheduling data inconsistent.")
		}

		out := "# 0: " + pair[0] + "\n"
		out += "# co-scheduled with \n"
		out += "# 1: " + pair[1] + "\n"
		out += "# L3(0) Runtime(0) Std.Dev.(0) Runtime(1) Std.Dev(1)\n"

		// TODO extract as a function
		var sortedKeys0 []uint64
		for k, _ := range *r0 {
			sortedKeys0 = append(sortedKeys0, k)
		}
		sort.Slice(sortedKeys0[:], func(i, j int) bool {
			return bits.OnesCount64(sortedKeys0[i]) < bits.OnesCount64(sortedKeys0[j])
		})
		var sortedKeys1 []uint64
		for k, _ := range *r1 {
			sortedKeys1 = append(sortedKeys1, k)
		}
		sort.Slice(sortedKeys1[:], func(i, j int) bool {
			return bits.OnesCount64(sortedKeys1[i]) > bits.OnesCount64(sortedKeys1[j])
		})

		if len(sortedKeys0) != len(sortedKeys1) {
			log.Fatalf("CAT co-scheduling data inconsistent: len: %v - %v\n", len(sortedKeys0), len(sortedKeys1))
		}

		for i, _ := range sortedKeys0 {
			k0 := sortedKeys0[i]
			k1 := sortedKeys1[i]
			v0, exist := (*r0)[k0]
			if !exist {
				log.Fatalf("Could not find key %v. Should never happen.", k0)
			}
			v1, exist := (*r1)[k1]
			if !exist {
				log.Fatalf("Could not find key %v. Should never happen.\n", k1)
			}
			out += strconv.FormatFloat(1.5*float64(bits.OnesCount64(k0)), 'E', -1, 64) + " "
			out += strconv.FormatFloat(v0.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(v0.Stddev, 'E', -1, 64) + " "
			out += strconv.FormatFloat(v1.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(v1.Stddev, 'E', -1, 64) + "\n"
		}

		ref0 := stats.GetCoSchedRuntimes(pair[0], pair[1])
		ref1 := stats.GetCoSchedRuntimes(pair[1], pair[0])
		// TODO fix hardcoded 20
		out += strconv.FormatFloat(1.5*20, 'E', -1, 64) + " "
		out += strconv.FormatFloat(ref0.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref0.Stddev, 'E', -1, 64) + " "
		out += strconv.FormatFloat(ref1.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref1.Stddev, 'E', -1, 64) + "\n"

		filename := coSchedCATDatFilename(pair[0], pair[1])
		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.Fatalf("Error while write file %v: %v", filename, err)
		}

		ret = append(ret, filename)
	}

	return ret
}

// Returns the filenames
func createIndvCATDatFiles(apps []string) []string {
	log.Println("Creating dat files for individual CAT runs.")

	ret := make([]string, 0)

	for _, app := range apps {
		catRuntime := stats.GetIndvCATRuntimes(app)
		if catRuntime == nil {
			continue
		}

		out := "# " + app + "\n"
		out += "# L3 Runtime Std.Dev.\n"

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

		filename := indvCATDatFilename(app)
		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.Fatalf("Error while write file %v: %v", filename, err)
		}

		ret = append(ret, filename)
	}

	return ret
}

func indvCATDatFilename(app string) string {
	// TODO check if filename already in use
	return prettyAppCmd(app) + "-cat.dat"
}

func coSchedCATDatFilename(app0 string, app1 string) string {
	// TODO check if filename already in use
	return prettyAppCmd(app0) + "-" + prettyAppCmd(app1) + "-cosched-cat.dat"
}
