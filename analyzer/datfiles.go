package main

import (
	"io/ioutil"

	"sort"
	"strconv"

	"github.com/jbreitbart/coBench/commands"
	"github.com/jbreitbart/coBench/stats"
	log "github.com/sirupsen/logrus"
)

func sortedKeys(r *map[int]stats.RuntimeT) []int {
	var sortedKeys []int
	for k := range *r {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Slice(sortedKeys[:], func(i, j int) bool {
		return sortedKeys[i] < sortedKeys[j]
	})

	return sortedKeys
}

func createCoSchedCATDatFiles(pairs [][2]string) []string {
	log.Infoln("Creating dat files for co-scheduling CAT runs.")

	ret := make([]string, 0)

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

		sortedKeys0 := sortedKeys(r0)
		sortedKeys1 := sortedKeys(r1)

		for i := range sortedKeys0 {
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
			out += strconv.FormatFloat(1.5*float64(k0), 'E', -1, 64) + " "
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
			log.WithError(err).WithFields(log.Fields{
				"filename": filename,
			}).Fatalln("Error while write file")
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

		sortedKeys := sortedKeys(catRuntime)

		for _, k := range sortedKeys {
			v, exist := (*catRuntime)[k]
			if !exist {
				log.Fatalln("Could not find key. Should never happen.")
			}
			out += strconv.FormatFloat(1.5*float64(k), 'E', -1, 64) + " " + strconv.FormatFloat(v.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(v.Stddev, 'E', -1, 64) + "\n"
		}

		ref := stats.GetReferenceRuntime(app)
		// TODO fix hardcoded 20
		out += strconv.FormatFloat(1.5*20, 'E', -1, 64) + " " + strconv.FormatFloat(ref.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref.Stddev, 'E', -1, 64) + "\n"

		filename := indvCATDatFilename(app)
		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"filename": filename,
			}).Fatalln("Error while write file")
		}

		ret = append(ret, filename)
	}

	return ret
}

func indvCATDatFilename(app string) string {
	// TODO check if filename already in use
	return commands.Pretty(app) + "-cat.dat"
}

func coSchedCATDatFilename(app0 string, app1 string) string {
	// TODO check if filename already in use
	return commands.Pretty(app0) + "-" + commands.Pretty(app1) + "-cosched-cat.dat"
}
