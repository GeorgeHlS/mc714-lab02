package clock

import "sync"

type LamportClock struct {
	mu   sync.Mutex
	time int
}

// cria um novo relogio logico inicializado em 0.
func NewLamportClock() *LamportClock {
	return &LamportClock{time: 0}
}

// incrementa o relogio para um evento local somando 1 ao relogio local
func (lc *LamportClock) Tick() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

// incrementa o relogio e retorna o timestamp para anexar na mensagem.
func (lc *LamportClock) SendTick() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.time++
	return lc.time
}

// ajusta o relogio ao receber uma mensagem
func (lc *LamportClock) ReceiveTick(receivedTimestamp int) int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	if receivedTimestamp > lc.time {
		lc.time = receivedTimestamp
	}
	lc.time++
	return lc.time
}

// retorna o valor atual do relogio
func (lc *LamportClock) Time() int {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	return lc.time
}
