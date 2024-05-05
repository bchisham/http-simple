package service

import (
	"github.com/gorilla/sessions"
)

type Session struct {
	Options sessions.Options
	Session *sessions.Session
	id      string
	store   sessions.Store
}

func NewSession(id string, options sessions.Options, store sessions.Store) *Session {
	return &Session{Options: options, id: id, store: store}
}

func (s *Session) Start(request Request) (*Session, error) {
	session, err := s.store.Get(request.httpRequest, request.sessionName)
	if err != nil {
		return nil, err
	}
	s.Session = session
	return s, nil
}

func (s *Session) Save(request Request) error {
	err := s.Session.Save(request.httpRequest, request.writer)
	if err != nil {
		return err
	}
	return nil
}
