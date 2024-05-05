package service

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bchisham/collections-go/contracts"
	"github.com/bchisham/collections-go/stream"
	"log/slog"
	"net/http"
)

// ResponseDataFunc is a function that returns the response data. It is used to defer the execution of the response data.
type ResponseDataFunc func() ([]byte, error)

// BinaryData returns a ResponseDataFunc that returns the provided data
func BinaryData(data []byte) ResponseDataFunc {
	return func() ([]byte, error) {
		return data, nil
	}
}

// StringData returns a ResponseDataFunc that returns the provided data
func StringData(data string) ResponseDataFunc {
	return func() ([]byte, error) {
		return []byte(data), nil
	}
}

// JSONData returns a ResponseDataFunc that returns the provided data as JSON
func JSONData(data interface{}) ResponseDataFunc {
	return func() ([]byte, error) {
		return json.Marshal(data)
	}
}

// BinaryStreamData returns a ResponseDataFunc that returns the provided data
func BinaryStreamData(ctx context.Context, request Request, ch chan []byte) ResponseDataFunc {
	s := stream.FromChanWithContext(ch, ctx)
	w := request.Writer()
	defer func() {
		err := s.Close()
		if err != nil {
			slog.ErrorContext(ctx, "error closing stream", err)
		}
	}()
	return func() ([]byte, error) {
		var endOfStream []byte
		err := s.Each(func(v []byte) error {
			_, err := w.Write(v)
			if err != nil {
				slog.ErrorContext(ctx, "error writing to stream", err)
				return err
			}
			return nil
		})
		if err != nil && errors.Is(err, contracts.ErrContextDone) {
			slog.DebugContext(ctx, "context done")
			return endOfStream, nil

		}
		return endOfStream, err
	}
}

type ResponseBuilder interface {
	WithBody(body []byte) ResponseBuilder
	WithHeader(key, value string) ResponseBuilder
	WithBodyFunc(bodyFunc ResponseDataFunc) ResponseBuilder
}

type Response struct {
	state *responseBuilder
}

type responseBuilder struct {
	request  *Request
	bodyFunc ResponseDataFunc
}

func (r *responseBuilder) WithHeader(key, value string) ResponseBuilder {
	r.request.Writer().Header().Set(key, value)
	return r
}

func (r *responseBuilder) WithStatus(status int) ResponseBuilder {
	r.request.Writer().WriteHeader(status)
	return r
}

func (r *responseBuilder) WithBody(body []byte) ResponseBuilder {
	r.bodyFunc = BinaryData(body)
	return r
}

func (r *responseBuilder) WithBodyFunc(bodyFunc ResponseDataFunc) ResponseBuilder {
	r.bodyFunc = bodyFunc
	return r
}

func (s *service) BadRequest(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusBadRequest, "Bad Request")
}

func (s *service) NotFound(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusNotFound, "Not Found")
}

func (s *service) InternalServerError(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
}

func (s *service) Unauthorized(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusUnauthorized, "Unauthorized")
}

func (s *service) Forbidden(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusForbidden, "Forbidden")
}

func (s *service) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusMethodNotAllowed, "Method Not Allowed")
}

func (s *service) Conflict(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusConflict, "Conflict")
}

func (s *service) Gone(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusGone, "Gone")
}

func (s *service) TooManyRequests(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusTooManyRequests, "Too Many Requests")
}

func (s *service) NotImplemented(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusNotImplemented, "Not Implemented")
}

func (s *service) ServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusServiceUnavailable, "Service Unavailable")
}

func (s *service) GatewayTimeout(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusGatewayTimeout, "Gateway Timeout")
}

func (s *service) InsufficientStorage(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusInsufficientStorage, "Insufficient Storage")

}

func (s *service) LoopDetected(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusLoopDetected, "Loop Detected")
}

func (s *service) NotExtended(w http.ResponseWriter, r *http.Request) {
	s.ErrorResponse(w, r, http.StatusNotExtended, "Not Extended")
}

func (s *service) ErrorResponse(w http.ResponseWriter, r *http.Request, responseCode int, message string) {
	http.Error(w, message, responseCode)
}

func (r *Response) Send() error {
	body, err := r.state.bodyFunc()
	if err != nil {
		slog.ErrorContext(r.state.request.Context(), "error getting response body", err)
		return err
	}
	_, err = r.state.request.Writer().Write(body)
	if err != nil {
		slog.ErrorContext(r.state.request.Context(), "error writing response body", err)
		return err
	}
	return nil
}
