package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
	"github.com/pepecloud/go-homeworks/hw4/internal/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	repo := repository.NewRepository()
	if err := repo.LoadData(); err != nil {
		fmt.Printf("Ошибка загрузки данных: %v\n", err)
	}

	dataCh := make(chan interface{}, 10)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		service.ConsumeEntities(ctx, repo, dataCh)
	}()

	go func() {
		defer wg.Done()
		service.RunLogger(ctx, repo)
	}()

	go func() {
		defer wg.Done()
		defer close(dataCh)
		service.GenerateEntities(ctx, dataCh)
	}()

	<-sigChan
	fmt.Println("\n[MAIN] Получен сигнал завершения, начинаем graceful shutdown...")

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Println("[MAIN] Все горутины завершены корректно")
	case <-time.After(5 * time.Second):
		fmt.Println("[MAIN] Таймаут ожидания завершения горутин")
	}

	fmt.Println("\n=== Итого в репозитории ===")
	fmt.Printf("Всего заказов: %d\n", len(repo.GetOrders()))
	fmt.Printf("Всего транзакций: %d\n", len(repo.GetTransactions()))
}
