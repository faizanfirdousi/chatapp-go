package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type ClientList map[*Client]bool // goes as an argument in manager struct


// goes as an argument ins ClientList
type Client struct {
	connection *websocket.Conn
	manager *Manager

	//egress is used to avoid concurrent writes on the websocket connection
	egress chan []byte
}

// a constructor/factory function to create instance of Client
func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager: manager,
		egress: make(chan []byte),
	}
}


func (c *Client) readMessages(){
	defer func() {
		//cleanup connection
		c.manager.removeClient(c)
	}()

	for {
		messageType, payload, err := c.connection.ReadMessage()

		if err != nil{
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure){
				log.Printf("error reading message: %v",err)
			}
			break
		}

		for wsclient := range c.manager.clients{
			wsclient.egress <- payload
		}

		log.Println(messageType)
		log.Println(string(payload))
	}
}

func (c *Client) writeMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <- c.egress:
			if !ok{
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed")
				}
				return
			}

			if err := c.connection.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("failed to send message: %v",err)
			}
			log.Println("message sent")
		}
	}
}
