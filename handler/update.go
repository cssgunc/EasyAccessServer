package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var act int
var sat int

func (h *Handler) updateMajorInfo(w http.ResponseWriter, r *http.Request) {
	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		log.Fatalln(err)
	}

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	params.Add("fields", "school.name,latest.programs.cip_4_digit.code")
	//Limited to 100 per page max
	params.Add("per_page", "100")

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters
	log.Printf("Encoded URL is %q\n", baseURL.String())
	response, err := http.Get(baseURL.String())
	if err != nil {
		log.Fatalln(err)
	}

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
	}

	var scorecardColleges cipScoreCardResponse
	err = json.Unmarshal(resBody, &scorecardColleges)
	if err != nil {
		log.Fatalln(err)
	}
	var totalPages float64
	//only need top two CC results for safety just get one page
	//gets total amount of pages from metadata
	totalPages = math.Ceil(float64(scorecardColleges.Metadata.Total) / float64(scorecardColleges.Metadata.PerPage))

	//loops through remaining pages and takes in results and addes them to our array of colleges
	for i := 1; i < int(totalPages); i++ {
		a := strconv.Itoa(i)
		params.Add("page", ""+a)
		baseURL.RawQuery = params.Encode()
		response, err := http.Get(baseURL.String())
		if err != nil {
			log.Fatalln(err)
		}

		resBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var tempColleges cipScoreCardResponse
		err = json.Unmarshal(resBody, &tempColleges)
		if err != nil {
			log.Fatalln(err)
		}
		scorecardColleges.Results = append(scorecardColleges.Results, tempColleges.Results...)
	}
	var codeMap map[string][]string
	codeMap = make(map[string][]string)
	for _, c := range scorecardColleges.Results {
		if strings.Contains(c.SchoolName, "/") {
			c.SchoolName = strings.ReplaceAll(c.SchoolName, "/", " ")
		}
		for _, code := range c.Codes {
			codeMap[code.Code] = append(codeMap[code.Code], c.SchoolName)
		}
	}
	ctx := context.Background()
	for code, schools := range codeMap {
		code = strings.TrimPrefix(code, "0")

		_, err = client.Collection("majors").Doc(code).Set(ctx, map[string]interface{}{
			"schools": schools,
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	w.WriteHeader(http.StatusOK)
	return
}

type schoolCipCodes struct {
	Schools []string
}

type cipScoreCardResponse struct {
	Metadata Metadata    `json:"metadata"`
	Results  []cipResult `json:"results"`
}

type cipResult struct {
	SchoolName string `json:"school.name"`
	Codes      []struct {
		Code string `json:"code"`
	} `json:"latest.programs.cip_4_digit"`
}

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
		schoolName := strings.TrimSpace(record[0])
		needMet, err := strconv.Atoi(record[1])
		m := make(map[string]int)
		m["NeedMet"] = needMet
		_, err = client.Collection("NeedMet").Doc(schoolName).Set(ctx, m)
		if err != nil {
			log.Fatalln(err)
		}

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
