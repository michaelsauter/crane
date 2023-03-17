package crane

type LoggingParameters struct {
	RawDriver  string `json:"driver" yaml:"driver"`
	RawOptions any    `json:"options" yaml:"options"`
}

func (l LoggingParameters) Options() []string {
	return sliceOrMap2ExpandedSlice(l.RawOptions)
}

func (l LoggingParameters) Driver() string {
	return expandEnv(l.RawDriver)
}
