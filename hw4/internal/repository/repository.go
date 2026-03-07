package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	defaultMongoURI   = "mongodb://localhost:27017"
	defaultMongoDB    = "itemsdb"
	defaultRedisAddr  = "localhost:6379"
	defaultHistoryTTL = 24 * time.Hour
	opTimeout         = 5 * time.Second
)

type orderDoc struct {
	ID     int  `bson:"id"`
	Status bool `bson:"status"`
	Amount int  `bson:"amount"`
}

type transactionDoc struct {
	ID     int    `bson:"id"`
	Amount int    `bson:"amount"`
	Date   string `bson:"date"`
}

type historyRecord struct {
	EntityType string      `json:"entity_type"`
	Action     string      `json:"action"`
	EntityID   int         `json:"entity_id"`
	Payload    interface{} `json:"payload"`
	Timestamp  string      `json:"timestamp"`
}

type Repository struct {
	orders       []model.Order
	transactions []model.Transaction
	mu           sync.RWMutex

	mongoClient *mongo.Client
	ordersColl  *mongo.Collection
	txColl      *mongo.Collection

	redisClient *redis.Client
	historyTTL  time.Duration
}

func NewRepository(ctx context.Context) (*Repository, error) {
	mongoURI := getenv("MONGO_URI", defaultMongoURI)
	mongoDB := getenv("MONGO_DB", defaultMongoDB)

	mongoCtx, mongoCancel := context.WithTimeout(ctx, opTimeout)
	defer mongoCancel()

	mongoClient, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	if err := mongoClient.Ping(mongoCtx, nil); err != nil {
		return nil, fmt.Errorf("ошибка ping MongoDB: %w", err)
	}

	redisDB, err := strconv.Atoi(getenv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_ADDR", defaultRedisAddr),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})
	if err := redisClient.Ping(mongoCtx).Err(); err != nil {
		return nil, fmt.Errorf("ошибка подключения к Redis: %w", err)
	}

	repo := &Repository{
		orders:       []model.Order{},
		transactions: []model.Transaction{},
		mongoClient:  mongoClient,
		ordersColl:   mongoClient.Database(mongoDB).Collection("orders"),
		txColl:       mongoClient.Database(mongoDB).Collection("transactions"),
		redisClient:  redisClient,
		historyTTL:   getHistoryTTL(),
	}

	if err := repo.ensureIndexes(mongoCtx); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *Repository) Close(ctx context.Context) error {
	var resultErr error

	if r.redisClient != nil {
		if err := r.redisClient.Close(); err != nil {
			resultErr = fmt.Errorf("ошибка закрытия Redis: %w", err)
		}
	}

	if r.mongoClient != nil {
		if err := r.mongoClient.Disconnect(ctx); err != nil && resultErr == nil {
			resultErr = fmt.Errorf("ошибка закрытия MongoDB: %w", err)
		}
	}

	return resultErr
}

func (r *Repository) AddEntity(entity interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch v := entity.(type) {
	case model.Order:
		ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
		defer cancel()
		if _, err := r.ordersColl.InsertOne(ctx, orderFromModel(v)); err != nil {
			return err
		}
		r.orders = append(r.orders, v)
		if err := r.logChange(ctx, "order", "create", v.GetID(), orderFromModel(v)); err != nil {
			return err
		}
		fmt.Println("Добавлен Order")
	case model.Transaction:
		ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
		defer cancel()
		if _, err := r.txColl.InsertOne(ctx, txFromModel(v)); err != nil {
			return err
		}
		r.transactions = append(r.transactions, v)
		if err := r.logChange(ctx, "transaction", "create", v.GetID(), txFromModel(v)); err != nil {
			return err
		}
		fmt.Println("Добавлена Transaction")
	default:
		return fmt.Errorf("неизвестный тип сущности")
	}

	return nil
}

func (r *Repository) GetOrders() []model.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Order(nil), r.orders...)
}

func (r *Repository) GetTransactions() []model.Transaction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]model.Transaction(nil), r.transactions...)
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

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	res, err := r.ordersColl.ReplaceOne(ctx, bson.M{"id": id}, orderFromModel(order))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("заказ с id %d не найден", id)
	}

	for i := range r.orders {
		if r.orders[i].GetID() == id {
			r.orders[i] = order
			return r.logChange(ctx, "order", "update", id, orderFromModel(order))
		}
	}

	return fmt.Errorf("заказ с id %d не найден", id)
}

func (r *Repository) UpdateTransaction(id int, tx model.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	res, err := r.txColl.ReplaceOne(ctx, bson.M{"id": id}, txFromModel(tx))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("транзакция с id %d не найдена", id)
	}

	for i := range r.transactions {
		if r.transactions[i].GetID() == id {
			r.transactions[i] = tx
			return r.logChange(ctx, "transaction", "update", id, txFromModel(tx))
		}
	}

	return fmt.Errorf("транзакция с id %d не найдена", id)
}

func (r *Repository) DeleteOrder(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	res, err := r.ordersColl.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("заказ с id %d не найден", id)
	}

	for i := range r.orders {
		if r.orders[i].GetID() == id {
			r.orders = append(r.orders[:i], r.orders[i+1:]...)
			return r.logChange(ctx, "order", "delete", id, map[string]int{"id": id})
		}
	}

	return fmt.Errorf("заказ с id %d не найден", id)
}

func (r *Repository) DeleteTransaction(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()
	res, err := r.txColl.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return fmt.Errorf("транзакция с id %d не найдена", id)
	}

	for i := range r.transactions {
		if r.transactions[i].GetID() == id {
			r.transactions = append(r.transactions[:i], r.transactions[i+1:]...)
			return r.logChange(ctx, "transaction", "delete", id, map[string]int{"id": id})
		}
	}

	return fmt.Errorf("транзакция с id %d не найдена", id)
}

func (r *Repository) LoadData() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.loadOrdersFromMongo(); err != nil {
		return fmt.Errorf("ошибка загрузки заказов: %v", err)
	}

	if err := r.loadTransactionsFromMongo(); err != nil {
		return fmt.Errorf("ошибка загрузки транзакций: %v", err)
	}

	return nil
}

func (r *Repository) loadOrdersFromMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	cursor, err := r.ordersColl.Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var orders []model.Order
	for cursor.Next(ctx) {
		var doc orderDoc
		if err := cursor.Decode(&doc); err != nil {
			return err
		}
		order := model.NewOrder(doc.ID, doc.Status, doc.Amount)
		orders = append(orders, order)
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	r.orders = orders
	return nil
}

func (r *Repository) loadTransactionsFromMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	cursor, err := r.txColl.Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var transactions []model.Transaction
	for cursor.Next(ctx) {
		var doc transactionDoc
		if err := cursor.Decode(&doc); err != nil {
			return err
		}
		tx := model.NewTransaction(doc.ID, doc.Amount, doc.Date)
		transactions = append(transactions, tx)
	}
	if err := cursor.Err(); err != nil {
		return err
	}

	r.transactions = transactions
	return nil
}

func (r *Repository) ensureIndexes(ctx context.Context) error {
	orderIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	txIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := r.ordersColl.Indexes().CreateOne(ctx, orderIndex); err != nil {
		return fmt.Errorf("ошибка создания индекса orders.id: %w", err)
	}
	if _, err := r.txColl.Indexes().CreateOne(ctx, txIndex); err != nil {
		return fmt.Errorf("ошибка создания индекса transactions.id: %w", err)
	}
	return nil
}

func (r *Repository) logChange(ctx context.Context, entityType, action string, entityID int, payload interface{}) error {
	record := historyRecord{
		EntityType: entityType,
		Action:     action,
		EntityID:   entityID,
		Payload:    payload,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("history:%s:%d:%d", entityType, entityID, time.Now().UnixNano())
	return r.redisClient.Set(ctx, key, data, r.historyTTL).Err()
}

func orderFromModel(o model.Order) orderDoc {
	return orderDoc{
		ID:     o.GetID(),
		Status: o.GetStatus(),
		Amount: o.GetAmount(),
	}
}

func txFromModel(t model.Transaction) transactionDoc {
	return transactionDoc{
		ID:     t.GetID(),
		Amount: t.GetAmount(),
		Date:   t.GetDate(),
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getHistoryTTL() time.Duration {
	ttlRaw := getenv("REDIS_HISTORY_TTL_SECONDS", "")
	if ttlRaw == "" {
		return defaultHistoryTTL
	}

	seconds, err := strconv.Atoi(ttlRaw)
	if err != nil || seconds <= 0 {
		return defaultHistoryTTL
	}
	return time.Duration(seconds) * time.Second
}
