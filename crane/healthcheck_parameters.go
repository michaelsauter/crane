package crane

type HealthcheckParameters struct {
	RawTest     string `json:"test" yaml:"test"`
	RawInterval string `json:"interval" yaml:"interval"`
	RawTimeout  string `json:"timeout" yaml:"timeout"`
	Retries     int    `json:"retries" yaml:"retries"`
	Disable     bool   `json:"disable" yaml:"disable"`
}

func (h HealthcheckParameters) Test() string {
	return expandEnv(h.RawTest)
}

func (h HealthcheckParameters) Interval() string {
	return expandEnv(h.RawInterval)
}

func (h HealthcheckParameters) Timeout() string {
	return expandEnv(h.RawTimeout)
}
