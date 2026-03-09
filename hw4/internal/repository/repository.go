package repository

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pepecloud/go-homeworks/hw4/internal/model"
)

const (
	defaultPostgresDSN = "postgres://postgres:postgres@localhost:5432/itemsdb?sslmode=disable"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ошибка ping PostgreSQL: %w", err)
	}

	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close(_ context.Context) error {
	if r.db == nil {
		return nil
	}
	if err := r.db.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия PostgreSQL: %w", err)
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
		return fmt.Errorf("неизвестный тип сущности")
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

func (r *Repository) GetOrders() []model.Order {
	rows, err := r.db.Query(`SELECT id, status, amount FROM orders ORDER BY id`)
	if err != nil {
		return []model.Order{}
	}
	defer rows.Close()

	orders := make([]model.Order, 0)
	for rows.Next() {
		var id, amount int
		var status bool
		if err := rows.Scan(&id, &status, &amount); err != nil {
			return []model.Order{}
		}
		orders = append(orders, model.NewOrder(id, status, amount))
	}
	return orders
}

func (r *Repository) GetTransactions() []model.Transaction {
	rows, err := r.db.Query(`SELECT id, amount, date FROM transactions ORDER BY id`)
	if err != nil {
		return []model.Transaction{}
	}
	defer rows.Close()

	transactions := make([]model.Transaction, 0)
	for rows.Next() {
		var id, amount int
		var date string
		if err := rows.Scan(&id, &amount, &date); err != nil {
			return []model.Transaction{}
		}
		transactions = append(transactions, model.NewTransaction(id, amount, date))
	}
	return transactions
}

func (r *Repository) GetOrderByID(id int) *model.Order {
	var amount int
	var status bool
	if err := r.db.QueryRow(`SELECT status, amount FROM orders WHERE id = $1`, id).Scan(&status, &amount); err != nil {
		return nil
	}
	order := model.NewOrder(id, status, amount)
	return &order
}

func (r *Repository) GetTransactionByID(id int) *model.Transaction {
	var amount int
	var date string
	if err := r.db.QueryRow(`SELECT amount, date FROM transactions WHERE id = $1`, id).Scan(&amount, &date); err != nil {
		return nil
	}
	tr := model.NewTransaction(id, amount, date)
	return &tr
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
	if err != nil || affected == 0 {
		return fmt.Errorf("заказ с id %d не найден", id)
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
	if err != nil || affected == 0 {
		return fmt.Errorf("транзакция с id %d не найдена", id)
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
	if err != nil || affected == 0 {
		return fmt.Errorf("заказ с id %d не найден", id)
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
	if err != nil || affected == 0 {
		return fmt.Errorf("транзакция с id %d не найдена", id)
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
	// Данные уже находятся в PostgreSQL, отдельная загрузка не нужна.
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

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка инициализации драйвера миграций: %w", err)
	}

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("ошибка загрузки migration-файлов: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("ошибка инициализации мигратора: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}
	return nil
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
