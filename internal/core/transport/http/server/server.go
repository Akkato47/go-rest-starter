package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Config struct {
	Port            string
	ShutdownTimeout time.Duration
	DrainDelay      time.Duration
	HardStopDelay   time.Duration
}

type Server struct {
	engine         *gin.Engine
	cfg            Config
	logger         *slog.Logger
	isShuttingDown atomic.Bool
}

func New(cfg Config, logger *slog.Logger, middleware ...gin.HandlerFunc) *Server {
	s := &Server{cfg: cfg, logger: logger}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.SetTrustedProxies(nil)
	engine.Use(gin.Recovery())
	engine.Use(middleware...)

	engine.Use(func(c *gin.Context) {
		if s.isShuttingDown.Load() {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "server shutting down"})
			return
		}
		c.Next()
	})

	engine.GET("/healthz", func(c *gin.Context) {
		if s.isShuttingDown.Load() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "shutting down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	s.engine = engine
	return s
}

func (s *Server) RegisterSwagger() {
	handler := ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json"))

	s.engine.GET("/swagger/*any", func(c *gin.Context) {
		if c.Param("any") == "/" {
			c.Redirect(http.StatusFound, "/swagger/index.html")
			return
		}
		handler(c)
	})
}

func (s *Server) RegisterGroups(groups ...*RouterGroup) {
	for _, g := range groups {
		g.register(s.engine)
	}
}

func (s *Server) Run(ctx context.Context) error {
	ongoingCtx, stopOngoing := context.WithCancel(context.Background())

	srv := &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      s.engine,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
	}

	ch := make(chan error, 1)
	go func() {
		s.logger.Info("starting server", slog.String("port", s.cfg.Port))
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		stopOngoing()
		return err
	case <-ctx.Done():
		s.isShuttingDown.Store(true)
		srv.SetKeepAlivesEnabled(false)
		s.logger.Info("shutdown signal received, draining...")

		time.Sleep(s.cfg.DrainDelay)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		stopOngoing()

		if err != nil {
			s.logger.Error("graceful shutdown failed, forcing close")
			_ = srv.Close()
			time.Sleep(s.cfg.HardStopDelay)
			return err
		}

		s.logger.Info("server stopped gracefully")
	}

	return nil
}
