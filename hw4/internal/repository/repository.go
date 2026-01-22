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

func (r *Repository) GetOrderByID(id int) *model.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.orders {
		if r.orders[i].GetID() == id {
			return &r.orders[i]
		}
	}
	return nil
}

func (r *Repository) GetTransactionByID(id int) *model.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := range r.transactions {
		if r.transactions[i].GetID() == id {
			return &r.transactions[i]
		}
	}
	return nil
}

func (r *Repository) UpdateOrder(id int, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.orders {
		if r.orders[i].GetID() == id {
			r.orders[i] = order
			return r.saveAllOrdersToCSV()
		}
	}
	return fmt.Errorf("заказ с id %d не найден", id)
}

func (r *Repository) UpdateTransaction(id int, tx model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.transactions {
		if r.transactions[i].GetID() == id {
			r.transactions[i] = tx
			return r.saveAllTransactionsToCSV()
		}
	}
	return fmt.Errorf("транзакция с id %d не найдена", id)
}

func (r *Repository) DeleteOrder(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.orders {
		if r.orders[i].GetID() == id {
			r.orders = append(r.orders[:i], r.orders[i+1:]...)
			return r.saveAllOrdersToCSV()
		}
	}
	return fmt.Errorf("заказ с id %d не найден", id)
}

func (r *Repository) DeleteTransaction(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.transactions {
		if r.transactions[i].GetID() == id {
			r.transactions = append(r.transactions[:i], r.transactions[i+1:]...)
			return r.saveAllTransactionsToCSV()
		}
	}
	return fmt.Errorf("транзакция с id %d не найдена", id)
}

func (r *Repository) saveAllOrdersToCSV() error {
	file, err := os.Create(r.ordersFile)
	if err != nil {
		return fmt.Errorf("ошибка создания файла orders.csv: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"id", "status", "amount"})

	for _, order := range r.orders {
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

	return nil
}

func (r *Repository) saveAllTransactionsToCSV() error {
	file, err := os.Create(r.txFile)
	if err != nil {
		return fmt.Errorf("ошибка создания файла transactions.csv: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"id", "amount", "date"})

	for _, tx := range r.transactions {
		writer.Write([]string{
			strconv.Itoa(tx.GetID()),
			strconv.Itoa(tx.GetAmount()),
			tx.GetDate(),
		})
	}

	return nil
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
