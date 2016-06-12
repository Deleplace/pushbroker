package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"golang.org/x/net/websocket"
)

type PeerID [8]byte

func (id PeerID) String() string {
	return string(id[:])
}

type GuardedConn struct {
	sync.Mutex
	*websocket.Conn
}

type Peers map[PeerID]GuardedConn

type GuardedPeers struct {
	sync.RWMutex
	Peers
}

func (peers *GuardedPeers) Add(A PeerID, ws *websocket.Conn) {
	peers.Lock()
	defer peers.Unlock()
	gws := GuardedConn{
		sync.Mutex{},
		ws,
	}
	peers.Peers[A] = gws
}

func (peers *GuardedPeers) Get(B PeerID) (GuardedConn, bool) {
	peers.Lock()
	defer peers.Unlock()
	gws, exists := peers.Peers[B]
	return gws, exists
}

var peers = GuardedPeers{
	sync.RWMutex{},
	Peers{},
}

func EnterServer(wsA *websocket.Conn) {
	addr := wsA.RemoteAddr().String()

	var A PeerID
	// Read exactly 8 bytes: the new participant's name
	n, err := wsA.Read(A[:])
	if err != nil && err != io.EOF {
		processError(wsA, "Error reading new peer name: "+err.Error())
		return
	}
	if n < len(A) {
		processError(wsA, "Invalid name: "+A.String())
		return
	}

	if _, existsA := peers.Get(A); existsA {
		processError(wsA, "Peer name "+A.String()+" already taken")
		return
	}

	fmt.Println(A.String() + " entered the party (from " + addr + ")")
	peers.Add(A, wsA)

	var B PeerID
	for {
		n, err := wsA.Read(B[:])
		if err != nil && err != io.EOF {
			processError(wsA, "Error reading destination peer name: "+err.Error())
			return
		}
		if n < len(B) {
			processError(wsA, "Invalid name for B: "+B.String())
			return
		}

		var buf bytes.Buffer
		_, err = buf.ReadFrom(wsA)
		if err != nil && err != io.EOF {
			processError(wsA, "Error reading message M to deliver: "+err.Error())
			return
		}
		M := buf.String()

		wsB, existsB := peers.Get(B)
		if !existsB {
			processError(wsA, "Target peer "+B.String()+" not currently connected")
			return
		}
		fmt.Println(A.String() + " entered the party (from " + addr + ")")
		fmt.Fprintln(wsB, M)
	}
}

func processError(ws *websocket.Conn, errmsg string) {
	fmt.Fprintln(os.Stderr, errmsg)
	fmt.Fprintln(ws, errmsg)
	ws.Close()
}

func main() {
	http.Handle("/enter", websocket.Handler(EnterServer))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic(err)
	}
}
