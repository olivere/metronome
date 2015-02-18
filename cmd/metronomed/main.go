package main

import (
	"flag"
	"log"
	"os"

	"github.com/olivere/metronome"
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
