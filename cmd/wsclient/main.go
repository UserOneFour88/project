package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type inbound struct {
	Type    string `json:"type"`
	Room    string `json:"room,omitempty"`
	Name    string `json:"name,omitempty"`
	Token   string `json:"token,omitempty"`
	Text    string `json:"text,omitempty"`
	Message string `json:"message,omitempty"`
}

type outbound struct {
	Type  string `json:"type"`
	Room  string `json:"room,omitempty"`
	Name  string `json:"name,omitempty"`
	Token string `json:"token,omitempty"`
	Text  string `json:"text,omitempty"`
}

func main() {
	url := flag.String("url", "ws://localhost:8081/ws", "websocket url")
	room := flag.String("room", "lobby", "room name")
	name := flag.String("name", "", "user name (used when requesting a token)")
	token := flag.String("token", "", "auth token (optional, for identification)")
	flag.Parse()

	conn, _, err := websocket.DefaultDialer.Dial(*url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	join := outbound{
		Type:  "join",
		Room:  *room,
		Name:  strings.TrimSpace(*name),
		Token: strings.TrimSpace(*token),
	}
	if err := conn.WriteJSON(join); err != nil {
		fmt.Fprintf(os.Stderr, "join send error: %v\n", err)
		os.Exit(1)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var in inbound
			if err := conn.ReadJSON(&in); err != nil {
				return
			}
			switch in.Type {
			case "token":
				fmt.Printf("[token] name=%s room=%s token=%s\n", in.Name, in.Room, in.Token)
			case "msg":
				fmt.Printf("[%s@%s] %s\n", in.Name, in.Room, in.Text)
			case "error":
				fmt.Printf("[error] %s\n", in.Message)
			default:
				b, _ := json.Marshal(in)
				fmt.Printf("[recv] %s\n", string(b))
			}
		}
	}()

	fmt.Println("Type messages, Ctrl+C to exit.")
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if err := conn.WriteJSON(outbound{Type: "msg", Text: line}); err != nil {
			fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			break
		}
	}

	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(time.Second))
	<-done
}
