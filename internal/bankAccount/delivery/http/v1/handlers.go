package v1

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/queries"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/constants"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/httpErrors"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/logger"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/middlewares"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/tracing"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/pkg/utils"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

type bankAccountHandlers struct {
	group              *echo.Group
	middlewareManager  middlewares.MiddlewareManager
	log                logger.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	validate           *validator.Validate
}

func NewBankAccountHandlers(
	group *echo.Group,
	middlewareManager middlewares.MiddlewareManager,
	log logger.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	validate *validator.Validate,
) *bankAccountHandlers {
	return &bankAccountHandlers{
		group:              group,
		middlewareManager:  middlewareManager,
		log:                log,
		cfg:                cfg,
		bankAccountService: bankAccountService,
		validate:           validate,
	}
}

// CreateBankAccount
// @Tags BankAccounts
// @Summary Create BankAccount
// @Description Create new Bank Account
// @Param BankAccount body commands.CreateBankAccountCommand true "create Bank Account"
// @Accept json
// @Produce json
// @Success 201 {string} id ""
// @Router /accounts [post]
func (h *bankAccountHandlers) CreateBankAccount() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.CreateBankAccount")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var command commands.CreateBankAccountCommand
		if err := c.Bind(&command); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		if err := h.validate.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		id := uuid.NewV4().String()
		err := h.bankAccountService.Commands.CreateBankAccount.Handle(ctx, command)
		if err != nil {
			h.log.Errorf("(CreateBankAccount.Handle) id: %s, err: %v", id, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(BankAccount created) id: %s", id)
		return c.JSON(http.StatusCreated, id)
	}
}

// DepositBalance
// @Tags BankAccounts
// @Summary Create BankAccount
// @Description Create new Bank Account
// @Param BankAccount body commands.DepositBalanceCommand true "create Bank Account"
// @Accept json
// @Produce json
// @Param id path string true "bank account ID"
// @Success 200 {string} id ""
// @Router /accounts/deposit/{id} [post]
func (h *bankAccountHandlers) DepositBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.DepositBalance")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var command commands.DepositBalanceCommand
		if err := c.Bind(&command); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}
		command.AggregateID = c.Param(constants.ID)

		if err := h.validate.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		err := h.bankAccountService.Commands.DepositBalance.Handle(ctx, command)
		if err != nil {
			h.log.Errorf("(DepositBalance.Handle) id: %s, err: %v", command.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(balance deposited) id: %s, amount: %d", command.AggregateID)
		return c.NoContent(http.StatusOK)
	}
}

// WithdrawBalance
// @Tags BankAccounts
// @Summary Withdraw balance amount
// @Description Withdraw balance amount
// @Param BankAccount body commands.WithdrawBalanceCommand true "Withdraw balance amount"
// @Accept json
// @Produce json
// @Param id path string true "bank account ID"
// @Success 200 {string} id ""
// @Router /accounts/withdraw/{id} [post]
func (h *bankAccountHandlers) WithdrawBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.WithdrawBalance")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var command commands.WithdrawBalanceCommand
		if err := c.Bind(&command); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}
		command.AggregateID = c.Param(constants.ID)

		if err := h.validate.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		err := h.bankAccountService.Commands.WithdrawBalance.Handle(ctx, command)
		if err != nil {
			h.log.Errorf("(WithdrawBalance.Handle) id: %s, err: %v", command.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(balance withdraw) id: %s, amount: %d", command.AggregateID)
		return c.NoContent(http.StatusOK)
	}
}

// ChangeEmail
// @Tags BankAccounts
// @Summary Change Email
// @Description Change account email
// @Param BankAccount body commands.WithdrawBalanceCommand true "Change account email"
// @Accept json
// @Produce json
// @Param id path string true "bank account ID"
// @Success 200 {string} id ""
// @Router /accounts/email/{id} [post]
func (h *bankAccountHandlers) ChangeEmail() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.WithdrawBalance")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var command commands.ChangeEmailCommand
		if err := c.Bind(&command); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}
		command.AggregateID = c.Param(constants.ID)

		if err := h.validate.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		err := h.bankAccountService.Commands.ChangeEmail.Handle(ctx, command)
		if err != nil {
			h.log.Errorf("(ChangeEmail.Handle) id: %s, err: %v", command.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(balance withdraw) id: %s, amount: %d", command.AggregateID)
		return c.NoContent(http.StatusOK)
	}
}

// GetByID
// @Tags BankAccounts
// @Summary Get bank account by id
// @Description Get bank account by id
// @Param BankAccount body queries.GetBankAccountByIDQuery true "Get bank account by id"
// @Accept json
// @Produce json
// @Param id path string true "bank account ID"
// @Success 200 {object} domain.BankAccountMongoProjection
// @Router /accounts/{id} [get]
func (h *bankAccountHandlers) GetByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.GetByID")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var query queries.GetBankAccountByIDQuery
		if err := c.Bind(&query); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}
		query.AggregateID = c.Param(constants.ID)

		if err := h.validate.StructCtx(ctx, query); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		bankAccountProjection, err := h.bankAccountService.Queries.GetBankAccountByID.Handle(ctx, query)
		if err != nil {
			h.log.Errorf("(ChangeEmail.Handle) id: %s, err: %v", bankAccountProjection.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(get bank account) id: %s, amount: %d", bankAccountProjection.AggregateID)
		return c.JSON(http.StatusOK, bankAccountProjection)
	}
}

// Search
// @Tags BankAccounts
// @Summary Get bank account by id
// @Description Get bank account by id
// @Param BankAccount body queries.GetBankAccountByIDQuery true "Get bank account by id"
// @Accept json
// @Produce json
// @Param id path string true "bank account ID"
// @Param search path string true "search term"
// @Success 200 {object} queries.SearchQueryResult
// @Router /accounts/search [get]
func (h *bankAccountHandlers) Search() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.Search")
		defer span.Finish()
		//h.metrics.CreateOrderHttpRequests.Inc()

		var query queries.SearchBankAccountsQuery
		if err := c.Bind(&query); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		query.QueryTerm = c.Param("search")
		query.Pagination = utils.NewPaginationFromQueryParams(c.QueryParam(constants.Size), c.QueryParam(constants.Page))

		if err := h.validate.StructCtx(ctx, query); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		searchResult, err := h.bankAccountService.Queries.SearchBankAccounts.Handle(ctx, query)
		if err != nil {
			h.log.Errorf("(SearchBankAccounts.Handle) id: %s, err: %v", query.QueryTerm, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(search) result: %+v", searchResult)
		return c.JSON(http.StatusOK, searchResult)
	}
}
