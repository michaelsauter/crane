package print

import (
	"github.com/fatih/color"
)

var Infof func(format string, a ...interface{})
var Noticef func(format string, a ...interface{})
var Errorf func(format string, a ...interface{})

func init() {
	Infof = color.New(color.FgBlue).PrintfFunc()
	Noticef = color.New(color.FgYellow).PrintfFunc()
	Errorf = color.New(color.FgRed).PrintfFunc()
}
