package account

type Account struct {
	username string
	email    string
	balance  float64
}

func NewAccount(username string, email string, balance float64) *Account {
	return &Account{
		username: username,
		email:    email,
		balance:  balance,
	}
}

func (a *Account) GetUsername() string {
	return a.username
}

func (a *Account) GetEmail() string {
	return a.email
}

func (a *Account) GetBalance() float64 {
	return a.balance
}
