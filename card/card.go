/*
	Information and operation about card
*/
package card

import (
	"fmt"
	"math/rand"
	"strconv"
)

//card information including color and number
//
//there are 108 cards in the game
//
//color:
//		  randnum%4,Wild's color and Wild Draw 4's color don't mater
//		  red
//		  yellow
//		  blue
//		  green
//number:
//		  randnum/4
//		  0     --->0
//		  1..9  --->1-9
//		  10..19--->1-9
//		  19..20--->Skip
//		  21..22--->Reverse
//		  23..24--->Draw Two(+2)
//		  25    --->Wild(change color)
//		  26    --->Wild Draw 4(change color and +4)
type CardInfo struct {
	Color  string
	Number int
}

//Record the state of the game
type GameState struct {
	Uno         int      //whether shouting uno
	LastCard    CardInfo //what your opponent put out
	NowCard     int      //what you put out
	Penalty     int      //the card you should get if you give up
	CardsNum    int
	Cards       []CardInfo
	Err         int    //error from your opponent,-1 when first round
	Skip        int    //you should skip this round
	ChangeColor string //when put out Wild or Wild Draw 4 to change color
}

//Create a new state of the game
func New() *GameState {
	return &GameState{Uno: 0, CardsNum: 0, Penalty: 0, Err: -1, Skip: 0}
}

//Get card when starting the game or be punished
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

//Check whether the card is right
func (gs *GameState) CheckCard() bool {
	if gs.NowCard == -1 { //give up
		return true
	}

	if gs.Err == -1 { //first round
		gs.Err = 0
		return true
	}

	nowcardinfo := gs.Cards[gs.NowCard]

	if nowcardinfo.Number == 26 { //Wild Draw 4
		return true
	}

	if gs.LastCard.Number != 26 {
		if nowcardinfo.Number == 25 { //Wild
			return true
		}

		if gs.LastCard.Color == nowcardinfo.Color || gs.LastCard.Number == nowcardinfo.Number { //either color or number is right
			return true
		}

		if gs.LastCard.Number >= 1 && gs.LastCard.Number <= 18 { //number card
			if gs.LastCard.Number+9 == nowcardinfo.Number || gs.LastCard.Number-9 == nowcardinfo.Number {
				return true
			}
		}

		if gs.LastCard.Number >= 19 && gs.LastCard.Number <= 20 { //Skip card
			if gs.LastCard.Number+1 == nowcardinfo.Number || gs.LastCard.Number-1 == nowcardinfo.Number {
				return true
			}
		}

		if gs.LastCard.Number >= 21 && gs.LastCard.Number <= 22 { //Reverse card
			if gs.LastCard.Number+1 == nowcardinfo.Number || gs.LastCard.Number-1 == nowcardinfo.Number {
				return true
			}
		}

		if gs.LastCard.Number >= 23 && gs.LastCard.Number <= 24 { //Draw 2 card
			if gs.LastCard.Number+1 == nowcardinfo.Number || gs.LastCard.Number-1 == nowcardinfo.Number {
				return true
			}
		}
	}

	return false
}

//Update the state after putting out one card
func (gs *GameState) UpdateState() int {
	var fine int = 0 //the number of the card you should get

	if gs.NowCard == -1 { //give up
		if gs.Penalty != 0 {
			fine = gs.Penalty
		} else {
			fine = 2
		}

		if gs.LastCard.Number == 26 { //after your opponent put out Wild Draw 4, it should be replaced the card with Wild, or you will have to put out Wild Draw 4 forever
			gs.LastCard.Number = 25
		}

		gs.Penalty = 0
	} else {
		nowcardinfo := gs.Cards[gs.NowCard]

		gs.LastCard.Color = nowcardinfo.Color
		gs.LastCard.Number = nowcardinfo.Number

		num := nowcardinfo.Number
		switch num {
		case 19, 20: //Skip card
			gs.Skip = 1
		case 23, 24: //Draw 2 card
			gs.Penalty += 2
		case 25: //Wild card
			gs.LastCard.Color = gs.ChangeColor
		case 26: //Wild Draw 4 card
			gs.Penalty += 4
			gs.LastCard.Color = gs.ChangeColor
		}

		gs.DeleteCard() //delete the card you put out

		if (gs.Uno == 1 && gs.CardsNum != 1) || (gs.Uno != 1 && gs.CardsNum == 1) {
			gs.Uno = 0
			fine = 2
		}
	}

	if gs.CardsNum == 0 { //you win
		fine = -1
	}

	if fine > 0 {
		gs.GetCard(fine)
	}

	return fine
}

//Get information after your opponent put out one card
func (gs *GameState) Reset(opstate GameState) {
	gs.Uno = opstate.Uno
	gs.LastCard = opstate.LastCard
	gs.Penalty = opstate.Penalty
	gs.Skip = opstate.Skip
}

//Delete a card you put out
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

//Whether player choose the right number of the card in client
func (gs *GameState) Correct() int {
	if gs.NowCard < -1 || gs.NowCard >= gs.CardsNum {
		return 0
	}
	return 1
}

//Show the card to choose for player
func (gs *GameState) Show() {
	fmt.Println("--------CardInfo--------")

	if gs.Uno == 1 {
		fmt.Println("Your opponent shouts UNO!")
	}

	fmt.Printf("You has/have %d card(s):\n", gs.CardsNum)

	fmt.Println("##0: give up")
	for i := 0; i < gs.CardsNum; i++ {
		fmt.Printf("##%d: ", i+1)
		if gs.Cards[i].Number < 25 {
			fmt.Printf("%s  ", gs.Cards[i].Color) //no color of Wild and Wild Draw 4
		}
		num := gs.Cards[i].Number
		ShowNum(num)
	}

	fmt.Printf("Lastround color:%s ", gs.LastCard.Color)
	ShowNum(gs.LastCard.Number)

	fmt.Println("Select a number of a card.")
	fmt.Println("------------------------")
	fmt.Printf("choose >>")
}

//Transform number to card information
func ShowNum(num int) {
	switch {
	case num <= 9:
		fmt.Println(strconv.Itoa(num))
	case num <= 18:
		fmt.Println(strconv.Itoa(num - 9))
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
