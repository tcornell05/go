package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/tcornell05/go/demos/tcp-chat/internal/chat"
)

func main() {
	// connect to server
	conn, err := net.Dial("tcp", ":8081")
	if err != nil {
		log.Fatalf("Dial Error: %v", err)
	}

	// Waits here until motd
	connReader := bufio.NewReader(conn)
	motd, _ := connReader.ReadString('\n')
	fmt.Print(chat.Colorize(motd+"\n", "green"))
	stdin := bufio.NewScanner(os.Stdin)
	fmt.Print(chat.Colorize(">> ", "bold_green"))
	if stdin.Scan() {
		// Name has been entered
		name := stdin.Text()
		chat.WriteToConn(&conn, name)

		// Wait for confirm
		intro, _ := connReader.ReadString('\n')
		fmt.Println(chat.Colorize(intro, "purple"))
		newLine()
	}

	// Continue reading proceeding messages from the server
	go func() {
		for {
			serverMsg, _ := connReader.ReadString('\n')
			fmt.Println(chat.Colorize(serverMsg, "purple"))

		}
	}()

	// continue reading from stdin
	for stdin.Scan() {
		newLine()
		msg := stdin.Text()

		// Client side validate but not today here
		if len(msg) > 0 {
			chat.WriteToConn(&conn, msg)
		}
	}
}

func newLine() {
	fmt.Print(chat.Colorize(">> ", "bold_green"))
}
