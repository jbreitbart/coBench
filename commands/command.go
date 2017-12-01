package commands

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

// GenerateIndv generates individual command list
func GenerateIndv(a []string) []string {
	ret := []string{}
	seen := map[string]bool{}
	for _, val := range a {
		if _, old := seen[val]; !old {
			ret = append(ret, val)
			seen[val] = true
		}
	}
	return ret
}

// GeneratePairs generates a pair of commands to be used with co-scheduling
func GeneratePairs(commands []string) [][2]string {
	seen := map[string]bool{}

	var ret [][2]string
	for i, c0 := range commands {
		for j, c1 := range commands {
			if i >= j {
				continue
			}
			if _, old := seen[c0+c1]; !old {
				ret = append(ret, [2]string{c0, c1})
				seen[c0+c1] = true
				seen[c1+c0] = true
			}
		}
	}
	return ret
}

// Read reads commands from files
func Read(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, errors.New("Error opening file " + filename + ": " + err.Error())
	}
	defer file.Close()

	var commands []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		temp := scanner.Text()
		temp = strings.TrimSpace(temp)
		if len(temp) > 0 && temp[0] != '#' {
			commands = append(commands, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.New("Error scanning commands: " + err.Error())
	}

	return commands, nil
}

func Pretty(app string) string {
	slash := strings.LastIndex(app, "/")
	space := strings.Index(app[slash+1:], " ")
	if space == -1 {
		space = len(app)
	}

	return app[slash+1 : space]
}
