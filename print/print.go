package print

import (
	"github.com/fatih/color"
)

var Notice func(format string, a ...interface{})
var Error func(format string, a ...interface{})

func init() {
	Notice = color.New(color.FgYellow).PrintfFunc()
	Error = color.New(color.FgRed).PrintfFunc()
}
