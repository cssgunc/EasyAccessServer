package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	firestore "cloud.google.com/go/firestore"
	"firebase.google.com/go/auth"
	"google.golang.org/api/iterator"
)

var user student

//Verify user
func Verify(idToken string) (*auth.Token, error) {
	ctx := context.Background()
	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	token, err := auth.VerifyIDTokenAndCheckRevoked(ctx, idToken)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func scoreStudent(UID string) int {
	ctx := context.Background()
	log.Println(user.UID)
	userInfo, err := client.Collection("users").Doc(UID).Get(ctx)
	if err != nil {
		log.Println(err)
	}

	var student student
	userInfo.DataTo(&student)
	log.Println("SAT: ", student.SAT, " ACT: ", student.ACT, " GPA: ", student.UnweightedGPA)
	var potentialScores []string
	iter := client.Collection("Selectivity").Where("LowGPA", "<=", student.UnweightedGPA).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		var s selectivity
		doc.DataTo(&s)
		if s.HighACT != 0 {
			if s.LowACT <= student.ACT && student.ACT <= s.HighACT && s.LowGPA <= float64(student.UnweightedGPA) && float64(student.UnweightedGPA) <= s.HighGPA {
				potentialScores = append(potentialScores, s.Score)
			}
		} else {
			if s.LowSAT <= student.SAT && student.SAT <= s.HighSAT && s.LowGPA <= float64(student.UnweightedGPA) && float64(student.UnweightedGPA) <= s.HighGPA {
				potentialScores = append(potentialScores, s.Score)
			}
		}
	}
	topScore := 0
	for _, score := range potentialScores {
		log.Println("Potential score: ", score)
		i, err := strconv.Atoi(score[len(score)-1:])
		if err != nil {
			log.Fatalln(err)
		}
		if i > topScore {
			topScore = i
		}
	}
	log.Println(topScore)
	return topScore
}

//AuthUser is
func (h *Handler) AuthUser(w http.ResponseWriter, r *http.Request) {
	log.Println("User Endpoint")
	ctx := context.Background()
	tokenBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var idToken string
	err = json.Unmarshal(tokenBody, &idToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	userInfo, err := client.Collection("users").Doc(idToken).Get(ctx)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	userInfo.DataTo(&user)

	//TODO output pastMatches to frontend
	pastMatches := loadUserMatches()
	log.Println(pastMatches)

	output, err := json.Marshal(userInfo.Data())
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	//w.Write(outputToken)
	w.Write(output)
	return
}

func (h *Handler) addUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tokenBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var idToken string
	err = json.Unmarshal(tokenBody, &idToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = Verify(idToken)
	if err != nil {
		log.Printf("error verifying ID token: %v\n", err)
		http.Error(w, err.Error(), 401)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var user student
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = client.Collection("users").Doc(user.UID).Set(ctx, user)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

type updateInfo struct {
	UID  string             `json:"uid"`
	Info []firestore.Update `json:"info"`
}

func (h *Handler) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	tokenBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var idToken string
	err = json.Unmarshal(tokenBody, &idToken)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	_, err = Verify(idToken)
	if err != nil {
		log.Printf("error verifying ID token: %v\n", err)
		http.Error(w, err.Error(), 401)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var newInfo *updateInfo
	err = json.Unmarshal(body, &newInfo)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	userRef := client.Collection("users").Doc(newInfo.UID)
	_, err = userRef.Update(ctx, newInfo.Info)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	return

}

// type college struct {
// 	AcceptanceRate float64 `json:"Acceptance Rate"`
// 	AverageGPA     float64 `json:"Average GPA"`
// 	AverageSAT     int64   `json:"Average SAT"`
// 	Diversity      float32 `json:"Diversity"`
// 	Name           string  `json:"Name"`
// 	Size           int64   `json:"Size"`
// 	Zip            int64   `json:"Zip Code"`
// }

type student struct {
	UID            string   `json:"uid"`
	FirstName      string   `json:"firstname"`
	LastName       string   `json:"lastname"`
	Email          string   `json:"email"`
	SchoolCode     string   `json:"schoolCode"`
	GraduationYear string   `json:"graduationYear"`
	WeightedGPA    float32  `json:"weightedGpa"`
	UnweightedGPA  float32  `json:"unweightedGpa"`
	ClassRank      int      `json:"classRank"`
	SAT            int      `json:"SAT"`
	ACT            int      `json:"ACT"`
	Size           string   `json:"size"`
	State          string   `json:"state"`
	Diversity      string   `json:"diversity"`
	Majors         []string `json:"majors"`
	Distance       string   `json:"distance"`
	Zip            string   `json:"zip"`
	AbilityToPay   int      `json:"abilityToPay"`
}
