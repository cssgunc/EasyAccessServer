package handler

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var act int
var sat int

func (h *Handler) updateSelectivityScores(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	act = 0
	sat = 0
	var sACT []selectivityACT
	var sSAT []selectivitySAT
	file, err := os.Open("handler/test.csv")
	if err != nil {

	}
	csvfile := csv.NewReader(file)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		lowGPA, err := strconv.ParseFloat(record[1], 64)
		highGPA, err := strconv.ParseFloat(record[2], 64)
		lowScore, err := strconv.Atoi(record[3])
		highScore, err := strconv.Atoi(record[4])

		var tempACT selectivityACT
		var tempSAT selectivitySAT
		actBool := false
		if highScore > 34 {
			id := strconv.Itoa(sat)
			sat = sat + 1
			tempSAT = selectivitySAT{
				Score:   id + "SAT" + record[0],
				LowSAT:  lowScore,
				HighSAT: highScore,
				LowGPA:  lowGPA,
				HighGPA: highGPA,
			}
		} else {
			id := strconv.Itoa(act)
			act = act + 1
			tempACT = selectivityACT{
				Score:   id + "ACT" + record[0],
				LowACT:  lowScore,
				HighACT: highScore,
				LowGPA:  lowGPA,
				HighGPA: highGPA,
			}
			actBool = true
		}
		if actBool {
			sACT = append(sACT, tempACT)
		} else {
			sSAT = append(sSAT, tempSAT)
		}

	}
	//log.Println(sSAT)
	for _, record := range sACT {
		_, err := client.Collection("Selectivity").Doc(record.Score).Set(ctx, record)
		if err != nil {
			log.Fatalln(err)
		}
	}
	for _, record := range sSAT {
		_, err := client.Collection("Selectivity").Doc(record.Score).Set(ctx, record)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (h *Handler) updateSchoolNeedMet(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	file, err := os.Open("handler/school.csv")
	if err != nil {

	}
	csvfile := csv.NewReader(file)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		schoolName := record[0]
		needMet, err := strconv.Atoi(record[1])
		m := make(map[string]int)
		m[schoolName] = needMet
		_, err = client.Collection("NeedMet").Doc(schoolName).Set(ctx, m)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(schoolName, needMet)
	}
}

type selectivityACT struct {
	Score   string
	LowACT  int
	HighACT int
	LowGPA  float64
	HighGPA float64
}

type selectivitySAT struct {
	Score   string
	LowSAT  int
	HighSAT int
	LowGPA  float64
	HighGPA float64
}
