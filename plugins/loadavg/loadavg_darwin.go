// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

// +build darwin

package loadavg

import (
	"strconv"
)

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
