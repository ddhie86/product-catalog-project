// Package response menyeragamkan bentuk JSON response API.
package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
)

type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

func NewMeta(page, limit, total int) *Meta {
	pages := 0
	if limit > 0 {
		pages = (total + limit - 1) / limit
	}
	return &Meta{Page: page, Limit: limit, TotalItems: total, TotalPages: pages}
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Envelope{Success: true, Data: data})
}

func Paginated(c *gin.Context, data any, meta *Meta) {
	c.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Error memetakan error ke response; error tak dikenal menjadi 500.
func Error(c *gin.Context, err error) {
	ae, ok := apperror.As(err)
	if !ok {
		ae = apperror.Internal(err)
	}
	if ae.HTTPStatus >= http.StatusInternalServerError {
		slog.Error("request failed", "path", c.FullPath(), "err", err)
	}
	c.AbortWithStatusJSON(ae.HTTPStatus, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: ae.Code, Message: ae.Message},
	})
}
