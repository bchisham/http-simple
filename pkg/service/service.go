package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/bchisham/collections-go/sequence"
	"log"
	"net/http"
	"os"
	"time"
)

type Service interface {
	Start()
	Stop()
}

type Options struct {
	hostname              string
	port                  int
	requireTLS            bool
	requestTimeout        time.Duration
	certFile              string
	keyFile               string
	sessionKey            []byte
	disableOptionsHandler bool
	disableHealthHandler  bool
}

type Option func(*Options)

func WithHostname(hostname string) Option {
	return func(o *Options) {
		o.hostname = hostname
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.port = port
	}
}

func WithRequireTLS(requireTLS bool) Option {
	return func(o *Options) {
		o.requireTLS = requireTLS
	}
}

func WithRequestTimeout(requestTimeout time.Duration) Option {
	return func(o *Options) {
		o.requestTimeout = requestTimeout
	}
}

func WithCertFile(certFile string) Option {
	return func(o *Options) {
		o.certFile = certFile
	}

}

func WithKeyFile(keyFile string) Option {
	return func(o *Options) {
		o.keyFile = keyFile
	}
}

func WithDisableOptionsHandler(disableOptionsHandler bool) Option {
	return func(o *Options) {
		o.disableOptionsHandler = disableOptionsHandler
	}
}

func WithDisableHealthHandler(disableHealthHandler bool) Option {
	return func(o *Options) {
		o.disableHealthHandler = disableHealthHandler
	}
}

type service struct {
	Options
	ctx        context.Context
	cancelFunc context.CancelFunc
	srv        *http.Server
	mux        *http.ServeMux
}

func NewService(opts ...Option) Service {
	options := Options{
		hostname:       "localhost",
		port:           8080,
		requireTLS:     false,
		requestTimeout: 30 * time.Second,
	}

	_ = sequence.FromSlice(opts).Each(func(opt Option) error {
		opt(&options)
		return nil
	})
	srv, err := options.buildServer()
	if err != nil {
		log.Fatal(err)
	}
	return &service{
		Options: options,
		srv:     srv,
	}
}

func (o Options) hostAddr() string {
	return o.hostname + ":" + fmt.Sprintf("%d", o.port)
}

func (o Options) buildServer() (*http.Server, error) {
	// Build the server
	server := &http.Server{
		DisableGeneralOptionsHandler: o.disableOptionsHandler,
		Addr:                         o.hostAddr(),
		ReadTimeout:                  o.requestTimeout,
		WriteTimeout:                 o.requestTimeout,
		IdleTimeout:                  o.requestTimeout,
	}
	if o.requireTLS {
		tlsConfig, err := o.buildTLSConfig()
		if err != nil {
			return nil, err
		}
		server.TLSConfig = tlsConfig
	}
	return server, nil
}

func (o Options) buildTLSConfig() (*tls.Config, error) {
	// Build the TLS configuration
	cert, err := os.ReadFile(o.certFile)
	if err != nil {
		return nil, err
	}
	key, err := os.ReadFile(o.keyFile)
	if err != nil {
		return nil, err
	}
	certificate, err := tls.X509KeyPair(cert, key)
	return &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}, nil
}

func (s *service) Start() {
	// Start the service
	if s.ctx != nil {
		log.Fatal("Service already started")
	}
	s.ctx = context.Background()
	s.ctx, s.cancelFunc = context.WithCancel(s.ctx)
	s.mux = http.NewServeMux()
	s.mux.Handle("/", s)
	if !s.disableHealthHandler {
		s.mux.HandleFunc("/health", handleHealth)
	}
	if s.requireTLS {
		if err := s.srv.ListenAndServeTLS(s.certFile, s.keyFile); err != nil {
			log.Fatal(err)
		}
	} else {
		err := s.srv.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}

}

func (s *service) Stop() {
	// Stop the service
	if s.ctx == nil {
		log.Fatal("Service already stopped")
	}
	s.cancelFunc()
	s.ctx = nil
	if err := s.srv.Close(); err != nil {
		log.Fatal(err)
	}
}
