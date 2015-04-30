// Copyright 2012-2015 Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package metronome

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client connects to a server via Websockets.
type Client struct {
	mu       sync.Mutex
	ws       *websocket.Conn
	addr     string
	username string
	password string

	// Connected is used to indicate a successful connection.
	Connected chan bool

	// Disconnected is used to indicate a server disconnect.
	Disconnected chan bool

	// Incoming has messages sent from server to client.
	Incoming chan []byte
}

// NewClient returns a client that connects to a server via Websockets.
func NewClient(addr, username, password string) (*Client, error) {
	c := &Client{
		addr:         addr,
		username:     username,
		password:     password,
		Connected:    make(chan bool),
		Disconnected: make(chan bool),
		Incoming:     make(chan []byte),
	}

	go c.autoconnect()

	return c, nil
}

func (c *Client) header() http.Header {
	hdr := http.Header{}
	if c.username != "" || c.password != "" {
		credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.username, c.password)))
		hdr.Add("Authorization", fmt.Sprintf("Basic %s", credentials))
	}
	return hdr
}

// autoconnect tries to connect to the server periodically.
func (c *Client) autoconnect() {
	ticker := time.NewTicker(10 * time.Second)

	c.connect()

	for {
		select {
		case <-ticker.C:
			c.connect()
			break
		}
	}
}

// connect connects to the server. It is a no-op if the client is already
// connected.
func (c *Client) connect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ws == nil {
		header := c.header()
		ws, _, err := websocket.DefaultDialer.Dial(c.addr, header)
		if err != nil {
			log.Printf("cannot connect: %v", err)
		} else {
			c.ws = ws
			go c.readPump()
		}
	}
}

// readPump is a goroutine waiting for messages from the server.
func (c *Client) readPump() {
	defer func() {
		c.mu.Lock()
		c.Disconnected <- true // indicate that we are disconnected
		c.ws.Close()
		c.ws = nil
		c.mu.Unlock()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))

	// Client should return a PongMessage when asked.
	// Ping/Pong is used to keep both parties alive.
	c.ws.SetPingHandler(func(string) error {
		c.ws.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(pingPeriod))
		return nil
	})

	// Indicate that we are connected.
	c.Connected <- true

	for {
		// Take a message via websocket
		typ, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}

		switch typ {
		case websocket.CloseMessage:
			// Server decided to close the connection.
			return
		case websocket.TextMessage:
			// Server sent a message, so pass it on to the app using this client
			c.Incoming <- msg
			break
		}
	}
}
