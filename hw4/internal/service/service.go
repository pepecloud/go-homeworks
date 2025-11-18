package service

import (
	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
)

// Функция, которая создаёт структуры и передаёт в репозиторий
func ProcessEntities(repo *repository.Repository) {
	order1 := model.NewOrder(1, true, 100)
	order2 := model.NewOrder(2, false, 1000)

	// Создаём несколько транзакций
	tx1 := model.NewTransaction(100, 500, "2025-11-08")
	tx2 := model.NewTransaction(100, 500, "2025-11-19")

	// Передаём их в репозиторий
	repo.AddEntity(order1)
	repo.AddEntity(order2)
	repo.AddEntity(tx1)
	repo.AddEntity(tx2)
}
