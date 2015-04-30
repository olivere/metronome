// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package metronome

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olivere/metronome/plugins"
)

var (
	defaultUpdateInterval = time.Duration(5) * time.Second
)

// Server gathers information via plugins and sends it to registered
// clients via websockets.
type Server struct {
	mu sync.Mutex

	mux *http.ServeMux

	updateInterval time.Duration

	conns        map[*wsConn]bool
	register     chan *wsConn         // for new clients joining
	unregister   chan *wsConn         // for clients leaving
	statusUpdate chan json.RawMessage // for status updates

	Addr               string
	Username, Password string
	Logger             *log.Logger
}

// NewServer creates a new Metronome server. Use Start to start it up.
func NewServer() *Server {
	return &Server{
		Addr:           "127.0.0.1:8999",
		register:       make(chan *wsConn),
		unregister:     make(chan *wsConn),
		statusUpdate:   make(chan json.RawMessage),
		conns:          make(map[*wsConn]bool),
		updateInterval: defaultUpdateInterval,
	}
}

// UpdateInterval specifies the time between two snapshots.
func (s *Server) UpdateInterval(interval time.Duration) *Server {
	s.updateInterval = interval
	return s
}

// Start starts the server.
func (s *Server) Start() {
	if err := s.initPlugins(); err != nil {
		s.printf("error initializing plugins: %v\n", err)
		os.Exit(1)
	}

	if err := s.initMux(); err != nil {
		s.printf("error initializing mux: %v\n", err)
		os.Exit(1)
	}

	go s.startHub()

	go s.startUpdate()

	//go metrics.Log(metrics.DefaultRegistry, 1*time.Second, log.New(os.Stdout, "", log.Lmicroseconds))
	//go s.log()

	httpSrv := &http.Server{
		Addr:    s.Addr,
		Handler: s,
	}
	err := httpSrv.ListenAndServe()
	if err != nil {
		s.printf("error starting http server: %v\n", err)
		os.Exit(1)
	}
}

// ServeHTTP handles HTTP requests.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Username != "" || s.Password != "" {
		// Check for authentication.
		u, p, ok := r.BasicAuth()
		if !ok {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		if u != s.Username || p != s.Password {
			http.Error(w, "", http.StatusForbidden)
			return
		}
	}

	// Use the muxer to handle different requests.
	s.mux.ServeHTTP(w, r)
}

func (s *Server) printf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (s *Server) fprintf(format string, args ...interface{}) {
	if s.Logger != nil {
		s.Logger.Fatalf(format, args...)
	} else {
		log.Fatalf(format, args...)
	}
}

func (s *Server) initMux() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.home)
	mux.HandleFunc("/stats", s.stats)

	s.mu.Lock()
	s.mux = mux
	s.mu.Unlock()

	return nil
}

// initPlugins initializes all registered plugins.
func (s *Server) initPlugins() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plugins := plugins.Plugins()
	if len(plugins) == 0 {
		return errors.New("no plugins registered")
	}

	return nil
}

// startUpdate periodically gathers metrics and sends updates to
// registered clients. Use UpdateInterval to specify how often an
// update happens.
func (s *Server) startUpdate() {
	ticker := time.NewTicker(s.updateInterval)
	for {
		select {
		case <-ticker.C:
			s.update()
		}
	}
}

// startHub watches for websocket connections (joining clients, leaving
// clients, and sending status updates).
func (s *Server) startHub() {
	var lastStatusMsg []byte
	for {
		select {
		case c := <-s.register:
			// New client joins
			s.printf("registered client on %s", c.ws.RemoteAddr())
			s.mu.Lock()
			s.conns[c] = true
			s.mu.Unlock()
			c.send <- lastStatusMsg // send last known message to new client
			break
		case c := <-s.unregister:
			// Client leaves
			s.printf("unregistered client on %s", c.ws.RemoteAddr())
			s.mu.Lock()
			delete(s.conns, c)
			s.mu.Unlock()
			close(c.send)
			break
		case st := <-s.statusUpdate:
			// Send status update to all registered clients
			lastStatusMsg = st
			s.mu.Lock()
			for c := range s.conns {
				c.send <- st
			}
			s.mu.Unlock()
			break
		}
	}
}

// update asks all plugins to return a snapshot of their watched data,
// then sends it to all clients.
func (s *Server) update() {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := &Status{Metrics: make(map[string]interface{})}

	// Take a snapshot from each client.
	for _, plugin := range plugins.Plugins() {
		data, err := plugin.Snapshot()
		if err != nil {
			continue
		}
		if data != nil {
			msg.Metrics[plugin.Name()] = data
		}
	}

	// Convert the map into a JSON structure, then pass it to the WS handler.
	data, err := json.Marshal(msg)
	if err != nil {
		log.Print(err)
		return
	}
	s.statusUpdate <- data // startHub handles the sending (see above)
}

// home is the home page on / and returns {}.
func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{}")
}

// stats is the websocket endpoint on /stats.
//
// It tries to do a WebSocket upgrade/handshake and starts a new
// read/write pump for the new client.
func (s *Server) stats(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		s.printf("%v", err)
		return
	}
	c := &wsConn{
		ws:     ws,
		send:   make(chan []byte, 256),
		server: s,
	}
	s.register <- c
	go c.writePump()
	c.readPump()
}
