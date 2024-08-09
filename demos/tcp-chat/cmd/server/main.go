package main

import (
	"log"

	"github.com/tcornell05/go/demos/tcp-chat/internal/chat"
)

func main() {
	chatService := chat.New()
	if err := chatService.StartServer(8081); err != nil {
		log.Fatal(err)
	}
}
