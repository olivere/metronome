// +build darwin

package metronome

import (
	"strconv"
	"strings"
)

// GetSwap returns Swap.
func GetSwap() (*Swap, error) {
	swap := &Swap{}

	vm, err := Sysctl("vm.swapusage")
	if err != nil {
		return nil, err
	}

	totals := strings.Replace(vm[2], "M", "", 1)
	useds := strings.Replace(vm[5], "M", "", 1)
	frees := strings.Replace(vm[8], "M", "", 1)

	total, err := strconv.ParseFloat(totals, 64)
	if err != nil {
		return nil, err
	}
	used, err := strconv.ParseFloat(useds, 64)
	if err != nil {
		return nil, err
	}
	free, err := strconv.ParseFloat(frees, 64)
	if err != nil {
		return nil, err
	}

	swap.Total = int64(total * 1024 * 1024)
	swap.Used = int64(used * 1024 * 1024)
	swap.Free = int64(free * 1024 * 1024)

	swap.UsedPercent = (used / total) * 100.0

	return swap, nil
}
