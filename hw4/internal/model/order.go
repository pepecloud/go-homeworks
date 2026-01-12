package model

import "fmt"

type Order struct {

	//ID не должен быть меньше 0
	id int

	//STATUS может быть любым хоть true хоть false (выполнен не выполнен)
	status bool

	//AMOUNT или же сумма сделки не может быть меньше или равна 0
	amount int //RUB
}

func NewOrder(
	id int,
	status bool,
	amount int,
) Order {
	//ПРОВЕРКА НА ВАЛИДНОСТЬ ID
	if id < 0 {
		fmt.Println("id не валдиен!")
		return Order{}
	}

	//ПРОВЕРКА НА ВАЛИДНОСТЬ AMOUNT
	if amount <= 0 {
		fmt.Println("Сумма не валидна!")
		return Order{}
	}

	//ЕСЛИ ВСЕ ХОРОШО ТО СОЗДАЕМ СТРУКТУРУ
	return Order{
		id:     id,
		status: status,
		amount: amount,
	}
}

//---------МЕТОДЫ ДЛЯ СМЕНЫ ЗНАЧЕНИЙ---------(МОЖЕТ ТУТ ВОВСЕ НЕ НУЖНЫ НУ Я САМ ТАКОЕ ЗАДАНИЕ ОПРЕДЕЛИЛ)
func (o *Order) ChangeId(NewId int) {
	if NewId < 0 {
		fmt.Println("Новый ID Order не валиден!")
		return
	}

	o.id = NewId
}

//НУ К ПРИМЕРУ БАГ В ПРОГРАММЕ А СТАТУС ЗАЧЕЛСЯ, НО А МЫ ЕГО ИЗМЕНИМ
func (o *Order) ChangeStatus(NewStatus bool) {
	o.status = NewStatus
	fmt.Println("Статус сделки сменился на новый:", NewStatus)
}

func (o *Order) ChangeAmount(NewAmount int) {
	if NewAmount <= 0 {
		fmt.Println("Новый Amount Order не валиден!")
		return
	}

	o.amount = NewAmount
}

func (o *Order) GetID() int {
	return o.id
}

func (o *Order) GetStatus() bool {
	return o.status
}

func (o *Order) GetAmount() int {
	return o.amount
}