package usecase_test

import (
	"errors"
	"testing"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/pepecloud/go-homeworks/hw4/internal/repository"
	"github.com/pepecloud/go-homeworks/hw4/internal/usecase"
)

type fakeRepo struct {
	orders       []model.Order
	transactions []model.Transaction
}

func (f *fakeRepo) AddEntity(entity interface{}) error {
	switch v := entity.(type) {
	case model.Order:
		f.orders = append(f.orders, v)
	case model.Transaction:
		f.transactions = append(f.transactions, v)
	}
	return nil
}

func (f *fakeRepo) GetOrders() ([]model.Order, error) { return f.orders, nil }

func (f *fakeRepo) GetTransactions() ([]model.Transaction, error) { return f.transactions, nil }

func (f *fakeRepo) GetOrderByID(id int) (*model.Order, error) {
	for i := range f.orders {
		if f.orders[i].GetID() == id {
			return &f.orders[i], nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *fakeRepo) GetTransactionByID(id int) (*model.Transaction, error) {
	for i := range f.transactions {
		if f.transactions[i].GetID() == id {
			return &f.transactions[i], nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *fakeRepo) UpdateOrder(id int, order model.Order) error {
	for i := range f.orders {
		if f.orders[i].GetID() == id {
			f.orders[i] = order
			return nil
		}
	}
	return repository.ErrNotFound
}

func (f *fakeRepo) UpdateTransaction(id int, tx model.Transaction) error {
	for i := range f.transactions {
		if f.transactions[i].GetID() == id {
			f.transactions[i] = tx
			return nil
		}
	}
	return repository.ErrNotFound
}

func (f *fakeRepo) DeleteOrder(id int) error {
	for i := range f.orders {
		if f.orders[i].GetID() == id {
			f.orders = append(f.orders[:i], f.orders[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}

func (f *fakeRepo) DeleteTransaction(id int) error {
	for i := range f.transactions {
		if f.transactions[i].GetID() == id {
			f.transactions = append(f.transactions[:i], f.transactions[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}

func TestServiceCreateOrder(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func() *fakeRepo
		id        int
		status    bool
		amount    int
		wantErr   error
	}{
		{
			name: "success",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      1,
			status:  true,
			amount:  100,
			wantErr: nil,
		},
		{
			name: "invalid id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      -1,
			status:  true,
			amount:  100,
			wantErr: usecase.ErrInvalidID,
		},
		{
			name: "invalid amount",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      1,
			status:  true,
			amount:  0,
			wantErr: usecase.ErrInvalidAmount,
		},
		{
			name: "duplicate id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{orders: []model.Order{model.NewOrder(1, false, 10)}}
			},
			id:      1,
			status:  true,
			amount:  100,
			wantErr: usecase.ErrOrderExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := usecase.New(repo)

			order, err := svc.CreateOrder(tt.id, tt.status, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if order.GetID() != tt.id || order.GetAmount() != tt.amount || order.GetStatus() != tt.status {
					t.Fatalf("unexpected order: %+v", order)
				}
				if len(repo.orders) != 1 {
					t.Fatalf("expected order to be persisted")
				}
			}
		})
	}
}

func TestServiceCreateTransaction(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func() *fakeRepo
		id        int
		amount    int
		date      string
		wantErr   error
	}{
		{
			name: "success",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      1,
			amount:  100,
			date:    "2026-03-04",
			wantErr: nil,
		},
		{
			name: "invalid id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      -1,
			amount:  100,
			date:    "2026-03-04",
			wantErr: usecase.ErrInvalidID,
		},
		{
			name: "invalid amount",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      1,
			amount:  0,
			date:    "2026-03-04",
			wantErr: usecase.ErrInvalidAmount,
		},
		{
			name: "empty date",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			id:      1,
			amount:  100,
			date:    "",
			wantErr: usecase.ErrInvalidDate,
		},
		{
			name: "duplicate id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(1, 50, "2026-01-01")}}
			},
			id:      1,
			amount:  100,
			date:    "2026-03-04",
			wantErr: usecase.ErrTransactionExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := usecase.New(repo)

			tx, err := svc.CreateTransaction(tt.id, tt.amount, tt.date)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if tx.GetID() != tt.id || tx.GetAmount() != tt.amount || tx.GetDate() != tt.date {
					t.Fatalf("unexpected transaction: %+v", tx)
				}
				if len(repo.transactions) != 1 {
					t.Fatalf("expected transaction to be persisted")
				}
			}
		})
	}
}

func TestServiceUpdateOrder(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func() *fakeRepo
		pathID    int
		payloadID int
		status    bool
		amount    int
		wantErr   error
	}{
		{
			name: "success",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{orders: []model.Order{model.NewOrder(10, false, 100)}}
			},
			pathID:    10,
			payloadID: 10,
			status:    true,
			amount:    200,
		},
		{
			name: "invalid payload id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{orders: []model.Order{model.NewOrder(10, false, 100)}}
			},
			pathID:    10,
			payloadID: -1,
			status:    true,
			amount:    200,
			wantErr:   usecase.ErrInvalidID,
		},
		{
			name: "invalid amount",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{orders: []model.Order{model.NewOrder(10, false, 100)}}
			},
			pathID:    10,
			payloadID: 10,
			status:    true,
			amount:    0,
			wantErr:   usecase.ErrInvalidAmount,
		},
		{
			name: "id mismatch",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{orders: []model.Order{model.NewOrder(10, false, 100)}}
			},
			pathID:    10,
			payloadID: 11,
			status:    true,
			amount:    200,
			wantErr:   usecase.ErrIDMismatch,
		},
		{
			name: "not found",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			pathID:    10,
			payloadID: 10,
			status:    true,
			amount:    200,
			wantErr:   usecase.ErrEntityNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := usecase.New(repo)

			order, err := svc.UpdateOrder(tt.pathID, tt.payloadID, tt.status, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if order.GetID() != tt.payloadID || order.GetAmount() != tt.amount {
					t.Fatalf("unexpected order: %+v", order)
				}
			}
		})
	}
}

func TestServiceUpdateTransaction(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func() *fakeRepo
		pathID    int
		payloadID int
		amount    int
		date      string
		wantErr   error
	}{
		{
			name: "success",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(7, 100, "2026-01-01")}}
			},
			pathID:    7,
			payloadID: 7,
			amount:    250,
			date:      "2026-03-04",
		},
		{
			name: "invalid payload id",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(7, 100, "2026-01-01")}}
			},
			pathID:    7,
			payloadID: -1,
			amount:    250,
			date:      "2026-03-04",
			wantErr:   usecase.ErrInvalidID,
		},
		{
			name: "invalid amount",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(7, 100, "2026-01-01")}}
			},
			pathID:    7,
			payloadID: 7,
			amount:    0,
			date:      "2026-03-04",
			wantErr:   usecase.ErrInvalidAmount,
		},
		{
			name: "invalid date",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(7, 100, "2026-01-01")}}
			},
			pathID:    7,
			payloadID: 7,
			amount:    250,
			date:      "",
			wantErr:   usecase.ErrInvalidDate,
		},
		{
			name: "id mismatch",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{transactions: []model.Transaction{model.NewTransaction(7, 100, "2026-01-01")}}
			},
			pathID:    7,
			payloadID: 8,
			amount:    250,
			date:      "2026-03-04",
			wantErr:   usecase.ErrIDMismatch,
		},
		{
			name: "not found",
			setupRepo: func() *fakeRepo {
				return &fakeRepo{}
			},
			pathID:    7,
			payloadID: 7,
			amount:    250,
			date:      "2026-03-04",
			wantErr:   usecase.ErrEntityNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := usecase.New(repo)

			tx, err := svc.UpdateTransaction(tt.pathID, tt.payloadID, tt.amount, tt.date)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil {
				if tx.GetID() != tt.payloadID || tx.GetAmount() != tt.amount || tx.GetDate() != tt.date {
					t.Fatalf("unexpected transaction: %+v", tx)
				}
			}
		})
	}
}

func TestServiceGetOrderAndGetTransaction(t *testing.T) {
	repo := &fakeRepo{
		orders:       []model.Order{model.NewOrder(1, true, 11)},
		transactions: []model.Transaction{model.NewTransaction(2, 22, "2026-03-04")},
	}
	svc := usecase.New(repo)

	order, err := svc.GetOrder(1)
	if err != nil || order == nil || order.GetID() != 1 {
		t.Fatalf("expected order by id, got order=%v err=%v", order, err)
	}
	if _, err := svc.GetOrder(999); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found for order, got %v", err)
	}

	tx, err := svc.GetTransaction(2)
	if err != nil || tx == nil || tx.GetID() != 2 {
		t.Fatalf("expected transaction by id, got tx=%v err=%v", tx, err)
	}
	if _, err := svc.GetTransaction(999); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found for transaction, got %v", err)
	}
}

func TestServiceDeleteOrderAndDeleteTransaction(t *testing.T) {
	repo := &fakeRepo{
		orders:       []model.Order{model.NewOrder(3, false, 30)},
		transactions: []model.Transaction{model.NewTransaction(4, 40, "2026-03-04")},
	}
	svc := usecase.New(repo)

	if err := svc.DeleteOrder(3); err != nil {
		t.Fatalf("expected delete order success, got %v", err)
	}
	if err := svc.DeleteOrder(3); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found on second delete order, got %v", err)
	}

	if err := svc.DeleteTransaction(4); err != nil {
		t.Fatalf("expected delete transaction success, got %v", err)
	}
	if err := svc.DeleteTransaction(4); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found on second delete transaction, got %v", err)
	}
}

func TestServiceGetDeleteAndList(t *testing.T) {
	repo := &fakeRepo{
		orders:       []model.Order{model.NewOrder(1, false, 10)},
		transactions: []model.Transaction{model.NewTransaction(2, 20, "2026-03-04")},
	}
	svc := usecase.New(repo)

	item, err := svc.GetItem(1)
	if err != nil {
		t.Fatalf("expected found item, got %v", err)
	}
	if _, ok := item.(model.Order); !ok {
		t.Fatalf("expected order type, got %T", item)
	}

	item, err = svc.GetItem(2)
	if err != nil {
		t.Fatalf("expected found item, got %v", err)
	}
	if _, ok := item.(model.Transaction); !ok {
		t.Fatalf("expected transaction type, got %T", item)
	}
	if _, err := svc.GetItem(99); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found for item, got %v", err)
	}

	if err := svc.DeleteItem(1); err != nil {
		t.Fatalf("expected delete order success, got %v", err)
	}
	if err := svc.DeleteItem(2); err != nil {
		t.Fatalf("expected delete transaction success, got %v", err)
	}
	if err := svc.DeleteItem(99); !errors.Is(err, usecase.ErrEntityNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}

	orders, txs, err := svc.ListItems()
	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if len(orders) != 0 || len(txs) != 0 {
		t.Fatalf("expected empty repo after deletions, got %d orders and %d txs", len(orders), len(txs))
	}
}
