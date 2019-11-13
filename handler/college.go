package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	// "os"
	// firebase "firebase.google.com/go"
	// "google.golang.org/api/iterator"
	// firestore "cloud.google.com/go/firestore"
)

type selectivity struct {
	Score   string
	LowACT  int
	HighACT int
	LowGPA  float64
	HighGPA float64
	LowSAT  int
	HighSAT int
}

func (h *Handler) collegeMajors(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("handler/majors.csv")
	if err != nil {

	}
	csvfile := csv.NewReader(file)
	var majors []string
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		majors = append(majors, record[0][19:])
	}
	output, err := json.Marshal(majors)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return

	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	// var college string
	// err = json.Unmarshal(body, &college)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// }

	// baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	// if err != nil {
	// 	fmt.Println("Malformed URL: ", err.Error())
	// 	return
	// }

	// // Prepare Query Parameters
	// params := url.Values{}
	// params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	// params.Add("school.name", "Chapel Hill")
	// //params.Add("fields", "latest.academics.program_percentage.computer,school.name,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall")

	// baseURL.RawQuery = params.Encode() // Escape Query Parameters
	// log.Printf("Encoded URL is %q\n", baseURL.String())

	// response, err := http.Get(baseURL.String())
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// }

	// resBody, err := ioutil.ReadAll(response.Body)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	//
	// var colleges majorResponse
	// err = json.Unmarshal(resBody, &colleges)
	// if err != nil {
	// 	http.Error(w, err.Error(), 500)
	// 	return
	// }

	// for _, college := range colleges.Results {
	// 	log.Println(college.Latest)
	// }

	// buf := &bytes.Buffer{}
	// gob.NewEncoder(buf).Encode(colleges.Results)
	// bs := buf.Bytes()
	// w.Header().Set("content-type", "application/json")
	// w.Write(bs)
	// return
}

func getCollegeRanges(score int) ([]CollegeSelectivityInfo, error) {
	ctx := context.Background()
	targetScore := strconv.Itoa(score)
	targetInfo, err := client.Collection("Selectivity").Doc(targetScore).Get(ctx)
	if err != nil {
		return nil, err
	}
	reachScore := strconv.Itoa(score + 1)
	log.Println("ReachScore: ", reachScore)
	reachInfo, err := client.Collection("Selectivity").Doc(reachScore).Get(ctx)
	if err != nil {
		return nil, err
	}
	var target CollegeSelectivityInfo
	var reach CollegeSelectivityInfo
	var safety CollegeSelectivityInfo
	if score != 1 {
		safetyScore := strconv.Itoa(score - 1)
		safetyInfo, err := client.Collection("Selectivity").Doc(safetyScore).Get(ctx)
		if err != nil {
			return nil, err
		}
		safetyInfo.DataTo(&safety)
	}

	targetInfo.DataTo(&target)
	reachInfo.DataTo(&reach)

	info := []CollegeSelectivityInfo{safety, target, reach}

	return info, nil
}

func queryColleges(selectivityInfo *CollegeSelectivityInfo, queryParams collegeParams) ([]Result, error) {

	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		return nil, err
	}

	//sets ranges of possible scores to limit query
	lowAct := strconv.Itoa(selectivityInfo.ACT[0])
	highAct := strconv.Itoa(selectivityInfo.ACT[1])
	lowSat := strconv.Itoa(selectivityInfo.SAT[0])
	highSat := strconv.Itoa(selectivityInfo.SAT[1])
	rate := strconv.Itoa(selectivityInfo.Rate)

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	params.Add("school.region_id", queryParams.Region)
	params.Add("school.degrees_awarded.highest__range", "3..")
	params.Add("fields", "school.name,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall")
	params.Add("per_page", "100")
	params.Add("latest.admissions.act_scores.midpoint.cumulative__range", lowAct+".."+highAct)
	params.Add("latest.admissions.admission_rate.overall__range", ".."+rate)
	params.Add("latest.admissions.sat_scores.average.overall__range", lowSat+".."+highSat)

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters
	log.Printf("Encoded URL is %q\n", baseURL.String())
	response, err := http.Get(baseURL.String())
	if err != nil {
		return nil, err
	}

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var colleges ScoreCardResponse
	err = json.Unmarshal(resBody, &colleges)
	if err != nil {
		return nil, err
	}
	log.Println("total: ", colleges.Metadata.Total)
	log.Println("page: ", colleges.Metadata.Page)

	totalPages := math.Ceil(float64(colleges.Metadata.Total) / float64(colleges.Metadata.PerPage))
	log.Println("totalPages: ", totalPages)
	for i := 1; i < int(totalPages); i++ {
		a := strconv.Itoa(i)
		params.Add("page", ""+a)
		baseURL.RawQuery = params.Encode()
		response, err := http.Get(baseURL.String())
		if err != nil {
			return nil, err
		}

		resBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		var tempColleges ScoreCardResponse
		err = json.Unmarshal(resBody, &tempColleges)
		if err != nil {
			return nil, err
		}
		colleges.Results = append(colleges.Results, tempColleges.Results...)
	}

	return colleges.Results, nil
}

func (h *Handler) getMatches(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	log.Println(ctx)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//change this to firestore token not string
	var queryParams collegeParams
	err = json.Unmarshal(body, &queryParams)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//scores the student from 1-5
	score := scoreStudent()

	selectivityInfo, err := getCollegeRanges(score)
	if err != nil {
		fmt.Println("bad SelectivityInfo ", err)
		return
	}
	var safety []Result
	if selectivityInfo[0].Rate != 0 {
		log.Println("Safety")
		safety, err = queryColleges(&selectivityInfo[0], queryParams)
	}
	log.Println("Target")
	target, err := queryColleges(&selectivityInfo[1], queryParams)
	log.Println("Reach")
	reach, err := queryColleges(&selectivityInfo[2], queryParams)

	// safetyResults := sortColleges(safety)
	// targetResults := sortColleges(target)
	// reachResults := sortColleges(reach)

	results := SafetyTargetReach{
		Safety: safety,
		Target: target,
		Reach:  reach,
	}

	output, err := json.Marshal(results)
	if err != nil {
		log.Fatalln(err)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func sortColleges(colleges []Result) []Result {
	return nil
}

//SafetyTargetReach structure
type SafetyTargetReach struct {
	Safety []Result
	Target []Result
	Reach  []Result
}

//CollegeSelectivityInfo structure
type CollegeSelectivityInfo struct {
	ACT  []int `json:"act"`
	SAT  []int `json:"sat"`
	Rate int   `json:"rate"`
}

// ScoreCardResponse structure
type ScoreCardResponse struct {
	Metadata Metadata `json:"metadata"`
	Results  []Result `json:"results"`
}

// Result structure
type Result struct {
	SchoolName     string  `json:"school.name"`
	AvgACT         float32 `json:"latest.admissions.act_scores.midpoint.cumulative"`
	AvgSat         float32 `json:"latest.admissions.sat_scores.average.overall"`
	AdmissionsRate float32 `json:"latest.admissions.admission_rate.overall"`
}

// ScoreCardResponse structure
type majorResponse struct {
	Metadata Metadata      `json:"metadata"`
	Results  []majorResult `json:"results"`
}

type majorResult struct {
	Latest map[string]float32 `json:"latest.academics.program_percentage"`
}

//Metadata structure
type Metadata struct {
	Total   int `json:"total"`
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Location = city/large, suburbs/midsize
type collegeParams struct {
	ZIP      string `json:"ZIP"`
	State    string
	Region   string
	Majors   []string
	Location int
}
