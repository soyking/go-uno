/*
	Server program of the game
*/
package main

import (
	"fmt"
	"net"
	"os"
	"uno/handler"
	"uno/player"
)

var playerlist *player.PlayerList

const PLAYERSLENGTH = 30 //most number of online players

//Checking err in connecting loop
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(0)
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port\n", os.Args[0])
		os.Exit(1)
	}

	//playerlist to record online players
	playerlist = player.New(PLAYERSLENGTH)
	fmt.Printf("**Create playerlist with %d\n", PLAYERSLENGTH)

	//connect and listen
	service := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckErr(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	CheckErr(err)
	fmt.Printf("**Listen on host:port: %s\n", service)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handler.NewPlayer(conn, playerlist)
	}
}
