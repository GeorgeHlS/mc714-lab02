# Lab 2 - Algoritmos Distribuidos | MC714 - UNICAMP

Implementacao de tres algoritmos fundamentais de sistemas distribuidos usando Go e Docker:

1. **Relogio Logico de Lamport** - Ordenacao causal de eventos
2. **Exclusao Mutua (Ricart-Agrawala)** - Acesso exclusivo a recurso compartilhado
3. **Eleicao de Lider (Bully)** - Eleicao automatica quando o lider falha

## Arquitetura

- **Linguagem:** Go 1.22+
- **Comunicacao:** Sockets TCP com mensagens JSON
- **Ambiente:** Docker Compose com 5 containers (nodes) na mesma rede bridge
- **Processos:** 5 nos distribuidos (node1..node5), cada um rodando os 3 algoritmos

## Estrutura do Projeto

```
lab2/
├── docker-compose.yml          # Define os 5 containers
├── Dockerfile                  # Multi-stage build (compila + imagem minima)
├── go.mod                      # Modulo Go
├── cmd/
│   └── node/
│       └── main.go             # Ponto de entrada
├── internal/
│   ├── clock/
│   │   └── lamport.go          # Relogio logico de Lamport
│   ├── mutex/
│   │   └── ricart_agrawala.go  # Exclusao mutua distribuida
│   ├── election/
│   │   └── bully.go            # Eleicao de lider
│   ├── network/
│   │   ├── transport.go        # Camada de comunicacao TCP/JSON
│   │   └── message.go          # Tipos de mensagem
│   └── config/
│       └── config.go           # Configuracao via variaveis de ambiente
└── README.md
```

## Como Executar

### Pre-requisitos

- Docker e Docker Compose instalados

### Comandos

```bash
# Construir e iniciar todos os 5 nos
docker-compose up --build

# Em outro terminal, acompanhar logs de um no especifico
docker logs -f node1

# Simular falha do lider (node5)
docker stop node5

# Observar nos logs que uma eleicao ocorre e node4 assume como lider

# Recuperar o lider original
docker start node5

# node5 retoma lideranca automaticamente (Bully)

# Parar tudo
docker-compose down
```

## Cenarios de Demonstracao

### 1. Relogio de Lamport
- Observe nos logs os timestamps `[Clock: N]` avancando
- Quando um no envia uma mensagem, o timestamp e anexado
- Quando um no recebe uma mensagem, ajusta: `max(local, recebido) + 1`

### 2. Exclusao Mutua (Ricart-Agrawala)
- Multiplos nos solicitam a secao critica periodicamente
- Observe que apenas UM no por vez esta na secao critica
- Logs `SECAO CRITICA` nunca se sobrepoem entre nos

### 3. Eleicao de Lider (Bully)
- `docker stop node5` -> eleicao automatica -> node4 vira lider
- `docker start node5` -> node5 retoma lideranca
- `docker stop node4 node5` -> node3 vira lider

## Protocolo de Mensagens

| Tipo          | Algoritmo         | Descricao                            |
|---------------|-------------------|--------------------------------------|
| `REQUEST`     | Ricart-Agrawala   | Pedido para entrar na secao critica  |
| `REPLY`       | Ricart-Agrawala   | Permissao concedida                  |
| `ELECTION`    | Bully             | Inicio de eleicao                    |
| `OK`          | Bully             | "Estou vivo e tenho ID maior"        |
| `COORDINATOR` | Bully             | "Eu sou o novo lider"                |
| `HEARTBEAT`   | Bully             | Sinal periodico de vida do lider     |

## Referencias

1. Lamport, L. "Time, Clocks, and the Ordering of Events in a Distributed System." Communications of the ACM, 1978.
2. Ricart, G. e Agrawala, A. K. "An Optimal Algorithm for Mutual Exclusion in Computer Networks." Communications of the ACM, 1981.
3. Garcia-Molina, H. "Elections in a Distributed Computing System." IEEE Transactions on Computers, 1982.
4. Coulouris, G. et al. Distributed Systems: Concepts and Design. 5th Ed., 2011.
5. Tanenbaum, A. S., Van Steen, M. Distributed Systems: Principles and Paradigms. 3rd Ed., 2017.
6. Go Documentation - Effective Go. https://go.dev/doc/effective_go
