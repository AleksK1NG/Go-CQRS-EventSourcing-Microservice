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

func NewBankAccount(id string) *BankAccount {
	return &BankAccount{ID: id}
}

func (b *BankAccount) DepositBalance(amount float64) {
	b.Balance += amount
}

func (b *BankAccount) WithdrawBalance(amount float64) {
	b.Balance -= amount
}
