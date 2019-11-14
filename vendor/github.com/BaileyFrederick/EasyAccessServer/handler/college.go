package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"sync"

	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	// "os"
	// firebase "firebase.google.com/go"
	// "google.golang.org/api/iterator"
	// firestore "cloud.google.com/go/firestore"
)

var wg sync.WaitGroup

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

func queryColleges(selectivityInfo *CollegeSelectivityInfo, queryParams collegeParams, c chan []college) ([]college, error) {

	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		return nil, err
	}
	majorString := ""
	for _, major := range queryParams.Majors {
		majorString = majorString + ",latest.academics.program_percentage." + major
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
	params.Add("fields", "school.name,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall"+majorString)
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

	var scorecardColleges ScoreCardResponse
	err = json.Unmarshal(resBody, &scorecardColleges)
	if err != nil {
		return nil, err
	}
	log.Println("total: ", scorecardColleges.Metadata.Total)
	log.Println("page: ", scorecardColleges.Metadata.Page)

	totalPages := math.Ceil(float64(scorecardColleges.Metadata.Total) / float64(scorecardColleges.Metadata.PerPage))
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
		scorecardColleges.Results = append(scorecardColleges.Results, tempColleges.Results...)
	}
	var colleges []college
	for _, c := range scorecardColleges.Results {
		majors := make(map[string]float32)
		majors["History"] = c.History
		temp := college{
			c.SchoolName,
			c.AvgACT,
			c.AvgSat,
			c.AdmissionsRate,
			majors,
		}
		colleges = append(colleges, temp)
	}
	println("Sending results to chan")
	if c != nil {
		c <- colleges
		wg.Done()
	}
	return colleges, nil
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
	var safety []college
	if selectivityInfo[0].Rate != 0 {
		log.Println("Safety")
		safety, err = queryColleges(&selectivityInfo[0], queryParams, nil)
	}

	//XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX WORKING HERE
	c1 := make(chan []college)
	c2 := make(chan []college)
	log.Println("Target")
	wg.Add(1)
	go queryColleges(&selectivityInfo[1], queryParams, c1)
	log.Println("Reach")
	wg.Add(1)
	go queryColleges(&selectivityInfo[2], queryParams, c2)
	target := <-c1
	reach := <-c2
	println(target)

	println("Waiting")
	wg.Wait()
	println("Done")
	_ = sortColleges(reach, queryParams)
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

func sortColleges(colleges []college, queryParams collegeParams) []Result {
	// var sortedColleges []college
	//major
	sort.SliceStable(colleges, func(i, j int) bool {
		return colleges[i].Majors[queryParams.Majors[0]] > colleges[j].Majors[queryParams.Majors[0]]
	})

	for _, c := range colleges {
		log.Println(c.Majors[queryParams.Majors[0]])
	}

	//Size

	//Location: City/large

	//in/out of state
	return nil
}

//SafetyTargetReach structure
type SafetyTargetReach struct {
	Safety []college
	Target []college
	Reach  []college
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
	SchoolName                        string  `json:"school.name"`
	AvgACT                            float32 `json:"latest.academics.act_scores.midpoint.cumulative"`
	AvgSat                            float32 `json:"latest.academics.sat_scores.average.overall"`
	AdmissionsRate                    float32 `json:"latest.academics.admission_rate.overall"`
	Agriculture                       float32 `json:"latest.academics.program_percentage.agriculture"`
	Resources                         float32 `json:"latest.academics.program_percentage.resources"`
	Architecture                      float32 `json:"latest.academics.program_percentage.architecture"`
	EthnicCulturalGender              float32 `json:"latest.academics.program_percentage.ethnic_cultural_gender"`
	Communication                     float32 `json:"latest.academics.program_percentage.communication"`
	CommunicationsTechnology          float32 `json:"latest.academics.program_percentage.communications_technology"`
	Computer                          float32 `json:"latest.academics.program_percentage.computer"`
	PersonalCulinary                  float32 `json:"latest.academics.program_percentage.personal_culinary"`
	Education                         float32 `json:"latest.academics.program_percentage.education"`
	Engineering                       float32 `json:"latest.academics.program_percentage.engineering"`
	EngineeringTechnology             float32 `json:"latest.academics.program_percentage.engineering_technology"`
	Language                          float32 `json:"latest.academics.program_percentage.language"`
	FamilyConsumerScience             float32 `json:"latest.academics.program_percentage.family_consumer_science"`
	Legal                             float32 `json:"latest.academics.program_percentage.legal"`
	English                           float32 `json:"latest.academics.program_percentage.english"`
	Humanities                        float32 `json:"latest.academics.program_percentage.humanities"`
	Library                           float32 `json:"latest.academics.program_percentage.library"`
	Biological                        float32 `json:"latest.academics.program_percentage.biological"`
	Mathematics                       float32 `json:"latest.academics.program_percentage.mathematics"`
	Military                          float32 `json:"latest.academics.program_percentage.military"`
	Multidiscipline                   float32 `json:"latest.academics.program_percentage.multidiscipline"`
	ParksRecreationFitness            float32 `json:"latest.academics.program_percentage.parks_recreation_fitness"`
	PhilosophyReligious               float32 `json:"latest.academics.program_percentage.philosophy_religious"`
	TheologyReligiousVocation         float32 `json:"latest.academics.program_percentage.theology_religious_vocation"`
	PhysicalScience                   float32 `json:"latest.academics.program_percentage.physical_science"`
	ScienceTechnology                 float32 `json:"latest.academics.program_percentage.science_technology"`
	Psychology                        float32 `json:"latest.academics.program_percentage.psychology"`
	SecurityLawEnforcement            float32 `json:"latest.academics.program_percentage.security_law_enforcement"`
	PublicAdministrationSocialService float32 `json:"latest.academics.program_percentage.public_administration_social_service"`
	SocialScience                     float32 `json:"latest.academics.program_percentage.social_science"`
	Construction                      float32 `json:"latest.academics.program_percentage.construction"`
	MechanicRepairTechnology          float32 `json:"latest.academics.program_percentage.mechanic_repair_technology"`
	PrecisionProduction               float32 `json:"latest.academics.program_percentage.precision_production"`
	Transportation                    float32 `json:"latest.academics.program_percentage.transportation"`
	VisualPerforming                  float32 `json:"latest.academics.program_percentage.visual_performing"`
	Health                            float32 `json:"latest.academics.program_percentage.health"`
	BusinessMarketing                 float32 `json:"latest.academics.program_percentage.business_marketing"`
	History                           float32 `json:"latest.academics.program_percentage.history"`
}

type college struct {
	SchoolName     string
	AvgACT         float32
	AvgSat         float32
	AdmissionsRate float32
	Majors         map[string]float32
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
	Region   string
	Majors   []string
	Location int
}

type selectivity struct {
	Score   string
	LowACT  int
	HighACT int
	LowGPA  float64
	HighGPA float64
	LowSAT  int
	HighSAT int
}
