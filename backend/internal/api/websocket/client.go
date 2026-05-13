package websocket

import "github.com/gorilla/websocket"

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for msg := range c.send {
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, _, err := c.conn.ReadMessage()

		if err != nil {
			break
		}
	}
}
