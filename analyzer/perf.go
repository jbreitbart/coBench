package main

import (
	"sort"
	"strconv"
	"strings"

	"github.com/jbreitbart/coBench/stats"
	mstats "github.com/montanaflynn/stats"
	log "github.com/sirupsen/logrus"
)

type perfDataT struct {
	Name   string
	Mean   float64
	Stddev float64
	Vari   float64
}

func extractPerfData(dat *stats.RuntimeT) []perfDataT {
	coll := make(map[string][]float64)
	for _, temp := range *dat.RawRuntimesByMask {
		for _, runs := range temp {
			stdout := runs.Output
			/*
				Performance counter stats for '/global/work/share/npb/bt.C.x':

				  31162368.00 Bytes intel_cqm/llc_occupancy/
				   7313911210       LLC-load-misses
			*/
			posHead := strings.Index(stdout, "Performance counter stats for")
			if posHead == -1 {
				continue
			}
			posHEadLinebreak := posHead + strings.Index(stdout[posHead:], "\n")
			if posHEadLinebreak == -1 {
				continue
			}
			lineStart := posHEadLinebreak + 2 // 2 linebreaks
			for {
				/* 				  31162368.00 Bytes intel_cqm/llc_occupancy/ */
				endOfLine := strings.Index(stdout[lineStart:], "\n")
				if endOfLine == -1 {
					break
				}

				line := strings.TrimSpace(stdout[lineStart : lineStart+endOfLine])
				if line == "" {
					break
				}
				endNumber := strings.Index(line, " ")
				if endNumber == -1 {
					break
				}

				if value, err := strconv.ParseFloat(line[:endNumber], 64); err == nil {
					beginName := strings.LastIndex(line, " ")
					name := strings.TrimSpace(line[beginName:])
					log.WithField("value", value).WithField("name", name).Debugln("found value")
					coll[name] = append(coll[name], value)
				} else {
					log.WithError(err).WithField("line", line).Errorln("Could not extract value from perf")
				}
				lineStart = lineStart + endOfLine + 1
			}
		}
	}
	var sortedCollKeys []string
	for k := range coll {
		sortedCollKeys = append(sortedCollKeys, k)
	}
	sort.Slice(sortedCollKeys[:], func(i, j int) bool {
		return sortedCollKeys[i] < sortedCollKeys[j]
	})

	// TODO verify that sortedCollKeys is identical for every call

	perfData := make([]perfDataT, 0)
	for _, v := range sortedCollKeys {
		var p perfDataT
		var err error
		p.Name = v
		p.Mean, err = mstats.Mean(coll[v])
		if err != nil {
			log.WithError(err).Errorln("Error while computing mean")
		}
		p.Stddev, err = mstats.StandardDeviation(coll[v])
		if err != nil {
			log.WithError(err).Errorln("Error while computing stddev")
		}
		p.Vari, err = mstats.Variance(coll[v])
		if err != nil {
			log.WithError(err).Errorln("Error while computing variance")
		}
		perfData = append(perfData, p)
		log.WithField("compute perf", p).Debugln()
	}
	return perfData
}
