package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type MessageHandler func(msg Message)

// define as operacoes de rede necessarias para os algoritmos
type Transporter interface {
	Send(targetID int, msg Message) error
	Broadcast(msg Message)
	On(msgType string, handler MessageHandler)
}

// encapsula a camada de comunicacao TCP/JSON de um no distribuido
type Transport struct {
	nodeID     int
	port       int
	totalNodes int
	handlers   map[string]MessageHandler
}

// cria a camada de rede de um no
func NewTransport(nodeID, port, totalNodes int) *Transport {
	return &Transport{
		nodeID:     nodeID,
		port:       port,
		totalNodes: totalNodes,
		handlers:   make(map[string]MessageHandler),
	}
}

// inicia o servidor TCP que escuta conexoes de outros nos
func (t *Transport) StartServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.port))
	if err != nil {
		log.Fatalf("[Node %d] Erro ao iniciar servidor: %v", t.nodeID, err)
	}
	log.Printf("[Node %d] Servidor TCP escutando na porta %d", t.nodeID, t.port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("[Node %d] Erro ao aceitar conexao: %v", t.nodeID, err)
				continue
			}
			go t.handleConnection(conn)
		}
	}()
}

// le uma mensagem JSON da conexao e despacha para o handler registrado
func (t *Transport) handleConnection(conn net.Conn) {
	defer conn.Close()
	var msg Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil {
		return
	}
	if handler, ok := t.handlers[msg.Type]; ok {
		handler(msg)
	} else {
		log.Printf("[Node %d] Mensagem de tipo desconhecido: %s", t.nodeID, msg.Type)
	}
}

// envia uma mensagem para um no especifico identificado por targetID
func (t *Transport) Send(targetID int, msg Message) error {
	addr := fmt.Sprintf("node%d:5000", targetID)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return json.NewEncoder(conn).Encode(msg)
}

// envia uma mensagem para todos os outros nos
func (t *Transport) Broadcast(msg Message) {
	for id := 1; id <= t.totalNodes; id++ {
		if id != t.nodeID {
			targetID := id
			go func() {
				if err := t.Send(targetID, msg); err != nil {
				}
			}()
		}
	}
}

// envia uma mensagem apenas para nos com ID maior que o proprio
func (t *Transport) SendToHigherNodes(msg Message) bool {
	sentToAny := false
	for id := t.nodeID + 1; id <= t.totalNodes; id++ {
		if err := t.Send(id, msg); err == nil {
			sentToAny = true
		}
	}
	return sentToAny
}

// registra um handler para um tipo de mensagem
func (t *Transport) On(msgType string, handler MessageHandler) {
	t.handlers[msgType] = handler
}

// retorna o ID deste no
func (t *Transport) NodeID() int {
	return t.nodeID
}
