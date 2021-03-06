// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package metronome

// Status is a status update sent to registered clients.
type Status struct {
	// Metrics data.
	Metrics map[string]interface{} `json:"metrics"`
}
