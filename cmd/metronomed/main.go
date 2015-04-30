// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/olivere/metronome"
	"github.com/olivere/metronome/plugins"
	"github.com/olivere/metronome/plugins/elasticsearch"
	"github.com/olivere/metronome/plugins/loadavg"
	"github.com/olivere/metronome/plugins/mem"
	"github.com/olivere/metronome/plugins/swap"
)

var (
	addr     = flag.String("http", "", "HTTP server address (e.g. ':8999')")
	username = flag.String("username", "", "Username for authentication")
	password = flag.String("password", "", "Password for authentication")
	logfile  = flag.String("log", "", "Log file")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	// TODO use a config file for loading plugins

	loadavgPlugin, err := loadavg.NewPlugin()
	if err != nil {
		log.Fatalf("error initializing loadavg plugin: %v", err)
		os.Exit(1)
	}
	plugins.Register(loadavgPlugin)

	memPlugin, err := mem.NewPlugin()
	if err != nil {
		log.Fatalf("error initializing mem plugin: %v", err)
		os.Exit(1)
	}
	plugins.Register(memPlugin)

	swapPlugin, err := swap.NewPlugin()
	if err != nil {
		log.Fatalf("error initializing swap plugin: %v", err)
		os.Exit(1)
	}
	plugins.Register(swapPlugin)

	esConfig := &elasticsearch.Config{Urls: []string{"http://localhost:9200"}}
	esPlugin, err := elasticsearch.NewPlugin("elasticsearch", esConfig)
	if err != nil {
		log.Fatalf("error initializing elasticsearch plugin: %v", err)
		os.Exit(1)
	}
	plugins.Register(esPlugin)

	// Initialize server

	srv := metronome.NewServer()

	if *addr != "" {
		srv.Addr = *addr
	}
	srv.Username = *username
	srv.Password = *password
	if *logfile != "" {
		f, err := os.OpenFile(*logfile, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("cannot open file %q: %v", *logfile, err)
			os.Exit(1)
		}
		defer f.Close()
		srv.Logger = log.New(f, "", log.Lshortfile|log.Lmicroseconds)
	}

	go srv.Start()

	select {}
}
