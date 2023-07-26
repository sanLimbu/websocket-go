package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (

	/**
	websocketUpgrader is used to upgrade incomming HTTP requests into a persitent websocket connection
	*/

	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 2014,
	}
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc

type Manager struct {
	clients ClientList

	// Using a syncMutex here to be able to lcok state before editing clients
	// Could also use Channels to block
	sync.RWMutex
}

// NewManager is used to initalize all the values inside the manager

func NewManager() *Manager {
	return &Manager{
		clients: make(ClientList),
	}
}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New connection")

	// Begin by upgrading the HTTP request
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//Create a new client
	client := NewClient(conn, m)

	//Add the newly created client to the manager
	m.addClient(client)

	//Start the read / write processes
	go client.readMessages()
	go client.writeMessages()

}

// Addclient will add clients to our clientlist
func (m *Manager) addClient(client *Client) {

	//Lock so we can manipulate
	m.Lock()
	defer m.Unlock()

	//add client
	m.clients[client] = true
}

//removeClient will remove the client and clean up

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	//check if the client exists, then delete it
	if _, ok := m.clients[client]; ok {

		//close connection
		client.connection.Close()

		//remove
		delete(m.clients, client)
	}
}
