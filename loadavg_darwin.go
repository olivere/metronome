// +build darwin

package metronome

import "strconv"

func GetLoadAvg() (*LoadAvg, error) {
	values, err := Sysctl("vm.loadavg")
	if err != nil {
		return nil, err
	}
	loadavg := &LoadAvg{}
	loadavg.Last1Min, _ = strconv.ParseFloat(values[0], 64)
	loadavg.Last5Min, _ = strconv.ParseFloat(values[1], 64)
	loadavg.Last15Min, _ = strconv.ParseFloat(values[2], 64)
	return loadavg, nil
}
