package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Room struct {
	sync.Mutex
	members map[string]*net.UDPAddr
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("[Server] Сигнальный сервер запущен на порту :%s\n", port)

	room := &Room{members: make(map[string]*net.UDPAddr)}
	buf := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}

		msg := strings.TrimSpace(string(buf[:n]))

		if msg == "JOIN" {
			room.Lock()
			room.members[clientAddr.String()] = clientAddr
			fmt.Printf("[+] Игрок подключился: %s (Всего в сети: %d)\n", clientAddr.String(), len(room.members))

			var list []string
			for _, mAddr := range room.members {
				list = append(list, mAddr.String())
			}
			peerListMsg := "PEERS:" + strings.Join(list, ",")

			for _, mAddr := range room.members {
				conn.WriteToUDP([]byte(peerListMsg), mAddr)
			}
			room.Unlock()
		}
	}
}
