package service

import (
	"context"
	"fmt"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
)

// Функция, которая генерирует данные и отправляет их в канал
func GenerateEntities(ctx context.Context, ch chan<- interface{}) {
	for i := 1; i <= 5; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
			return
		default:
			order := model.NewOrder(i, i%2 == 0, 100*i)
			select {
			case ch <- order:
			case <-ctx.Done():
				fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
				return
			}
			select {
			case <-time.After(100 * time.Millisecond):
			case <-ctx.Done():
				fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
				return
			}
		}
	}

	select {
	case <-ctx.Done():
		fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
		return
	default:
	}

	tx1 := model.NewTransaction(100, 500, "2025-11-08")
	select {
	case ch <- tx1:
	case <-ctx.Done():
		fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
		return
	}

	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
		return
	}

	select {
	case <-ctx.Done():
		fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
		return
	default:
	}

	tx2 := model.NewTransaction(101, 600, "2025-11-19")
	select {
	case ch <- tx2:
	case <-ctx.Done():
		fmt.Println("[GENERATOR] Получен сигнал завершения, прекращаем генерацию")
		return
	}
}

// Функция, которая читает из канала и сохраняет в репозиторий
func ConsumeEntities(ctx context.Context, repo *repository.Repository, ch <-chan interface{}) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("[CONSUMER] Получен сигнал завершения, прекращаем обработку")
			return
		case entity, ok := <-ch:
			if !ok {
				fmt.Println("[CONSUMER] Канал закрыт, завершаем обработку")
				return
			}
			repo.AddEntity(entity)
		}
	}
}

// Логгер, работающий в отдельной горутине
func RunLogger(ctx context.Context, repo *repository.Repository) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var prevOrderCount, prevTxCount int

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[LOGGER] Получен сигнал завершения, останавливаем логгер")
			return
		case <-ticker.C:
			orders := repo.GetOrders()
			currentOrderCount := len(orders)

			if currentOrderCount > prevOrderCount {
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
}
