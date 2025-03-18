package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

// ping pong messages are heartbeats of a websocket connection
var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager

	// egress is used to avoid concurrent writes on the websocket connection
	egress chan Event
}

type Manager struct {
	clients ClientList
	sync.RWMutex

	handlers map[string]EventHandler
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func NewManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessage
}

func SendMessage(event Event, c *Client) error {
	var chatEvent SendMessageEvent

	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadMessage NewMessageEvent
	broadMessage.Sent = time.Now()
	broadMessage.Message = chatEvent.Message
	broadMessage.From = chatEvent.From

	data, err := json.Marshal(broadMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal the broadcast: %v", err)
	}

	outgoingEvent := Event{
		Payload: data,
		Type:    EventNewMessage,
	}

	for client := range c.manager.clients {
		client.egress <- outgoingEvent
	}

	return nil
}

var e error

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		log.Println("There is no such event type")
		return e
	}

}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Check if the user is authorized or not

	log.Println("New connection")

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := NewClient(conn, m)
	m.addClient(client)

	// Start client processes

	go client.readMessages()
	go client.writeMessages()
}

func (m *Manager) addClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	m.clients[client] = true
	log.Println("Client added:", client)
}

func (m *Manager) removeClient(client *Client) {
	m.Lock()
	defer m.Unlock()

	_, ok := m.clients[client]
	if ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

// The websocket connection from gorilla package only allows one concurrent writer

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	// The pong message deadline setter
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}

	// to handle jumbo frames ^^/
	c.connection.SetReadLimit(512)

	// pong handling function which handles a function of the client
	c.connection.SetPongHandler(c.pongHandler)

	for {
		// msgTypes : ping, pong, data, control, binary etc.
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println(err)
			}
			break
		}
		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Println(err)
			break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println(err)
		}

	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	ticker := time.NewTicker(pingInterval)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)

			if err != nil {
				log.Println(err)
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
				return
			}
			log.Println("message sent")

		case <-ticker.C:
			log.Println("ping")

			//send a ping to the client
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Println("Writing error: ", err)
				return
			}
		}

	}
}

// pong message handler
func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

// Create a manager(Hub) to manage all the users present
// Create the users(clients) and manager(hub) in models
// Establish a websocket connection
// Add the users to the connection
// Read Write handlers
// Keeping the websocket server alive
// Manage the origin of requests i.e. checkOrigin
// Remove the users
// Create the events like send message change user etc.
