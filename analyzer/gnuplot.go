package main

import (
	"io/ioutil"
	"strings"

	"github.com/jbreitbart/coBench/commands"
	log "github.com/sirupsen/logrus"
)

func writeGNUPlotCATCoSchedFile(pairs [][2]string, filename []string) {
	if len(pairs) == 0 {
		return
	}
	if len(pairs) != len(filename) {
		log.Fatalln("Not all apps have CAT data? Not supported atm.")
	}

	log.Infoln("Creating plot file for co-scheduling CAT runs.")

	var ret string
	ret += "set terminal pdf\n"
	ret += "set output 'co-sched-cat.pdf'\n"

	ret += "set yrange [0:*]\n"
	ret += "set key right bottom\n"
	ret += "unset x2tics\n"
	ret += "unset y2tics\n"
	ret += "set border 3\n"

	ret += "set xlabel 'L3 Cache (MB)'\n"
	ret += "set ylabel 'Runtime (s)'\n"

	ret += "shadecolor = '#80E0A080'\n"

	for i, pair := range pairs {
		ret += "plot '" + filename[i]
		ret += "' using 1:($2+$3):($2-$3) with filledcurve fc rgb shadecolor title 'Std. dev. (" + gnuplotEscape(commands.Pretty(pair[0])) + ")', "
		ret += "'' using 1:2 smooth mcspline lw 2 title 'Mean runtime (" + gnuplotEscape(commands.Pretty(pair[0])) + ")', "
		ret += "'' using 1:($4+$5):($4-$5) with filledcurve fc rgb shadecolor title 'Std. dev. (" + gnuplotEscape(commands.Pretty(pair[1])) + ")', "
		ret += "'' using 1:4 smooth mcspline lw 2 title 'Mean runtime (" + gnuplotEscape(commands.Pretty(pair[1])) + ")'\n"
	}

	err := ioutil.WriteFile("co-sched-cat.plot", []byte(ret), 0644)
	if err != nil {
		log.WithError(err).Fatalln("Error while write file co-sched-cat.plot")
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
	ret += "set terminal pdf\n"
	ret += "set output 'indv-cat.pdf'\n"

	ret += "set yrange [0:*]\n"
	ret += "set key right bottom\n"
	ret += "unset x2tics\n"
	ret += "unset y2tics\n"
	ret += "set border 3\n"

	ret += "set xlabel 'L3 Cache (MB)'\n"
	ret += "set ylabel 'Runtime (s)'\n"

	ret += "shadecolor = '#80E0A080'\n"

	for i, app := range apps {
		ret += "plot '" + filename[i] + "' using 1:($2+$3):($2-$3) with filledcurve fc rgb shadecolor title 'Std. dev.', '' using 1:2 smooth mcspline lw 2 title 'Mean runtime (" + gnuplotEscape(commands.Pretty(app)) + ")'\n"
	}

	err := ioutil.WriteFile("indv-cat.plot", []byte(ret), 0644)
	if err != nil {
		log.WithError(err).Fatalln("Error while write file indv-cat.plot")
	}
}

// only supports some characters
func gnuplotEscape(s string) string {
	ret := strings.Replace(s, "_", "\\_", -1)

	return ret
}
