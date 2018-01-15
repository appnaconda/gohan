package gohan

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/appnaconda/gohan/database"
	"github.com/appnaconda/gohan/logger"
	"github.com/appnaconda/gohan/logger/logrus"
	"github.com/appnaconda/gohan/logger/option"
	"github.com/appnaconda/gohan/router"
	"github.com/rs/cors"
)

type HandlerFunc func(*ServiceContext, http.ResponseWriter, *http.Request)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Option interface {
	Apply(*Service) error
}

type Service struct {
	Context    context.Context
	Name       string
	Version    string
	Logger     logger.Logger
	router     *router.Router
	db         *sql.DB
	HttpClient *http.Client
	Tracer     Tracer
}

func New(ctx context.Context, opts ...Option) (*Service, error) {
	logLevel := logger.DEBUG
	logFormat := logger.JSON_FORMAT

	logLevelValue := os.Getenv("LOG_LEVEL")
	if logLevelValue != "" {
		if level, ok := logger.ParseLevel(logLevelValue); ok {
			logLevel = level
		}
	}

	logFormatValue := os.Getenv("LOG_FORMAT")
	if logFormatValue != "" {
		if format, ok := logger.ParseFormat(logFormatValue); ok {
			logFormat = format
		}
	}

	service := &Service{
		Context: ctx,
		Logger: logrus.New(
			option.WithLevel(logLevel),
			option.WithFormat(logFormat),
		),
		router:     router.New(),
		HttpClient: http.DefaultClient,
	}

	if _, found := os.LookupEnv("DB_CONN_STR"); found {
		db, err := database.New(service.Context)
		if err != nil {
			return nil, err
		}

		service.db = db
	}

	for _, opt := range opts {
		if err := opt.Apply(service); err != nil {
			return nil, err
		}
	}

	return service, nil
}

func (s *Service) GET(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodGet, path, handle, middlewares...)
}

func (s *Service) HEAD(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodHead, path, handle, middlewares...)
}

func (s *Service) OPTIONS(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodOptions, path, handle, middlewares...)
}

func (s *Service) POST(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodPost, path, handle, middlewares...)
}

func (s *Service) PUT(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodPut, path, handle, middlewares...)
}

func (s *Service) PATCH(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodPatch, path, handle, middlewares...)
}

func (s *Service) DELETE(path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	s.Handle(http.MethodDelete, path, handle, middlewares...)
}

func (s *Service) Handle(method string, path string, handle HandlerFunc, middlewares ...MiddlewareFunc) {
	for _, middleware := range middlewares {
		handle = middleware(handle)
	}
	s.router.Handle(method, path, s.wrapHandle(handle))
}

func (s *Service) wrapHandle(h HandlerFunc) router.HandlerFunc {
	next := func(w http.ResponseWriter, req *http.Request) {

		requestUuid, err := NewUUID()
		if err != nil {
			s.Logger.Warnf("failed generating a new request UUID: %+v", err)
		}

		traceUuid, err := NewUUID()
		if err != nil {
			s.Logger.Warnf("failed generating a new trace UUID: %+v", err)
		}

		serviceContext := &ServiceContext{
			Logger: s.Logger.With(logger.Fields{
				"request_uuid": requestUuid,
				"trace_id":     traceUuid,
				"handler":      GetFuncName(h),
			}),
			db:         s.db,
			Context:    s.Context,
			HttpClient: s.HttpClient,
		}

		h(serviceContext, w, req)

		return
	}

	return next
}

func (s *Service) Run(port int) {
	var handler http.Handler

	if s.Tracer != nil {
		handler = s.Tracer.HTTPHandler(s.router)
	} else {
		handler = s.router
	}

	handler = cors.AllowAll().Handler(handler)

	s.Logger.Debugf("Starting Service on port %d", port)

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: handler}

	go func() {
		server.ListenAndServe()
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTTIN)

	sig := <-c
	s.Logger.Debugf("Signal received: %+v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		s.Logger.Errorf("failed shutting down the http server: %+v", err)
	} else {
		s.Logger.Debugf("the http server was shutdown gracefully")
	}
}

func (s *Service) Close() {
	s.Logger.Debug("shutting down the service")
	if s.db != nil {
		err := s.db.Close()
		if err != nil {
			s.Logger.Errorf("failed closing the database connection: %s", err)
		}
	}

	s.Logger.Debug("the service was shutdown gracefully")
}
