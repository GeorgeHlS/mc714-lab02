package network

// representa uma mensagem trocada entre nos distribuidos via TCP/JSON
type Message struct {
	Type      string `json:"type"`
	SenderID  int    `json:"sender_id"`
	Timestamp int    `json:"timestamp"`
	Data      string `json:"data,omitempty"`
}
