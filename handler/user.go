package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
)

type message struct {
	UID string `json:"uid"`
}

func (h *Handler) authUser(w http.ResponseWriter, r *http.Request) {
	ProjectID := os.Getenv("ProjectID")
	fmt.Println("Test GET endpoint is being hit now!")
	ctx := context.Background()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("1")
		http.Error(w, err.Error(), 500)
		return
	}

	var idToken string
	err = json.Unmarshal(body, &idToken)
	if err != nil {
		log.Println("2")
		http.Error(w, err.Error(), 500)
		return
	}

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

	// auth, err := app.Auth(ctx)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// token, err := client.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	// if err != nil {
	// 	log.Fatalf("error verifying ID token: %v\n", err)
	// }

	// userInfo, err := client.Collection("users").Doc(token.UID).Get(ctx)
	output, err := json.Marshal(body)
	if err != nil {
		log.Println("3")
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func (h *Handler) userInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Info Endpoint")

}

type student struct {
	UID            string   `json:"uid"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	SchoolCode     string   `json:"schoolCode"`
	GraduationYear string   `json:"graduationYear"`
	WeightedGPA    float32  `json:"weightedGpa"`
	UnweightedGPA  float32  `json:"unweightedGpa"`
	ClassRank      int      `json:"classRank"`
	SAT            int      `json:"SAT"`
	ACT            int      `json:"ACT"`
	Size           string   `json:"size"`
	Location       string   `json:"location"`
	Diversity      string   `json:"diversity"`
	Majors         []string `json:"majors"`
	Distance       string   `json:"distance"`
	Zip            string   `json:"zip"`
	Matches        []string `json:"matches"`
}