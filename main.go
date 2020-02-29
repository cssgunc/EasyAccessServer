package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BaileyFrederick/EasyAccessServer/handler"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	// loads values from .env into the system
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	println("GOPATH set up correctly and project is working")

	err := setHandler()
	if err != nil {
		log.Println(err)
	}
}

func setHandler() error {
	// set up our global handler
	log.Println("setHandler")
	h, err := handler.New(handler.Config{
		Logger: log,
	})
	if err != nil {
		return errors.Wrap(err, "handler new")
	}

	log.Println(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	port := os.Getenv("PORT")
	server := &http.Server{
		Handler: h,
		Addr:    fmt.Sprintf(":%v", port),
	}

	// do graceful server shutdown
	go gracefulShutdown(server, time.Second*30)

	log.Infof("listening on port %v", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "cannot start a server")
	}
	return nil
}

// gracefulShutdown shuts down our server in a graceful way.
func gracefulShutdown(server *http.Server, timeout time.Duration) {
	sigStop := make(chan os.Signal)

	signal.Notify(sigStop, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	<-sigStop

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Fatal("graceful shutdown failed")
	}

	log.Info("graceful shutdown complete")
}
