package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pepecloud/go-homeworks/hw4/internal/model"
)

const (
	defaultPostgresDSN = "postgres://postgres:postgres@localhost:5432/itemsdb?sslmode=disable"
)

var ErrNotFound = errors.New("entity not found")

type historyRecord struct {
	EntityType string          `json:"entity_type"`
	Action     string          `json:"action"`
	EntityID   int             `json:"entity_id"`
	Payload    json.RawMessage `json:"payload"`
}

type Repository struct {
	db *sql.DB
}

func NewRepository(ctx context.Context) (*Repository, error) {
	dsn := getenv("POSTGRES_DSN", defaultPostgresDSN)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping PostgreSQL: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close(_ context.Context) error {
	if r.db == nil {
		return nil
	}
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("close PostgreSQL: %w", err)
	}
	return nil
}

func (r *Repository) AddEntity(entity interface{}) error {
	switch v := entity.(type) {
	case model.Order:
		return r.createOrderWithHistory(v)
	case model.Transaction:
		return r.createTransactionWithHistory(v)
	default:
		return fmt.Errorf("unknown entity type")
	}
}

func (r *Repository) createOrderWithHistory(order model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO orders (id, status, amount) VALUES ($1, $2, $3)`,
		order.GetID(), order.GetStatus(), order.GetAmount(),
	)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]interface{}{
		"id":     order.GetID(),
		"status": order.GetStatus(),
		"amount": order.GetAmount(),
	})
	if err != nil {
		return err
	}

	if err := r.insertHistoryTx(tx, "order", "create", order.GetID(), payload); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) createTransactionWithHistory(tr model.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT INTO transactions (id, amount, date) VALUES ($1, $2, $3)`,
		tr.GetID(), tr.GetAmount(), tr.GetDate(),
	)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]interface{}{
		"id":     tr.GetID(),
		"amount": tr.GetAmount(),
		"date":   tr.GetDate(),
	})
	if err != nil {
		return err
	}

	if err := r.insertHistoryTx(tx, "transaction", "create", tr.GetID(), payload); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *Repository) GetOrders() ([]model.Order, error) {
	rows, err := r.db.Query(`SELECT id, status, amount FROM orders ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query orders: %w", err)
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var id, amount int
		var status bool
		if err := rows.Scan(&id, &status, &amount); err != nil {
			return nil, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, model.NewOrder(id, status, amount))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate orders: %w", err)
	}
	return orders, nil
}

func (r *Repository) GetTransactions() ([]model.Transaction, error) {
	rows, err := r.db.Query(`SELECT id, amount, date FROM transactions ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]model.Transaction, 0)
	for rows.Next() {
		var id, amount int
		var date string
		if err := rows.Scan(&id, &amount, &date); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, model.NewTransaction(id, amount, date))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return transactions, nil
}

func (r *Repository) GetOrderByID(id int) (*model.Order, error) {
	var amount int
	var status bool
	if err := r.db.QueryRow(`SELECT status, amount FROM orders WHERE id = $1`, id).Scan(&status, &amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query order by id=%d: %w", id, err)
	}
	order := model.NewOrder(id, status, amount)
	return &order, nil
}

func (r *Repository) GetTransactionByID(id int) (*model.Transaction, error) {
	var amount int
	var date string
	if err := r.db.QueryRow(`SELECT amount, date FROM transactions WHERE id = $1`, id).Scan(&amount, &date); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query transaction by id=%d: %w", id, err)
	}
	tr := model.NewTransaction(id, amount, date)
	return &tr, nil
}

func (r *Repository) UpdateOrder(id int, order model.Order) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`UPDATE orders SET status = $1, amount = $2, updated_at = NOW() WHERE id = $3`,
		order.GetStatus(), order.GetAmount(), id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("affected rows for order id=%d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: order id %d", ErrNotFound, id)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"id":     id,
		"status": order.GetStatus(),
		"amount": order.GetAmount(),
	})
	if err != nil {
		return err
	}
	if err := r.insertHistoryTx(tx, "order", "update", id, payload); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) UpdateTransaction(id int, transaction model.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`UPDATE transactions SET amount = $1, date = $2, updated_at = NOW() WHERE id = $3`,
		transaction.GetAmount(), transaction.GetDate(), id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("affected rows for transaction id=%d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: transaction id %d", ErrNotFound, id)
	}

	payload, err := json.Marshal(map[string]interface{}{
		"id":     id,
		"amount": transaction.GetAmount(),
		"date":   transaction.GetDate(),
	})
	if err != nil {
		return err
	}
	if err := r.insertHistoryTx(tx, "transaction", "update", id, payload); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteOrder(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("affected rows for order id=%d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: order id %d", ErrNotFound, id)
	}

	payload, err := json.Marshal(map[string]int{"id": id})
	if err != nil {
		return err
	}
	if err := r.insertHistoryTx(tx, "order", "delete", id, payload); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteTransaction(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("affected rows for transaction id=%d: %w", id, err)
	}
	if affected == 0 {
		return fmt.Errorf("%w: transaction id %d", ErrNotFound, id)
	}

	payload, err := json.Marshal(map[string]int{"id": id})
	if err != nil {
		return err
	}
	if err := r.insertHistoryTx(tx, "transaction", "delete", id, payload); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) LoadData() error {
	return nil
}

func (r *Repository) insertHistoryTx(tx *sql.Tx, entityType, action string, entityID int, payload []byte) error {
	record := historyRecord{
		EntityType: entityType,
		Action:     action,
		EntityID:   entityID,
		Payload:    payload,
	}

	_, err := tx.Exec(
		`INSERT INTO change_history (entity_type, action, entity_id, payload) VALUES ($1, $2, $3, $4)`,
		record.EntityType, record.Action, record.EntityID, record.Payload,
	)
	return err
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
