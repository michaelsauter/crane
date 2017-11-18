package crane

type BuildParameters struct {
	RawContext    string      `json:"context" yaml:"context"`
	RawFile       string      `json:"file" yaml:"file"`
	RawDockerfile string      `json:"dockerfile" yaml:"dockerfile"`
	RawBuildArgs  interface{} `json:"build-arg" yaml:"build-arg"`
	RawArgs       interface{} `json:"args" yaml:"args"`
}

func (b BuildParameters) Context() string {
	return expandEnv(b.RawContext)
}

func (b BuildParameters) File() string {
	if len(b.RawFile) > 0 {
		return expandEnv(b.RawFile)
	}
	return expandEnv(b.RawDockerfile)
}

func (b BuildParameters) BuildArgs() []string {
	buildArgs := sliceOrMap2ExpandedSlice(b.RawBuildArgs)
	if len(buildArgs) == 0 {
		return sliceOrMap2ExpandedSlice(b.RawArgs)
	}
	return buildArgs
}
