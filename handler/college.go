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

func (h *Handler) scoreStudent() {
	ctx := context.Background()
	userInfo, err := client.Collection("users").Doc(UUID).Get(ctx)
	if err != nil {
		// http.Error(err.Error(), 404)
	}
	// selectivityInfo, err := client.Collection("info").Doc("Selectivity").Get(ctx)
	// if err != nil {
	// 	//handle
	// }
	var student student
	userInfo.DataTo(&student)
	if student.SAT > student.ACT {
		//use Sat and GPA
	} else {
		// use ACT and GPA
	}
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

	// //change this to firestore token not string
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

func (h *Handler) queryColleges(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit")
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
	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		fmt.Println("Malformed URL: ", err.Error())
		return
	}

	log.Println(string(queryParams.DegreeType))
	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	params.Add("school.region_id", queryParams.Region)
	params.Add("school.degrees_awarded.highest__range", string(queryParams.DegreeType)+"..")
	params.Add("fields", "school.name,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall")
	params.Add("per_page", "100")

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters
	log.Printf("Encoded URL is %q\n", baseURL.String())

	response, err := http.Get(baseURL.String())
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	resBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//change this to firestore token not string
	var colleges ScoreCardResponse
	err = json.Unmarshal(resBody, &colleges)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
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
			http.Error(w, err.Error(), 500)
		}

		resBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var tempColleges ScoreCardResponse
		err = json.Unmarshal(resBody, &tempColleges)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		colleges.Results = append(colleges.Results, tempColleges.Results...)
	}

	//Use colleges to narrow down list based on academics/cost
	for _, college := range colleges.Results {
		if college.AdmissionsRate < 30 {

		}
	}

	output, err := json.Marshal(colleges.Results)
	if err != nil {
		log.Fatalln(err)
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
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
	ZIP        string `json:"ZIP"`
	State      string
	Region     string
	ACT        int
	SAT        int
	DegreeType string
	GPA        int32 `json:"GPA"`
	Majors     []string
	Location   int
}
