// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

// +build linux

package swap

import "syscall"

func GetSwap() (*Swap, error) {
	sysinfo := &syscall.Sysinfo_t{}

	if err := syscall.Sysinfo(sysinfo); err != nil {
		return nil, err
	}
	mem := &Swap{
		Total: int64(sysinfo.Totalswap),
		Free:  int64(sysinfo.Freeswap),
	}
	mem.Used = mem.Total - mem.Free
	if mem.Total != 0 {
		mem.UsedPercent = float64(mem.Total-mem.Free) / float64(mem.Total) * 100.0
	} else {
		mem.UsedPercent = 0
	}
	return mem, nil
}
