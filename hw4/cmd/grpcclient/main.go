package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pepecloud/go-homeworks/hw4/internal/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcAddr = "localhost:9090"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("не удалось подключиться к gRPC-серверу: %v", err)
	}
	defer conn.Close()

	authClient := pb.NewAuthServiceClient(conn)
	orderClient := pb.NewOrderServiceClient(conn)
	txClient := pb.NewTransactionServiceClient(conn)

	// 1. Логин и получение токена (для демонстрации, но в текущей реализации сервер сам работает без токена).
	loginResp, err := authClient.Login(ctx, &pb.LoginRequest{
		Login:    "admin",
		Password: "admin",
	})
	if err != nil {
		log.Printf("ошибка логина (вероятно, не настроены переменные окружения): %v", err)
	} else {
		fmt.Printf("Успешный логин, токен: %s\n", loginResp.GetToken())
	}

	// 2. Создание заказа.
	fmt.Println("=== Создаём Order ===")
	createOrderResp, err := orderClient.CreateOrder(ctx, &pb.CreateOrderRequest{
		Order: &pb.Order{
			Id:     1,
			Status: true,
			Amount: 1000,
		},
	})
	if err != nil {
		log.Fatalf("ошибка CreateOrder: %v", err)
	}
	fmt.Printf("Создан заказ: %+v\n", createOrderResp.GetOrder())

	// 3. Получение заказа по ID.
	fmt.Println("=== Получаем Order по ID ===")
	getOrderResp, err := orderClient.GetOrder(ctx, &pb.OrderIdRequest{Id: 1})
	if err != nil {
		log.Fatalf("ошибка GetOrder: %v", err)
	}
	fmt.Printf("Получен заказ: %+v\n", getOrderResp.GetOrder())

	// 4. Список заказов.
	fmt.Println("=== Список Orders ===")
	listOrdersResp, err := orderClient.ListOrders(ctx, &pb.ListOrdersRequest{})
	if err != nil {
		log.Fatalf("ошибка ListOrders: %v", err)
	}
	for _, o := range listOrdersResp.GetOrders() {
		fmt.Printf("Order: %+v\n", o)
	}

	// 5. Создание транзакции.
	fmt.Println("=== Создаём Transaction ===")
	createTxResp, err := txClient.CreateTransaction(ctx, &pb.CreateTransactionRequest{
		Transaction: &pb.Transaction{
			Id:     100,
			Amount: 500,
			Date:   "2025-11-08",
		},
	})
	if err != nil {
		log.Fatalf("ошибка CreateTransaction: %v", err)
	}
	fmt.Printf("Создана транзакция: %+v\n", createTxResp.GetTransaction())

	// 6. Получение транзакции по ID.
	fmt.Println("=== Получаем Transaction по ID ===")
	getTxResp, err := txClient.GetTransaction(ctx, &pb.TransactionIdRequest{Id: 100})
	if err != nil {
		log.Fatalf("ошибка GetTransaction: %v", err)
	}
	fmt.Printf("Получена транзакция: %+v\n", getTxResp.GetTransaction())

	// 7. Список транзакций.
	fmt.Println("=== Список Transactions ===")
	listTxResp, err := txClient.ListTransactions(ctx, &pb.ListTransactionsRequest{})
	if err != nil {
		log.Fatalf("ошибка ListTransactions: %v", err)
	}
	for _, t := range listTxResp.GetTransactions() {
		fmt.Printf("Transaction: %+v\n", t)
	}
}

