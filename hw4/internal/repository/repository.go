package repository

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
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
	ordersFile   string
	txFile       string
}

func NewRepository() *Repository {
	return &Repository{
		orders:       []model.Order{},
		transactions: []model.Transaction{},
		ordersFile:   "orders.csv",
		txFile:       "transactions.csv",
	}
}

func (r *Repository) AddEntity(entity interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch v := entity.(type) {
	case model.Order:
		r.orders = append(r.orders, v)
		r.saveOrderToCSV(v)
		fmt.Println("Добавлен Order")
	case model.Transaction:
		r.transactions = append(r.transactions, v)
		r.saveTransactionToCSV(v)
		fmt.Println("Добавлена Transaction")
	default:
		fmt.Println("Неизвестный тип")
	}
}

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

func (r *Repository) saveOrderToCSV(order model.Order) {
	fileExists := true
	if _, err := os.Stat(r.ordersFile); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(r.ordersFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Ошибка открытия файла orders.csv: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		writer.Write([]string{"id", "status", "amount"})
	}

	statusStr := "false"
	if order.GetStatus() {
		statusStr = "true"
	}

	writer.Write([]string{
		strconv.Itoa(order.GetID()),
		statusStr,
		strconv.Itoa(order.GetAmount()),
	})
}

func (r *Repository) saveTransactionToCSV(tx model.Transaction) {
	fileExists := true
	if _, err := os.Stat(r.txFile); os.IsNotExist(err) {
		fileExists = false
	}

	file, err := os.OpenFile(r.txFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Ошибка открытия файла transactions.csv: %v\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		writer.Write([]string{"id", "amount", "date"})
	}

	writer.Write([]string{
		strconv.Itoa(tx.GetID()),
		strconv.Itoa(tx.GetAmount()),
		tx.GetDate(),
	})
}

func (r *Repository) LoadData() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.loadOrdersFromCSV(); err != nil {
		return fmt.Errorf("ошибка загрузки заказов: %v", err)
	}

	if err := r.loadTransactionsFromCSV(); err != nil {
		return fmt.Errorf("ошибка загрузки транзакций: %v", err)
	}

	return nil
}

func (r *Repository) loadOrdersFromCSV() error {
	file, err := os.Open(r.ordersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) <= 1 {
		return nil
	}

	var orders []model.Order
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 3 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		status, err := strconv.ParseBool(record[1])
		if err != nil {
			continue
		}

		amount, err := strconv.Atoi(record[2])
		if err != nil {
			continue
		}

		if id < 0 || amount <= 0 {
			continue
		}
		order := model.NewOrder(id, status, amount)
		orders = append(orders, order)
	}

	r.orders = orders
	return nil
}

func (r *Repository) loadTransactionsFromCSV() error {
	file, err := os.Open(r.txFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) <= 1 {
		return nil
	}

	var transactions []model.Transaction
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 3 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		amount, err := strconv.Atoi(record[1])
		if err != nil {
			continue
		}

		date := record[2]
		if date == "" {
			continue
		}

		if id < 0 || amount <= 0 {
			continue
		}
		tx := model.NewTransaction(id, amount, date)
		transactions = append(transactions, tx)
	}

	r.transactions = transactions
	return nil
}
