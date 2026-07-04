package mutex

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

func TestRicartAgrawala_HandleRequest_Released(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()
	ra := NewRicartAgrawala(1, 3, clk, mockNet)

	reqMsg := network.Message{
		Type:      "REQUEST",
		SenderID:  2,
		Timestamp: 10,
	}

	ra.state = Released

	handler := mockNet.Handlers["REQUEST"]
	handler(reqMsg)

	time.Sleep(50 * time.Millisecond)

	if len(mockNet.SentMessages) != 1 {
		t.Fatalf("Esperava 1 mensagem enviada, obteve %d", len(mockNet.SentMessages))
	}

	replyMsg := mockNet.SentMessages[0]
	if replyMsg.Type != "REPLY" {
		t.Errorf("Esperava mensagem do tipo REPLY, obteve %s", replyMsg.Type)
	}
}

func TestRicartAgrawala_HandleRequest_Held(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()
	ra := NewRicartAgrawala(1, 3, clk, mockNet)

	ra.state = Held

	reqMsg := network.Message{Type: "REQUEST", SenderID: 2, Timestamp: 10}

	handler := mockNet.Handlers["REQUEST"]
	handler(reqMsg)
	time.Sleep(10 * time.Millisecond)

	if len(mockNet.SentMessages) > 0 {
		t.Errorf("Nao esperava envio de REPLY enquanto HELD")
	}

	if len(ra.deferredReplies) != 1 || ra.deferredReplies[0] != 2 {
		t.Errorf("Esperava Node 2 enfileirado")
	}
}

func TestRicartAgrawala_HandleRequest_WantedPriority(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()
	ra := NewRicartAgrawala(1, 3, clk, mockNet)

	ra.state = Wanted
	ra.requestTimestamp = 5

	reqMsg := network.Message{Type: "REQUEST", SenderID: 2, Timestamp: 10}

	handler := mockNet.Handlers["REQUEST"]
	handler(reqMsg)
	time.Sleep(10 * time.Millisecond)

	if len(mockNet.SentMessages) > 0 {
		t.Errorf("Nao esperava envio de REPLY, pois Node 1 tem prioridade")
	}
	if len(ra.deferredReplies) != 1 {
		t.Errorf("Esperava Node 2 enfileirado")
	}
}

func TestRicartAgrawala_HandleRequest_WantedConcede(t *testing.T) {
	clk := clock.NewLamportClock()
	mockNet := NewMockTransport()
	ra := NewRicartAgrawala(1, 3, clk, mockNet)

	ra.state = Wanted
	ra.requestTimestamp = 10

	reqMsg := network.Message{Type: "REQUEST", SenderID: 2, Timestamp: 5}

	handler := mockNet.Handlers["REQUEST"]
	handler(reqMsg)
	time.Sleep(50 * time.Millisecond)

	if len(mockNet.SentMessages) != 1 || mockNet.SentMessages[0].Type != "REPLY" {
		t.Errorf("Esperava envio de REPLY, pois Node 2 pediu antes")
	}
}
