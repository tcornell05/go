package chat

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type ChatService struct {
	server  net.Listener
	clients map[string]*net.Conn
	mu      sync.Mutex
}

func New() *ChatService {
	return &ChatService{
		clients: make(map[string]*net.Conn),
	}
}

func (c *ChatService) StartServer(p int) error {
	port := fmt.Sprintf(":%v", p)
	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	defer l.Close()

	c.server = l

	fmt.Printf("Server started on port: %v\n", p)

	for {
		if conn, err := l.Accept(); err == nil {
			// run in go routinues so mutliple conns can be handled at once
			go func() {
				// Ask for name
				WriteToConn(&conn, "Welcome! Please enter your name: ")

				connReader := bufio.NewReader(conn)
				// Will hang until name is sent back for client
				name, _ := connReader.ReadString('\n')
				name = strings.TrimSpace(name)
				fmt.Println(name)
				// name entered. Register clients
				c.mu.Lock()
				c.clients[name] = &conn
				c.mu.Unlock()

				WriteToConn(&conn, "Welcome, "+name+"! Send a message via the format toUser:Message")
				// now we just handle future messages.

				for {

					clientMsg, _ := connReader.ReadString('\n')
					parts := strings.SplitN(clientMsg, ":", 2)
					if len(parts) == 0 {
						continue
					}
					fmt.Println(parts)
					if len(parts) < 2 {
						WriteToConn(&conn, "Invalid Format, please use toUser:Message")
						continue
					}
					// we have a valid message, how exciting!
					to := strings.TrimSpace(parts[0])
					from := strings.TrimSpace(name)
					body := strings.TrimSpace(parts[1])

					clientConn := c.clients[to]
					if clientConn == nil {
						WriteToConn(&conn, Colorize("Message not sent. This user does not exist.", "red"))
						continue
					}

					WriteToConn(clientConn, Colorize(from+": ", "bold_green")+body)
				}
			}()
			//
		}
	}
}

func WriteToConn(conn *net.Conn, msg string) error {
	_, err := fmt.Fprint(*conn, msg+"\n")
	if err != nil {
		return err
	}

	return nil
}
