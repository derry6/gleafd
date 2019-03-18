package server

import (
	"net"
	"net/http"
	"sync/atomic"

	"github.com/derry6/gleafd/pkg/log"
)

type Server struct {
	// logger
	logger log.Logger
	// http server
	httpSvr *http.Server
	//
	svc    Service
	closed int32
}

func (s *Server) Closed() bool {
	return atomic.LoadInt32(&s.closed) == 1
}

func (s *Server) ListenAndServe(addr string) error {
	defer func() {
		s.Close()
	}()
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		s.logger.Errorw("Can not listen", "addr", addr, "err", err)
		return err
	}
	if err = s.httpSvr.Serve(lis); err != nil {
		s.logger.Errorw("Server serve error", "err", err)
	}
	return err
}

func (s *Server) Close() (err error) {
	if atomic.CompareAndSwapInt32(&s.closed, 0, 1) {
		if err = s.httpSvr.Close(); err != nil {
			s.logger.Errorw("Server close error", "err", err)
		}
	}
	return nil
}

func New(svc Service, logger log.Logger) (*Server, error) {
	if logger == nil {
		logger = log.DefaultLogger
	}
	hdlr := NewHttpHandler(svc, logger)
	s := &Server{
		logger:  logger,
		svc:     svc,
		httpSvr: &http.Server{Handler: hdlr},
	}
	return s, nil
}
