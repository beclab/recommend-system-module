package json

import (
	"encoding/json"
	"errors"
	"net/http"

	"bytetrade.io/web3os/RSSync/common"
	"bytetrade.io/web3os/RSSync/http/response"

	"go.uber.org/zap"
)

const contentTypeHeader = `application/json`

// OK creates a new JSON response with a 200 status code.
func OK(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// Created sends a created response to the client.
func Created(w http.ResponseWriter, r *http.Request, body interface{}) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusCreated)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSON(body))
	builder.Write()
}

// NoContent sends a no content response to the client.
func NoContent(w http.ResponseWriter, r *http.Request) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusNoContent)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.Write()
}

// ServerError sends an internal error to the client.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusInternalServerError)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
	builder.Write()
}

// BadRequest sends a bad request error to the client.
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {

	builder := response.New(w, r)
	builder.WithStatus(http.StatusBadRequest)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(err))
	builder.Write()
}

// Unauthorized sends a not authorized error to the client.
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	builder := response.New(w, r)
	builder.WithStatus(http.StatusUnauthorized)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access Unauthorized")))
	builder.Write()
}

// Forbidden sends a forbidden error to the client.
func Forbidden(w http.ResponseWriter, r *http.Request) {

	builder := response.New(w, r)
	builder.WithStatus(http.StatusForbidden)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("access Forbidden")))
	builder.Write()
}

// NotFound sends a page not found error to the client.
func NotFound(w http.ResponseWriter, r *http.Request) {
	common.Logger.Error("[HTTP:Not Found]")

	builder := response.New(w, r)
	builder.WithStatus(http.StatusNotFound)
	builder.WithHeader("Content-Type", contentTypeHeader)
	builder.WithBody(toJSONError(errors.New("resource Not Found")))
	builder.Write()
}

func toJSONError(err error) []byte {
	type errorMsg struct {
		ErrorMessage string `json:"error_message"`
	}

	return toJSON(errorMsg{ErrorMessage: err.Error()})
}

func toJSON(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		common.Logger.Error("[HTTP:JSON] ", zap.Error(err))
		return []byte("")
	}

	return b
}
