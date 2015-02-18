package metronome

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rcrowley/go-metrics"
)

type Server struct {
	mu sync.Mutex

	mux *http.ServeMux

	conns        map[*wsConn]bool
	register     chan *wsConn
	unregister   chan *wsConn
	statusUpdate chan json.RawMessage

	load struct {
		Last1min  metrics.GaugeFloat64
		Last5min  metrics.GaugeFloat64
		Last15min metrics.GaugeFloat64
	}

	mem struct {
		Total       metrics.Gauge
		Used        metrics.Gauge
		UsedPercent metrics.GaugeFloat64
		Free        metrics.Gauge
	}

	swap struct {
		Total       metrics.Gauge
		Used        metrics.Gauge
		UsedPercent metrics.GaugeFloat64
		Free        metrics.Gauge
	}

	Addr               string
	Username, Password string
	Logger             *log.Logger
}

func NewServer() *Server {
	return &Server{
		Addr:         ":8999",
		register:     make(chan *wsConn),
		unregister:   make(chan *wsConn),
		statusUpdate: make(chan json.RawMessage),
		conns:        make(map[*wsConn]bool),
	}
}

func (s *Server) Start() {
	if err := s.initMetrics(); err != nil {
		s.printf("error initializing metrics: %v\n", err)
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

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Username != "" || s.Password != "" {
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

func (s *Server) initMetrics() error {
	s.mu.Lock()

	s.load.Last1min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last1min", s.load.Last1min)
	s.load.Last5min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last5min", s.load.Last5min)
	s.load.Last15min = metrics.NewGaugeFloat64()
	metrics.Register("loadavg.last15min", s.load.Last15min)

	s.mem.Total = metrics.NewGauge()
	metrics.Register("mem.total", s.mem.Total)
	s.mem.Free = metrics.NewGauge()
	metrics.Register("mem.free", s.mem.Free)
	s.mem.Used = metrics.NewGauge()
	metrics.Register("mem.used", s.mem.Used)
	s.mem.UsedPercent = metrics.NewGaugeFloat64()
	metrics.Register("mem.usedpercent", s.mem.UsedPercent)

	s.swap.Total = metrics.NewGauge()
	metrics.Register("swap.total", s.swap.Total)
	s.swap.Free = metrics.NewGauge()
	metrics.Register("swap.free", s.swap.Free)
	s.swap.Used = metrics.NewGauge()
	metrics.Register("swap.used", s.swap.Used)
	s.swap.UsedPercent = metrics.NewGaugeFloat64()
	metrics.Register("swap.usedpercent", s.swap.UsedPercent)

	s.mu.Unlock()
	return nil
}

func (s *Server) startUpdate() {
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			s.update()
			s.log()
		}
	}
}

// startHub watches for websocket connections.

func (s *Server) startHub() {
	var lastStatusMsg []byte
	for {
		select {
		case c := <-s.register:
			s.printf("registered client on %s", c.ws.RemoteAddr())
			s.mu.Lock()
			s.conns[c] = true
			s.mu.Unlock()
			c.send <- lastStatusMsg
			break
		case c := <-s.unregister:
			s.printf("unregistered client on %s", c.ws.RemoteAddr())
			s.mu.Lock()
			delete(s.conns, c)
			s.mu.Unlock()
			close(c.send)
			break
		case st := <-s.statusUpdate:
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

func (s *Server) update() {
	s.mu.Lock()

	// Loadavg
	loadavg, err := GetLoadAvg()
	if err == nil {
		s.load.Last1min.Update(loadavg.Last1Min)
		s.load.Last5min.Update(loadavg.Last5Min)
		s.load.Last15min.Update(loadavg.Last15Min)
	}

	// Mem
	mem, err := GetMem()
	if err == nil {
		s.mem.Total.Update(mem.Total)
		s.mem.Free.Update(mem.Free)
		s.mem.Used.Update(mem.Used)
		s.mem.UsedPercent.Update(mem.UsedPercent)
	}

	// Swap
	swap, err := GetSwap()
	if err == nil {
		s.swap.Total.Update(swap.Total)
		s.swap.Free.Update(swap.Free)
		s.swap.Used.Update(swap.Used)
		s.swap.UsedPercent.Update(swap.UsedPercent)
	}

	s.mu.Unlock()
}

func (s *Server) log() {
	msg := &Status{}
	msg.LoadAvg.Load1Min = s.load.Last1min.Value()
	msg.LoadAvg.Load5Min = s.load.Last5min.Value()
	msg.LoadAvg.Load15Min = s.load.Last15min.Value()
	msg.Mem.Total = s.mem.Total.Value()
	msg.Mem.Free = s.mem.Free.Value()
	msg.Mem.Used = s.mem.Used.Value()
	msg.Swap.Total = s.swap.Total.Value()
	msg.Swap.Free = s.swap.Free.Value()
	msg.Swap.Used = s.swap.Used.Value()
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.statusUpdate <- data
}

// home is the home page on /.
func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{}")
}

// stats is the websocket endpoint on /stats.
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
