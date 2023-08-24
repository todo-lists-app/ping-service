package service

import (
	"fmt"
	validate "github.com/todo-lists-app/go-validate-user"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/keloran/go-healthcheck"
	"github.com/todo-lists-app/ping-service/internal/config"
	"github.com/todo-lists-app/ping-service/internal/ping"
	pb "github.com/todo-lists-app/protobufs/generated/ping/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Service struct {
	*config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		Config: cfg,
	}
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go startHTTP(s.Config, errChan)
	go startGRPC(s.Config, errChan)

	return <-errChan
}

func (s *Service) Health() error {
	os.Exit(0)
	return nil
}

func startHTTP(cfg *config.Config, errChan chan error) {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://api.todo-list.app",
		"https://todo-list.app",
		"https://beta.todo-list.app",
	}

	if cfg.Local.Development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			"GET",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-User-Subject",
			"X-User-Access-Token",
		},
		MaxAge: 300,
	}))
	r.Get("/health", healthcheck.HTTP)

	r.Route("/ping", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			subject := r.Header.Get("X-User-Subject")
			if subject == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token := r.Header.Get("X-User-Access-Token")
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			v, err := validate.NewValidate(r.Context(), cfg.Identity.Service, cfg.Local.Development).GetClient()
			valid, err := v.ValidateUser(token, subject)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errChan <- err
				return
			}
			if !valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			p := ping.NewPingService(r.Context(), *cfg, subject, &ping.RealMongoOperations{})
			if err := p.Ping(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				errChan <- err
				return
			}

			w.WriteHeader(http.StatusOK)
			return
		})
	})

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Local.HTTPPort),
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}
	logs.Local().Infof("starting http server on port %d", cfg.Local.HTTPPort)
	if err := srv.ListenAndServe(); err != nil {
		errChan <- err
	}
}

func startGRPC(cfg *config.Config, errChan chan error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Local.GRPCPort))
	if err != nil {
		errChan <- err
	}
	gs := grpc.NewServer()
	reflection.Register(gs)
	pb.RegisterPingServiceServer(gs, &ping.Server{
		Config: cfg,
	})
	logs.Local().Infof("starting grpc server on port %d", cfg.Local.GRPCPort)
	if err := gs.Serve(lis); err != nil {
		errChan <- err
	}
}
