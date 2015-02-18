package metronome

type Status struct {
	LoadAvg struct {
		Load1Min  float64 `json:"load1min"`
		Load5Min  float64 `json:"load5min"`
		Load15Min float64 `json:"load15min"`
	} `json:"loadavg"`

	Mem struct {
		Total int64 `json:"total"`
		Free  int64 `json:"free"`
		Used  int64 `json:"used"`
	} `json:"mem"`

	Swap struct {
		Total int64 `json:"total"`
		Free  int64 `json:"free"`
		Used  int64 `json:"used"`
	} `json:"swap"`
}
