package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
	"github.com/pepecloud/go-homeworks/hw4/internal/service"
)

func main() {
	repo := repository.NewRepository()

	// Канал для передачи данных от генератора к потребителю
	dataCh := make(chan interface{}, 10)

	// Используем WaitGroup, чтобы дождаться обработки всех данных
	var wg sync.WaitGroup
	wg.Add(1)

	// 1. Запускаем функцию-потребителя (Consumer) в горутине
	// Она будет читать из канала, пока он не закроется
	go func() {
		defer wg.Done()
		service.ConsumeEntities(repo, dataCh)
	}()

	// 2. Запускаем логгер в горутине
	// Он работает независимо с интервалом 200мс
	go service.RunLogger(repo)

	// 3. Запускаем функцию-генератора (Producer) в горутине
	go func() {
		service.GenerateEntities(dataCh)
		// После отправки всех данных закрываем канал
		close(dataCh)
	}()

	// Ждем, пока Consumer обработает все сообщения и выйдет из цикла (после закрытия канала)
	wg.Wait()

	// Небольшая пауза, чтобы логгер успел дописать последние изменения, если они были только что добавлены
	time.Sleep(300 * time.Millisecond)

	// Вывод итоговое содержимое репозитория
	fmt.Println("\n=== Итого в репозитории ===")
	fmt.Printf("Всего заказов: %d\n", len(repo.GetOrders()))
	fmt.Printf("Всего транзакций: %d\n", len(repo.GetTransactions()))
}
