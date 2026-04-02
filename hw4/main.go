// API заказов и транзакций: CRUD по /api/item, авторизация по JWT.
//
//	@title          Items API
//	@version        1.0
//	@description    CRUD для заказов и транзакций. Мутирующие запросы требуют JWT (получить через POST /api/login).
//	@host           localhost:8080
//	@BasePath       /
//	@securityDefinitions.apikey  BearerAuth
//	@in   header
//	@name Authorization
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
	"github.com/joho/godotenv"
	_ "github.com/pepecloud/go-homeworks/hw4/docs"
	"github.com/pepecloud/go-homeworks/hw4/internal/grpc"
	"github.com/pepecloud/go-homeworks/hw4/internal/handlers"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
	"github.com/pepecloud/go-homeworks/hw4/internal/service"
	"github.com/pepecloud/go-homeworks/hw4/internal/usecase"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	_ = godotenv.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	repo, err := repository.NewRepository(ctx)
	if err != nil {
		log.Fatalf("Ошибка инициализации репозитория: %v", err)
	}
	defer func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := repo.Close(closeCtx); err != nil {
			log.Printf("Ошибка закрытия репозитория: %v\n", err)
		}
	}()

	if err := repo.LoadData(); err != nil {
		fmt.Printf("Ошибка загрузки данных: %v\n", err)
	}

	itemsUsecase := usecase.New(repo)
	h := handlers.NewHandlers(itemsUsecase)
	router := mux.NewRouter()

	router.HandleFunc("/api/login", h.Login).Methods("POST")
	router.HandleFunc("/api/item", h.CreateItem).Methods("POST")
	router.HandleFunc("/api/item/{id}", h.UpdateItem).Methods("PUT")
	router.HandleFunc("/api/items", h.GetItems).Methods("GET")
	router.HandleFunc("/api/item/{id}", h.GetItem).Methods("GET")
	router.HandleFunc("/api/item/{id}", h.DeleteItem).Methods("DELETE")

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

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

	// Запуск gRPC-сервера в отдельной горутине
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := grpc.StartGRPCServer(ctx, itemsUsecase, ":9090"); err != nil {
			log.Printf("Ошибка gRPC-сервера: %v\n", err)
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
	orders, err := repo.GetOrders()
	if err != nil {
		fmt.Printf("Ошибка чтения заказов: %v\n", err)
	} else {
		fmt.Printf("Всего заказов: %d\n", len(orders))
	}

	transactions, err := repo.GetTransactions()
	if err != nil {
		fmt.Printf("Ошибка чтения транзакций: %v\n", err)
	} else {
		fmt.Printf("Всего транзакций: %d\n", len(transactions))
	}
}
