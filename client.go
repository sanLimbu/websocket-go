package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// ClientList is a map used to help manage a map of clients
type ClientList map[*Client]bool

// Client is a websocket client, basically a frontend visitor
type Client struct {
	//the websocket connection
	connection *websocket.Conn

	// manager is the manager used to manage the client
	manager *Manager

	//egress is used to avoid concurrent writes on the websocket
	egress chan []byte
}

// NewClient is used to initialize a new Client with all required values initialized

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan []byte),
	}
}

// readMessages will start the client to read messages and handle them
// appropriatly.
// This is suppose to be ran as a goroutine
func (c *Client) readMessages() {
	defer func() {
		//Graceful close the connection once this func is done
		c.manager.removeClient(c)
	}()

	//Loop forver
	for {
		//Read message is used to read the next message in queue in the connection
		messageType, payload, err := c.connection.ReadMessage()

		if err != nil {
			//if connection is closed, we get error message here
			//log strange errors, not simpl disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break //break the loop to close connection and cleanup

		}
		log.Println("MessageType: ", messageType)
		log.Println("Payload: ", string(payload))

		//hack to test that writemessage works as intended will be replace soon
		for wsclient := range c.manager.clients {
			wsclient.egress <- payload
		}
	}

}

// writeMessages is a process that listens for new messages to output to the client
func (c *Client) writeMessages() {

	defer func() {
		//Graceful close if this triggers a closing
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			//ok will be false incase the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					//log that the connection is closed and the reason
					log.Println("connection closed:", err)
				}
				//return to close the goroutine
				return

			}
			//write a regular text message to the connection
			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		}
	}

}
