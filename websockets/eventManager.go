package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	// use the same string fields that you are using in javascript
	EventSendMessage = "send_message"
	EventNewMessage  = "new_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	site := os.Getenv("ORIGIN")

	switch origin {
	case site:
		return true
	default:
		return false
	}
}
