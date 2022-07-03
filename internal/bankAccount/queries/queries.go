package queries

type BankAccountQueries struct {
	GetBankAccountByID
	SearchBankAccounts
}

func NewBankAccountQueries(getBankAccountByID GetBankAccountByID, searchBankAccounts SearchBankAccounts) *BankAccountQueries {
	return &BankAccountQueries{GetBankAccountByID: getBankAccountByID, SearchBankAccounts: searchBankAccounts}
}
