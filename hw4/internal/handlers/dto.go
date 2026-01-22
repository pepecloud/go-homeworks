package handlers

// OrderDTO представляет заказ для JSON сериализации
type OrderDTO struct {
	ID     int  `json:"id"`
	Status bool `json:"status"`
	Amount int  `json:"amount"`
}

// TransactionDTO представляет транзакцию для JSON сериализации
type TransactionDTO struct {
	ID     int    `json:"id"`
	Amount int    `json:"amount"`
	Date   string `json:"date"`
}
