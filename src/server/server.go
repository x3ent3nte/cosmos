package server

import (
	"time"
	"sync"
	"log"
	"net/http"
    "github.com/gorilla/websocket"
    "concurrent"
)

type Message struct {
	Type string `json:"type"`
	Data interface{} `json:"data"`
}

type Client struct {
	id int64
	keycode int
	conn *websocket.Conn
}

func (client *Client) setKeycode(new_keycode int) {
	client.keycode = new_keycode
}

type Server struct {
	sync.RWMutex
	upgrader websocket.Upgrader
	ids concurrent.IdHandler
	clients map[*Client]bool
	data []string
}

func (server *Server) ServeData() {
	for {
		msg := Message{"update", server.data}
		for client := range server.clients {
			client.conn.WriteJSON(msg)
		}
		time.Sleep(time.Millisecond * 1)
	}
}

func (server *Server) SetData(data []string) {
	server.Lock()
	server.data = data
	server.Unlock()
}

func (server *Server) AddClient(client *Client) {
	server.Lock()
	server.clients[client] = true
	server.Unlock()
}

func (server *Server) RemoveClient(client *Client) {
	server.Lock()
	delete(server.clients, client)
	server.Unlock()
}

func (server *Server) GetClientsData() map[int64]int {
	//server.Lock()
	data := make(map[int64]int)
	for client := range server.clients {
		data[client.id] = client.keycode
	}
	//server.Unlock()
	return data
}

func (server *Server) handleConnections(write http.ResponseWriter, read *http.Request) {
	conn, err := server.upgrader.Upgrade(write, read, nil)
	if err != nil {
		log.Fatal(err)
	}

	client := Client{server.ids.NextId(), 0, conn}
	server.AddClient(&client)

	//id_msg := CreateMessage("id", client.id)
	id_msg := Message{"id", client.id}
	conn.WriteJSON(id_msg)
	
	go func (client *Client) {
		for {
			_, msg, err := client.conn.ReadMessage()
			if err != nil {
				log.Printf("error: %v", err)
				server.RemoveClient(client)
				break
			}
			keycode := 0;
			factor := 1
			for i := 0; i < len(msg); i++ {
				keycode += int(msg[i]) * factor
				factor *= 256
			}
			client.setKeycode(keycode)
		}
		log.Println("Client closed")
		conn.Close()
	}(&client)
}

func (server *Server) StartServer() {
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	http.HandleFunc("/ws", server.handleConnections)

	log.Println("http server started on :3000")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func CreateServer() Server{
	ws := websocket.Upgrader{ 
		ReadBufferSize:  1024,
    	WriteBufferSize: 1024,
    	CheckOrigin: func(r *http.Request) bool {
        return true
        },
    }
	return Server{sync.RWMutex{}, ws, concurrent.CreateIdHandler(), make(map[*Client]bool), make([]string, 0)}
}














