package model

type Order struct {
	orderID    int
	customerID int
	totalPrice float64
}

func NewOrder(orderID int, custcustomerID int, totalPrice float64) *Order {
	return &Order{
		orderID:    orderID,
		customerID: custcustomerID,
		totalPrice: totalPrice,
	}
}

func (o *Order) GetOrderID() int {
	return o.orderID
}

func (o *Order) GetCustomerID() int {
	return o.customerID
}

func (o *Order) GetTotalPrice() float64 {
	return o.totalPrice
}
