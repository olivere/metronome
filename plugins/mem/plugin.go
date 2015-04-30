// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package mem

import (
	"math"

	metrics "github.com/rcrowley/go-metrics"
)

// Plugin watches the memory usage of a machine.
type Plugin struct {
	total       metrics.Gauge
	used        metrics.Gauge
	usedPercent metrics.GaugeFloat64
	free        metrics.Gauge
}

// NewPlugin creates a Plugin that watches the memory usage of a machine.
func NewPlugin() (*Plugin, error) {
	p := &Plugin{}
	p.total = metrics.NewGauge()
	metrics.Register("mem.total", p.total)
	p.free = metrics.NewGauge()
	metrics.Register("mem.free", p.free)
	p.used = metrics.NewGauge()
	metrics.Register("mem.used", p.used)
	p.usedPercent = metrics.NewGaugeFloat64()
	metrics.Register("mem.usedpercent", p.usedPercent)
	return p, nil
}

// Name of the plugin.
func (p *Plugin) Name() string {
	return "mem"
}

// Snapshot returns a snapshot of the current memory usage.
func (p *Plugin) Snapshot() (interface{}, error) {
	mem, err := GetMem()
	if err != nil {
		return nil, err
	}

	// Update metrics
	p.total.Update(mem.Total)
	p.free.Update(mem.Free)
	p.used.Update(mem.Used)
	p.usedPercent.Update(mem.UsedPercent)

	usedPercent := p.usedPercent.Value()
	if math.IsNaN(usedPercent) {
		usedPercent = 0.0
	}

	// Return data
	return map[string]interface{}{
		"total":        p.total.Value(),
		"used":         p.used.Value(),
		"used_percent": usedPercent,
		"free":         p.free.Value(),
	}, nil
}
