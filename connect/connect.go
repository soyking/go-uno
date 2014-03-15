/*
	Connection between server and client
*/
package connect

import (
	"encoding/json"
	"errors"
	"go-uno/card"
	"net"
	"strconv"
	"strings"
)

const MAXTIME = 10 //max times of error

//Struct for transporting
type TransInfo struct {
	//state: 0 transport information in InfoString
	//		 1 request from server of puting out one card
	//		 2 response from client of a card the player put out
	State      int
	CardInfo   card.GameState
	InfoString string
}

//Create struct
func New() *TransInfo {
	cardinfo := card.New()
	return &TransInfo{State: 0, CardInfo: *cardinfo}
}

//Method of sending message to client/server
func (ti *TransInfo) Send(conn net.Conn, info string, state int) error {
	ti.State = state
	ti.InfoString = info
	b, err := json.Marshal(ti) //json type
	if err != nil {
		return err
	}

	//json type: #length#{}#length#{}...
	//avoid receiving two messages, such that something wrong when unmarshar
	length := []byte(strconv.Itoa(len(b)))
	length = append(length, '#')
	length = append([]byte("#"), length...)
	b = append(length, b...)

	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

//Call client to put out one card
func (ti *TransInfo) PutOutCard(conn net.Conn) error {
	times := 0

	receiveinfo := New()
	for times < MAXTIME {
		err := ti.Send(conn, "put out one card", 1)
		if err != nil {
			return errors.New(err.Error())
		}

		//receive response from client and check them
		transinfolist, err := Receive(conn)
		if err != nil {
			return errors.New(err.Error())
		}

		//assume there is just one transinfo from client
		if len(transinfolist) > 0 {
			receiveinfo = &transinfolist[0]
		} else {
			return errors.New("Receive error.")
		}

		if receiveinfo.CardInfo.CheckCard() == false {
			ti.Send(conn, "You put out a wrong card", 0) //reput out a card
			times++
		} else {
			break
		}
	}

	//whether put out a right card
	if times < MAXTIME {
		*ti = *receiveinfo
		return nil
	}
	return errors.New("Put out the wrong card for too many times.")
}

//Receive information from client/server,return information list
func Receive(conn net.Conn) ([]TransInfo, error) {
	transinfolist := make([]TransInfo, 0)

	var b [5000]byte
	n, _ := conn.Read(b[0:])

	//receive response from client and check them
	//split an unmarshal the information
	split := strings.Split(string(b[0:n]), "#")
	for i := 1; i < len(split); i += 2 {
		length, err := strconv.Atoi(split[i])
		if err != nil {
			return nil, errors.New("Information Length error.")
		}
		js := split[i+1]

		if len(js) != length {
			return nil, errors.New("Information transmission error.")
		}

		transinfo := New()
		err = json.Unmarshal([]byte(js), transinfo)
		if err != nil {
			return nil, err
		}
		transinfolist = append(transinfolist, *transinfo)
	}

	return transinfolist, nil
}
