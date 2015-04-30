// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package plugins

// Plugin is a component that watches a resource and can return metrics.
type Plugin interface {
	// Name of the plugin.
	Name() string

	// Snapshot asks the plugin to return a snapshot of the metrics.
	Snapshot() (interface{}, error)
}
