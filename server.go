package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net"
	"strconv"
	"strings"
)

// Structure storing server information
type server struct {
	rooms    map[int]*roomStruct
	commands chan command
}

func contains(array []string, item string) bool {
	// Check if an array contains an item
	for i := 0; i < len(array); i++ {
		if item == array[i] {
			return true
		}
	}
	return false
}
func initializeDB(db *sql.DB) {
	// Create blank database if one doesn't exist
	db.Exec("BEGIN TRANSACTION;")
	db.Exec("CREATE TABLE IF NOT EXISTS chatroomrel (username text NOT NULL, chatroomid integer NOT NULL);")
	db.Exec("CREATE TABLE IF NOT EXISTS chatrooms (id integer NOT NULL, name text NOT NULL);")
	db.Exec("CREATE TABLE IF NOT EXISTS messages (id integer NOT NULL, sender text NOT NULL, message text NOT NULL, chatroomid integer NOT NULL);")
	db.Exec("CREATE TABLE IF NOT EXISTS users (username text NOT NULL, password text NOT NULL);")
	db.Exec("COMMIT;")
}
func newServer() *server {
	// Initialize new server
	return &server{
		rooms:    make(map[int]*roomStruct),
		commands: make(chan command),
	}
}
func (server *server) deleteRoom(db *sql.DB, client *clientStruct, args []string) {
	// Check number of arguments
	if len(args) != 2 {
		client.msg("Room ID required\n")
		client.msg("Usage: /delete ROOM_ID\n")
		return
	}
	// Delete chatroom from database
	_, err := db.Exec("DELETE FROM chatroomrel WHERE chatroomrel.username = ? AND chatroomrel.chatroomid = ?;", client.username, args[1])
	if err != nil {
		client.msg("Error deleting chatroom\n")
	}
}
func (server *server) exit(client *clientStruct) {
	// Leave current chat rooms and disconnect from server
	log.Printf("%s logged off\n", client.username)
	server.quit(client)
	client.conn.Close()
}
func (server *server) help(db *sql.DB, client *clientStruct) {
	// Print list of commands with parameters
	client.msg("/delete ROOM_ID\n")
	client.msg("/exit\n")
	client.msg("/help\n")
	client.msg("/join ROOM_ID\n")
	client.msg("/make ROOM_NAME USER1 USER2 .... USERX\n")
	client.msg("/password PASSWORD\n")
	client.msg("/quit\n")
	client.msg("/rooms\n")
}
func (server *server) join(db *sql.DB, client *clientStruct, args []string) {
	// Check number of arguments
	if len(args) != 2 {
		client.msg("Room ID required\n")
		client.msg("Usage: /join ROOM_ID\n")
		return
	}
	// Select chatroom
	roomID, _ := strconv.Atoi(args[1])
	var roomName string
	err := db.QueryRow("SELECT chatrooms.name FROM chatrooms JOIN chatroomrel ON chatrooms.id = chatroomrel.chatroomid JOIN users ON chatroomrel.username = users.username WHERE chatrooms.id = ? AND users.username = ?", roomID, client.username).Scan(&roomName)
	if err != nil {
		client.msg("Invalid room ID\n")
		return
	}
	// Join chat room
	room, ok := server.rooms[roomID]
	if !ok {
		room = &roomStruct{
			id:      roomID,
			name:    roomName,
			members: make(map[net.Addr]*clientStruct),
		}
		server.rooms[roomID] = room
	}
	room.members[client.conn.RemoteAddr()] = client
	server.quit(client)
	client.room = room
	room.broadcast(client, fmt.Sprintf("%s joined the room\n", client.username))
	client.msg(fmt.Sprintf("welcome to %s\n", roomName))
	// Print message backlog in chat
	messages, _ := db.Query("SELECT messages.sender, messages.message FROM messages WHERE chatroomid = ? ORDER BY messages.id ASC;", client.room.id)
	var sender string
	var message string
	for messages.Next() {
		messages.Scan(&sender, &message)
		client.msg(fmt.Sprintf(sender + " : " + message + "\n"))
	}
}
func (server *server) listRooms(db *sql.DB, client *clientStruct) {
	// List rooms the user is in
	result, err := db.Query("SELECT chatrooms.id, chatrooms.name FROM chatrooms JOIN chatroomrel ON chatrooms.id = chatroomrel.chatroomid JOIN users ON chatroomrel.username = users.username WHERE users.username = ? ORDER BY chatrooms.id ASC", client.username)
	if err != nil {
		client.msg("Database error\n")
		return
	}
	client.msg("Available Rooms:\n")
	var id int
	var name string
	for result.Next() {
		result.Scan(&id, &name)
		client.msg(fmt.Sprintf(strconv.Itoa(id) + " : " + name + "\n"))
	}
}
func (server *server) login(db *sql.DB, client *clientStruct, args []string) {
	var username string

	// Add user to database if creating new user
	if client.newuser == true {
		db.Exec("INSERT INTO users VALUES(?, ?)", args[0], args[1])
		client.username = args[0]
		return
	}

	// Check if user exists in system
	err := db.QueryRow("SELECT users.username FROM users WHERE users.username = ?;", args[0]).Scan(&username)
	if err != nil {
		client.newuser = true
		client.username = args[0]
		return
	}

	// Check if username and password is correct
	err = db.QueryRow("SELECT users.username FROM users WHERE users.username = ? AND users.password = ?;", args[0], args[1]).Scan(&username)
	// Set username if login in valid
	if err != nil {
		client.username = "NULL"
		return
	}
	// Set username and log to server
	client.username = username
	log.Printf("%s logged in\n", client.username)
}
func (server *server) makeRoom(db *sql.DB, client *clientStruct, args []string) {
	// Check number of arguments
	if len(args) < 3 {
		client.msg("Room name and users are required.\n")
		client.msg("Usage: /make ROOM_NAME USER1 USER2 .... USERX\n")
		return
	}
	// Verify users exist in database and add to list
	var count int
	var users []string
	users = append(users, client.username)
	for i := 2; i < len(args); i++ {
		db.QueryRow("SELECT COUNT(*) FROM users WHERE users.username = ?", args[i]).Scan(&count)
		if !contains(users, args[i]) && count == 1 {
			users = append(users, args[i])
		}
	}
	// Enter new chatroom into database
	err := db.QueryRow("SELECT COUNT(*) FROM chatrooms").Scan(&count)
	count++
	_, err = db.Exec("INSERT INTO chatrooms VALUES(?, ?)", count, args[1])
	if err != nil {
		client.msg("Database error\n")
		return
	}
	// Enter relationships into database
	for i := 0; i < len(users); i++ {
		_, err := db.Exec("INSERT INTO chatroomrel VALUES(?, ?)", users[i], count)
		if err != nil {
			client.msg("Database error\n")
			continue
		}
	}
}
func (server *server) msg(db *sql.DB, client *clientStruct, args []string) {
	// Join arguments to form message
	msg := strings.Join(args[0:], " ")
	msg = strings.TrimSpace(strings.Trim(msg, "\r\n"))
	// Don't send message if empty or only containing whitespace
	if msg == "" || msg == " " {
		return
	}
	// Send message to chatroom and save to database
	client.room.broadcast(client, "\n"+client.username+": "+msg+"\n")
	db.Exec("INSERT INTO messages VALUES ((SELECT COUNT(*) from messages)+1, ?, ?, ?);", client.username, msg, client.room.id)
}
func (server *server) newClient(conn net.Conn) *clientStruct {
	log.Printf("New client connected")
	return &clientStruct{
		conn:     conn,
		username: "NULL",
		commands: server.commands,
	}
}
func (server *server) password(db *sql.DB, client *clientStruct, args []string) {
	// Check number of arguments
	if len(args) != 2 {
		client.msg("Password is required.\n")
		client.msg("Usage: /password PASSWORD\n")
		return
	}
	// Update password in database
	_, err := db.Exec("UPDATE users SET password = ? WHERE users.username = ?;", args[1], client.username)
	if err != nil {
		client.msg("Error updating password\n")
		return
	}
	client.msg("Password updated\n")
}
func (server *server) quit(client *clientStruct) {
	// Quit the current chat room
	if client.room != nil {
		oldRoom := server.rooms[client.room.id]
		delete(server.rooms[client.room.id].members, client.conn.RemoteAddr())
		oldRoom.broadcast(client, fmt.Sprintf("%s has left the room\n", client.username))
	}
	client.room = nil
}
func (server *server) run() {
	// Open and initialize database
	db, _ := sql.Open("sqlite3", "./db.db")
	initializeDB(db)
	// Run the indexed command
	for cmd := range server.commands {
		switch cmd.id {
		case CMD_DELETE:
			server.deleteRoom(db, cmd.client, cmd.args)
		case CMD_EXIT:
			server.exit(cmd.client)
		case CMD_HELP:
			server.help(db, cmd.client)
		case CMD_JOIN:
			server.join(db, cmd.client, cmd.args)
		case CMD_LOGIN:
			server.login(db, cmd.client, cmd.args)
		case CMD_MAKE:
			server.makeRoom(db, cmd.client, cmd.args)
		case CMD_MSG:
			server.msg(db, cmd.client, cmd.args)
		case CMD_PASSWORD:
			server.password(db, cmd.client, cmd.args)
		case CMD_QUIT:
			server.quit(cmd.client)
		case CMD_ROOMS:
			server.listRooms(db, cmd.client)
		}
	}
}
