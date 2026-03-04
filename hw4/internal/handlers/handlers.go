package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pepecloud/go-homeworks/hw4/internal/usecase"
)

type Handlers struct {
	items *usecase.Service
}

func NewHandlers(items *usecase.Service) *Handlers {
	return &Handlers{items: items}
}

// CreateItem создаёт заказ или транзакцию по телу запроса (тип по полю date).
// CreateItem godoc
// @Summary      Создать сущность
// @Description  Создаёт Order или Transaction по JSON. Если есть поле "date" — транзакция, иначе заказ. Требуется JWT.
// @Tags         items
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  object  true  "Order: id, status, amount. Transaction: id, amount, date"
// @Success      201   {object}  object
// @Failure      400   {object}  object
// @Failure      401   {object}  object  "Требуется авторизация"
// @Failure      409   {object}  object  "Сущность с таким ID уже есть"
// @Router       /api/item [post]
func (h *Handlers) CreateItem(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
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

	if _, err := h.items.CreateOrder(dto.ID, dto.Status, dto.Amount); err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderExists):
			http.Error(w, "Заказ с таким ID уже существует", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

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

	if _, err := h.items.CreateTransaction(dto.ID, dto.Amount, dto.Date); err != nil {
		switch {
		case errors.Is(err, usecase.ErrTransactionExists):
			http.Error(w, "Транзакция с таким ID уже существует", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto)
}

// UpdateItem обновляет заказ или транзакцию по id в пути.
// UpdateItem godoc
// @Summary      Обновить сущность
// @Description  Обновляет Order или Transaction по id. Тело — как при создании. Требуется JWT.
// @Tags         items
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int     true  "ID сущности"
// @Param        body  body      object  true  "Order: id, status, amount. Transaction: id, amount, date"
// @Success      200   {object}  object
// @Failure      400   {object}  object
// @Failure      401   {object}  object  "Требуется авторизация"
// @Failure      404   {object}  object
// @Router       /api/item/{id} [put]
func (h *Handlers) UpdateItem(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
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
	order, _ := h.items.GetOrder(id)
	tx, _ := h.items.GetTransaction(id)

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

	if _, err := h.items.UpdateOrder(id, dto.ID, dto.Status, dto.Amount); err != nil {
		switch {
		case errors.Is(err, usecase.ErrEntityNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
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

	if _, err := h.items.UpdateTransaction(id, dto.ID, dto.Amount, dto.Date); err != nil {
		switch {
		case errors.Is(err, usecase.ErrEntityNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto)
}

// GetItems возвращает все заказы и транзакции одним списком.
// GetItems godoc
// @Summary      Список сущностей
// @Description  Возвращает все Order и Transaction в одном массиве.
// @Tags         items
// @Produce      json
// @Success      200   {array}   object
// @Router       /api/items [get]
func (h *Handlers) GetItems(w http.ResponseWriter, r *http.Request) {
	orders, transactions := h.items.ListItems()

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

// GetItem возвращает одну сущность по id (заказ или транзакция).
// GetItem godoc
// @Summary      Получить сущность по ID
// @Description  Ищет Order или Transaction по id в пути.
// @Tags         items
// @Produce      json
// @Param        id   path      int  true  "ID сущности"
// @Success      200  {object}  object
// @Failure      400  {object}  object
// @Failure      404  {object}  object
// @Router       /api/item/{id} [get]
func (h *Handlers) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}

	// Проверяем сначала Order, потом Transaction
	if order, _ := h.items.GetOrder(id); order != nil {
		dto := OrderDTO{
			ID:     order.GetID(),
			Status: order.GetStatus(),
			Amount: order.GetAmount(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(dto)
		return
	}

	if tx, _ := h.items.GetTransaction(id); tx != nil {
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

// DeleteItem удаляет заказ или транзакцию по id. Требуется JWT.
// DeleteItem godoc
// @Summary      Удалить сущность
// @Description  Удаляет Order или Transaction по id в пути. Требуется авторизация.
// @Tags         items
// @Security     BearerAuth
// @Param        id   path  int  true  "ID сущности"
// @Success      204
// @Failure      400  {object}  object
// @Failure      401  {object}  object  "Требуется авторизация"
// @Failure      404  {object}  object
// @Router       /api/item/{id} [delete]
func (h *Handlers) DeleteItem(w http.ResponseWriter, r *http.Request) {
	if !h.checkAuth(w, r) {
		return
	}
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Неверный формат ID", http.StatusBadRequest)
		return
	}
	if err := h.items.DeleteItem(id); err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Сущность не найдена", http.StatusNotFound)
}
