package mutex

import (
	"lab2/internal/clock"
	"lab2/internal/network"
	"log"
	"sync"
	"time"
)

type State int

const (
	Released State = iota
	Wanted
	Held
)

type RicartAgrawala struct {
	mu               sync.Mutex
	nodeID           int
	totalNodes       int
	state            State
	clock            *clock.LamportClock
	transport        network.Transporter
	requestTimestamp int
	deferredReplies  []int
	replyChan        chan struct{}
}

// cria uma nova instancia do algoritmo e registra os handlers de mensagem
func NewRicartAgrawala(nodeID, totalNodes int, clk *clock.LamportClock, transport network.Transporter) *RicartAgrawala {
	ra := &RicartAgrawala{
		nodeID:     nodeID,
		totalNodes: totalNodes,
		state:      Released,
		clock:      clk,
		transport:  transport,
		replyChan:  make(chan struct{}, totalNodes),
	}

	transport.On("REQUEST", ra.handleRequest)
	transport.On("REPLY", ra.handleReply)

	return ra
}

// solicita entrada na secao critica
func (ra *RicartAgrawala) RequestCS() {
	ra.mu.Lock()
	ra.state = Wanted
	ra.requestTimestamp = ra.clock.SendTick()
	ts := ra.requestTimestamp
	ra.mu.Unlock()

	log.Printf("[Node %d][Clock: %d] MUTEX | Solicitando secao critica (REQUEST broadcast, ts=%d)",
		ra.nodeID, ts, ts)

	needed := 0
	for id := 1; id <= ra.totalNodes; id++ {
		if id != ra.nodeID {
			targetID := id
			err := ra.transport.Send(targetID, network.Message{
				Type:      "REQUEST",
				SenderID:  ra.nodeID,
				Timestamp: ts,
			})
			if err == nil {
				needed++
			} else {
				log.Printf("[Node %d][Clock: %d] MUTEX | Falha ao enviar REQUEST para Node %d, assumindo falho",
					ra.nodeID, ra.clock.Time(), targetID)
			}
		}
	}

	for i := 0; i < needed; i++ {
		select {
		case <-ra.replyChan:
		case <-time.After(5 * time.Second):
			log.Printf("[Node %d][Clock: %d] MUTEX | Timeout aguardando REPLY, prosseguindo",
				ra.nodeID, ra.clock.Time())
		}
	}

	ra.mu.Lock()
	ra.state = Held
	ra.mu.Unlock()

	log.Printf("[Node %d][Clock: %d] MUTEX | >>> ENTRANDO na secao critica <<<",
		ra.nodeID, ra.clock.Time())
}

// sai da secao critica e envia REPLY para todos os requests enfileirados
func (ra *RicartAgrawala) ReleaseCS() {
	ra.mu.Lock()
	ra.state = Released
	deferred := make([]int, len(ra.deferredReplies))
	copy(deferred, ra.deferredReplies)
	ra.deferredReplies = nil
	ra.mu.Unlock()

	log.Printf("[Node %d][Clock: %d] MUTEX | SAINDO da secao critica, liberando %d requests enfileirados",
		ra.nodeID, ra.clock.Time(), len(deferred))

	for _, id := range deferred {
		targetID := id
		go func() {
			ra.transport.Send(targetID, network.Message{
				Type:      "REPLY",
				SenderID:  ra.nodeID,
				Timestamp: ra.clock.SendTick(),
			})
		}()
	}
}

// processa um REQUEST recebido de outro no
func (ra *RicartAgrawala) handleRequest(msg network.Message) {
	ra.clock.ReceiveTick(msg.Timestamp)

	ra.mu.Lock()
	defer ra.mu.Unlock()

	shouldDefer := false

	switch ra.state {
	case Held:
		shouldDefer = true
	case Wanted:
		if msg.Timestamp > ra.requestTimestamp {
			shouldDefer = true
		} else if msg.Timestamp == ra.requestTimestamp && msg.SenderID > ra.nodeID {
			shouldDefer = true
		}
	case Released:
	}

	if shouldDefer {
		ra.deferredReplies = append(ra.deferredReplies, msg.SenderID)
		log.Printf("[Node %d][Clock: %d] MUTEX | REQUEST de Node %d (ts=%d) ENFILEIRADO (meu estado=%d, meu ts=%d)",
			ra.nodeID, ra.clock.Time(), msg.SenderID, msg.Timestamp, ra.state, ra.requestTimestamp)
	} else {
		senderID := msg.SenderID
		go func() {
			ra.transport.Send(senderID, network.Message{
				Type:      "REPLY",
				SenderID:  ra.nodeID,
				Timestamp: ra.clock.SendTick(),
			})
		}()
		log.Printf("[Node %d][Clock: %d] MUTEX | REPLY enviado para Node %d",
			ra.nodeID, ra.clock.Time(), msg.SenderID)
	}
}

// processa um REPLY recebido
func (ra *RicartAgrawala) handleReply(msg network.Message) {
	ra.clock.ReceiveTick(msg.Timestamp)
	ra.replyChan <- struct{}{}
	log.Printf("[Node %d][Clock: %d] MUTEX | REPLY recebido de Node %d",
		ra.nodeID, ra.clock.Time(), msg.SenderID)
}
