package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/jbreitbart/coBench/bit"
)

func createDirsCAT(dirs []string) error {
	for _, dir := range dirs {
		err := os.Mkdir(dir, 0777)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("CAT: %v", err)
		}
	}

	return nil
}

func setupCAT() (minBits int, numBits int, err error) {
	// TODO read from resctrlPath
	minBits = 2
	numBits = 20

	// TODO hardcoded length
	dirs := []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	numbers := regexp.MustCompile("[0-9]+")

	if err = createDirsCAT(dirs); err != nil {
		return
	}

	for i, cpu := range cpus {
		cpuIDs := numbers.FindAllString(cpu, -1)
		if len(cpuIDs)%2 != 0 {
			err = fmt.Errorf("Unsupported CPU list: %v", cpu)
			return
		}

		var bitset int64

		// loop over every pair
		for i := 0; i < len(cpuIDs); i += 2 {
			var start, end uint64
			start, err = strconv.ParseUint(cpuIDs[i], 10, 64)
			if err != nil {
				err = fmt.Errorf("Parse number: %v", start)
				return
			}
			end, err = strconv.ParseUint(cpuIDs[i+1], 10, 64)
			if err != nil {
				err = fmt.Errorf("Parse number: %v", end)
				return
			}
			if end < start {
				start, end = end, start
			}
			for c := start; c <= end; c++ {
				// TODO validate by max cpu size
				bitset = bit.Set(bitset, c)
			}
		}

		var file *os.File

		file, err = os.OpenFile(dirs[i]+"/cpus", os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			err = fmt.Errorf("CAT could not open cpus file: %v", err)
			return
		}
		defer file.Close()

		_, err = file.WriteString(fmt.Sprintf("%v", bitset))
		if err != nil {
			err = fmt.Errorf("CAT could write to cpus file: %v", err)
			return
		}
	}

	return
}

func writeCATConfig(configs []int64) error {
	// TODO duplicated line
	dirs := []string{*resctrlPath + "/cobench0", *resctrlPath + "/cobench1"}

	if len(dirs) != len(configs) {
		return fmt.Errorf("Internal error")
	}

	for i, dir := range dirs {
		file, err := os.OpenFile(dir+"/schemata", os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			return fmt.Errorf("CAT could not open cpus file: %v", err)
		}
		defer file.Close()

		// TODO hardcoded string
		_, err = file.WriteString(fmt.Sprintf("L3:0=%x;1=%x\n", (uint)(configs[i]), (uint)(configs[i])))
		if err != nil {
			return fmt.Errorf("CAT could not write to schemata file: %v", err)
		}
	}
	return nil
}