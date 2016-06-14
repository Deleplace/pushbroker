# pushbroker
Message broker for web browsers

The purpose is to let any peer A send a text message to any peer B, where A and B may be web browsers.

This implementation uses an intermediate server between A and B : the "Broker".

The Broker is implemented using bi-directional WebSockets: each peer registers itself, then uses a long-lived WebSocket connection to send and receive messages.

![Broker](/broker-illustration.png)

Currently, the peer chooses its own name, acting as peer ID, before the connection. An equivalent strategy would be to have the server assign generated peer IDs to each incoming peer (but then it would be more difficult for a host to have a consistent ID on subsequent connections).

The Broker implementation in Go uses a global map to keep track of all known peers and their open connections. To prevent data races caused by concurrent messages handling, the global map is guarded by a RWMutex, and each map entry (outgoing peer connection) is guarded by a Mutex.

The communication from peer A to the Broker consists in 2 frames: 1 for the target peer name B and 1 for the message M. A simple refactoring would have the 2 pieces of information serialized in a single frame.

Because the server is small and executable, it is implemented in a single source file having a main func, declared in package main. A more fancy project structure would have a custom package name, and an executable command outside the package.