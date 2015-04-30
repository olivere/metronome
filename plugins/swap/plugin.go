// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package swap

import (
	"math"

	metrics "github.com/rcrowley/go-metrics"
)

// Plugin watches the swap usage of a machine.
type Plugin struct {
	total       metrics.Gauge
	used        metrics.Gauge
	usedPercent metrics.GaugeFloat64
	free        metrics.Gauge
}

// NewPlugin initializes a watcher that watches the swap usage of a machine.
func NewPlugin() (*Plugin, error) {
	p := &Plugin{}
	p.total = metrics.NewGauge()
	metrics.Register("swap.total", p.total)
	p.free = metrics.NewGauge()
	metrics.Register("swap.free", p.free)
	p.used = metrics.NewGauge()
	metrics.Register("swap.used", p.used)
	p.usedPercent = metrics.NewGaugeFloat64()
	metrics.Register("swap.used_percent", p.usedPercent)
	return p, nil
}

// Name of the plugin.
func (p *Plugin) Name() string {
	return "swap"
}

// Snapshot returns a snapshot of the current swap usage.
func (p *Plugin) Snapshot() (interface{}, error) {
	swap, err := GetSwap()
	if err != nil {
		return nil, err
	}

	// Update metrics
	p.total.Update(swap.Total)
	p.free.Update(swap.Free)
	p.used.Update(swap.Used)
	p.usedPercent.Update(swap.UsedPercent)

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
