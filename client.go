package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

var (
	//pongWait is how long we will await a pong rsponse from client
	pongWait = 10 * time.Second

	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
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
	egress chan Event

	// chatroom is used to know what room user is in
	chatroom string
}

// NewClient is used to initialize a new Client with all required values initialized

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
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

	//Set the max sige of messages in bytes
	c.connection.SetReadLimit(1024)

	//Configure Wait time for Pong response, use Current time+pongWait.
	//This has to be done here to set the first initial timer.
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	//Configure how to handle Pong response
	c.connection.SetPongHandler(c.pongHandler)

	//Loop forver
	for {
		//Read message is used to read the next message in queue in the connection
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			//if connection is closed, we get error message here
			//log strange errors, not simpl disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break //break the loop to close connection and cleanup

		}

		//Marshal incoming data into a Event struct
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error marshalling message : %v", err)
			break
		}
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handeling Message: ", err)
		}
	}

}

// pongHandler is used to handle PongMessages for the client
func (c *Client) pongHandler(pongMsg string) error {
	//Current time + Pong Wait time
	log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

// writeMessages is a process that listens for new messages to output to the client
func (c *Client) writeMessages() {

	//Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)

	defer func() {
		ticker.Stop()
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

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return //close the connection
			}

			//write a regular text message to the connection
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")
		case <-ticker.C:
			log.Println("ping")
			//Send the ping
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("writemsg :", err)
				return //return to break this goroutine triggering cleanup
			}
		}
	}

}
