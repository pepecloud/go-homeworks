package usecase

import (
	"errors"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
)

var (
	ErrInvalidID         = errors.New("id не может быть меньше 0")
	ErrInvalidAmount     = errors.New("amount должен быть больше 0")
	ErrInvalidDate       = errors.New("date обязателен для заполнения")
	ErrOrderExists       = errors.New("заказ с таким ID уже существует")
	ErrTransactionExists = errors.New("транзакция с таким ID уже существует")
	ErrEntityNotFound    = errors.New("сущность с таким ID не найдена")
	ErrIDMismatch        = errors.New("ID в пути не совпадает с ID в теле запроса")
	ErrInvalidOrderData  = errors.New("ошибка создания заказа: невалидные данные")
	ErrInvalidTxData     = errors.New("ошибка создания транзакции: невалидные данные")
)

type Repository interface {
	AddEntity(entity interface{}) error
	GetOrders() []model.Order
	GetTransactions() []model.Transaction
	GetOrderByID(id int) *model.Order
	GetTransactionByID(id int) *model.Transaction
	UpdateOrder(id int, order model.Order) error
	UpdateTransaction(id int, tx model.Transaction) error
	DeleteOrder(id int) error
	DeleteTransaction(id int) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateOrder(id int, status bool, amount int) (model.Order, error) {
	if id < 0 {
		return model.Order{}, ErrInvalidID
	}
	if amount <= 0 {
		return model.Order{}, ErrInvalidAmount
	}
	if existing := s.repo.GetOrderByID(id); existing != nil {
		return model.Order{}, ErrOrderExists
	}

	order := model.NewOrder(id, status, amount)
	if order.GetID() != id {
		return model.Order{}, ErrInvalidOrderData
	}

	if err := s.repo.AddEntity(order); err != nil {
		return model.Order{}, err
	}
	return order, nil
}

func (s *Service) CreateTransaction(id, amount int, date string) (model.Transaction, error) {
	if id < 0 {
		return model.Transaction{}, ErrInvalidID
	}
	if amount <= 0 {
		return model.Transaction{}, ErrInvalidAmount
	}
	if date == "" {
		return model.Transaction{}, ErrInvalidDate
	}
	if existing := s.repo.GetTransactionByID(id); existing != nil {
		return model.Transaction{}, ErrTransactionExists
	}

	tx := model.NewTransaction(id, amount, date)
	if tx.GetID() != id {
		return model.Transaction{}, ErrInvalidTxData
	}

	if err := s.repo.AddEntity(tx); err != nil {
		return model.Transaction{}, err
	}
	return tx, nil
}

func (s *Service) UpdateOrder(pathID, payloadID int, status bool, amount int) (model.Order, error) {
	if payloadID < 0 {
		return model.Order{}, ErrInvalidID
	}
	if amount <= 0 {
		return model.Order{}, ErrInvalidAmount
	}
	if pathID != payloadID {
		return model.Order{}, ErrIDMismatch
	}

	order := model.NewOrder(payloadID, status, amount)
	if order.GetID() != payloadID {
		return model.Order{}, ErrInvalidOrderData
	}

	if err := s.repo.UpdateOrder(pathID, order); err != nil {
		return model.Order{}, ErrEntityNotFound
	}
	return order, nil
}

func (s *Service) UpdateTransaction(pathID, payloadID, amount int, date string) (model.Transaction, error) {
	if payloadID < 0 {
		return model.Transaction{}, ErrInvalidID
	}
	if amount <= 0 {
		return model.Transaction{}, ErrInvalidAmount
	}
	if date == "" {
		return model.Transaction{}, ErrInvalidDate
	}
	if pathID != payloadID {
		return model.Transaction{}, ErrIDMismatch
	}

	tx := model.NewTransaction(payloadID, amount, date)
	if tx.GetID() != payloadID {
		return model.Transaction{}, ErrInvalidTxData
	}

	if err := s.repo.UpdateTransaction(pathID, tx); err != nil {
		return model.Transaction{}, ErrEntityNotFound
	}
	return tx, nil
}

func (s *Service) DeleteOrder(id int) error {
	if err := s.repo.DeleteOrder(id); err != nil {
		return ErrEntityNotFound
	}
	return nil
}

func (s *Service) DeleteTransaction(id int) error {
	if err := s.repo.DeleteTransaction(id); err != nil {
		return ErrEntityNotFound
	}
	return nil
}

func (s *Service) DeleteItem(id int) error {
	if err := s.repo.DeleteOrder(id); err == nil {
		return nil
	}
	if err := s.repo.DeleteTransaction(id); err == nil {
		return nil
	}
	return ErrEntityNotFound
}

func (s *Service) GetOrder(id int) (*model.Order, error) {
	order := s.repo.GetOrderByID(id)
	if order == nil {
		return nil, ErrEntityNotFound
	}
	return order, nil
}

func (s *Service) GetTransaction(id int) (*model.Transaction, error) {
	tx := s.repo.GetTransactionByID(id)
	if tx == nil {
		return nil, ErrEntityNotFound
	}
	return tx, nil
}

func (s *Service) GetItem(id int) (interface{}, error) {
	if order := s.repo.GetOrderByID(id); order != nil {
		return *order, nil
	}
	if tx := s.repo.GetTransactionByID(id); tx != nil {
		return *tx, nil
	}
	return nil, ErrEntityNotFound
}

func (s *Service) ListItems() ([]model.Order, []model.Transaction) {
	return s.repo.GetOrders(), s.repo.GetTransactions()
}
