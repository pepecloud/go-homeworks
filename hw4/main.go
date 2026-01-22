package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/pepecloud/go-homeworks/hw4/internal/handlers"
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

	// Настройка веб-сервера
	h := handlers.NewHandlers(repo)
	router := mux.NewRouter()

	// Универсальные маршруты для всех сущностей
	router.HandleFunc("/api/item", h.CreateItem).Methods("POST")
	router.HandleFunc("/api/item/{id}", h.UpdateItem).Methods("PUT")
	router.HandleFunc("/api/items", h.GetItems).Methods("GET")
	router.HandleFunc("/api/item/{id}", h.GetItem).Methods("GET")
	router.HandleFunc("/api/item/{id}", h.DeleteItem).Methods("DELETE")

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Запуск веб-сервера в отдельной горутине
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("Веб-сервер запущен на http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Ошибка веб-сервера: %v\n", err)
		}
	}()

	dataCh := make(chan interface{}, 10)

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

	// Graceful shutdown веб-сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Ошибка при остановке веб-сервера: %v\n", err)
	}

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
