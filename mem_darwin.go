// +build darwin

package metronome

import (
	"os/exec"
	"strconv"
	"strings"
)

func getPageSize() (int64, error) {
	out, err := exec.Command("pagesize").Output()
	if err != nil {
		return 0, err
	}
	o := strings.TrimSpace(string(out))
	p, err := strconv.ParseInt(o, 10, 64)
	if err != nil {
		return 0, err
	}
	return p, nil
}

// GetMem returns Mem.
func GetMem() (*Mem, error) {
	p, err := getPageSize()
	if err != nil {
		return nil, err
	}

	mem := &Mem{}

	total, err := Sysctl("hw.memsize")
	if err != nil {
		return nil, err
	}
	mem.Total, _ = strconv.ParseInt(total[0], 10, 64)
	mem.Total = mem.Total

	free, err := Sysctl("vm.page_free_count")
	if err != nil {
		return nil, err
	}
	mem.Free, _ = strconv.ParseInt(free[0], 10, 64)
	mem.Free = mem.Free * p

	mem.Used = mem.Total - mem.Free

	mem.UsedPercent = float64(mem.Total-mem.Free) / float64(mem.Total) * 100.0

	return mem, nil
}
