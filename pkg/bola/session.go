package bola

// Session is a connection to a remote peer
type Session struct {
}
type Sessions []Session

// Dispatch manages the Session pool to minimise spin-up time initialisation
type Dispatch struct {
}

// Listener is a single thread that pulls packets off the network interface and
// sends them to the Session they relate to
type Listener struct {
}
