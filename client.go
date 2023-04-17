package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// Structure to store client information
type clientStruct struct {
	conn     net.Conn
	username string
	newuser  bool
	room     *roomStruct
	commands chan<- command
}

func isInChatroom(client *clientStruct) bool {
	if client.room == nil {
		return false
	}
	return true
}
func (client *clientStruct) login() {
	// Loop through login process until a valid login is given
	for client.username == "NULL" {
		// Get username and password
		client.newuser = false
		client.msg("Username: ")
		username, _ := bufio.NewReader(client.conn).ReadString('\n')
		client.msg("Password: ")
		password, _ := bufio.NewReader(client.conn).ReadString('\n')
		// Clean up input and put in args
		username = strings.TrimSpace(strings.Trim(username, "\r\n"))
		password = strings.TrimSpace(strings.Trim(password, "\r\n"))
		args := []string{username, password}
		// Run command to authenticate login
		client.username = "Authenticating"
		client.commands <- command{
			id:     CMD_LOGIN,
			client: client,
			args:   args,
		}
		// Wait for login to finish
		for client.username == "Authenticating" {
			// Wait
		}
		if client.newuser == true {
			client.msg("Create new user (y/N): ")
			response, _ := bufio.NewReader(client.conn).ReadString('\n')
			response = strings.TrimSpace(strings.Trim(response, "\r\n"))
			if response == "Y" || response == "y" {
				client.commands <- command{
					id:     CMD_LOGIN,
					client: client,
					args:   args,
				}
			} else {
				client.username = "NULL"
			}
		}
		// Return if login was successful
		if client.username != "NULL" {
			client.msg(fmt.Sprintf("Logged in as %s\n", client.username))
			return
		}
		// Notify user that login failed
		client.msg("Login failed\n")
	}
}
func (client *clientStruct) msg(msg string) {
	// Send message to other users in chat room
	client.conn.Write([]byte(msg))
}
func (client *clientStruct) readInput() {
	client.login()
	for {
		// Get input
		input, err := bufio.NewReader(client.conn).ReadString('\n')
		if err != nil {
			return
		}
		// Parse input
		input = strings.Trim(input, "\r\n")
		args := strings.Split(input, " ")
		cmd := strings.TrimSpace(args[0])
		// Index and run command
		if isInChatroom(client) {
			switch cmd {
			case "/exit":
				client.commands <- command{
					id:     CMD_EXIT,
					client: client,
				}
				break
			case "/join":
				client.commands <- command{
					id:     CMD_JOIN,
					client: client,
					args:   args,
				}
				break
			case "/quit":
				client.commands <- command{
					id:     CMD_QUIT,
					client: client,
				}
				break
			default:
				client.commands <- command{
					id:     CMD_MSG,
					client: client,
					args:   args,
				}
			}
		} else {
			switch cmd {
			case "/delete":
				client.commands <- command{
					id:     CMD_DELETE,
					client: client,
					args:   args,
				}
				break
			case "/exit":
				client.commands <- command{
					id:     CMD_EXIT,
					client: client,
				}
				break
			case "/help":
				client.commands <- command{
					id:     CMD_HELP,
					client: client,
				}
				break
			case "/join":
				client.commands <- command{
					id:     CMD_JOIN,
					client: client,
					args:   args,
				}
				break
			case "/make":
				client.commands <- command{
					id:     CMD_MAKE,
					client: client,
					args:   args,
				}
				break
			case "/password":
				client.commands <- command{
					id:     CMD_PASSWORD,
					client: client,
					args:   args,
				}
				break
			case "/quit":
				client.commands <- command{
					id:     CMD_QUIT,
					client: client,
				}
				break
			case "/rooms":
				client.commands <- command{
					id:     CMD_ROOMS,
					client: client,
				}
				break
			default:
				client.msg("Invalid command, try a valid command or join a chatroom\n")
			}
		}
	}
}
