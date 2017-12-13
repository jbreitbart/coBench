package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jbreitbart/coBench/bit"
	log "github.com/sirupsen/logrus"
)

func generateCatConfigs(minBits uint64, numBits uint64) [][2]uint64 {
	pairs := make([][2]uint64, 0)

	if *cat {
		for bits := minBits; bits <= numBits-minBits; bits += *catBitChunk {
			if *inverseCat {
				pairs = append(pairs, [2]uint64{bit.SetLastN(0, bits, numBits), bit.SetFirstN(0, bits)})
			} else {
				pairs = append(pairs, [2]uint64{bit.SetFirstN(0, bits), bit.SetLastN(0, bits, numBits)})
			}
		}
	}

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

func readCATInfo() (minBits uint64, numBits uint64, err error) {
	minBitsFilename := *resctrlPath + "/info/L3/min_cbm_bits"
	minBitsByteTxt, err := ioutil.ReadFile(minBitsFilename)
	if err != nil {
		return
	}
	minBitsTxt := strings.TrimSpace(string(minBitsByteTxt))
	minBits, err = strconv.ParseUint(minBitsTxt, 10, 64)
	if err != nil {
		return
	}

	numBitsFilename := *resctrlPath + "/info/L3/cbm_mask"
	numBitsByteTxt, err := ioutil.ReadFile(numBitsFilename)
	if err != nil {
		return
	}
	numBitsTxt := strings.TrimSpace(string(numBitsByteTxt))
	numBitsTemp, err := strconv.ParseUint(numBitsTxt, 16, 64)
	if err != nil {
		return
	}

	numBits = 0
	for i := (uint64)(0); i < 64; i++ {
		if bit.Has(numBitsTemp, i) {
			numBits++
		}
	}

	log.WithFields(log.Fields{
		"Min CBM Bits":  minBits,
		"CBM Mask Bits": numBits,
	}).Infoln("CAT configuration")

	err = nil
	return
}

func setupCAT() (minBits uint64, numBits uint64, err error) {

	if minBits, numBits, err = readCATInfo(); err != nil {
		return
	}

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
		for p := 0; p < len(cpuIDs); p += 2 {
			var start, end uint64
			start, err = strconv.ParseUint(cpuIDs[p], 10, 64)
			if err != nil {
				err = fmt.Errorf("Parse number: %v", start)
				return
			}
			end, err = strconv.ParseUint(cpuIDs[p+1], 10, 64)
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

		_, err = file.WriteString(fmt.Sprintf("%x", bitset))
		if err != nil {
			err = fmt.Errorf("CAT could write to cpus file: %v", err)
			return
		}
	}

	return
}

func writeCATConfig(configs [2]uint64) error {

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
