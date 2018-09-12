package router

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Router interface {
	ConnectAndGetArp(username string, password string) (net.Conn, error)
}

type DlinkRouter struct{}

func (r *DlinkRouter) ConnectAndGetArp(username string, password string) (output string, err error) {
	conn, err := net.Dial("tcp", "192.168.0.1:23")
	defer conn.Close()

	if err != nil {
		return "", err
	}

	_, err = bufio.NewReader(conn).ReadString(':')
	fmt.Fprintf(conn, username+"\n")
	_, err = bufio.NewReader(conn).ReadString(':')
	fmt.Fprintf(conn, password+"\n")
	_, err = bufio.NewReader(conn).ReadString('#')
	fmt.Fprintf(conn, "cat /proc/net/arp\n")
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
