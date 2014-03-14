/*
	Client program of the game
*/
package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"uno/card"
	"uno/connect"
)

//Operation when a game is over
func EndGame(conn net.Conn, transinfo *connect.TransInfo, err string, op int) {
	fmt.Println("Information from server: ", err)
	transinfo.CardInfo.Err = 1
	if op == 0 { //something wrong in my client not my opponent's client
		_ = transinfo.Send(conn, err, 0)
	}
}

//Checking err in connecting loop
func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s\n", err.Error())
		os.Exit(1)
	}
}

//Select a card you want to put out
func SelectCard(cardinfo *card.GameState) {
	r := bufio.NewReader(os.Stdin)

	for {
		cardinfo.Show()
		selectnum, err := ReadLine(r)
		if err != nil {
			fmt.Println("Input a number.")
			continue
		}

		cardinfo.NowCard = selectnum - 1 //key [0..n-1]
		if cardinfo.Correct() == 0 {
			fmt.Println("You selected a wrong card.")
			continue
		}

		cardinfo.Uno = 0            //assume you don't shout uno
		if cardinfo.NowCard != -1 { //you don't give up
			fmt.Printf("The card you select: color:%s ", cardinfo.Cards[cardinfo.NowCard].Color)
			card.ShowNum(cardinfo.Cards[cardinfo.NowCard].Number)

			if cardinfo.Cards[cardinfo.NowCard].Number >= 25 { //Wild or Wild Draw 4
				if SelectColor(cardinfo) == 0 { //withdrawn this operation
					continue
				}
			}

			fmt.Println("Shouting UNO now?(yes or no)") //whether uno
			fmt.Printf("choose >>")
			rawLine, _, _ := r.ReadLine()
			line := string(rawLine)
			if line == "yes" {
				cardinfo.Uno = 1
			}
		}
		break
	}
}

//Select Color when you put out Wild or Wild Draw 4
func SelectColor(cardinfo *card.GameState) int {
	colors := map[int]string{0: "red", 1: "yellow", 2: "blue", 3: "green"}
	r := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("Choose one color(with number) or input q to reselect a card")
		fmt.Println("1: red; 2: yellow; 3: blue; 4: green")
		fmt.Printf("choose >>")

		rawLine, _, _ := r.ReadLine()
		line := string(rawLine)
		if line == "q" {
			return 0
		}

		selectnum, err := strconv.Atoi(line)
		if err != nil {
			fmt.Println("Input a number.")
			continue
		}

		if selectnum < 1 || selectnum > 4 {
			fmt.Println("Choose a number between 1 to 4.")
			continue
		}

		cardinfo.ChangeColor = colors[selectnum-1] //change to new color
		return 1
	}
}

//Read what player input, and transform to number
func ReadLine(r *bufio.Reader) (int, error) {
	rawLine, _, _ := r.ReadLine()
	line := string(rawLine)
	selectnum, err := strconv.Atoi(line)
	return selectnum, err
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

	//create connection with server
	transinfo := connect.New()
JLoop:
	for {
		transinfolist, err := connect.Receive(conn)
		if err != nil {
			break
		}

		//read transinfo in the list and handle
		for _, v := range transinfolist {
			transinfo = &v
			if transinfo.CardInfo.Err == 1 {
				EndGame(conn, transinfo, transinfo.InfoString, 1)
				break JLoop
			}

			if transinfo.State == 0 { //print information
				fmt.Println("Information from server: " + transinfo.InfoString)
			} else if transinfo.State == 1 { //put out one card
				SelectCard(&transinfo.CardInfo)
				err = transinfo.Send(conn, "Put out one card.", 2)
				if err != nil {
					EndGame(conn, transinfo, err.Error(), 0)
					break JLoop
				}
			}
		}
	}
}
