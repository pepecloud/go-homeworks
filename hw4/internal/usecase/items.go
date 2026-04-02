package usecase

import (
	"errors"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
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
	GetOrders() ([]model.Order, error)
	GetTransactions() ([]model.Transaction, error)
	GetOrderByID(id int) (*model.Order, error)
	GetTransactionByID(id int) (*model.Transaction, error)
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

	existing, err := s.repo.GetOrderByID(id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return model.Order{}, err
	}
	if existing != nil {
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

	existing, err := s.repo.GetTransactionByID(id)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return model.Transaction{}, err
	}
	if existing != nil {
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
		if errors.Is(err, repository.ErrNotFound) {
			return model.Order{}, ErrEntityNotFound
		}
		return model.Order{}, err
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
		if errors.Is(err, repository.ErrNotFound) {
			return model.Transaction{}, ErrEntityNotFound
		}
		return model.Transaction{}, err
	}
	return tx, nil
}

func (s *Service) DeleteOrder(id int) error {
	if err := s.repo.DeleteOrder(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEntityNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteTransaction(id int) error {
	if err := s.repo.DeleteTransaction(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrEntityNotFound
		}
		return err
	}
	return nil
}

func (s *Service) DeleteItem(id int) error {
	if err := s.repo.DeleteOrder(id); err == nil {
		return nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	if err := s.repo.DeleteTransaction(id); err == nil {
		return nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	return ErrEntityNotFound
}

func (s *Service) GetOrder(id int) (*model.Order, error) {
	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}
	return order, nil
}

func (s *Service) GetTransaction(id int) (*model.Transaction, error) {
	tx, err := s.repo.GetTransactionByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrEntityNotFound
		}
		return nil, err
	}
	return tx, nil
}

func (s *Service) GetItem(id int) (interface{}, error) {
	order, err := s.repo.GetOrderByID(id)
	if err == nil {
		return *order, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	tx, err := s.repo.GetTransactionByID(id)
	if err == nil {
		return *tx, nil
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	return nil, ErrEntityNotFound
}

func (s *Service) ListItems() ([]model.Order, []model.Transaction, error) {
	orders, err := s.repo.GetOrders()
	if err != nil {
		return nil, nil, err
	}

	transactions, err := s.repo.GetTransactions()
	if err != nil {
		return nil, nil, err
	}

	return orders, transactions, nil
}
