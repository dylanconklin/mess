SERVER_NAME=mess-server

all: mess mess-local mess-server

mess:
	cp mess.sh mess
	chmod +x mess

mess-local:
	echo "#!/usr/bin/bash" > mess-local
	echo "telnet -E localhost 8888" >> mess-local
	chmod +x mess-local

mess-server:
	go build -o mess-server client.go command.go main.go room.go server.go
	chmod +x mess-server

deps:
	go get github.com/mattn/go-sqlite3

clean:
	rm -f mess
	rm -f mess-local
	rm -f mess-server
