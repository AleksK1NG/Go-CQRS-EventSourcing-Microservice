package queries

type BankAccountQueries struct {
	GetBankAccountByID
}

func NewBankAccountQueries(getBankAccountByID GetBankAccountByID) *BankAccountQueries {
	return &BankAccountQueries{GetBankAccountByID: getBankAccountByID}
}
