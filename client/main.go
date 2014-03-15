/*
	Client program of the game
*/
package main

import (
	"errors"
	"fmt"
	"go-uno/play"
	"net"
	"os"
)

//Checking err in connecting loop
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port\n", os.Args[0])
		os.Exit(1)
	}
	err := errors.New("")

	//connect to service
	service := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckErr(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	CheckErr(err)

	play.NewGame(conn)
}
