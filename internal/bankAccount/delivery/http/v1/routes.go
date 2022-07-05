package v1

func (h *bankAccountHandlers) MapRoutes() {
	h.group.POST("", h.CreateBankAccount())
	h.group.PUT("/deposit/:id", h.DepositBalance())
	h.group.PUT("/withdraw/:id", h.WithdrawBalance())
	h.group.PUT("/email/:id", h.ChangeEmail())

	h.group.GET("/:id", h.GetByID())
	h.group.GET("/search", h.Search())
}
