// Copyright (c) 2026 Dedy Setiawan. All rights reserved.
// Package apperror menyediakan error terstruktur yang dipetakan ke HTTP status.
package apperror

import (
	"errors"
	"net/http"
)

type Error struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func New(code, msg string, status int) *Error {
	return &Error{Code: code, Message: msg, HTTPStatus: status}
}

func NotFound(msg string) *Error     { return New("NOT_FOUND", msg, http.StatusNotFound) }
func Conflict(msg string) *Error     { return New("CONFLICT", msg, http.StatusConflict) }
func BadRequest(msg string) *Error   { return New("BAD_REQUEST", msg, http.StatusBadRequest) }
func Unauthorized(msg string) *Error { return New("UNAUTHORIZED", msg, http.StatusUnauthorized) }
func Forbidden(msg string) *Error    { return New("FORBIDDEN", msg, http.StatusForbidden) }
func InsufficientStock(msg string) *Error {
	return New("INSUFFICIENT_STOCK", msg, http.StatusUnprocessableEntity)
}

// Internal membungkus error tak terduga; pesan asli tidak dibocorkan ke client.
func Internal(err error) *Error {
	return &Error{Code: "INTERNAL_ERROR", Message: "internal server error", HTTPStatus: http.StatusInternalServerError, Err: err}
}

// Code mengembalikan kode error aplikasi, atau "" jika bukan *Error.
func Code(err error) string {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

// As mencoba meng-unwrap err menjadi *Error.
func As(err error) (*Error, bool) {
	var ae *Error
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
