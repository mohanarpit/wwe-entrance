package router

import (
	"net"
	"time"
)

// Router interface must be implemented for all Router implementations
// This is because some routers connect over Telnet while others over SSH
// The commands for ARP Output may also vary from Router to Router.
// A default implementation for a DLink router is given below
type Router interface {
	Connect(username string, password string, delay time.Duration) (*net.Conn, error)
	GetArpOutput(conn net.Conn) (output string, err error)
}
