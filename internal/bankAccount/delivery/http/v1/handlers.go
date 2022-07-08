package v1

import (
	"github.com/AleksK1NG/go-cqrs-eventsourcing/config"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/commands"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/queries"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/bankAccount/service"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/mappers"
	"github.com/AleksK1NG/go-cqrs-eventsourcing/internal/metrics"
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
	"strconv"
)

type bankAccountHandlers struct {
	group              *echo.Group
	middlewareManager  middlewares.MiddlewareManager
	log                logger.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	validate           *validator.Validate
	metrics            *metrics.ESMicroserviceMetrics
}

func NewBankAccountHandlers(
	group *echo.Group,
	middlewareManager middlewares.MiddlewareManager,
	log logger.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	validate *validator.Validate,
	metrics *metrics.ESMicroserviceMetrics,
) *bankAccountHandlers {
	return &bankAccountHandlers{
		group:              group,
		middlewareManager:  middlewareManager,
		log:                log,
		cfg:                cfg,
		bankAccountService: bankAccountService,
		validate:           validate,
		metrics:            metrics,
	}
}

func (h *bankAccountHandlers) CreateBankAccount() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.CreateBankAccount")
		defer span.Finish()
		h.metrics.HttpCreateBankAccountRequests.Inc()

		var command commands.CreateBankAccountCommand
		if err := c.Bind(&command); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		command.AggregateID = uuid.NewV4().String()

		if err := h.validate.StructCtx(ctx, command); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		err := h.bankAccountService.Commands.CreateBankAccount.Handle(ctx, command)
		if err != nil {
			h.log.Errorf("(CreateBankAccount.Handle) id: %s, err: %v", command.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(BankAccount created) id: %s", command.AggregateID)
		return c.JSON(http.StatusCreated, command.AggregateID)
	}
}

func (h *bankAccountHandlers) DepositBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.DepositBalance")
		defer span.Finish()
		h.metrics.HttpDepositBalanceRequests.Inc()

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

func (h *bankAccountHandlers) WithdrawBalance() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.WithdrawBalance")
		defer span.Finish()
		h.metrics.HttpWithdrawBalanceRequests.Inc()

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

func (h *bankAccountHandlers) ChangeEmail() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.WithdrawBalance")
		defer span.Finish()
		h.metrics.HttpChangeEmailRequests.Inc()

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

func (h *bankAccountHandlers) GetByID() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.GetByID")
		defer span.Finish()
		h.metrics.HttpGetBuIdRequests.Inc()

		var query queries.GetBankAccountByIDQuery
		if err := c.Bind(&query); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		query.AggregateID = c.Param(constants.ID)

		fromStore := c.QueryParam("store")
		if fromStore != "" {
			isFromStore, err := strconv.ParseBool(fromStore)
			if err != nil {
				h.log.Errorf("strconv.ParseBool err: %v", tracing.TraceWithErr(span, err))
				return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
			}
			query.FromEventStore = isFromStore
		}

		if err := h.validate.StructCtx(ctx, query); err != nil {
			h.log.Errorf("(validate) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		bankAccountProjection, err := h.bankAccountService.Queries.GetBankAccountByID.Handle(ctx, query)
		if err != nil {
			h.log.Errorf("(ChangeEmail.Handle) id: %s, err: %v", query.AggregateID, tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		h.log.Infof("(get bank account) id: %s", bankAccountProjection.AggregateID)
		return c.JSON(http.StatusOK, mappers.BankAccountMongoProjectionToHttp(bankAccountProjection))
	}
}

func (h *bankAccountHandlers) Search() echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, span := tracing.StartHttpServerTracerSpan(c, "bankAccountHandlers.Search")
		defer span.Finish()
		h.metrics.HttpSearchRequests.Inc()

		var query queries.SearchBankAccountsQuery
		if err := c.Bind(&query); err != nil {
			h.log.Errorf("(Bind) err: %v", tracing.TraceWithErr(span, err))
			return httpErrors.ErrorCtxResponse(c, err, h.cfg.Http.DebugErrorsResponse)
		}

		query.QueryTerm = c.QueryParam("search")
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
		response := mappers.SearchResultToHttp(searchResult.List, searchResult.PaginationResponse)

		h.log.Infof("(search) result: %+v", response)
		return c.JSON(http.StatusOK, response)
	}
}
