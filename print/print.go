package print

import (
	"github.com/fatih/color"
	"os"
)

var Infof func(format string, a ...interface{})
var Noticef func(format string, a ...interface{})
var Errorf func(format string, a ...interface{})

func init() {
	color.Output = os.Stderr
	Infof = color.New(color.FgBlue).PrintfFunc()
	Noticef = color.New(color.FgYellow).PrintfFunc()
	Errorf = color.New(color.FgRed).PrintfFunc()
}
