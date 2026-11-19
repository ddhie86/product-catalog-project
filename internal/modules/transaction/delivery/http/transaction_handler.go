// Package http (transaction): HTTP handlers & routes modul Transaction.
package http

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	product "github.com/dedysetiawan/product-catalog-project/internal/modules/product/domain"
	"github.com/dedysetiawan/product-catalog-project/internal/modules/transaction/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

type TransactionHandler struct{ uc domain.ReportUsecase }

func NewTransactionHandler(uc domain.ReportUsecase) *TransactionHandler {
	return &TransactionHandler{uc: uc}
}

// RegisterAdminRoutes: rg harus sudah dipasangi middleware JWT.
func (h *TransactionHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/transactions", h.List)
	rg.GET("/transactions/:id", h.Detail)
	rg.GET("/reports/summary", h.Summary)
}

func (h *TransactionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, total, err := h.uc.GetTransactionList(c.Request.Context(), domain.ListFilter{
		Status: domain.Status(c.Query("status")), Page: page, Limit: limit,
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	page, limit = product.NormalizePage(page, limit)
	response.Paginated(c, items, response.NewMeta(page, limit, total))
}

func (h *TransactionHandler) Detail(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, apperror.BadRequest("invalid transaction id"))
		return
	}
	t, err := h.uc.GetTransactionDetail(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, t)
}

func (h *TransactionHandler) Summary(c *gin.Context) {
	s, err := h.uc.GetDashboardSummary(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, s)
}
