package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Protocol:
// Client -> Server:
//   {"type":"join","room":"lobby","name":"alice","token":""}
//   {"type":"msg","text":"hello"}
//
// Server -> Client:
//   {"type":"token","token":"...","name":"alice","room":"lobby"}
//   {"type":"msg","name":"alice","room":"lobby","text":"hello"}
//   {"type":"error","message":"..."}

type inbound struct {
	Type  string `json:"type"`
	Room  string `json:"room,omitempty"`
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
	Text  string `json:"text,omitempty"`
}

type outbound struct {
	Type    string `json:"type"`
	Room    string `json:"room,omitempty"`
	Name    string `json:"name,omitempty"`
	Token   string `json:"token,omitempty"`
	Text    string `json:"text,omitempty"`
	Message string `json:"message,omitempty"`
}

type identity struct {
	Name  string
	Token string
	Room  string
}

type client struct {
	conn *websocket.Conn
	send chan []byte
	id   identity
}

type roomHub struct {
	name       string
	clients    map[*client]struct{}
	register   chan *client
	unregister chan *client
	broadcast  chan outbound
}

func newRoomHub(name string) *roomHub {
	return &roomHub{
		name:       name,
		clients:    make(map[*client]struct{}),
		register:   make(chan *client),
		unregister: make(chan *client),
		broadcast:  make(chan outbound, 128),
	}
}

func (h *roomHub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.broadcast <- outbound{Type: "msg", Room: h.name, Name: "system", Text: fmt.Sprintf("%s joined", c.id.Name)}
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
				h.broadcast <- outbound{Type: "msg", Room: h.name, Name: "system", Text: fmt.Sprintf("%s left", c.id.Name)}
			}
		case msg := <-h.broadcast:
			b, _ := json.Marshal(msg)
			for c := range h.clients {
				select {
				case c.send <- b:
				default:
					delete(h.clients, c)
					close(c.send)
					_ = c.conn.Close()
				}
			}
		}
	}
}

type server struct {
	upgrader websocket.Upgrader

	mu      sync.Mutex
	rooms   map[string]*roomHub
	tokens  map[string]string // token -> name
	nameSeq map[string]int    // baseName -> last suffix
}

func newServer() *server {
	return &server{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		rooms:   make(map[string]*roomHub),
		tokens:  make(map[string]string),
		nameSeq: make(map[string]int),
	}
}

func (s *server) getOrCreateRoom(room string) *roomHub {
	s.mu.Lock()
	defer s.mu.Unlock()

	if h, ok := s.rooms[room]; ok {
		return h
	}
	h := newRoomHub(room)
	s.rooms[room] = h
	go h.run()
	return h
}

func (s *server) issueToken(name string) string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	token := hex.EncodeToString(buf)

	s.mu.Lock()
	s.tokens[token] = name
	s.mu.Unlock()
	return token
}

func (s *server) nameFromToken(token string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name, ok := s.tokens[token]
	return name, ok
}

func (s *server) uniqueName(requested string) string {
	base := strings.TrimSpace(requested)
	if base == "" {
		base = fmt.Sprintf("user-%d", time.Now().Unix()%100000)
	}
	base = strings.ReplaceAll(base, " ", "_")

	s.mu.Lock()
	defer s.mu.Unlock()

	// Very lightweight uniqueness: if base already has a sequence, increment.
	if last, ok := s.nameSeq[base]; ok {
		last++
		s.nameSeq[base] = last
		return fmt.Sprintf("%s_%d", base, last)
	}
	s.nameSeq[base] = 0
	return base
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "websocket upgrade failed", http.StatusBadRequest)
		return
	}

	c := &client{
		conn: conn,
		send: make(chan []byte, 64),
	}

	go writePump(c)
	readPump(s, c)
}

func writePump(c *client) {
	for msg := range c.send {
		_ = c.conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func readPump(s *server, c *client) {
	defer func() {
		_ = c.conn.Close()
	}()

	var hub *roomHub
	joined := false

	sendErr := func(message string) {
		b, _ := json.Marshal(outbound{Type: "error", Message: message})
		select {
		case c.send <- b:
		default:
		}
	}

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if joined && hub != nil {
				hub.unregister <- c
			}
			return
		}

		var in inbound
		if err := json.Unmarshal(data, &in); err != nil {
			sendErr("invalid json")
			continue
		}

		switch in.Type {
		case "join":
			if joined {
				sendErr("already joined")
				continue
			}
			room := strings.TrimSpace(in.Room)
			if room == "" {
				room = "lobby"
			}

			name := strings.TrimSpace(in.Name)
			token := strings.TrimSpace(in.Token)
			if token != "" {
				if tokenName, ok := s.nameFromToken(token); ok {
					name = tokenName
				} else {
					sendErr("unknown token")
					continue
				}
			}
			name = s.uniqueName(name)

			c.id = identity{
				Name:  name,
				Token: s.issueToken(name),
				Room:  room,
			}

			hub = s.getOrCreateRoom(room)
			hub.register <- c
			joined = true

			b, _ := json.Marshal(outbound{
				Type:  "token",
				Room:  room,
				Name:  name,
				Token: c.id.Token,
			})
			c.send <- b

		case "msg":
			if !joined || hub == nil {
				sendErr("join first")
				continue
			}
			txt := strings.TrimSpace(in.Text)
			if txt == "" {
				continue
			}
			hub.broadcast <- outbound{
				Type: "msg",
				Room: hub.name,
				Name: c.id.Name,
				Text: txt,
			}
		default:
			sendErr("unknown message type")
		}
	}
}

func main() {
	addr := flag.String("addr", ":8081", "http listen address")
	flag.Parse()

	s := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ws", s.handleWS)

	log.Printf("http server listening on %s", *addr)
	log.Printf("health: http://localhost%s/health", *addr)
	log.Printf("ws: ws://localhost%s/ws", *addr)

	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
