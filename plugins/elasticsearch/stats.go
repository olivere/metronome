// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elasticsearch

import "github.com/olivere/elastic"

// Stats contains all information gathered of a cluster periodically.
type Stats struct {
	Cluster struct {
		Name  string
		State string
	}

	NumNodes     int64
	NumDataNodes int64
	Shards       struct {
		Active       int64
		Relocating   int64
		Initializing int64
		Unassigned   int64
	}
	NumPendingTasks int64

	NumIndices int64

	HeapUsed    int64
	HeapMax     int64
	HeapPercent float64

	CPUPercent float64

	OFDMin int64
	OFDMax int64
	OFDAvg int64
}

// GetStats gathers a snapshot of the cluster.
func GetStats(client *elastic.Client) (*Stats, error) {
	health, err := client.ClusterHealth().Do()
	if err != nil {
		return nil, err
	}
	cs, err := client.ClusterStats().Do()
	if err != nil {
		return nil, err
	}

	stats := &Stats{}
	stats.Cluster.Name = health.ClusterName
	stats.Cluster.State = health.Status

	stats.NumNodes = int64(health.NumberOfNodes)
	stats.NumDataNodes = int64(health.NumberOfDataNodes)
	stats.Shards.Active = int64(health.ActiveShards)
	stats.Shards.Relocating = int64(health.RelocatingShards)
	stats.Shards.Initializing = int64(health.InitializingShards)
	stats.Shards.Unassigned = int64(health.UnassignedShards)
	stats.NumPendingTasks = int64(health.NumberOfPendingTasks)

	stats.NumIndices = int64(cs.Indices.Count)

	stats.HeapUsed = cs.Nodes.JVM.Mem.HeapUsedInBytes
	stats.HeapMax = cs.Nodes.JVM.Mem.HeapMaxInBytes
	if stats.HeapMax > 0 {
		stats.HeapPercent = float64(100*stats.HeapUsed) / float64(stats.HeapMax)
	}

	stats.CPUPercent = cs.Nodes.Process.CPU.Percent

	stats.OFDMin = cs.Nodes.Process.OpenFileDescriptors.Min
	stats.OFDMax = cs.Nodes.Process.OpenFileDescriptors.Max
	stats.OFDAvg = cs.Nodes.Process.OpenFileDescriptors.Avg

	return stats, nil
}
