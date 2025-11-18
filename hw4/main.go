package main

import (
	"fmt"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
	"github.com/pepecloud/go-homeworks/hw4/internal/service"
)

func main() {
	repo := repository.NewRepository()

	// Запуск по интервалу (например, каждые 2 секунды)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for i := 0; i < 3; i++ { // для примера 3 раза
		service.ProcessEntities(repo)
		<-ticker.C
	}

	// Вывод итоговое содержимое репозитория
	fmt.Println("\n=== Итого в репозитории ===")
	fmt.Printf("Всего заказов: %d\n", len(repo.GetOrders()))
	fmt.Printf("Всего транзакций: %d\n", len(repo.GetTransactions()))

}
