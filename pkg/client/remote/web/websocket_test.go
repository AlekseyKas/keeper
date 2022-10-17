package web

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewStream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(handlerToBeTested))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("cannot make websocket connection: %v", err)
	}
	err = conn.WriteMessage(websocket.BinaryMessage, []byte("world"))
	if err != nil {
		log.Fatalf("cannot write message: %v", err)
	}
	_, p, err := conn.ReadMessage()
	if err != nil {
		log.Fatalf("cannot read message: %v", err)
	}
	fmt.Printf("success: received response: %q\n", p)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handlerToBeTested(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
	}
	mt, p, err := conn.ReadMessage()
	if err != nil {
		log.Printf("cannot read message: %v", err)
		return
	}
	conn.WriteMessage(mt, []byte("hello "+string(p)))
}
