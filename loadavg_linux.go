// +build linux

package metronome

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func GetLoadAvg() (*LoadAvg, error) {
	b, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return nil, err
	}
	content := string(b)
	values := strings.Fields(content)
	loadavg := &LoadAvg{}
	loadavg.Last1Min, _ = strconv.ParseFloat(values[0], 64)
	loadavg.Last5Min, _ = strconv.ParseFloat(values[1], 64)
	loadavg.Last15Min, _ = strconv.ParseFloat(values[2], 64)
	return loadavg, nil
}
