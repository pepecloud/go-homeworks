package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
)

type Handlers struct {
	repo *repository.Repository
}

func NewHandlers(repo *repository.Repository) *Handlers {
	return &Handlers{repo: repo}
}

// Универсальные handlers для /api/item

func (h *Handlers) CreateItem(w http.ResponseWriter, r *http.Request) {
	// Пытаемся определить тип сущности по JSON
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Если есть поле "date", это Transaction, иначе Order
	if _, hasDate := raw["date"]; hasDate {
		h.createTransactionFromMap(w, raw)
	} else {
		h.createOrderFromMap(w, raw)
	}
}

func (h *Handlers) createOrderFromMap(w http.ResponseWriter, raw map[string]interface{}) {
	var dto OrderDTO

	// Извлечение и проверка полей
	idVal, ok := raw["id"]
	if !ok {
		http.Error(w, "Поле 'id' обязательно", http.StatusBadRequest)
		return
	}
	idFloat, ok := idVal.(float64)
	if !ok {
		http.Error(w, "Поле 'id' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.ID = int(idFloat)

	statusVal, ok := raw["status"]
	if !ok {
		http.Error(w, "Поле 'status' обязательно", http.StatusBadRequest)
		return
	}
	dto.Status, ok = statusVal.(bool)
	if !ok {
		http.Error(w, "Поле 'status' должно быть булевым", http.StatusBadRequest)
		return
	}

	amountVal, ok := raw["amount"]
	if !ok {
		http.Error(w, "Поле 'amount' обязательно", http.StatusBadRequest)
		return
	}
	amountFloat, ok := amountVal.(float64)
	if !ok {
		http.Error(w, "Поле 'amount' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.Amount = int(amountFloat)

	// Валидация обязательных полей
	if dto.ID < 0 {
		http.Error(w, "ID не может быть меньше 0", http.StatusBadRequest)
		return
	}
	if dto.Amount <= 0 {
		http.Error(w, "Amount должен быть больше 0", http.StatusBadRequest)
		return
	}

	// Проверяем, не существует ли уже заказ с таким ID
	if existing := h.repo.GetOrderByID(dto.ID); existing != nil {
		http.Error(w, "Заказ с таким ID уже существует", http.StatusConflict)
		return
	}

	order := model.NewOrder(dto.ID, dto.Status, dto.Amount)
	if order.GetID() != dto.ID {
		http.Error(w, "Ошибка создания заказа: невалидные данные", http.StatusBadRequest)
		return
	}

	h.repo.AddEntity(order)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

func (h *Handlers) createTransactionFromMap(w http.ResponseWriter, raw map[string]interface{}) {
	var dto TransactionDTO

	// Извлечение и проверка полей
	idVal, ok := raw["id"]
	if !ok {
		http.Error(w, "Поле 'id' обязательно", http.StatusBadRequest)
		return
	}
	idFloat, ok := idVal.(float64)
	if !ok {
		http.Error(w, "Поле 'id' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.ID = int(idFloat)

	amountVal, ok := raw["amount"]
	if !ok {
		http.Error(w, "Поле 'amount' обязательно", http.StatusBadRequest)
		return
	}
	amountFloat, ok := amountVal.(float64)
	if !ok {
		http.Error(w, "Поле 'amount' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.Amount = int(amountFloat)

	dateVal, ok := raw["date"]
	if !ok {
		http.Error(w, "Поле 'date' обязательно", http.StatusBadRequest)
		return
	}
	dto.Date, ok = dateVal.(string)
	if !ok {
		http.Error(w, "Поле 'date' должно быть строкой", http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if dto.ID < 0 {
		http.Error(w, "ID не может быть меньше 0", http.StatusBadRequest)
		return
	}
	if dto.Amount <= 0 {
		http.Error(w, "Amount должен быть больше 0", http.StatusBadRequest)
		return
	}
	if dto.Date == "" {
		http.Error(w, "Date обязателен для заполнения", http.StatusBadRequest)
		return
	}

	// Проверяем, не существует ли уже транзакция с таким ID
	if existing := h.repo.GetTransactionByID(dto.ID); existing != nil {
		http.Error(w, "Транзакция с таким ID уже существует", http.StatusConflict)
		return
	}

	tx := model.NewTransaction(dto.ID, dto.Amount, dto.Date)
	if tx.GetID() != dto.ID {
		http.Error(w, "Ошибка создания транзакции: невалидные данные", http.StatusBadRequest)
		return
	}

	h.repo.AddEntity(tx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

func (h *Handlers) UpdateItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	// Пытаемся определить тип сущности по JSON
	var raw map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		http.Error(w, "Неверный формат JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Проверяем, существует ли Order или Transaction с таким ID
	order := h.repo.GetOrderByID(id)
	tx := h.repo.GetTransactionByID(id)

	if order != nil {
		h.updateOrderFromMap(w, id, raw)
	} else if tx != nil {
		h.updateTransactionFromMap(w, id, raw)
	} else {
		http.Error(w, "Сущность с таким ID не найдена", http.StatusNotFound)
		return
	}
}

func (h *Handlers) updateOrderFromMap(w http.ResponseWriter, id int, raw map[string]interface{}) {
	var dto OrderDTO

	// Извлечение и проверка полей
	idVal, ok := raw["id"]
	if !ok {
		http.Error(w, "Поле 'id' обязательно", http.StatusBadRequest)
		return
	}
	idFloat, ok := idVal.(float64)
	if !ok {
		http.Error(w, "Поле 'id' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.ID = int(idFloat)

	statusVal, ok := raw["status"]
	if !ok {
		http.Error(w, "Поле 'status' обязательно", http.StatusBadRequest)
		return
	}
	dto.Status, ok = statusVal.(bool)
	if !ok {
		http.Error(w, "Поле 'status' должно быть булевым", http.StatusBadRequest)
		return
	}

	amountVal, ok := raw["amount"]
	if !ok {
		http.Error(w, "Поле 'amount' обязательно", http.StatusBadRequest)
		return
	}
	amountFloat, ok := amountVal.(float64)
	if !ok {
		http.Error(w, "Поле 'amount' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.Amount = int(amountFloat)

	// Валидация обязательных полей
	if dto.ID < 0 {
		http.Error(w, "ID не может быть меньше 0", http.StatusBadRequest)
		return
	}
	if dto.Amount <= 0 {
		http.Error(w, "Amount должен быть больше 0", http.StatusBadRequest)
		return
	}

	if dto.ID != id {
		http.Error(w, "ID в пути не совпадает с ID в теле запроса", http.StatusBadRequest)
		return
	}

	order := model.NewOrder(dto.ID, dto.Status, dto.Amount)
	if order.GetID() != dto.ID {
		http.Error(w, "Ошибка создания заказа: невалидные данные", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateOrder(id, order); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

func (h *Handlers) updateTransactionFromMap(w http.ResponseWriter, id int, raw map[string]interface{}) {
	var dto TransactionDTO

	// Извлечение и проверка полей
	idVal, ok := raw["id"]
	if !ok {
		http.Error(w, "Поле 'id' обязательно", http.StatusBadRequest)
		return
	}
	idFloat, ok := idVal.(float64)
	if !ok {
		http.Error(w, "Поле 'id' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.ID = int(idFloat)

	amountVal, ok := raw["amount"]
	if !ok {
		http.Error(w, "Поле 'amount' обязательно", http.StatusBadRequest)
		return
	}
	amountFloat, ok := amountVal.(float64)
	if !ok {
		http.Error(w, "Поле 'amount' должно быть числом", http.StatusBadRequest)
		return
	}
	dto.Amount = int(amountFloat)

	dateVal, ok := raw["date"]
	if !ok {
		http.Error(w, "Поле 'date' обязательно", http.StatusBadRequest)
		return
	}
	dto.Date, ok = dateVal.(string)
	if !ok {
		http.Error(w, "Поле 'date' должно быть строкой", http.StatusBadRequest)
		return
	}

	// Валидация обязательных полей
	if dto.ID < 0 {
		http.Error(w, "ID не может быть меньше 0", http.StatusBadRequest)
		return
	}
	if dto.Amount <= 0 {
		http.Error(w, "Amount должен быть больше 0", http.StatusBadRequest)
		return
	}
	if dto.Date == "" {
		http.Error(w, "Date обязателен для заполнения", http.StatusBadRequest)
		return
	}

	if dto.ID != id {
		http.Error(w, "ID в пути не совпадает с ID в теле запроса", http.StatusBadRequest)
		return
	}

	tx := model.NewTransaction(dto.ID, dto.Amount, dto.Date)
	if tx.GetID() != dto.ID {
		http.Error(w, "Ошибка создания транзакции: невалидные данные", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateTransaction(id, tx); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

func (h *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	orders := h.repo.GetOrders()
	transactions := h.repo.GetTransactions()

	// Объединяем все сущности в один массив
	items := make([]interface{}, 0, len(orders)+len(transactions))

	for _, order := range orders {
		items = append(items, OrderDTO{
			ID:     order.GetID(),
			Status: order.GetStatus(),
			Amount: order.GetAmount(),
		})
	}

	for _, tx := range transactions {
		items = append(items, TransactionDTO{
			ID:     tx.GetID(),
			Amount: tx.GetAmount(),
			Date:   tx.GetDate(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	// Проверяем сначала Order, потом Transaction
	if order := h.repo.GetOrderByID(id); order != nil {
		dto := OrderDTO{
			ID:     order.GetID(),
			Status: order.GetStatus(),
			Amount: order.GetAmount(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dto)
		return
	}

	if tx := h.repo.GetTransactionByID(id); tx != nil {
		dto := TransactionDTO{
			ID:     tx.GetID(),
			Amount: tx.GetAmount(),
			Date:   tx.GetDate(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dto)
		return
	}

	http.Error(w, "Сущность не найдена", http.StatusNotFound)
}

func (h *Handlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	// Пытаемся удалить сначала Order, потом Transaction
	if err := h.repo.DeleteOrder(id); err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.repo.DeleteTransaction(id); err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Сущность не найдена", http.StatusNotFound)
}
