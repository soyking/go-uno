package player

import (
	"errors"
	"fmt"
	"uno/card"
)

type Player struct {
	AnyPlayer chan int
	LastRound chan card.GameState
	Used      int
	Free      int
}

type PlayerList struct {
	Players []Player
}

func (pl *PlayerList) New() (int, *Player, error) {
	for k, _ := range pl.Players {
		temppl := &pl.Players[k]
		if temppl.Used != 1 {
			temppl.Used = 1
			temppl.Free = 1
			return k, temppl, nil
		}
	}
	return 0, nil, errors.New("There are too many players.Please Wait.")
}

func (pl *PlayerList) FreePlayer(key int) (int, error) {
	for k, _ := range pl.Players {
		temppl := &pl.Players[k]
		if k != key && temppl.Used == 1 && temppl.Free == 1 {
			fmt.Printf("finding free player,key:%d opkey:%d\n", key, k)
			temppl.Free = 0
			temppl.AnyPlayer <- key
			pl.Players[key].Free = 0
			fmt.Printf("**finding free player,key:%d opkey:%d\n", key, k)
			return k, nil
		}
	}
	return 0, errors.New("There are not free players.Please Wait.")
}

func (pl *PlayerList) Recover(key int) {
	pl.Players[key].Used = 0
}

func (pl *PlayerList) Show() {
	fmt.Println("------show of playerlist------")
	for k, v := range pl.Players {
		fmt.Printf("key:%d used:%d free:%d\n", k, v.Used, v.Free)
	}
	fmt.Println("------------------------------")
}

func New(number int) *PlayerList {
	player := make([]Player, number)
	return &PlayerList{Players: player}
}
