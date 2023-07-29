package main

import "encoding/json"

type Event struct {

	//TYpe is the message type sent
	Type string `json:"type"`

	//Payload is the data based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(event Event, c *Client) error

const (
	//EventSendMessage is the event name for new chat message sent
	EventSendMessage = "send_message"
)

// SendMessageEvent is the payload sent in the send_message event
type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}
