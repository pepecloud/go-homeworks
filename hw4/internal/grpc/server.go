package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pepecloud/go-homeworks/hw4/internal/grpc/pb"
	"github.com/pepecloud/go-homeworks/hw4/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// jwtTTL определяет время жизни JWT.
const jwtTTL = 24 * time.Hour

type jwtClaims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

// Server реализует все gRPC-сервисы.
type Server struct {
	pb.UnimplementedAuthServiceServer
	pb.UnimplementedOrderServiceServer
	pb.UnimplementedTransactionServiceServer

	items *usecase.Service
}

// NewServer создаёт gRPC-сервер поверх usecase.
func NewServer(items *usecase.Service) *Server {
	return &Server{items: items}
}

// StartGRPCServer поднимает gRPC-сервер на указанном адресе.
func StartGRPCServer(ctx context.Context, items *usecase.Service, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("не удалось открыть порт для gRPC: %w", err)
	}

	s := grpc.NewServer()
	srv := NewServer(items)

	pb.RegisterAuthServiceServer(s, srv)
	pb.RegisterOrderServiceServer(s, srv)
	pb.RegisterTransactionServiceServer(s, srv)

	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	fmt.Println("gRPC-сервер запущен на", addr)
	if err := s.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("ошибка работы gRPC-сервера: %w", err)
	}
	return nil
}

// ---- AuthService ----

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	expectedLogin := os.Getenv("API_LOGIN")
	expectedPassword := os.Getenv("API_PASSWORD")
	if expectedLogin == "" {
		expectedLogin = os.Getenv("LOGIN")
	}
	if expectedPassword == "" {
		expectedPassword = os.Getenv("PASSWORD")
	}
	if expectedLogin == "" || expectedPassword == "" {
		return nil, status.Error(codes.Internal, "авторизация не настроена")
	}

	if req.GetLogin() != expectedLogin || req.GetPassword() != expectedPassword {
		return nil, status.Error(codes.Unauthenticated, "неверный логин или пароль")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	claims := jwtClaims{
		Login: req.GetLogin(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка выдачи токена")
	}

	return &pb.LoginResponse{Token: tokenStr}, nil
}

// ---- OrderService ----

func (s *Server) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.GetOrderResponse, error) {
	o := req.GetOrder()
	if o == nil {
		return nil, status.Error(codes.InvalidArgument, "order is required")
	}
	if o.Id < 0 {
		return nil, status.Error(codes.InvalidArgument, "id не может быть меньше 0")
	}
	if o.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount должен быть больше 0")
	}

	if _, err := s.items.CreateOrder(int(o.Id), o.Status, int(o.Amount)); err != nil {
		switch {
		case errors.Is(err, usecase.ErrOrderExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.GetOrderResponse{
		Order: &pb.Order{
			Id:     o.Id,
			Status: o.Status,
			Amount: o.Amount,
		},
	}, nil
}

func (s *Server) UpdateOrder(ctx context.Context, req *pb.UpdateOrderRequest) (*pb.GetOrderResponse, error) {
	o := req.GetOrder()
	if o == nil {
		return nil, status.Error(codes.InvalidArgument, "order is required")
	}
	if o.Id < 0 {
		return nil, status.Error(codes.InvalidArgument, "id не может быть меньше 0")
	}
	if o.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount должен быть больше 0")
	}

	if _, err := s.items.UpdateOrder(int(o.Id), int(o.Id), o.Status, int(o.Amount)); err != nil {
		switch {
		case errors.Is(err, usecase.ErrEntityNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.GetOrderResponse{Order: o}, nil
}

func (s *Server) DeleteOrder(ctx context.Context, req *pb.DeleteOrderRequest) (*emptypb.Empty, error) {
	if err := s.items.DeleteOrder(int(req.GetId())); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetOrder(ctx context.Context, req *pb.OrderIdRequest) (*pb.GetOrderResponse, error) {
	order, err := s.items.GetOrder(int(req.GetId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "заказ не найден")
	}
	return &pb.GetOrderResponse{
		Order: &pb.Order{
			Id:     int32(order.GetID()),
			Status: order.GetStatus(),
			Amount: int32(order.GetAmount()),
		},
	}, nil
}

func (s *Server) ListOrders(ctx context.Context, _ *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders, _ := s.items.ListItems()
	resp := &pb.ListOrdersResponse{
		Orders: make([]*pb.Order, 0, len(orders)),
	}
	for _, o := range orders {
		resp.Orders = append(resp.Orders, &pb.Order{
			Id:     int32(o.GetID()),
			Status: o.GetStatus(),
			Amount: int32(o.GetAmount()),
		})
	}
	return resp, nil
}

// ---- TransactionService ----

func (s *Server) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.GetTransactionResponse, error) {
	t := req.GetTransaction()
	if t == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction is required")
	}
	if t.Id < 0 {
		return nil, status.Error(codes.InvalidArgument, "id не может быть меньше 0")
	}
	if t.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount должен быть больше 0")
	}
	if t.Date == "" {
		return nil, status.Error(codes.InvalidArgument, "date обязателен для заполнения")
	}

	if _, err := s.items.CreateTransaction(int(t.Id), int(t.Amount), t.Date); err != nil {
		switch {
		case errors.Is(err, usecase.ErrTransactionExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Id:     t.Id,
			Amount: t.Amount,
			Date:   t.Date,
		},
	}, nil
}

func (s *Server) UpdateTransaction(ctx context.Context, req *pb.UpdateTransactionRequest) (*pb.GetTransactionResponse, error) {
	t := req.GetTransaction()
	if t == nil {
		return nil, status.Error(codes.InvalidArgument, "transaction is required")
	}
	if t.Id < 0 {
		return nil, status.Error(codes.InvalidArgument, "id не может быть меньше 0")
	}
	if t.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount должен быть больше 0")
	}
	if t.Date == "" {
		return nil, status.Error(codes.InvalidArgument, "date обязателен для заполнения")
	}

	if _, err := s.items.UpdateTransaction(int(t.Id), int(t.Id), int(t.Amount), t.Date); err != nil {
		switch {
		case errors.Is(err, usecase.ErrEntityNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.GetTransactionResponse{Transaction: t}, nil
}

func (s *Server) DeleteTransaction(ctx context.Context, req *pb.DeleteTransactionRequest) (*emptypb.Empty, error) {
	if err := s.items.DeleteTransaction(int(req.GetId())); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *pb.TransactionIdRequest) (*pb.GetTransactionResponse, error) {
	tx, err := s.items.GetTransaction(int(req.GetId()))
	if err != nil {
		return nil, status.Error(codes.NotFound, "транзакция не найдена")
	}
	return &pb.GetTransactionResponse{
		Transaction: &pb.Transaction{
			Id:     int32(tx.GetID()),
			Amount: int32(tx.GetAmount()),
			Date:   tx.GetDate(),
		},
	}, nil
}

func (s *Server) ListTransactions(ctx context.Context, _ *pb.ListTransactionsRequest) (*pb.ListTransactionsResponse, error) {
	_, txs := s.items.ListItems()
	resp := &pb.ListTransactionsResponse{
		Transactions: make([]*pb.Transaction, 0, len(txs)),
	}
	for _, t := range txs {
		resp.Transactions = append(resp.Transactions, &pb.Transaction{
			Id:     int32(t.GetID()),
			Amount: int32(t.GetAmount()),
			Date:   t.GetDate(),
		})
	}
	return resp, nil
}
