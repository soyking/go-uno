/*
	Server program of the game
*/
package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"uno/card"
	"uno/connect"
	"uno/player"
)

var playerlist *player.PlayerList

const PLAYERSLENGTH = 30 //most number of online players
const INITCARD = 7       //the number of cards that you should get first

//Checking err in connecting loop
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		os.Exit(0)
	}
}

//Operation when a player leaves
func EndPlayer(conn net.Conn, transinfo *connect.TransInfo, err *error) {
	fmt.Println("Information:", (*err).Error())
	transinfo.CardInfo.Err = 1 //there is something wrong, stop connecting
	_ = transinfo.Send(conn, (*err).Error(), 0)
	conn.Close()
}

//Adding new player and deal with routines
func NewPlayer(conn net.Conn) {
	//connection with client
	transinfo := connect.New()
	err := errors.New("")
	defer EndPlayer(conn, transinfo, &err)

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

	//finding opponent
	opkey, err := playerlist.FreePlayer(key)
	pre := 1 //put out one card first
	if err != nil {
		transinfo.Send(conn, err.Error(), 0)
		transinfo.CardInfo.Err = 0    //not the first round
		opkey = <-newplayer.AnyPlayer //waiting someone to join in
		pre = 0
	}

	info := "Playing game now!"
	transinfo.Send(conn, info, 0)
	fmt.Println("**key " + strconv.Itoa(key) + ": new game with " + strconv.Itoa(opkey) + " whether pre:" + strconv.Itoa(pre))

	//recording game state, get cards firts
	// cardinfo := card.New()
	// cardinfo.GetCard(INITCARD)
	// transinfo.CardInfo = *cardinfo
	transinfo.CardInfo.GetCard(INITCARD)

	//first card
	if pre == 1 {
		//receive a card the player put out
		err = transinfo.PutOutCard(conn)
		if err != nil {
			transinfo.CardInfo.Err = 1
			playerlist.Players[opkey].LastRound <- transinfo.CardInfo
			return
		}
		fine := transinfo.CardInfo.UpdateState() //number of cards have to get
		if fine > 0 {
			finestr := "You have to get " + strconv.Itoa(fine) + " cards."
			transinfo.Send(conn, finestr, 0)
		}

		fmt.Printf("**key %d: put out one card: color:%s ", key, transinfo.CardInfo.LastCard.Color)
		card.ShowNum(transinfo.CardInfo.LastCard.Number)

		playerlist.Players[opkey].LastRound <- transinfo.CardInfo //change turn
	}

	//game loop
	for {
		opcardinfo := <-playerlist.Players[key].LastRound

		if opcardinfo.Err == 1 {
			err = errors.New("There is something wrong from your opponent.")
			return
		}

		//game over
		if opcardinfo.CardsNum == 0 {
			err = errors.New("You lose.T T")
			return
		}

		if opcardinfo.Skip == 1 { //skip this round
			transinfo.CardInfo.Reset(opcardinfo)
			transinfo.CardInfo.Uno = 0
			transinfo.CardInfo.Skip = 0
			transinfo.Send(conn, "You shoud skip this round.", 0)
		} else {
			transinfo.CardInfo.Reset(opcardinfo) //update state
			err = transinfo.PutOutCard(conn)
			if err != nil {
				transinfo.CardInfo.Err = 1
				playerlist.Players[opkey].LastRound <- transinfo.CardInfo
				return
			}

			fine := transinfo.CardInfo.UpdateState() //number of cards have to get
			fmt.Printf("**key %d: fine: %d\n", key, fine)
			if fine > 0 {
				finestr := "You have to get " + strconv.Itoa(fine) + " cards."
				transinfo.Send(conn, finestr, 0)
			} else if fine < 0 {
				playerlist.Players[opkey].LastRound <- transinfo.CardInfo
				err = errors.New("You win!") // game over
				return
			}
			fmt.Printf("**key %d: put out one card: color:%s ", key, transinfo.CardInfo.LastCard.Color)
			card.ShowNum(transinfo.CardInfo.LastCard.Number)
		}

		playerlist.Players[opkey].LastRound <- transinfo.CardInfo //change turn
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
		go NewPlayer(conn)
	}
}
