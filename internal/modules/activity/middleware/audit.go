// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/dedysetiawan/product-catalog-project/internal/modules/activity/domain"
	"github.com/dedysetiawan/product-catalog-project/pkg/apperror"
	"github.com/dedysetiawan/product-catalog-project/pkg/middleware"
	"github.com/dedysetiawan/product-catalog-project/pkg/response"
)

func Audit(repo domain.Repository, permissions middleware.PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		entity, action := requestActivity(c)
		if entity == "" {
			c.Next()
			return
		}

		actorID, err := uuid.Parse(c.GetString(middleware.CtxUserID))
		if err != nil {
			response.Error(c, apperror.Unauthorized("invalid admin identity"))
			return
		}
		permission, ok := requiredPermission(entity, action)
		if !ok {
			c.Next()
			return
		}
		allowed, err := permissions.HasPermission(c.Request.Context(), actorID, permission)
		if err != nil {
			response.Error(c, apperror.Internal(err))
			return
		}
		if !allowed {
			response.Error(c, apperror.Forbidden("insufficient permission"))
			return
		}

		var before json.RawMessage
		id, _ := uuid.Parse(c.Param("id"))
		if id != uuid.Nil {
			before, err = repo.Snapshot(c.Request.Context(), entity, id)
			if err != nil {
				response.Error(c, apperror.Internal(err))
				return
			}
		}

		originalWriter := c.Writer
		capture := &responseCapture{ResponseWriter: originalWriter}
		c.Writer = capture
		c.Next()
		c.Writer = originalWriter

		status := capture.Status()
		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			if action == "CREATE" {
				id = responseID(capture.body.Bytes())
			}
			var after json.RawMessage
			if action != "DELETE" && id != uuid.Nil {
				after, err = repo.Snapshot(c.Request.Context(), entity, id)
				if err != nil {
					response.Error(c, apperror.Internal(err))
					return
				}
			}
			if err := repo.Record(c.Request.Context(), actorID, action, entity, before, after); err != nil {
				slog.Error("record admin activity", "entity", entity, "action", action, "err", err)
				response.Error(c, apperror.Internal(err))
				return
			}
		}

		originalWriter.WriteHeader(status)
		if capture.body.Len() > 0 {
			_, _ = originalWriter.Write(capture.body.Bytes())
		} else {
			originalWriter.WriteHeaderNow()
		}
	}
}

func requestActivity(c *gin.Context) (entity, action string) {
	switch c.Request.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
	default:
		return "", ""
	}
	path := c.Request.URL.Path
	adminPath := "/admin/"
	index := strings.Index(path, adminPath)
	if index < 0 {
		return "", ""
	}
	resource := strings.SplitN(path[index+len(adminPath):], "/", 2)[0]
	switch resource {
	case "products", "categories", "users", "roles":
		entity = resource
	case "contact-messages":
		entity = "contact_messages"
	default:
		return "", ""
	}

	switch {
	case entity == "contact_messages":
		action = "REPLY"
	case c.Request.Method == http.MethodPost:
		action = "CREATE"
	case c.Request.Method == http.MethodDelete:
		action = "DELETE"
	default:
		action = "UPDATE"
	}
	return entity, action
}

func requiredPermission(entity, action string) (string, bool) {
	switch action {
	case "CREATE":
		return entity + ":create", true
	case "UPDATE":
		return entity + ":edit", true
	case "DELETE":
		return entity + ":delete", true
	case "REPLY":
		return "contact_messages:reply", true
	default:
		return "", false
	}
}

func responseID(body []byte) uuid.UUID {
	var envelope struct {
		Data struct {
			ID uuid.UUID `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return uuid.Nil
	}
	return envelope.Data.ID
}

type responseCapture struct {
	gin.ResponseWriter
	status  int
	written bool
	body    bytes.Buffer
}

func (w *responseCapture) WriteHeader(status int) {
	if !w.written {
		w.status = status
		w.written = true
	}
}

func (w *responseCapture) WriteHeaderNow() {
	if !w.written {
		w.status = http.StatusOK
		w.written = true
	}
}

func (w *responseCapture) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	return w.body.Write(data)
}

func (w *responseCapture) WriteString(data string) (int, error) {
	w.WriteHeaderNow()
	return w.body.WriteString(data)
}

func (w *responseCapture) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *responseCapture) Size() int { return w.body.Len() }

func (w *responseCapture) Written() bool { return w.written }
