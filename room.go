package main

import (
	"net"
)

// Structure for holding room information
type roomStruct struct {
	id      int
	name    string
	members map[net.Addr]*clientStruct
}

// Send message to all users in chatroom
func (room *roomStruct) broadcast(sender *clientStruct, msg string) {
	for addr, member := range room.members {
		if sender.conn.RemoteAddr() != addr {
			member.msg(msg)
		}
	}
}
