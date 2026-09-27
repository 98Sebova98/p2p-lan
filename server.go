package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключение любых клиентов
	},
}

type Client struct {
	conn *websocket.Conn
	addr string
}

type Room struct {
	sync.Mutex
	clients map[string]*Client
}

var room = &Room{
	clients: make(map[string]*Client),
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка Upgrade: %v", err)
		return
	}
	defer ws.Close()

	clientAddr := ws.RemoteAddr().String()

	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			room.Lock()
			delete(room.clients, clientAddr)
			room.Unlock()
			break
		}

		msg := strings.TrimSpace(string(msgBytes))

		if strings.HasPrefix(msg, "JOIN") {
			// Если клиент передает свой локальный/публичный адрес, используем его, иначе RemoteAddr
			parts := strings.Split(msg, " ")
			if len(parts) > 1 {
				clientAddr = parts[1]
			}

			room.Lock()
			room.clients[clientAddr] = &Client{conn: ws, addr: clientAddr}
			fmt.Printf("[+] Игрок подключился: %s (Всего в сети: %d)\n", clientAddr, len(room.clients))

			var list []string
			for addr := range room.clients {
				list = append(list, addr)
			}
			peerListMsg := "PEERS:" + strings.Join(list, ",")

			for _, client := range room.clients {
				client.conn.WriteMessage(websocket.TextMessage, []byte(peerListMsg))
			}
			room.Unlock()
		}
	}
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	http.HandleFunc("/", healthCheck)           // Для Health Check сканера Render
	http.HandleFunc("/ws", handleConnections)   // WebSocket эндпоинт

	fmt.Printf("[Server] Сигнальный сервер запущен на порту :%s\n", port)
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
