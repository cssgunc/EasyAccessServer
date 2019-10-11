package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
  "time"
  "encoding/json"

  firebase "firebase.google.com/go"
	"github.com/BaileyFrederick/EasyAccessServer/handler"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
  "github.com/sirupsen/logrus"
  // "google.golang.org/api/option"
  "golang.org/x/oauth2/google"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	// loads values from .env into the system
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
}

func createServiceAccount() *google.Credentials{
  Type := os.Getenv("TYPE")
  ProjectID := os.Getenv("PROJECT_ID")
  PrivateKeyID := os.Getenv("PRIVATE_KEY_ID")
  PrivateKey := os.Getenv("PRIVATE_KEY")
  ClientEmail := os.Getenv("CLIENT_EMAIL")
  ClientID := os.Getenv("CLIENT_ID")
  AuthURI := os.Getenv("AUTH_URI")
  TokenURI := os.Getenv("TOKEN_URI")
  AuthProviderX509CertURL := os.Getenv("AUTH_PROVIDER_X509_CERT_URL")
  ClientX509CertURL := os.Getenv("CLIENT_X509_CERT_URL")
  acc := serviceAccount{
    Type: Type,
    project_id: ProjectID,
    private_key_id: PrivateKeyID,
    private_key: PrivateKey,
    client_email: ClientEmail,
    client_id: ClientID,
    auth_uri: AuthURI,
    token_uri:TokenURI,
    auth_provider_x509_cert_url: AuthProviderX509CertURL,
    client_x509_cert_url: ClientX509CertURL,
  }
  data, err := json.Marshal(acc)
  if err != nil {
    log.Fatalln("err reading service account")
  }
  var account *google.Credentials
  err = json.Unmarshal(data, &account)
  return account
}

func main() {
  ctx := context.Background()
  ProjectID := os.Getenv("PROJECT_ID")
  
	println("GOPATH set up correctly and project is working")
	
	conf := &firebase.Config{ProjectID: ProjectID}
	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()

	auth, err := app.Auth(ctx)
  userRecord, err := auth.GetUserByEmail(ctx, "frederickbailey18@gmail.com")
  if err != nil {
    log.Fatalln(err)
  }

	println(userRecord.UID)

	//test to change info in firestore
	p := user{
		Name: "Our app is hosted MF",
	}
	//Changes the name of the specific user based on UID to ALICE
	_, err = client.Collection("users").Doc("755O4T422rS1CgngVpI8").Set(ctx, p)
	if err != nil {
		log.Fatal(err)
	}

	err = setHandler()
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
  println(port)
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

type user struct {
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


type serviceAccount struct{
  Type string
  project_id string
  private_key_id string
  private_key string
  client_email string
  client_id string
  auth_uri string
  token_uri string
  auth_provider_x509_cert_url string
  client_x509_cert_url string
}