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
	// ProjectId := os.Getenv("ProjectID")

	// println("GOPATH set up correctly amd project is working")
	// ctx := context.Background()

	// conf := &firebase.Config{ProjectID: ProjectId}
	// app, err := firebase.NewApp(ctx, conf)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// client, err := app.Firestore(ctx)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// defer client.Close()

	// auth, err := app.Auth(ctx)
	// userRecord, err := auth.GetUserByEmail(ctx, "FrederickBailey18@gmail.com")

	// println(userRecord.UID)

	// //test to change info in firestore
	// p := User{
	// 	Name: "TEST",
	// }
	// //Changes the name of the specific user based on UID to ALICE
	// _, err = client.Collection("users").Doc("755O4T422rS1CgngVpI8").Set(ctx, p)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err := setHandler()
	if err != nil {
		log.Println(err)
	}
}

func setHandler() error {
	// set up our global handler
	h, err := handler.New(handler.Config{
		Logger: log,
	})
	if err != nil {
		return errors.Wrap(err, "handler new")
	}

	println(h)
  port := os.Getenv("PORT")
	server := &http.Server{
		Handler: h,
		Addr:    fmt.Sprintf(":%d", port),
	}

	// do graceful server shutdown
	go gracefulShutdown(server, time.Second*30)
  
	log.Infof("listening on port %d", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return errors.Wrap(err, "cannot start a server")
	}
	return nil
}

type User struct {
	Name string
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
