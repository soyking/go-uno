/*
	Handle the connection from tcp socket
*/
package handler

import (
	"errors"
	"fmt"
	"go-uno/card"
	"go-uno/connect"
	"go-uno/player"
	"net"
	"strconv"
)

const INITCARD = 7 //the number of cards that you should get first

//Operation when a player leaves
func EndPlayer(conn net.Conn, transinfo *connect.TransInfo, err *error) {
	fmt.Println("Information:", (*err).Error())
	transinfo.CardInfo.Err = 1 //there is something wrong, stop connecting
	_ = transinfo.Send(conn, (*err).Error(), 0)
	conn.Close()
}

//Adding new player and deal with routines
func NewPlayer(conn net.Conn, playerlist *player.PlayerList) {
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

	transinfo.CardInfo.GetCard(INITCARD)

	//first card
	if pre == 1 {
		//receive a card the player put out
		err = Request(playerlist, conn, transinfo, key, opkey)
		if err != nil {
			return
		}

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
			transinfo.CardInfo.Reset(opcardinfo)
			err = Request(playerlist, conn, transinfo, key, opkey)
			if err != nil {
				return
			}
		}

		playerlist.Players[opkey].LastRound <- transinfo.CardInfo //change turn
	}
}

//Send request to client and receive the card the player puts out
func Request(playerlist *player.PlayerList, conn net.Conn, transinfo *connect.TransInfo, key int, opkey int) error {
	err := transinfo.PutOutCard(conn)
	if err != nil {
		transinfo.CardInfo.Err = 1
		playerlist.Players[opkey].LastRound <- transinfo.CardInfo
		return err
	}

	fine := transinfo.CardInfo.UpdateState() //number of cards have to get
	if fine > 0 {
		finestr := "You have to get " + strconv.Itoa(fine) + " cards."
		transinfo.Send(conn, finestr, 0)
	} else if fine < 0 {
		playerlist.Players[opkey].LastRound <- transinfo.CardInfo
		err = errors.New("You win!") // game over
		return err
	}

	fmt.Printf("**key %d: put out one card: color:%s ", key, transinfo.CardInfo.LastCard.Color)
	card.ShowNum(transinfo.CardInfo.LastCard.Number)

	return nil
}
