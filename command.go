package main

// Value for command to be run
type commandID int

// Constants to index command to be run
const (
	CMD_LOGIN commandID = iota
	CMD_DELETE
	CMD_EXIT
	CMD_HELP
	CMD_JOIN
	CMD_MAKE
	CMD_MSG
	CMD_PASSWORD
	CMD_ROOMS
	CMD_QUIT
)

// Structure for command and parameters to pass to server
type command struct {
	id     commandID
	client *clientStruct
	args   []string
}
