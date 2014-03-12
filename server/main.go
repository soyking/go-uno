package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"uno/card"
	"uno/connect"
	"uno/player"
)

var playerlist *player.PlayerList

const PLAYERSNUM = 30 //most number of online players

func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(0)
	}
}

func EndPlayer(conn net.Conn, transinfo *connect.TransInfo, err error) {
	fmt.Println(err)
	_ = transinfo.Send(conn, err.Error(), 0)
	conn.Close()
}

func NewPlayer(conn net.Conn) {
	//create a package for connect with client
	transinfo := connect.New()
	err := errors.New("")
	defer EndPlayer(conn, transinfo, err)

	err = transinfo.Send(conn, "Connect successful.", 0)
	if err != nil {
		return
	}

	//register a new player
	key, newplayer, err := playerlist.New()
	if err != nil {
		transinfo.Send(conn, err.Error(), 0)
	}
	fmt.Printf("**New player ,key:%d\n", key)
	defer playerlist.Recover(key)

	opkey, err := playerlist.FreePlayer(key)
	pre := 1
	if err != nil {
		transinfo.Send(conn, err.Error(), 0)
		fmt.Printf("player key:%d is waiting \n", key)
		opkey = <-newplayer.AnyPlayer
		pre = 0
		fmt.Printf("player key:%d stop waiting ,opkey:%d\n", key, opkey)
	}

	info := "connect to user key:" + string(opkey)
	transinfo.Send(conn, info, 0)
	fmt.Println("new game between " + string(key) + " and " + string(opkey))

	cardinfo := card.New()
	cardinfo.GetCard(7)
	if pre == 1 {
		transinfo.CardInfo = *cardinfo
		err = PutOutCard(transinfo, conn)
		if err != nil {
			transinfo.CardInfo.Err = 1
			playerlist.Players[opkey].LastRound <- transinfo.CardInfo
			return
		}
		_ = transinfo.CardInfo.UpdateState()
		if err != nil {
			transinfo.CardInfo.Err = 1
			playerlist.Players[opkey].LastRound <- transinfo.CardInfo
			return
		}
		playerlist.Players[opkey].LastRound <- transinfo.CardInfo
	}

	for {
		opcardinfo := <-playerlist.Players[key].LastRound
		if opcardinfo.Err == 1 {
			err = errors.New("There is something wrong from our opponent.")
			return
		}
		if opcardinfo.CardsNum == 0 {
			err = errors.New("You lose.T T")
			return
		}
		if opcardinfo.Skip == 1 { //skip this round
			transinfo.CardInfo.Reset(opcardinfo)
			transinfo.CardInfo.Uno = 0
			transinfo.CardInfo.Skip = 0
			transinfo.Send(conn, "You shoud Skip this round.", 0)
		} else {
			transinfo.CardInfo.Reset(opcardinfo)
			err = PutOutCard(transinfo, conn)
			if err != nil {
				transinfo.CardInfo.Err = 1
				playerlist.Players[opkey].LastRound <- transinfo.CardInfo
				return
			}
			fine := transinfo.CardInfo.UpdateState()
			if fine > 0 {
				finestr := "You have to get " + string(fine) + " cards."
				transinfo.Send(conn, finestr, 0)
			} else if fine < 0 {
				err = errors.New("You win!")
				return
			}
		}
		playerlist.Players[opkey].LastRound <- transinfo.CardInfo
	}
}

func PutOutCard(transinfo *connect.TransInfo, conn net.Conn) error {
	times := 0
	err := errors.New("")
	for times < 10 {
		err = transinfo.Send(conn, "put out one card", 1)
		if err != nil {
			return err
		}
		err = transinfo.Receive(conn)
		if err != nil {
			return err
		}
		if transinfo.CardInfo.CheckCard() == false {
			times++
		} else {
			return nil
		}
	}
	err = errors.New("Put out the wrong card for too many times.")
	return err
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port\n", os.Args[0])
		os.Exit(1)
	}

	//playerlist to record online players
	playerlist = player.New(PLAYERSNUM)
	fmt.Printf("**Create playerlist with %d\n", PLAYERSNUM)

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
		go NewPlayer(conn)
	}
}
