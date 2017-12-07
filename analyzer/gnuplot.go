package main

import (
	"io/ioutil"
	"strings"

	"github.com/jbreitbart/coBench/commands"
	log "github.com/sirupsen/logrus"
)

func writeGNUPlotCATCoSchedFile(pairs [][2]string, filenames []string, paired bool) {
	if len(pairs) == 0 {
		return
	}
	if len(pairs) != len(filenames) {
		log.Fatalln("Not all apps have CAT data? Not supported atm")
	}

	log.WithField("paired", paired).Infoln("Creating plot file for co-scheduling CAT runs")

	var ret string
	if !paired {
		ret += "set output 'co-sched-cat.pdf'\n"
	} else {
		ret += "set output 'co-sched-cat-paired.pdf'\n"
	}

	ret += gnuplotHeader()

	if !paired {
		ret += "set xlabel 'L3 Cache (MB)'\n"
	} else {
		ret += "set xlabel 'L3 Cache (MB) for app0'\n"
	}

	for i, pair := range pairs {
		ret += "set title '" + gnuplotEscape(commands.Pretty(pair[0])) + " + " + gnuplotEscape(commands.Pretty(pair[1])) + "'\n"
		ret += "plot '" + filenames[i]
		ret += "' using 1:2:3 w yerrorbars ls 1 title '', "
		ret += "'' using 1:2 with linespoints ls 1 title 'Ø runtime (" + gnuplotEscape(commands.Pretty(pair[0])) + ")', "
		ret += "'' using 1:4:5 w yerrorbars ls 2 title '', "
		ret += "'' using 1:4 with linespoints ls 2 title 'Ø runtime (" + gnuplotEscape(commands.Pretty(pair[1])) + ")'\n"
	}

	filename := "co-sched-cat.plot"
	if paired {
		filename = "co-sched-cat-paired.plot"
	}
	err := ioutil.WriteFile(filename, []byte(ret), 0644)
	if err != nil {
		log.WithError(err).WithField("filename", filename).Fatalln("Error while write file")
	}
}

func writeGNUPlotCATIndvFile(apps []string, filename []string) {
	if len(apps) == 0 {
		return
	}
	if len(apps) != len(filename) {
		log.Fatalln("Not all apps have CAT data? Not supported atm.")
	}

	log.Infoln("Creating plot file for individual CAT runs.")

	var ret string
	ret += "set output 'indv-cat.pdf'\n"
	ret += gnuplotHeader()
	ret += "set xlabel 'L3 Cache (MB)'\n"

	for i, app := range apps {
		ret += "set title '" + gnuplotEscape(commands.Pretty(app)) + "'\n"
		ret += "plot '" + filename[i] + "' "
		ret += "using 1:2:3 w yerrorbars ls 1 title '', "
		ret += "'' using 1:2 with linespoints ls 1 title 'Ø runtime (" + gnuplotEscape(commands.Pretty(app)) + ")'\n"

		// using 1:($2+$3):($2-$3) with filledcurve fc rgb shadecolor title 'Std. dev.', '' using 1:2 smooth mcspline lw 2 title 'Mean runtime (" + gnuplotEscape(commands.Pretty(app)) + ")'\n"
	}

	err := ioutil.WriteFile("indv-cat.plot", []byte(ret), 0644)
	if err != nil {
		log.WithError(err).Fatalln("Error while write file indv-cat.plot")
	}
}

func gnuplotHeader() string {
	ret := "set terminal pdf\n"
	ret += "set yrange [0:*]\n"
	ret += "set key right bottom\n"
	ret += "unset x2tics\n"
	ret += "unset y2tics\n"
	ret += "set border 3\n"
	ret += "set ylabel 'Runtime (s)'\n"

	ret += "set style line 1 lc rgb '#0060ad' lt 1 lw 2 pt 7 ps 0.1   # --- blue\n"
	ret += "set style line 2 lc rgb '#dd181f' lt 1 lw 2 pt 5 ps 0.1   # --- red\n"
	ret += "set pointintervalbox 3\n"
	//ret += "shadecolor = '#80E0A080'\n"

	return ret
}

// only supports some characters
func gnuplotEscape(s string) string {
	ret := strings.Replace(s, "_", "\\_", -1)

	return ret
}
