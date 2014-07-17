package print

import (
	"github.com/fatih/color"
)

var Noticef func(format string, a ...interface{})
var Errorf func(format string, a ...interface{})

func init() {
	Noticef = color.New(color.FgYellow).PrintfFunc()
	Errorf = color.New(color.FgRed).PrintfFunc()
}
