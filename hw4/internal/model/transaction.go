package model

import "fmt"

type Transaction struct {
	//ID не долен быть равен меньше 0
	id int

	//AMOUNT не должен быть меньше или равен 0
	amount int

	//DATE не может быть не записана т.е != ""
	date string
}

func NewTransaction(
	id int,
	amount int,
	date string,
) Transaction {
	//ПРОВЕРКА НА ВАЛИДНОСТЬ ID
	if id < 0 {
		fmt.Println("id не валдиен!")
		return Transaction{}
	}

	//ПРОВЕРКА НА ВАЛИДНОСТЬ AMOUNT
	if amount <= 0 {
		fmt.Println("Сумма не валидна!")
		return Transaction{}
	}

	//ПРОВЕРКА НА ВАЛИДНОСТЬ DATE
	if date == "" {
		fmt.Println("Дата не валидна!")
		return Transaction{}
	}

	//ЕСЛИ ВСЕ ХОРОШО ТО СОЗДАЕМ СТРУКТУРУ
	return Transaction{
		id:     id,
		amount: amount,
		date:   date,
	}
}

// ---------МЕТОДЫ ДЛЯ СМЕНЫ ЗНАЧЕНИЙ---------
func (t *Transaction) ChangeId(NewId int) {
	if NewId < 0 {
		fmt.Println("Новый ID Transaction не валиден!")
		return
	}

	t.id = NewId
}

func (t *Transaction) ChangeAmount(NewAmount int) {
	if NewAmount <= 0 {
		fmt.Println("Новый Amount Transaction не валиден!")
		return
	}

	t.amount = NewAmount
}

func (t *Transaction) NewDate(NewDate string) {
	if NewDate == "" {
		fmt.Println("Новая дата Transaction не валидна!")
		return
	}

	t.date = NewDate
}
