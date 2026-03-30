package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type client struct {
	name string
	conn net.Conn
	send chan string
}

type messageReq struct {
	from *client
	text string
}

type joinReq struct {
	client *client
	room   string
}

type privateReq struct {
	from   *client
	toName string
	text   string
}

type usersReq struct {
	reply chan []string
}

type hub struct {
	clients      map[*client]struct{}
	users        map[string]*client
	rooms        map[*client]string
	register     chan *client
	unregister   chan *client
	broadcastAll chan string
	publicMsg    chan messageReq
	joinRoom     chan joinReq
	privateMsg   chan privateReq
	usersList    chan usersReq
}

func newHub() *hub {
	return &hub{
		clients:      make(map[*client]struct{}),
		users:        make(map[string]*client),
		rooms:        make(map[*client]string),
		register:     make(chan *client),
		unregister:   make(chan *client),
		broadcastAll: make(chan string, 128),
		publicMsg:    make(chan messageReq, 128),
		joinRoom:     make(chan joinReq, 64),
		privateMsg:   make(chan privateReq, 128),
		usersList:    make(chan usersReq),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.register:
			c.name = h.uniqueName(c.name)
			h.clients[c] = struct{}{}
			h.users[c.name] = c
			h.rooms[c] = "lobby"
			h.safeSend(c, fmt.Sprintf("[system] welcome, %s! current room: lobby", c.name))
			h.safeSend(c, "[system] commands: /join <room>, /msg <user> <text>, /who, /quit")
			h.sendToRoom("lobby", fmt.Sprintf("[system] %s joined room lobby", c.name), c)
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				oldRoom := h.rooms[c]
				delete(h.clients, c)
				delete(h.users, c.name)
				delete(h.rooms, c)
				close(c.send)
				h.sendToRoom(oldRoom, fmt.Sprintf("[system] %s left room %s", c.name, oldRoom), c)
			}
		case msg := <-h.broadcastAll:
			for c := range h.clients {
				h.safeSend(c, msg)
			}
		case req := <-h.publicMsg:
			room := h.rooms[req.from]
			h.sendToRoom(room, fmt.Sprintf("[%s@%s] %s", req.from.name, room, req.text), nil)
		case req := <-h.joinRoom:
			h.handleJoin(req)
		case req := <-h.privateMsg:
			h.handlePrivate(req)
		case req := <-h.usersList:
			users := make([]string, 0, len(h.users))
			for name := range h.users {
				users = append(users, name)
			}
			sort.Strings(users)
			req.reply <- users
		}
	}
}

func (h *hub) handleJoin(req joinReq) {
	if _, ok := h.clients[req.client]; !ok {
		return
	}
	newRoom := strings.TrimSpace(req.room)
	if newRoom == "" {
		h.safeSend(req.client, "[system] room name cannot be empty")
		return
	}
	oldRoom := h.rooms[req.client]
	if oldRoom == newRoom {
		h.safeSend(req.client, fmt.Sprintf("[system] you are already in room %s", newRoom))
		return
	}

	h.rooms[req.client] = newRoom
	h.sendToRoom(oldRoom, fmt.Sprintf("[system] %s left room %s", req.client.name, oldRoom), req.client)
	h.safeSend(req.client, fmt.Sprintf("[system] joined room %s", newRoom))
	h.sendToRoom(newRoom, fmt.Sprintf("[system] %s joined room %s", req.client.name, newRoom), req.client)
}

func (h *hub) handlePrivate(req privateReq) {
	target, ok := h.users[req.toName]
	if !ok {
		h.safeSend(req.from, fmt.Sprintf("[system] user %q not found", req.toName))
		return
	}
	h.safeSend(target, fmt.Sprintf("[pm from %s] %s", req.from.name, req.text))
	h.safeSend(req.from, fmt.Sprintf("[pm to %s] %s", target.name, req.text))
}

func (h *hub) sendToRoom(room string, msg string, skip *client) {
	for c := range h.clients {
		if h.rooms[c] != room {
			continue
		}
		if skip != nil && c == skip {
			continue
		}
		h.safeSend(c, msg)
	}
}

func (h *hub) safeSend(c *client, msg string) {
	select {
	case c.send <- msg:
	default:
		// Slow/broken client: disconnect to protect server.
		delete(h.clients, c)
		delete(h.users, c.name)
		delete(h.rooms, c)
		close(c.send)
		_ = c.conn.Close()
	}
}

func (h *hub) uniqueName(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = fmt.Sprintf("user-%d", time.Now().Unix()%100000)
	}
	if _, exists := h.users[base]; !exists {
		return base
	}
	for i := 1; ; i++ {
		candidate := base + "_" + strconv.Itoa(i)
		if _, exists := h.users[candidate]; !exists {
			return candidate
		}
	}
}

func main() {
	addr := flag.String("addr", ":9000", "TCP address to listen on")
	flag.Parse()

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen error: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Printf("chat server listening on %s\n", *addr)
	fmt.Println("connect with: telnet 127.0.0.1 9000")

	h := newHub()
	go h.run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "accept error: %v\n", err)
			continue
		}
		go handleConn(h, conn)
	}
}

func handleConn(h *hub, conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	_, _ = conn.Write([]byte("Enter your name: "))
	nameRaw, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	name := strings.TrimSpace(nameRaw)
	if name == "" {
		name = fmt.Sprintf("user-%d", time.Now().Unix()%100000)
	}

	c := &client{
		name: name,
		conn: conn,
		send: make(chan string, 32),
	}

	h.register <- c
	defer func() { h.unregister <- c }()
	go writeLoop(c)

	_, _ = conn.Write([]byte("Type messages and press Enter.\n"))
	_, _ = conn.Write([]byte("Use /quit to leave.\n"))

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch {
		case line == "/quit":
			return
		case strings.HasPrefix(line, "/join "):
			room := strings.TrimSpace(strings.TrimPrefix(line, "/join "))
			h.joinRoom <- joinReq{client: c, room: room}
		case strings.HasPrefix(line, "/msg "):
			payload := strings.TrimSpace(strings.TrimPrefix(line, "/msg "))
			parts := strings.SplitN(payload, " ", 2)
			if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
				c.send <- "[system] usage: /msg <user> <text>"
				continue
			}
			h.privateMsg <- privateReq{
				from:   c,
				toName: strings.TrimSpace(parts[0]),
				text:   strings.TrimSpace(parts[1]),
			}
		case line == "/who":
			reply := make(chan []string, 1)
			h.usersList <- usersReq{reply: reply}
			users := <-reply
			c.send <- fmt.Sprintf("[system] users online: %s", strings.Join(users, ", "))
		default:
			h.publicMsg <- messageReq{from: c, text: line}
		}
	}
}

func writeLoop(c *client) {
	for msg := range c.send {
		_, err := c.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
}
