package tests

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestE2E_DockerCompose(t *testing.T) {
	if testing.Short() {
		t.Skip("Pulando teste E2E pois -short foi passado")
	}

	t.Log("Subindo containers com docker-compose up --build...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// verifica se o docker-compose esta disponivel
	if _, err := exec.LookPath("docker-compose"); err != nil {
		t.Skip("docker-compose nao encontrado no PATH (provavelmente rodando dentro do Docker). Pulando teste E2E.")
	}

	// inicia os containers
	cmdUp := exec.CommandContext(ctx, "docker-compose", "up", "--build", "-d")
	cmdUp.Dir = ".."
	if output, err := cmdUp.CombinedOutput(); err != nil {
		t.Fatalf("Erro ao subir docker-compose: %v\nSaída:\n%s", err, string(output))
	}

	// garante que o ambiente seja limpo ao final
	defer func() {
		t.Log("Desligando containers...")
		cmdDown := exec.Command("docker-compose", "down")
		cmdDown.Dir = ".."
		cmdDown.Run()
	}()

	// aguarda os sistemas iniciarem e gerarem alguns logs
	t.Log("Aguardando 15 segundos para geracao de logs...")
	time.Sleep(15 * time.Second)

	// captura os logs de todos os containers
	cmdLogs := exec.Command("docker-compose", "logs")
	cmdLogs.Dir = ".."
	var out bytes.Buffer
	cmdLogs.Stdout = &out
	if err := cmdLogs.Run(); err != nil {
		t.Fatalf("Erro ao buscar logs: %v", err)
	}

	logs := out.String()

	// Validacao 1: Ao menos algum log de seção crítica
	if !strings.Contains(logs, "SECAO CRITICA") {
		t.Errorf("Esperava encontrar logs de SECAO CRITICA, mas nao foram gerados")
	}

	// Validacao 2: Eleição funcionou (Leader inicial assumiu)
	if !strings.Contains(logs, "Sou o lider inicial, iniciando eleicao para confirmar") {
		t.Errorf("Esperava que o Node 5 iniciasse como lider")
	}

	// Validacao 3: Relogio de Lamport presente
	if !strings.Contains(logs, "LAMPORT") {
		t.Errorf("Esperava encontrar logs do Relogio de Lamport")
	}
}
