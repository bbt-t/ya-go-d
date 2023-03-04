package handlers

import (
	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/controller"
	"github.com/bbt-t/ya-go-d/internal/usecase"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type GopherMartHandler struct {
	s   *usecase.GopherMartService
	cfg *config.Config
}

func NewGopherMartRoutes(s *usecase.GopherMartService, cfg *config.Config) *GopherMartHandler {
	return &GopherMartHandler{
		s:   s,
		cfg: cfg,
	}
}

func (g GopherMartHandler) InitRoutes() *chi.Mux {
	/*
		Initialize the server, setting preferences and add routes.
	*/
	router := chi.NewRouter()
	router.Use(
		middleware.RealIP, // <- (!) Only if a reverse proxy is used (e.g. nginx) (!)
		middleware.Logger,
		middleware.Recoverer,
		// Compress:
		middleware.Compress(5, "text/html", "text/css"),
		// Working with paths:
		middleware.CleanPath,
	)

	router.Post("/api/user/register", g.reg)
	router.Post("/api/user/login", g.login)

	router.Group(func(r chi.Router) {
		r.Use(controller.RequireAuthentication(g.cfg))

		r.Post("/api/user/orders", g.order)
		r.Get("/api/user/orders", g.ordersAll)

		r.Post("/api/user/balance/withdraw", g.wd)
		r.Get("/api/user/withdrawals", g.wdAll)

		r.Get("/api/user/balance", g.getBalance)
	})

	return router
}
