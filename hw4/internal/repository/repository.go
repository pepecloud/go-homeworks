package repository

import (
	"fmt"
	"sync"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
)

type Entity interface {
	GetID() int
}

type Repository struct {
	orders       []model.Order
	transactions []model.Transaction
	mu           sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{
		orders:       []model.Order{},
		transactions: []model.Transaction{},
	}
}

func (r *Repository) AddEntity(entity interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch v := entity.(type) {
	case model.Order:
		r.orders = append(r.orders, v)
		fmt.Println("Добавлен Order")
	case model.Transaction:
		r.transactions = append(r.transactions, v)
		fmt.Println("Добавлена Transaction")
	default:
		fmt.Println("Неизвестный тип")
	}
}

// Методы для получения данных
func (r *Repository) GetOrders() []model.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.orders
}

func (r *Repository) GetTransactions() []model.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.transactions
}
