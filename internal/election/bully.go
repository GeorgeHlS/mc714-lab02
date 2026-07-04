package election

import (
	"lab2/internal/clock"
	"lab2/internal/network"
	"log"
	"sync"
	"time"
)

const (
	HeartbeatInterval = 2 * time.Second // Intervalo entre heartbeats do lider
	HeartbeatTimeout  = 6 * time.Second // lider como morto
	ElectionTimeout   = 3 * time.Second // aguardar OK na eleicao
)

type BullyElection struct {
	mu                 sync.Mutex
	nodeID             int
	totalNodes         int
	clock              *clock.LamportClock
	transport          network.Transporter
	leaderID           int
	electionInProgress bool
	lastHeartbeat      time.Time
	okReceived         chan struct{}
}

// cria uma nova instancia do algoritmo Bully
func NewBullyElection(nodeID, totalNodes int, clk *clock.LamportClock, transport network.Transporter) *BullyElection {
	be := &BullyElection{
		nodeID:        nodeID,
		totalNodes:    totalNodes,
		clock:         clk,
		transport:     transport,
		leaderID:      totalNodes,
		lastHeartbeat: time.Now(),
		okReceived:    make(chan struct{}, totalNodes),
	}

	transport.On("ELECTION", be.handleElection)
	transport.On("OK", be.handleOK)
	transport.On("COORDINATOR", be.handleCoordinator)
	transport.On("HEARTBEAT", be.handleHeartbeat)

	return be
}

// inicia o processo de eleicao Bully
func (be *BullyElection) StartElection() {
	be.mu.Lock()
	if be.electionInProgress {
		be.mu.Unlock()
		return
	}
	be.electionInProgress = true
	be.mu.Unlock()

	ts := be.clock.Tick()
	log.Printf("[Node %d][Clock: %d] ELECTION | Iniciando eleicao...", be.nodeID, ts)

	sentToHigher := false
	for id := be.nodeID + 1; id <= be.totalNodes; id++ {
		err := be.transport.Send(id, network.Message{
			Type:      "ELECTION",
			SenderID:  be.nodeID,
			Timestamp: be.clock.SendTick(),
		})
		if err == nil {
			sentToHigher = true
			log.Printf("[Node %d][Clock: %d] ELECTION | Enviado ELECTION para Node %d",
				be.nodeID, be.clock.Time(), id)
		} else {
			log.Printf("[Node %d][Clock: %d] ELECTION | Node %d inacessivel",
				be.nodeID, be.clock.Time(), id)
		}
	}

	if !sentToHigher {
		be.declareVictory()
		return
	}

	select {
	case <-be.okReceived:
		log.Printf("[Node %d][Clock: %d] ELECTION | Recebeu OK, aguardando COORDINATOR...",
			be.nodeID, be.clock.Time())
		be.mu.Lock()
		be.electionInProgress = false
		be.mu.Unlock()
	case <-time.After(ElectionTimeout):
		be.declareVictory()
	}
}

func (be *BullyElection) declareVictory() {
	be.mu.Lock()
	be.leaderID = be.nodeID
	be.electionInProgress = false
	be.mu.Unlock()

	ts := be.clock.Tick()
	log.Printf("[Node %d][Clock: %d] ELECTION | *** Sou o novo LIDER! *** Enviando COORDINATOR para todos",
		be.nodeID, ts)

	for id := 1; id <= be.totalNodes; id++ {
		if id != be.nodeID {
			targetID := id
			go func() {
				be.transport.Send(targetID, network.Message{
					Type:      "COORDINATOR",
					SenderID:  be.nodeID,
					Timestamp: be.clock.SendTick(),
				})
			}()
		}
	}
	go be.startHeartbeat()
}

// recebe ELECTION de um no com ID menor
func (be *BullyElection) handleElection(msg network.Message) {
	be.clock.ReceiveTick(msg.Timestamp)

	if be.nodeID > msg.SenderID {
		log.Printf("[Node %d][Clock: %d] ELECTION | Recebeu ELECTION de Node %d, respondendo OK",
			be.nodeID, be.clock.Time(), msg.SenderID)

		go be.transport.Send(msg.SenderID, network.Message{
			Type:      "OK",
			SenderID:  be.nodeID,
			Timestamp: be.clock.SendTick(),
		})

		go be.StartElection()
	}
}

// recebe OK de um no com ID maior
func (be *BullyElection) handleOK(msg network.Message) {
	be.clock.ReceiveTick(msg.Timestamp)
	log.Printf("[Node %d][Clock: %d] ELECTION | Recebeu OK de Node %d",
		be.nodeID, be.clock.Time(), msg.SenderID)

	select {
	case be.okReceived <- struct{}{}:
	default:
	}
}

// recebe o anuncio do novo lider
func (be *BullyElection) handleCoordinator(msg network.Message) {
	be.clock.ReceiveTick(msg.Timestamp)

	be.mu.Lock()
	oldLeader := be.leaderID
	be.leaderID = msg.SenderID
	be.electionInProgress = false
	be.lastHeartbeat = time.Now()
	be.mu.Unlock()

	log.Printf("[Node %d][Clock: %d] ELECTION | Novo lider: Node %d (anterior: Node %d)",
		be.nodeID, be.clock.Time(), msg.SenderID, oldLeader)
}

// envia HEARTBEAT periodicamente enquanto este no for o lider
func (be *BullyElection) startHeartbeat() {
	ticker := time.NewTicker(HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		be.mu.Lock()
		isLeader := be.leaderID == be.nodeID
		be.mu.Unlock()

		if !isLeader {
			return
		}

		be.transport.Broadcast(network.Message{
			Type:      "HEARTBEAT",
			SenderID:  be.nodeID,
			Timestamp: be.clock.SendTick(),
		})
	}
}

// atualiza o timestamp do ultimo heartbeat recebido
func (be *BullyElection) handleHeartbeat(msg network.Message) {
	be.clock.ReceiveTick(msg.Timestamp)
	be.mu.Lock()
	be.lastHeartbeat = time.Now()
	be.mu.Unlock()
}

// roda em background e detecta falha do lider via timeout de heartbeat
func (be *BullyElection) MonitorLeader() {
	for {
		time.Sleep(HeartbeatInterval)

		be.mu.Lock()
		isLeader := be.leaderID == be.nodeID
		elapsed := time.Since(be.lastHeartbeat)
		inElection := be.electionInProgress
		be.mu.Unlock()

		if !isLeader && !inElection && elapsed > HeartbeatTimeout {
			log.Printf("[Node %d][Clock: %d] ELECTION | Lider nao responde ha %v! Iniciando eleicao...",
				be.nodeID, be.clock.Time(), elapsed.Round(time.Millisecond))
			go be.StartElection()
		}
	}
}

// retorna o ID do lider atual
func (be *BullyElection) LeaderID() int {
	be.mu.Lock()
	defer be.mu.Unlock()
	return be.leaderID
}

// retorna true se este no e o lider atual
func (be *BullyElection) IsLeader() bool {
	be.mu.Lock()
	defer be.mu.Unlock()
	return be.leaderID == be.nodeID
}
