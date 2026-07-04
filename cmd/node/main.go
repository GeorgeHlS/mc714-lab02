package main

import (
	"fmt"
	"lab2/internal/clock"
	"lab2/internal/config"
	"lab2/internal/election"
	"lab2/internal/mutex"
	"lab2/internal/network"
	"log"
	"math/rand"
	"time"
)

func main() {
	// Configuracao
	cfg := config.LoadConfig()
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Printf("[Node %d] Iniciando... (total de nos: %d)", cfg.NodeID, cfg.TotalNodes)

	// Camada de rede -- servidor TCP
	transport := network.NewTransport(cfg.NodeID, cfg.Port, cfg.TotalNodes)
	transport.StartServer()

	// aguarda todos os nos subirem (tempo para containers Docker iniciarem)
	log.Printf("[Node %d] Aguardando outros nos iniciarem...", cfg.NodeID)
	time.Sleep(5 * time.Second)

	// Relogio logico de Lamport
	clk := clock.NewLamportClock()
	log.Printf("[Node %d][Clock: %d] LAMPORT | Relogio logico inicializado", cfg.NodeID, clk.Time())

	// Exclusao Mutua -- Ricart-Agrawala
	ra := mutex.NewRicartAgrawala(cfg.NodeID, cfg.TotalNodes, clk, transport)
	log.Printf("[Node %d] MUTEX | Ricart-Agrawala inicializado", cfg.NodeID)

	// Eleicao de Lider -- Bully
	bully := election.NewBullyElection(cfg.NodeID, cfg.TotalNodes, clk, transport)
	go bully.MonitorLeader()
	log.Printf("[Node %d] ELECTION | Bully inicializado (lider assumido: Node %d)", cfg.NodeID, cfg.TotalNodes)

	if cfg.NodeID == cfg.TotalNodes {
		log.Printf("[Node %d] ELECTION | Sou o lider inicial, iniciando eleicao para confirmar", cfg.NodeID)
		go bully.StartElection()
	}

	time.Sleep(3 * time.Second)

	// loop principal: periodicamente solicita acesso a secao critica
	rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(cfg.NodeID)))

	for {
		interval := time.Duration(5+rng.Intn(6)) * time.Second
		time.Sleep(interval)

		// incrementa relogio
		ts := clk.Tick()
		log.Printf("[Node %d][Clock: %d] LAMPORT | Evento local: decidindo acessar recurso compartilhado",
			cfg.NodeID, ts)

		// solicita acesso a secao critica
		ra.RequestCS()

		// SECAO CRITICA
		log.Printf("[Node %d][Clock: %d] ====== SECAO CRITICA ======", cfg.NodeID, clk.Time())
		log.Printf("[Node %d][Clock: %d] Lider atual: Node %d", cfg.NodeID, clk.Time(), bully.LeaderID())

		// simula trabalho na secao critica
		criticalWork(cfg.NodeID, clk)

		log.Printf("[Node %d][Clock: %d] ====== FIM SECAO CRITICA ======", cfg.NodeID, clk.Time())

		// libera a secao critica
		ra.ReleaseCS()
	}
}

// simula trabalho dentro da secao critica
func criticalWork(nodeID int, clk *clock.LamportClock) {
	for i := 1; i <= 3; i++ {
		ts := clk.Tick()
		log.Printf("[Node %d][Clock: %d] RECURSO | Operacao %d/3 no recurso compartilhado",
			nodeID, ts, i)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Printf(">>> Node %d usou o recurso compartilhado com sucesso <<<\n", nodeID)
}
