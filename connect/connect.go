package connect

import (
	"encoding/json"
	"net"
	"uno/card"
)

type TransInfo struct {
	//info type
	State      int
	CardInfo   card.GameState
	InfoString string
}

func New() *TransInfo {
	return &TransInfo{State: 0}
}

func (ti *TransInfo) Send(conn net.Conn, info string, state int) error {
	ti.State = state
	ti.InfoString = info
	b, err := json.Marshal(ti)
	if err != nil {
		return err
	}
	_, err = conn.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (ti *TransInfo) Receive(conn net.Conn) error {
	var b [1000]byte
	n, _ := conn.Read(b[0:])
	err := json.Unmarshal(b[0:n], &ti)
	if err != nil {
		return err
	}
	return nil
}
