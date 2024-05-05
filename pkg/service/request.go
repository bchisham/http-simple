package service

import (
	"context"
	"github.com/google/uuid"
	"net/http"
)

const (
	// SessionName is the name of the session
	SessionName = "http-session"
)

type Request struct {
	id          uuid.UUID
	sessionName string
	httpRequest *http.Request
	writer      http.ResponseWriter
	ctx         context.Context
}

func (r *Request) ID() uuid.UUID {
	return r.id
}

func (r *Request) Context() context.Context {
	return r.ctx
}

func (r *Request) HTTPRequest() *http.Request {
	return r.httpRequest
}

func NewRequest(ctx context.Context, httpRequest *http.Request, writer http.ResponseWriter) *Request {
	return &Request{id: uuid.New(), sessionName: SessionName, ctx: ctx, httpRequest: httpRequest, writer: writer}
}

func (r *Request) WithSessionName(sessionName string) *Request {
	r.sessionName = sessionName
	return r
}

func (r *Request) SessionName() string {
	return r.sessionName
}

func (r *Request) Writer() http.ResponseWriter {
	return r.writer
}

func (r *Request) ResponseBuilder() ResponseBuilder {
	return &responseBuilder{request: r}
}
