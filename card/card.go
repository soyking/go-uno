package card

import (
	"fmt"
	"math/rand"
)

type CardInfo struct {
	Color  string
	Number int
}

type GameState struct {
	Uno         int
	LastCard    CardInfo
	NowCard     int
	Penalty     int
	CardsNum    int
	Cards       []CardInfo
	Err         int //error from your opponent
	Skip        int //you should skip this round
	ChangeColor string
}

func (gs *GameState) GetCard(num int) {
	colors := map[int]string{0: "red", 1: "yellow", 2: "blue", 3: "green"}
	for i := 0; i < num; i++ {
		randcard := rand.Intn(108)
		color := randcard % 4
		number := randcard / 4
		tempcard := CardInfo{Color: colors[color], Number: number}
		gs.Cards = append(gs.Cards, tempcard)
		gs.CardsNum++
	}
}

func (gs *GameState) CheckCard() bool {
	nowcardinfo := gs.Cards[gs.NowCard]
	if gs.LastCard.Color == "first" { //first card
		return true
	}

	if nowcardinfo.Color == "giveup" { //player give up this round
		return true
	}

	if nowcardinfo.Number == 26 { //wild or wild+4
		return true
	}

	if gs.LastCard.Number != 26 && (gs.LastCard.Color == nowcardinfo.Color || gs.LastCard.Number == nowcardinfo.Number || gs.LastCard.Number == 25) {
		return true
	}

	return false
}

func (gs *GameState) UpdateState() int {
	nowcardinfo := gs.Cards[gs.NowCard]
	fine := 0
	if nowcardinfo.Color == "giveup" {
		if gs.Penalty != 0 {
			fine = gs.Penalty
		} else {
			fine = 2
		}
		gs.Penalty = 0
	} else {
		num := nowcardinfo.Number
		switch num {
		case 19, 20:
			gs.Skip = 1
		case 23, 24:
			gs.Penalty += 2
		case 25:
			gs.LastCard.Color = gs.ChangeColor
		case 26:
			gs.Penalty += 4
			gs.LastCard.Color = gs.ChangeColor
		}
		gs.LastCard.Number = nowcardinfo.Number
		if (gs.Uno == 1 && gs.CardsNum != 1) || (gs.Uno != 1 && gs.CardsNum == 1) {
			gs.Uno = 0
			fine = 2
		}
	}
	gs.DeleteCard()
	if gs.CardsNum == 0 { //win
		fine = -1
	}
	if fine > 0 {
		gs.GetCard(fine)
	}
	return fine
}

func (gs *GameState) Reset(opstate GameState) {
	gs.Uno = opstate.Uno
	gs.LastCard = opstate.LastCard
	gs.Penalty = opstate.Penalty
	gs.Skip = opstate.Skip
}

func (gs *GameState) DeleteCard() {
	if gs.CardsNum == 1 {
		gs.Cards = make([]CardInfo, 0)
	} else if gs.NowCard == 0 {
		gs.Cards = gs.Cards[1:]
	} else if gs.NowCard == gs.CardsNum {
		gs.Cards = gs.Cards[0 : gs.CardsNum-1]
	} else {
		gs.Cards = append(gs.Cards[:gs.NowCard], gs.Cards[gs.NowCard+1:gs.CardsNum]...)
	}
	gs.CardsNum--
}

func (gs *GameState) Correct() int {
	if gs.NowCard < 0 || gs.NowCard >= gs.CardsNum {
		return 0
	}
	return 1
}

func (gs *GameState) Show() {
	fmt.Println("--------CardInfo--------")
	if gs.Uno == 1 {
		fmt.Println("Your opponent shouts UNO!")
	}
	fmt.Printf("You has/have %d cards:\n", gs.CardsNum)
	for i := 0; i < gs.CardsNum; i++ {
		fmt.Printf("##%d: %s  ", i+1, gs.Cards[i].Color)
		num := gs.Cards[i].Number
		switch {
		case num <= 9:
			fmt.Println(string(num))
		case num <= 18:
			fmt.Println(string(num - 9))
		case num <= 20:
			fmt.Println("Skip")
		case num <= 22:
			fmt.Println("Reverse")
		case num <= 24:
			fmt.Println("Draw Two(+2)")
		case num == 25:
			fmt.Println("Wild(change color)")
		case num == 26:
			fmt.Println("Wild Draw 4(change color and +4)")
		}
	}
	fmt.Println("Select a number of a card.")
	fmt.Println("------------------------")
	fmt.Printf("choose >>")
}

func New() *GameState {
	lastcard := CardInfo{Color: "first"}
	return &GameState{Uno: 0, CardsNum: 0, Penalty: 0, LastCard: lastcard, Err: 0, Skip: 0}
}
