package commands

type BankAccountCommands struct {
	ChangeEmail
	DepositBalance
	CreateBankAccount
}

func NewBankAccountCommands(changeEmail ChangeEmail, depositBalance DepositBalance, createBankAccount CreateBankAccount) *BankAccountCommands {
	return &BankAccountCommands{ChangeEmail: changeEmail, DepositBalance: depositBalance, CreateBankAccount: createBankAccount}
}
