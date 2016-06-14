package main

import (
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

// This test uses the global infrastructure of broker.go :
// the maps of peers.
//
// It starts its own server and its own clients.

func TestBrokerMessages(t *testing.T) {
	serve := func() {
		http.Handle("/enter", websocket.Handler(EnterServer))
		err := http.ListenAndServe(":12345", nil)
		if err != nil {
			panic(err)
		}
	}
	go serve()
	// Giving time for the server to init (is there a better way??)
	time.Sleep(time.Second)

	url := "ws://localhost:12345/enter"
	origin := "http://localhost/"

	checkerr := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	connect := func(A string) *websocket.Conn {
		wsA, err := websocket.Dial(url, "", origin)
		checkerr(err)
		err = websocket.Message.Send(wsA, A)
		checkerr(err)
		return wsA
	}

	send := func(wsA *websocket.Conn, B, M string) {
		err := websocket.Message.Send(wsA, B)
		checkerr(err)
		err = websocket.Message.Send(wsA, M)
		checkerr(err)
	}

	receive := func(wsB *websocket.Conn) string {
		var inbox string
		err := websocket.Message.Receive(wsB, &inbox)
		checkerr(err)
		return inbox
	}

	// Alice sends "Hello Bob" to Bob.
	// Bob sends "Hi Carol" to Carol.
	// Carol sends "I'm fabulous" to herself.
	// Carol sends "Who are you?" to unknown peer Malcolm.
	alice := connect("Alice")
	bob := connect("Bob")
	carol := connect("Carol")
	send(alice, "Bob", "Hello Bob")
	send(bob, "Carol", "Hi Carol")
	send(carol, "Carol", "I'm fabulous")
	send(carol, "Malcolm", "Who are you?")

	// Bob receives "Hello Bob".
	got := receive(bob)
	want := "Hello Bob"
	if got != want {
		t.Errorf("receive(bob) == %q, want %q", got, want)
	}

	// Carol receives "I'm fabulous" and "Hi Carol", in any order.
	got1, got2 := receive(carol), receive(carol)
	switch {
	case got1 == "I'm fabulous" && got2 == "Hi Carol":
		// OK
	case got1 == "Hi Carol" && got2 == "I'm fabulous":
		// OK
	default:
		t.Errorf("unexpected carold receives %q, %q", got1, got2)
	}
}
