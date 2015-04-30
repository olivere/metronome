// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elasticsearch

import (
	"errors"
	"fmt"

	"github.com/olivere/elastic"
	metrics "github.com/rcrowley/go-metrics"
)

// Config is the configuration for the Elasticsearch plugin.
type Config struct {
	// Urls of the cluster to watch with the plugin.
	Urls []string
}

// Plugin that watches an Elasticsearch cluster.
type Plugin struct {
	name   string          // cluster name
	urls   []string        // URLs of the cluster
	client *elastic.Client // Elastic client

	NumNodes     metrics.Gauge // number of nodes in the cluster
	NumDataNodes metrics.Gauge // number of data nodes in the cluster
	Shards       struct {
		Active       metrics.Gauge // active shards
		Relocating   metrics.Gauge // relocating shards
		Initializing metrics.Gauge // initializing shards
		Unassigned   metrics.Gauge // unassigned shards
	}
	NumPendingTasks metrics.Gauge // number of pending shards

	NumIndices metrics.Gauge // number of indices

	HeapUsed    metrics.Gauge        // heap used (on all nodes of the cluster)
	HeapMax     metrics.Gauge        // max heap size (of all nodes in the cluster)
	HeapPercent metrics.GaugeFloat64 // percentage of heap used (all nodes)

	CPUPercent metrics.GaugeFloat64 // CPU usage across all nodes of the cluster

	OFDMin metrics.Gauge // min open file descriptors (all nodes)
	OFDMax metrics.Gauge // max open file descriptors (all nodes)
	OFDAvg metrics.Gauge // avg open file descriptors (all nodes)
}

// NewPlugin initializes a new watcher for an Elasticsearch cluster.
// Pass a name to differentiate between different clusters.
func NewPlugin(name string, config *Config) (*Plugin, error) {
	if name == "" {
		return nil, errors.New("no name specified")
	}
	if config == nil {
		return nil, errors.New("no configuration specified")
	}

	client, err := elastic.NewClient(elastic.SetURL(config.Urls...))
	if err != nil {
		return nil, err
	}
	plugin := &Plugin{
		name:   name,
		urls:   config.Urls,
		client: client,
	}

	plugin.NumNodes = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.num_nodes", plugin.name), plugin.NumNodes)
	plugin.NumDataNodes = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.num_data_nodes", plugin.name), plugin.NumDataNodes)
	plugin.Shards.Active = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.shards.active", plugin.name), plugin.Shards.Active)
	plugin.Shards.Relocating = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.shards.relocating", plugin.name), plugin.Shards.Relocating)
	plugin.Shards.Initializing = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.shards.initializing", plugin.name), plugin.Shards.Initializing)
	plugin.Shards.Unassigned = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.shards.unassigned", plugin.name), plugin.Shards.Unassigned)
	plugin.NumPendingTasks = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.num_pending_tasks", plugin.name), plugin.NumPendingTasks)

	plugin.NumIndices = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.num_indices", plugin.name), plugin.NumIndices)

	plugin.HeapUsed = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.heap_used", plugin.name), plugin.HeapUsed)
	plugin.HeapMax = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.heap_max", plugin.name), plugin.HeapMax)
	plugin.HeapPercent = metrics.NewGaugeFloat64()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.heap_percent", plugin.name), plugin.HeapPercent)

	plugin.CPUPercent = metrics.NewGaugeFloat64()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.cpu_percent", plugin.name), plugin.CPUPercent)

	plugin.OFDMin = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.open_file_descriptors.min", plugin.name), plugin.OFDMin)
	plugin.OFDMax = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.open_file_descriptors.max", plugin.name), plugin.OFDMax)
	plugin.OFDAvg = metrics.NewGauge()
	metrics.Register(fmt.Sprintf("elasticsearch.%s.open_file_descriptors.avg", plugin.name), plugin.OFDAvg)

	return plugin, nil
}

// Name of the plugin.
func (p *Plugin) Name() string {
	return p.name
}

// Snapshot returns a snapshot of the current cluster metrics.
func (p *Plugin) Snapshot() (interface{}, error) {
	stats, err := GetStats(p.client)
	if err != nil {
		return nil, err
	}

	// Update metrics
	p.NumNodes.Update(stats.NumNodes)
	p.NumDataNodes.Update(stats.NumDataNodes)
	p.Shards.Active.Update(stats.Shards.Active)
	p.Shards.Relocating.Update(stats.Shards.Relocating)
	p.Shards.Initializing.Update(stats.Shards.Initializing)
	p.Shards.Unassigned.Update(stats.Shards.Unassigned)
	p.NumPendingTasks.Update(stats.NumPendingTasks)
	p.NumIndices.Update(stats.NumIndices)
	p.HeapUsed.Update(stats.HeapUsed)
	p.HeapMax.Update(stats.HeapMax)
	p.HeapPercent.Update(stats.HeapPercent)
	p.CPUPercent.Update(stats.CPUPercent)
	p.OFDMin.Update(stats.OFDMin)
	p.OFDMax.Update(stats.OFDMax)
	p.OFDAvg.Update(stats.OFDAvg)

	// Return data
	return map[string]interface{}{
		"num_nodes":           p.NumNodes.Value(),
		"num_data_nodes":      p.NumDataNodes.Value(),
		"shards_active":       p.Shards.Active.Value(),
		"shards_relocating":   p.Shards.Relocating.Value(),
		"shards_initializing": p.Shards.Initializing.Value(),
		"shards_unassigned":   p.Shards.Unassigned.Value(),
		"num_pending_tasks":   p.NumPendingTasks.Value(),
		"num_indices":         p.NumIndices.Value(),
		"heap_current":        p.HeapUsed.Value(),
		"heap_max":            p.HeapMax.Value(),
		"heap_percent":        p.HeapPercent.Value(),
		"cpu_percent":         p.CPUPercent.Value(),
		"ofd_min":             p.OFDMin.Value(),
		"ofd_max":             p.OFDMax.Value(),
		"ofd_avg":             p.OFDAvg.Value(),
	}, nil
}
