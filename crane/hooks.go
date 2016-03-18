package crane

type HooksCommander interface {
	PreBuild() string
	PostBuild() string
	PreStart() string
	PostStart() string
	PreStop() string
	PostStop() string
}

type Hooks struct {
	RawPreBuild  string `json:"pre-build" yaml:"pre-build"`
	RawPostBuild string `json:"post-build" yaml:"post-build"`
	RawPreStart  string `json:"pre-start" yaml:"pre-start"`
	RawPostStart string `json:"post-start" yaml:"post-start"`
	RawPreStop   string `json:"pre-stop" yaml:"pre-stop"`
	RawPostStop  string `json:"post-stop" yaml:"post-stop"`
	// until we have a very long list, it's probably easier
	// to do 4 changes in that file for each new event than
	// using `go generate`
}

func (h *Hooks) PreBuild() string {
	return expandEnv(h.RawPreBuild)
}

func (h *Hooks) PostBuild() string {
	return expandEnv(h.RawPostBuild)
}

func (h *Hooks) PreStart() string {
	return expandEnv(h.RawPreStart)
}

func (h *Hooks) PostStart() string {
	return expandEnv(h.RawPostStart)
}

func (h *Hooks) PreStop() string {
	return expandEnv(h.RawPreStop)
}

func (h *Hooks) PostStop() string {
	return expandEnv(h.RawPostStop)
}

// Merge another set of hooks into the existing object. Existing
// hooks will be overriden if the corresponding hooks from the
// source struct are defined. Returns true if some content was
// overiden in the process.
func (h *Hooks) CopyFrom(source Hooks) (overriden bool) {
	overrideIfFromNotEmpty := func(from string, to *string) {
		if from != "" {
			overriden = overriden || *to != ""
			*to = from
		}
	}
	overrideIfFromNotEmpty(source.RawPreBuild, &h.RawPreBuild)
	overrideIfFromNotEmpty(source.RawPostBuild, &h.RawPostBuild)
	overrideIfFromNotEmpty(source.RawPreStart, &h.RawPreStart)
	overrideIfFromNotEmpty(source.RawPostStart, &h.RawPostStart)
	overrideIfFromNotEmpty(source.RawPreStop, &h.RawPreStop)
	overrideIfFromNotEmpty(source.RawPostStop, &h.RawPostStop)
	return
}
