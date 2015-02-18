// +build linux

package metronome

import (
	"io/ioutil"
	"strconv"
	"strings"
)

func GetMem() (*Mem, error) {
	data, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")

	mem := &Mem{}
	for _, line := range lines {
		values := strings.Split(line, ":")
		if len(values) != 2 {
			continue
		}
		key := strings.TrimSpace(values[0])
		value := strings.TrimSpace(values[1])
		value = strings.Replace(value, " kB", "", -1)

		t, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return nil, err
		}
		switch key {
		case "MemTotal":
			mem.Total = t * 1000
		case "MemFree":
			mem.Free = t * 1000
		}
	}
	mem.Used = mem.Total - mem.Free
	mem.UsedPercent = float64(mem.Total-mem.Free) / float64(mem.Total) * 100.0

	return mem, nil
}
