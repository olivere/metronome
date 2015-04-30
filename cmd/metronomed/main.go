// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"

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
	conffile = flag.String("c", "metronomed.toml", "Configuration file")
)

func main() {
	log.SetFlags(0)
	flag.Parse()

	if err := registerPlugins(*conffile); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

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

type configuration struct {
	LoadAvg       interface{} `toml:"loadavg"`
	Mem           interface{}
	Swap          interface{}
	Elasticsearch map[string]*esconf
}

type esconf struct {
	Urls []string
}

func registerPlugins(conffile string) error {
	var config configuration
	_, err := toml.DecodeFile(conffile, &config)
	if err != nil {
		return err
	}

	// LoadAvg
	if config.LoadAvg != nil {
		loadavgPlugin, err := loadavg.NewPlugin()
		if err != nil {
			return fmt.Errorf("error initializing loadavg plugin: %v", err)
		}
		plugins.Register(loadavgPlugin)
	}

	// Mem
	if config.Mem != nil {
		memPlugin, err := mem.NewPlugin()
		if err != nil {
			return fmt.Errorf("error initializing mem plugin: %v", err)
		}
		plugins.Register(memPlugin)
	}

	// Swap
	if config.Swap != nil {
		swapPlugin, err := swap.NewPlugin()
		if err != nil {
			return fmt.Errorf("error initializing swap plugin: %v", err)
		}
		plugins.Register(swapPlugin)
	}

	// Elasticsearch
	if config.Elasticsearch != nil {
		for name, escfg := range config.Elasticsearch {
			esConfig := &elasticsearch.Config{Urls: escfg.Urls}
			esPlugin, err := elasticsearch.NewPlugin(name, esConfig)
			if err != nil {
				return fmt.Errorf("error initializing elasticsearch plugin: %v", err)
			}
			plugins.Register(esPlugin)
		}
	}

	return nil
}
