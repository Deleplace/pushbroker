package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"golang.org/x/net/websocket"
)

func main() {
	http.Handle("/enter", websocket.Handler(EnterServer))
	const addr = ":12345"
	fmt.Println("Listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}

// GuardedConn is a websocket Conn, guarded by a Mutex
type GuardedConn struct {
	sync.Mutex
	*websocket.Conn
}

// Peers is a set of currently connected peers.
// Key is the peer name.
// Value is the open websocket connection to the peer.
// Value type is a pointer, because the value Mutex inside GuardedConn must not be copied.
type Peers map[string]*GuardedConn

// GuardedPeers is a set of currently connected peers, guarded by a Mutex
type GuardedPeers struct {
	sync.RWMutex
	Peers
}

func (peers *GuardedPeers) Add(A string, ws *websocket.Conn) {
	peers.Lock()
	defer peers.Unlock()
	gws := GuardedConn{
		sync.Mutex{},
		ws,
	}
	peers.Peers[A] = &gws
	fmt.Println(A + " entered the party")
}

func (peers *GuardedPeers) Get(B string) (*GuardedConn, bool) {
	peers.RLock()
	defer peers.RUnlock()
	gws, exists := peers.Peers[B]
	return gws, exists
}

func (peers *GuardedPeers) Remove(A string) {
	peers.Lock()
	defer peers.Unlock()
	fmt.Println(A + " has left the party")
	delete(peers.Peers, A)
	// We don't need to lock the *GuardedConn here. Just removing its ref from the map.
}

// peers is the only instance of Peers for this server.
var peers = GuardedPeers{
	sync.RWMutex{},
	Peers{},
}

// EnterServer is the handler for the total lifetime of
// the persistent WebSocket connection of a peer.
func EnterServer(wsA *websocket.Conn) {
	addr := wsA.RemoteAddr().String()
	fmt.Println("New connection from remote host " + addr)

	var A string
	err := websocket.Message.Receive(wsA, &A)
	if err != nil {
		processError(wsA, "Error reading new peer name: "+err.Error())
		return
	}
	if _, existsA := peers.Get(A); existsA {
		processError(wsA, "Peer name "+A+" already taken")
		return
	}

	peers.Add(A, wsA)
	defer peers.Remove(A)

	// A now offically exists! (in the global map)
	err = websocket.Message.Send(wsA, "OK")
	if err != nil {
		processError(wsA, "Error acknowledging new peer "+A+": "+err.Error())
	}

	// The infinite loop waits for A to send messages to other peers.
	var B string
	for {
		err := websocket.Message.Receive(wsA, &B)
		if err == io.EOF {
			// A has quit. No Problem.
			return
		}
		if err != nil {
			processError(wsA, "Error reading destination peer name: "+err.Error())
			return
		}

		var M string
		err = websocket.Message.Receive(wsA, &M)
		if err != nil {
			processError(wsA, "Error reading message M to deliver: "+err.Error())
			return
		}

		wsB, existsB := peers.Get(B)
		if !existsB {
			fmt.Fprintln(os.Stderr, "Message from source peer "+A+" can't be delivered because target peer "+B+" is not currently known")
			// The target peer is not known, so the message won't be delivered.
			// This is not an exceptional case, it doesn't stop the server nor the peer A.
			// The source peer A is not notified (per requirement).
			continue
		}

		// There is a small TOCTOU here, but inherently it is always
		// possible that B leaves surreptitiously when we try to talk
		// to her through the network.

		wsB.Lock()
		fmt.Println("Delivering message from " + A + " to " + B + ": " + M)
		// Note that the target peer B won't notified of the source peer name A.
		// A may decide, if it wishes, to include its name inside message M.
		err = websocket.Message.Send(wsB.Conn, M)
		if err != nil {
			processError(wsA, "Error sending message ["+M+"] to "+B+": "+err.Error())
		}
		wsB.Unlock()
	}
}

// processError, in case of unexpected error, prints error on
// server stderr, sends the error message to origin websocket,
// and closes origin websocket.
func processError(ws *websocket.Conn, errmsg string) {
	fmt.Fprintln(os.Stderr, errmsg)
	fmt.Fprintln(ws, errmsg)
	ws.Close()
}
