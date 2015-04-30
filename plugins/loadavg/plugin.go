// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package loadavg

import (
	"math"

	metrics "github.com/rcrowley/go-metrics"
)

// Plugin watches the load of a machine.
type Plugin struct {
	last1min  metrics.GaugeFloat64
	last5min  metrics.GaugeFloat64
	last15min metrics.GaugeFloat64
}

// NewPlugin initializes a new Plugin to watch the load of a machine.
func NewPlugin() (*Plugin, error) {
	p := &Plugin{}
	p.last1min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last1min", p.last1min)
	p.last5min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last5min", p.last5min)
	p.last15min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last15min", p.last15min)
	return p, nil
}

// Name is the name of the plugin.
func (p *Plugin) Name() string {
	return "loadavg"
}

// Snapshot returns a snapshot of the current load.
func (p *Plugin) Snapshot() (interface{}, error) {
	loadavg, err := GetLoadAvg()
	if err != nil {
		return nil, err
	}

	// Update metrics
	p.last1min.Update(loadavg.Last1Min)
	p.last5min.Update(loadavg.Last5Min)
	p.last15min.Update(loadavg.Last15Min)

	load1Min := p.last1min.Value()
	if math.IsNaN(load1Min) {
		load1Min = 0.0
	}

	load5Min := p.last5min.Value()
	if math.IsNaN(load5Min) {
		load5Min = 0.0
	}

	load15Min := p.last15min.Value()
	if math.IsNaN(load15Min) {
		load15Min = 0.0
	}

	// Return data
	return map[string]interface{}{
		"load1min":  load1Min,
		"load5min":  load5Min,
		"load15min": load15Min,
	}, nil
}
