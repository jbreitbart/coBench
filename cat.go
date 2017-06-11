package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/jbreitbart/coBench/bit"
)

func generateCatConfigs(minBits uint64, numBits uint64) [][]uint64 {
	pairs := make([][]uint64, 0)

	if *cat {
		for bits := minBits; bits <= numBits-minBits; bits += *catBitChunk {
			pairs = append(pairs, []uint64{bit.SetFirstN(0, bits), bit.SetLastN(0, bits, numBits)})
		}
	}

	// Added as a last entry, so that after programm execution cat is disabled
	pairs = append(pairs, []uint64{bit.SetFirstN(0, numBits), bit.SetFirstN(0, numBits)})

	return pairs
}

func createDirsCAT() error {
	for _, dir := range catDirs {
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

func removeDirsCAT() error {
	for _, dir := range catDirs {
		err := os.Remove(dir)
		if err != nil {
			return fmt.Errorf("Cannot remove dir %v: %v", dir, err)
		}
	}

	return nil
}

func resetCAT() error {
	return removeDirsCAT()
}

func setupCAT() (minBits uint64, numBits uint64, err error) {
	// TODO read from resctrlPath
	minBits = 2
	numBits = 20

	if err = createDirsCAT(); err != nil {
		return
	}

	numbers := regexp.MustCompile("[0-9]+")

	for i, cpu := range cpus {
		cpuIDs := numbers.FindAllString(cpu, -1)
		if len(cpuIDs)%2 != 0 {
			err = fmt.Errorf("Unsupported CPU list: %v", cpu)
			return
		}

		var bitset uint64

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

		file, err = os.OpenFile(catDirs[i]+"/cpus", os.O_WRONLY|os.O_TRUNC, 0777)
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

func writeCATConfig(configs []uint64) error {

	if len(catDirs) != len(configs) {
		return fmt.Errorf("Internal error")
	}

	for i, dir := range catDirs {
		file, err := os.OpenFile(dir+"/schemata", os.O_WRONLY|os.O_TRUNC, 0777)
		if err != nil {
			return fmt.Errorf("CAT could not open cpus file: %v", err)
		}
		defer file.Close()

		// TODO hardcoded string
		_, err = file.WriteString(fmt.Sprintf("L3:0=%x;1=%x\n", configs[i], configs[i]))
		if err != nil {
			return fmt.Errorf("CAT could not write to schemata file: %v", err)
		}
	}
	return nil
}
