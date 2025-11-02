package main

import (
	"fmt"
	"hw3/internal/model/account"
	"hw3/internal/model/book"
	"hw3/internal/model/order"
)

func main() {

	b := book.NewBook(1, "GOLang вместе с Otus", 1500.50)
	fmt.Printf("Индефикатор книги: %d\n", b.GetID())
	fmt.Printf("Название: %s\n", b.GetTitle())
	fmt.Printf("Цена: %.2f\n", b.GetPrice())
	fmt.Println("")

	a := account.NewAccount("josh", "testemail@email.com", 12300.29)
	fmt.Printf("Имя юзера: %s\n", a.GetUsername())
	fmt.Printf("Почта: %s\n", a.GetEmail())
	fmt.Printf("Баланс: %.2f\n", a.GetBalance())
	fmt.Println("")

	o := order.NewOrder(123, 1001023, 1000123)
	fmt.Printf("Номер заказа: %d\n", o.GetOrderID())
	fmt.Printf("Уникальный номер заказа: %d\n", o.GetCustomerID())
	fmt.Printf("Цена: %.2f\n", o.GetTotalPrice())
	fmt.Println("")
}
