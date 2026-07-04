package config

import (
	"log"
	"os"
	"strconv"
)

// armazena as configuracoes de um no distribuido
type Config struct {
	NodeID     int
	TotalNodes int
	Port       int
}

// le as variaveis de ambiente e retorna a configuracao do no
func LoadConfig() Config {
	nodeID, err := strconv.Atoi(os.Getenv("NODE_ID"))
	if err != nil {
		log.Fatal("NODE_ID invalido ou ausente")
	}

	totalNodes, err := strconv.Atoi(os.Getenv("TOTAL_NODES"))
	if err != nil {
		log.Fatal("TOTAL_NODES invalido ou ausente")
	}

	return Config{
		NodeID:     nodeID,
		TotalNodes: totalNodes,
		Port:       5000,
	}
}
