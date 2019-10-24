package handler

import (
	"context"
	"encoding/json"

	// "fmt"
	"io/ioutil"
	"log"
	"net/http"
	// "os"
	// firebase "firebase.google.com/go"
	// "google.golang.org/api/iterator"
	// firestore "cloud.google.com/go/firestore"
)

func (h *Handler) queryColleges(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log.Println(ctx)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//change this to firestore token not string
	var idToken string
	err = json.Unmarshal(body, &idToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// Location = city/large, suburbs/midsize
type collegeParams struct {
	GPA      int32  `json:"GPA"`
	ZIP      string `json:"ZIP"`
	Distance string
	Size     int
	Location int
	Lat      float32
	Long     float32
}
