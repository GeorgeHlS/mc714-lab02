package clock

import (
	"sync"
	"testing"
)

func TestLamportClock_Tick(t *testing.T) {
	clk := NewLamportClock()

	if clk.Time() != 0 {
		t.Errorf("Tempo inicial esperado 0, obtido %d", clk.Time())
	}

	ts := clk.Tick()
	if ts != 1 || clk.Time() != 1 {
		t.Errorf("Tempo esperado apos Tick() = 1, obtido %d", ts)
	}

	ts = clk.Tick()
	if ts != 2 || clk.Time() != 2 {
		t.Errorf("Tempo esperado apos segundo Tick() = 2, obtido %d", ts)
	}
}

func TestLamportClock_SendTick(t *testing.T) {
	clk := NewLamportClock()
	clk.Tick()

	ts := clk.SendTick()
	if ts != 2 {
		t.Errorf("Tempo esperado apos SendTick() = 2, obtido %d", ts)
	}
}

func TestLamportClock_ReceiveTick(t *testing.T) {
	clk := NewLamportClock()
	clk.Tick()

	ts := clk.ReceiveTick(5)
	if ts != 6 {
		t.Errorf("Tempo esperado apos receber msg com ts=5 eh 6, obtido %d", ts)
	}

	ts = clk.ReceiveTick(2)
	if ts != 7 {
		t.Errorf("Tempo esperado apos receber msg com ts=2 eh 7, obtido %d", ts)
	}
}

func TestLamportClock_Concurrent(t *testing.T) {
	clk := NewLamportClock()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clk.Tick()
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clk.ReceiveTick(10)
		}()
	}

	wg.Wait()
	if clk.Time() < 200 {
		t.Errorf("Tempo esperado minimo apos concorrência = 200, obtido %d", clk.Time())
	}
}
