package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"uno/card"
	"uno/connect"
)

func CheckErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s\n", err.Error())
		os.Exit(1)
	}
}

func EndGame(conn net.Conn, transinfo *connect.TransInfo, err error) {
	transinfo.CardInfo.Err = 1
	transinfo.Send(conn, err.Error(), 0)
	fmt.Println("error:", err)
	conn.Close()
}

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
		if cardinfo.Cards[cardinfo.NowCard].Number >= 25 {
			if SelectColor(cardinfo) == 0 { //withdrawn this operation
				continue
			}
		}
		fmt.Println("Shouting UNO now?(yes or no)")
		fmt.Println("choose >>")
		rawLine, _, _ := r.ReadLine()
		line := string(rawLine)
		if line == "yes" {
			cardinfo.Uno = 1
		}
		break
	}
}

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
		cardinfo.ChangeColor = colors[selectnum-1]
		return 1
	}
}

func ReadLine(r *bufio.Reader) (selectnum int, err error) {
	rawLine, _, _ := r.ReadLine()
	line := string(rawLine)
	selectnum, err = strconv.Atoi(line)
	return
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage:%s host:port\n", os.Args[0])
		os.Exit(1)
	}
	service := os.Args[1]
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	CheckErr(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	CheckErr(err)

	transinfo := connect.New()
	defer EndGame(conn, transinfo, err)

	for {
		err = transinfo.Receive(conn)
		if err != nil {
			return
		}
		if transinfo.State == 0 {
			fmt.Println(transinfo.InfoString)
		} else if transinfo.State == 1 {
			SelectCard(&transinfo.CardInfo)
			transinfo.Send(conn, "Put out one card.", 2)
		}
	}
}
