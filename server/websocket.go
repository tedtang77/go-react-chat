package main

import (
	// "fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
	// needed to allow connections from any origin for :3000 -> :8081
	CheckOrigin: func(r *http.Request) bool { return true },
}

type JsonData struct {
	Text string `json:"text"`
	Client string `json:"client"`
	Timestamp string `json:"timestamp"`
}

type Client struct {
	hub *ConnHub
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) readPump() {
	// schedule client to be disconnected
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// init client connection
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// handle connection read
	for {
		// read JSON data from connection
		// message := JsonData{}
		// if err := c.conn.ReadJSON(&message); err != nil {
		// 	fmt.Println("Error reading JSON", err)
		// }
		// fmt.Printf("Get response: %#v\n", message)

		// // queue message for writing
		// c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// channel has been closed by the hub
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// coalesce pending messages into one message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		// send ping over websocket
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// handle /ws route, upgrade HTTP request and begin handling of client conn
func wsHandler(hub *ConnHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	// init new client, register to hub
	client := &Client{
		hub: hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	// separate reads and writes to conform to WebSocket standard of one concurrent reader and writer
	// go client.writePump()
	// go client.readPump()
}
