package app

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbt-t/ya-go-d/internal/adapter/storage"
	"github.com/bbt-t/ya-go-d/internal/app/accrualservice"
	"github.com/bbt-t/ya-go-d/internal/config"
	"github.com/bbt-t/ya-go-d/internal/controller"
	"github.com/bbt-t/ya-go-d/internal/controller/handlers"
	"github.com/bbt-t/ya-go-d/internal/usecase"
	"github.com/bbt-t/ya-go-d/pkg"
)

func Run(cfg *config.Config) {
	/*
		Creating usable objects via constructors for layers and start app.
	*/
	defer pkg.Log.Close()

	repo := storage.NewStorage(cfg)
	service := usecase.NewGopherMart(repo)
	h := handlers.NewGopherMartRoutes(service, cfg)
	server := controller.NewHTTPServer(cfg.ServerAddress, h.InitRoutes())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go newWorkerPool(ctx, cfg, repo, accrualservice.NewAccrualSystem(*cfg))

	go func() {
		pkg.Log.Err(server.UP())
	}()
	// Graceful shutdown:
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-gracefulStop

	ctxGrace, cancelGrace := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelGrace()

	if err := server.Stop(ctxGrace); err != nil {
		pkg.Log.Err(err)
	} else {
		pkg.Log.Err(errors.New("! SERVER STOPPED !"))
	}
}
