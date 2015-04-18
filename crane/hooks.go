package crane

import (
	"os"
)

type Hooks interface {
	PreStart() string
	PostStart() string
	PreStop() string
	PostStop() string
}

type hooks struct {
	RawPreStart  string `json:"pre-start" yaml:"pre-start"`
	RawPostStart string `json:"post-start" yaml:"post-start"`
	RawPreStop   string `json:"pre-stop" yaml:"pre-stop"`
	RawPostStop  string `json:"post-stop" yaml:"post-stop"`
	// until we have a very long list, it's probably easier
	// to do 4 changes in that file for each new event than
	// using `go generate`
}

func (h *hooks) PreStart() string {
	return os.ExpandEnv(h.RawPreStart)
}

func (h *hooks) PostStart() string {
	return os.ExpandEnv(h.RawPostStart)
}

func (h *hooks) PreStop() string {
	return os.ExpandEnv(h.RawPreStop)
}

func (h *hooks) PostStop() string {
	return os.ExpandEnv(h.RawPostStop)
}

// Merge another set of hooks into the existing object. Existing
// hooks will be overriden if the corresponding hooks from the
// source struct are defined. Returns true if some content was
// overiden in the process.
func (h *hooks) CopyFrom(source hooks) (overriden bool) {
	overrideIfFromNotEmpty := func(from string, to *string) {
		if from != "" {
			overriden = overriden || *to != ""
			*to = from
		}
	}
	overrideIfFromNotEmpty(source.RawPreStart, &h.RawPreStart)
	overrideIfFromNotEmpty(source.RawPostStart, &h.RawPostStart)
	overrideIfFromNotEmpty(source.RawPreStop, &h.RawPreStop)
	overrideIfFromNotEmpty(source.RawPostStop, &h.RawPostStop)
	return
}
