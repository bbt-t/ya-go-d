package app

import (
	"context"
	"log"
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
)

func Run(cfg *config.Config) {
	/*
		Creating usable objects via constructors for layers and start app.
	*/
	repo := storage.NewStorage(cfg)
	service := usecase.NewGopherMart(repo)
	h := handlers.NewGopherMartRoutes(service, cfg)
	server := controller.NewHTTPServer(cfg.ServerAddress, h.InitRoutes())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go newWorkerPool(ctx, cfg, repo, accrualservice.NewAccrualSystem(*cfg))

	go func() {
		log.Println(server.UP())
	}()
	// Graceful shutdown:
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-gracefulStop

	ctxGrace, cancelGrace := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelGrace()

	if err := server.Stop(ctxGrace); err != nil {
		log.Printf("! Error shutting down server: !\n%v", err)
	} else {
		log.Println("! SERVER STOPPED !")
	}
}
