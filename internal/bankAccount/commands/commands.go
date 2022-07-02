package commands

type BankAccountCommands struct {
	ChangeEmail
	DepositBalance
	CreateBankAccount
	WithdrawBalance
}

func NewBankAccountCommands(
	changeEmail ChangeEmail,
	depositBalance DepositBalance,
	createBankAccount CreateBankAccount,
	withdrawBalance WithdrawBalance,
) *BankAccountCommands {
	return &BankAccountCommands{
		ChangeEmail:       changeEmail,
		DepositBalance:    depositBalance,
		CreateBankAccount: createBankAccount,
		WithdrawBalance:   withdrawBalance,
	}
}
