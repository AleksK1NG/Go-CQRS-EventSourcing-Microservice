package domain

type BankAccount struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Address   string  `json:"address"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
}

func NewBankAccount() *BankAccount {
	return &BankAccount{}
}
