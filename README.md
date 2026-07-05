# Lab 2 - Algoritmos Distribuidos | MC714 - UNICAMP

Implementação de três algoritmos fundamentais de sistemas distribuídos usando Go e Docker:

1. **Relógio Lógico de Lamport** - Ordenação causal de eventos
2. **Exclusão Mútua (Ricart-Agrawala)** - Acesso exclusivo a recurso compartilhado
3. **Eleição de Líder (Bully)** - Eleição automática quando o líder falha

## Arquitetura

- **Linguagem:** Go 1.22+
- **Comunicação:** Sockets TCP com mensagens JSON
- **Ambiente:** Docker Compose com 5 containers (nodes) na mesma rede bridge
- **Processos:** 5 nós distribuídos (node1..node5), cada um rodando os 3 algoritmos

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
│   │   └── lamport.go          # Relógio lógico de Lamport
│   ├── mutex/
│   │   └── ricart_agrawala.go  # Exclusão mútua distribuída
│   ├── election/
│   │   └── bully.go            # Eleição de líder
│   ├── network/
│   │   ├── transport.go        # Camada de comunicação TCP/JSON
│   │   └── message.go          # Tipos de mensagem
│   └── config/
│       └── config.go           # Configuração via variáveis de ambiente
└── README.md
```

## Como Executar

### Pré-requisitos

- Docker e Docker Compose instalados

### Comandos

```bash
# Construir e iniciar todos os 5 nós
docker-compose up --build

# Em outro terminal, acompanhar logs de um nó específico
docker logs -f node1

# Simular falha do líder (node5)
docker stop node5

# Observar nos logs que uma eleição ocorre e node4 assume como líder

# Recuperar o líder original
docker start node5

# node5 retoma liderança automaticamente (Bully)

# Parar tudo
docker-compose down
```

## Cenários de Demonstração

### 1. Relógio de Lamport
- Observe nos logs os timestamps `[Clock: N]` avançando
- Quando um nó envia uma mensagem, o timestamp é anexado
- Quando um nó recebe uma mensagem, ajusta: `max(local, recebido) + 1`

### 2. Exclusão Mútua (Ricart-Agrawala)
- Múltiplos nós solicitam a seção crítica periodicamente
- Observe que apenas UM nó por vez está na seção crítica
- Logs `SECAO CRITICA` nunca se sobrepõem entre nós

### 3. Eleição de Líder (Bully)
- `docker stop node5` -> eleição automática -> node4 vira líder
- `docker start node5` -> node5 retoma liderança
- `docker stop node4 node5` -> node3 vira líder

## Protocolo de Mensagens

| Tipo          | Algoritmo         | Descrição                            |
|---------------|-------------------|--------------------------------------|
| `REQUEST`     | Ricart-Agrawala   | Pedido para entrar na seção crítica  |
| `REPLY`       | Ricart-Agrawala   | Permissão concedida                  |
| `ELECTION`    | Bully             | Início de eleição                    |
| `OK`          | Bully             | "Estou vivo e tenho ID maior"        |
| `COORDINATOR` | Bully             | "Eu sou o novo líder"                |
| `HEARTBEAT`   | Bully             | Sinal periódico de vida do líder     |

## Configuração

O comportamento dos nós pode ser customizado alterando as variáveis de ambiente no arquivo `docker-compose.yml`:

- `NODE_ID`: O identificador numérico único do nó (ex: 1, 2, 3...). O nó com maior ID tem prioridade na eleição (Algoritmo Bully).
- `TOTAL_NODES`: O número total de nós configurados na rede. Utilizado para estabelecer as conexões iniciais.
- O DNS interno do Docker resolve automaticamente os hostnames (`node1`, `node2`, etc.) para que os nós se encontrem na mesma rede.

## Observações de Implementação

- **Conexões**: Cada nó atua simultaneamente como servidor e cliente, estabelecendo conexões TCP diretas e persistentes com todos os outros nós.
- **Tratamento de Falhas**: No algoritmo Bully, se um *heartbeat* não é recebido dentro do tempo limite (*timeout*) ou se a conexão com o líder atual é perdida, uma nova eleição é disparada imediatamente pelo nó que detectou a falha.
- **Integração de Algoritmos**: O Relógio Lógico de Lamport é utilizado de forma integrada com o algoritmo de Exclusão Mútua de Ricart-Agrawala. Ele garante a ordenação causal dos pedidos `REQUEST`, assegurando que o acesso à seção crítica seja justo (*fairness*).
