package election

import (
	"lab2/internal/clock"
	"lab2/internal/network"
	"testing"
	"time"
)

type MockTransport struct {
	Handlers     map[string]network.MessageHandler
	SentMessages []network.Message
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Handlers:     make(map[string]network.MessageHandler),
		SentMessages: make([]network.Message, 0),
	}
}

func (m *MockTransport) Send(targetID int, msg network.Message) error {
	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

func (m *MockTransport) Broadcast(msg network.Message) {
	m.SentMessages = append(m.SentMessages, msg)
}

func (m *MockTransport) On(msgType string, handler network.MessageHandler) {
	m.Handlers[msgType] = handler
}

func TestBullyElection_StartElection_SendToHigher(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()

	be := NewBullyElection(3, 5, clk, mockNet)

	go be.StartElection()

	time.Sleep(50 * time.Millisecond)

	if len(mockNet.SentMessages) != 2 {
		t.Errorf("Esperava 2 mensagens enviadas, obteve %d", len(mockNet.SentMessages))
	}

	for _, msg := range mockNet.SentMessages {
		if msg.Type != "ELECTION" {
			t.Errorf("Esperava msg ELECTION, obteve %s", msg.Type)
		}
	}
}

func TestBullyElection_HandleElection_SendOK(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()

	be := NewBullyElection(5, 5, clk, mockNet)
	_ = be

	reqMsg := network.Message{Type: "ELECTION", SenderID: 3, Timestamp: 10}

	handler := mockNet.Handlers["ELECTION"]
	handler(reqMsg)
	time.Sleep(50 * time.Millisecond)

	hasOK := false
	for _, msg := range mockNet.SentMessages {
		if msg.Type == "OK" && msg.SenderID == 5 {
			hasOK = true
		}
	}

	if !hasOK {
		t.Errorf("Esperava responder com OK para o ELECTION do Node 3")
	}
}

func TestBullyElection_HandleCoordinator(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()

	be := NewBullyElection(3, 5, clk, mockNet)

	if be.LeaderID() != 5 {
		t.Errorf("Lider inicial deveria ser 5")
	}

	msg := network.Message{Type: "COORDINATOR", SenderID: 4, Timestamp: 20}
	handler := mockNet.Handlers["COORDINATOR"]
	handler(msg)

	if be.LeaderID() != 4 {
		t.Errorf("Novo lider deveria ser atualizado para 4")
	}
}
