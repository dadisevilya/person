package skeleton

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// HTTPHandler base interface to be implemented by all HTTP handlers
type HTTPHandler interface {
	RegisterRoutes(router *mux.Router)
}

// BaseHTTPHandler common functionality for HTTP handlers
type BaseHTTPHandler struct {
	Render *render.Render
}

// JSON used to return JSON response
func (h *BaseHTTPHandler) JSON(writer io.Writer, httpCode int, v interface{}) {
	_ = h.Render.JSON(writer, httpCode, v)
}

// CloseBody used to close body Reader in the end of request
func (h *BaseHTTPHandler) CloseBody(body io.ReadCloser) {
	_ = body.Close()
}

// BadRequestError helper for returning BadRequest HTTP status and error
func (h *BaseHTTPHandler) BadRequestError(err error) APIError {
	return NewAPIError(
		http.StatusText(http.StatusBadRequest),
		err,
	)
}
