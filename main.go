package main

import (
	"log"
	"net/http"
)

func main() {
	setupAPI()
	log.Fatal(http.ListenAndServe(":3000", nil))

}

// setupAPI will start all Routes and their Handlers

func setupAPI() {

	//Create a manager instance used to handle websocket connection
	manager := NewManager()

	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/ws", manager.serveWS)
}
