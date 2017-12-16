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

func createCoSchedCATDatFiles(pairs [][2]string, matchpairs bool) ([]string, []string) {
	log.WithField("paired", matchpairs).Infoln("Creating dat files for co-scheduling CAT runs")

	ret := make([]string, 0)
	perfName := make([]string, 0)

	for _, pair := range pairs {
		r0 := stats.GetCoSchedCATRuntimes(pair[0], pair[1])
		r1 := stats.GetCoSchedCATRuntimes(pair[1], pair[0])

		if r0 == nil && r1 == nil {
			log.WithFields(log.Fields{
				"app0": pair[0],
				"app1": pair[1],
			}).Infoln("No CAT co-scheduling data")
			continue
		}

		if r0 == nil || r1 == nil {
			log.WithFields(log.Fields{
				"app0": pair[0],
				"app1": pair[1],
			}).Errorln("CAT co-scheduling data inconsistent")
			continue
		}

		out := "# 0: " + pair[0] + "\n"
		out += "# co-scheduled with \n"
		out += "# 1: " + pair[1] + "\n"
		if matchpairs {
			out += "L3 limit correct only for 0, 1 uses rest\n"
		}

		sortedKeys0 := sortedKeys(r0)
		sortedKeys1 := sortedKeys(r1)
		arbRun := (*r0)[sortedKeys0[0]]

		out += "# L3(0) Runtime(0) Std.Dev.(0) "
		for _, temp := range extractPerfData(&arbRun) {
			out += temp.Name + "(0) "
			out += "Std.Dev " + temp.Name + "(0) "
			if pair == pairs[0] {
				perfName = append(perfName, temp.Name)
			}
		}

		out += "Runtime(1) Std.Dev(1) "
		for _, temp := range extractPerfData(&arbRun) {
			out += temp.Name + "(1) "
			out += "Std.Dev " + temp.Name + "(1) "
		}
		out += "\n"

		for i := range sortedKeys0 {
			k0 := sortedKeys0[i]
			k1 := sortedKeys1[i]
			if matchpairs {
				k1 = sortedKeys1[len(sortedKeys1)-i-1]
			}
			v0, exist := (*r0)[k0]
			if !exist {
				log.Fatalf("Could not find key %v. Should never happen.", k0)
			}
			v1, exist := (*r1)[k1]
			if !exist {
				log.Fatalf("Could not find key %v. Should never happen.\n", k1)
			}
			out += coSchedRuntimeToString(k0, &v0, &v1, extractPerfData(&v0), extractPerfData(&v1))
		}

		ref0 := stats.GetCoSchedRuntimes(pair[0], pair[1])
		ref1 := stats.GetCoSchedRuntimes(pair[1], pair[0])
		// TODO fix hardcoded 20
		out += coSchedRuntimeToString(20, ref0, ref1, extractPerfData(ref0), extractPerfData(ref1))

		filename := coSchedCATDatFilename(pair[0], pair[1], matchpairs)
		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"filename": filename,
			}).Fatalln("Error while write file")
		}

		ret = append(ret, filename)
	}

	return ret, perfName
}

// Returns the filenames
func createIndvCATDatFiles(apps []string) ([]string, []string) {
	log.Println("Creating dat files for individual CAT runs.")

	ret := make([]string, 0)
	perfName := make([]string, 0)

	for _, app := range apps {
		catRuntime := stats.GetCATRuntimes(app)
		if catRuntime == nil {
			continue
		}

		out := "# " + app + "\n"

		sortedKeys := sortedKeys(catRuntime)
		arbRun := (*catRuntime)[sortedKeys[0]]

		out += "# L3 Runtime Std.Dev. "
		for _, temp := range extractPerfData(&arbRun) {
			out += temp.Name
			out += "Std.Dev " + temp.Name
			if app == apps[0] {
				perfName = append(perfName, temp.Name)
			}
		}
		out += "\n"

		for _, k := range sortedKeys {
			log.WithField("app", app).WithField("cat", k).Debugln("Currently analysing")
			v, exist := (*catRuntime)[k]
			if !exist {
				log.Fatalln("Could not find key. Should never happen.")
			}
			out += coSchedRuntimeToString(k, &v, nil, extractPerfData(&v), nil)
		}

		log.WithField("app", app).WithField("cat", "no cat").Debugln("Currently analysing")
		ref := stats.GetReferenceRuntime(app)
		// TODO fix hardcoded 20
		out += coSchedRuntimeToString(20, ref, nil, extractPerfData(ref), nil)

		filename := indvCATDatFilename(app)
		err := ioutil.WriteFile(filename, []byte(out), 0644)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"filename": filename,
			}).Fatalln("Error while write file")
		}

		ret = append(ret, filename)
	}

	return ret, perfName
}

func coSchedRuntimeToString(CATChunks int, ref0, ref1 *stats.RuntimeT, perf0, perf1 []perfDataT) string {
	out := strconv.FormatFloat(1.5*float64(CATChunks), 'E', -1, 64) + " "
	out += strconv.FormatFloat(ref0.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref0.Stddev, 'E', -1, 64)
	for _, p := range perf0 {
		out += " " + strconv.FormatFloat(p.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(p.Stddev, 'E', -1, 64)
	}

	if ref1 != nil {
		out += " "
		out += strconv.FormatFloat(ref1.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(ref1.Stddev, 'E', -1, 64)
		for _, p := range perf1 {
			out += " " + strconv.FormatFloat(p.Mean, 'E', -1, 64) + " " + strconv.FormatFloat(p.Stddev, 'E', -1, 64)
		}
	}
	out += "\n"
	return out
}

func indvCATDatFilename(app string) string {
	// TODO check if filename already in use
	return commands.Pretty(app) + "-cat.dat"
}

func coSchedCATDatFilename(app0 string, app1 string, matchpairs bool) string {
	// TODO check if filename already in use

	ret := commands.Pretty(app0) + "-" + commands.Pretty(app1) + "-cosched-cat"
	if matchpairs {
		ret += "-paired"
	}
	ret += ".dat"
	return ret
}
