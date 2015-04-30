// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package plugins

import "sync"

var (
	pluginsMu sync.RWMutex
	plugins   []Plugin
)

// Register a plugin. Use this function before starting a Metrononme server.
func Register(plugin Plugin) {
	if plugin == nil {
		panic("metronome: Register plugin is nil")
	}
	pluginsMu.Lock()
	defer pluginsMu.Unlock()
	if plugins == nil {
		plugins = make([]Plugin, 0)
	}
	plugins = append(plugins, plugin)
}

// Plugins returns a list of all registered plugins.
func Plugins() []Plugin {
	pluginsMu.RLock()
	defer pluginsMu.RUnlock()
	var list []Plugin
	for _, plugin := range plugins {
		list = append(list, plugin)
	}
	return list
}
