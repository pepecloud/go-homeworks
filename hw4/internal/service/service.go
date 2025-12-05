package service

import (
	"fmt"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
)

// Функция, которая генерирует данные и отправляет их в канал
func GenerateEntities(ch chan<- interface{}) {

	// Пример: отправляем несколько заказов и транзакций с паузами
	for i := 1; i <= 5; i++ {
		order := model.NewOrder(i, i%2 == 0, 100*i)
		ch <- order
		time.Sleep(100 * time.Millisecond) // Имитация задержки
	}

	tx1 := model.NewTransaction(100, 500, "2025-11-08")
	ch <- tx1
	time.Sleep(100 * time.Millisecond)

	tx2 := model.NewTransaction(101, 600, "2025-11-19")
	ch <- tx2
}

// Функция, которая читает из канала и сохраняет в репозиторий
func ConsumeEntities(repo *repository.Repository, ch <-chan interface{}) {
	for entity := range ch {
		repo.AddEntity(entity)
	}
}

// Логгер, работающий в отдельной горутине
func RunLogger(repo *repository.Repository) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var prevOrderCount, prevTxCount int

	for range ticker.C {
		// Получаем текущие срезы (под капотом RLock)
		orders := repo.GetOrders()
		currentOrderCount := len(orders)

		if currentOrderCount > prevOrderCount {
			// Есть новые заказы
			newOrders := orders[prevOrderCount:]
			for _, o := range newOrders {
				fmt.Printf("[LOGGER] Новый заказ: %+v\n", o)
			}
			prevOrderCount = currentOrderCount
		}

		transactions := repo.GetTransactions()
		currentTxCount := len(transactions)

		if currentTxCount > prevTxCount {
			// Есть новые транзакции
			newTransactions := transactions[prevTxCount:]
			for _, t := range newTransactions {
				fmt.Printf("[LOGGER] Новая транзакция: %+v\n", t)
			}
			prevTxCount = currentTxCount
		}
	}
}
