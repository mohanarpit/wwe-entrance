package router

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type DlinkRouter struct {
	ConnectionType string
	Command        string
}

func (r *DlinkRouter) Connect(username, password, host string, delay time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(r.ConnectionType, host, delay)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to router %v", err)
	}
	_, err = bufio.NewReader(conn).ReadString(':')
	fmt.Fprintf(conn, username+"\n")
	_, err = bufio.NewReader(conn).ReadString(':')
	fmt.Fprintf(conn, password+"\n")
	_, err = bufio.NewReader(conn).ReadString('#')

	return conn, err
}

func (r *DlinkRouter) GetArpOutput(conn net.Conn) (output string, err error) {

	fmt.Fprintf(conn, r.Command+"\n")
	_, err = bufio.NewReader(conn).ReadString('\n')

	var buf = make([]byte, 4096)

	_, err = bufio.NewReader(conn).Read(buf)
	if err != nil {
		log.Fatalln(err)
		return "", err
	}
	log.Println(string(buf))
	return string(buf), nil
}
