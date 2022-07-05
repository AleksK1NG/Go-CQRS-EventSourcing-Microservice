package dto

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
)

type HttpSearchResponse struct {
	List               []*HttpBankAccountResponse `json:"list"`
	PaginationResponse *utils.PaginationResponse  `json:"paginationResponse"`
}
