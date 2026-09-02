// Package mailer represents the services to sending emails.
package mailer

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// emailJob é a estrutura interna que trafega no canal
type emailJob struct {
	ToEmail string
	Token   string
}

type workerMailer struct {
	jobChan chan emailJob
	quit    chan struct{}
	wg      sync.WaitGroup
}

// NewWorkerMailer inicializa o pool de trabalhadores e a fila de mensagens.
func NewWorkerMailer(maxWorkers int, bufferSize int) Mailer {
	m := &workerMailer{
		jobChan: make(chan emailJob, bufferSize),
		quit:    make(chan struct{}),
	}

	// Inicia os workers
	for i := 1; i <= maxWorkers; i++ {
		m.wg.Add(1)
		go m.startWorker(i)
	}

	log.Printf("[Mailer] Worker pool iniciado com %d trabalhadores", maxWorkers)
	return m
}

// startWorker é o loop que consome a fila de e-mails
func (m *workerMailer) startWorker(workerID int) {
	defer m.wg.Done() // Avisa o WaitGroup quando o worker for encerrado

	for {
		select {
		case job := <-m.jobChan:
			link := fmt.Sprintf("https://seu-site.com.br/redefinir-senha?token=%s", job.Token)
			log.Printf("[Worker %d] Enviando e-mail para %s | Link: %s", workerID, job.ToEmail, link)

		case <-m.quit:
			// Se recebermos o sinal de quit, o worker encerra seu loop
			log.Printf("[Worker %d] Encerrando atividades...", workerID)
			return
		}
	}
}

// SendPasswordReset enfileira a solicitação no canal
func (m *workerMailer) SendPasswordReset(ctx context.Context, toEmail, token string) error {
	job := emailJob{
		ToEmail: toEmail,
		Token:   token,
	}

	select {
	case m.jobChan <- job:
		return nil
	case <-m.quit:
		// Previne que o sistema tente adicionar na fila durante um desligamento
		return fmt.Errorf("o serviço de e-mail está sendo encerrado")
	}
}

// Close bloqueia a execução até que todos os e-mails na fila sejam processados
func (m *workerMailer) Close() error {
	close(m.quit)
	m.wg.Wait()
	close(m.jobChan)
	return nil
}
