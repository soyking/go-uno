/*
	Operation about players, including creating list, creating player, searching free player, and reuse player when someone quit the game
*/
package player

import (
	"errors"
	"uno/card"
)

//Player information including game state, waiting opponent and getting your turn
type Player struct {
	AnyPlayer chan int            //waiting opponent
	LastRound chan card.GameState //waiting your opponent's card infomation
	Used      int                 //whether used
	Free      int                 //whether in game
}

//A list of player with fixed length
type PlayerList struct {
	Players []Player
}

//Create a playerlist with length
func New(length int) *PlayerList {
	player := make([]Player, 0)
	for i := 0; i < length; i++ {
		anyplayer := make(chan int)
		lastround := make(chan card.GameState)                                            //independent state for chan
		newplayer := Player{AnyPlayer: anyplayer, LastRound: lastround, Used: 0, Free: 0} //initial state
		player = append(player, newplayer)
	}
	return &PlayerList{Players: player}
}

//Create a new player in the not-full room
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

//Finding opponent for playerlist[key], which is playerlist[opkey]
func (pl *PlayerList) FreePlayer(key int) (int, error) {
	for opkey, _ := range pl.Players {
		temppl := &pl.Players[opkey]
		if opkey != key && temppl.Used == 1 && temppl.Free == 1 { //valid player :exist(used) and ready(free)
			temppl.AnyPlayer <- key //notice playerlist[opkey] that there is one opponent
			temppl.Free = 0
			pl.Players[key].Free = 0
			return opkey, nil
		}
	}
	return 0, errors.New("There are not free players.Please Wait.")
}

//Recover the playerlist when playerlist[key] leaves
func (pl *PlayerList) Recover(key int) {
	pl.Players[key].Used = 0
}
