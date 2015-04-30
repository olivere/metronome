// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/olivere/metronome"
)

var (
	addr     = flag.String("url", "ws://127.0.0.1:8999/stats", "Websocket server address (e.g. 'ws://127.0.0.1:8999/stats')")
	username = flag.String("username", "", "Username for authentication")
	password = flag.String("password", "", "Password for authentication")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	client, err := metronome.NewClient(*addr, *username, *password)
	if err != nil {
		log.Fatalf("%v", err)
	}

	for {
		select {
		case <-client.Connected:
			fmt.Fprintf(os.Stdout, "connected\n")
		case <-client.Disconnected:
			fmt.Fprintf(os.Stdout, "disconnected\n")
		case st := <-client.Incoming:
			fmt.Fprintf(os.Stdout, "%s\n", string(st))
			/*
				var msg metronome.Status
				if err := json.Unmarshal(st, &msg); err != nil {
					fmt.Fprintf(os.Stderr, "error decoding: %v", err)
				} else {
					memAvail := megabyte(msg.Mem.Total - msg.Mem.Used)
					swapAvail := megabyte(msg.Swap.Total - msg.Swap.Used)

					fmt.Fprintf(os.Stdout, "LoadAvg: %.2f, %.2f, %.2f, Mem: %.1fM total, %.1fM avail, %.1fM free, Swap: %.1fM total, %.1fM avail, %.1fM free\n",
						msg.LoadAvg.Load1Min,
						msg.LoadAvg.Load5Min,
						msg.LoadAvg.Load15Min,
						megabyte(msg.Mem.Total),
						memAvail,
						megabyte(msg.Mem.Free),
						megabyte(msg.Swap.Total),
						swapAvail,
						megabyte(msg.Swap.Used))
				}
			*/
		}
	}
}

func megabyte(b int64) float64 {
	return float64(b) / 1024 / 1024
}
