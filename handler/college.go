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
	"strings"
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
var needMap map[string]int
var statesMap map[string]int

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

func getCollegeRanges(score int) ([][]CollegeSelectivityInfo, error) {
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
	var target []CollegeSelectivityInfo
	var reach []CollegeSelectivityInfo
	var safety []CollegeSelectivityInfo
	if score != 2 {
		var temp info
		safetyScore := strconv.Itoa(score - 1)
		log.Println("SafetyScore: ", safetyScore)
		safetyInfo, err := client.Collection("Selectivity").Doc(safetyScore).Get(ctx)
		if err != nil {
			return nil, err
		}
		safetyInfo.DataTo(&temp)
		for _, info := range temp.Info {
			str := strings.Split(info, ",")
			lACT, err := strconv.Atoi(str[0])
			hACT, err := strconv.Atoi(str[1])
			lSAT, err := strconv.Atoi(str[2])
			hSAT, err := strconv.Atoi(str[3])
			lRate, err := strconv.ParseFloat(str[4], 64)
			hRate, err := strconv.ParseFloat(str[5], 64)
			if err != nil {

			}
			safetySelectivity := CollegeSelectivityInfo{
				Score:    score - 1,
				lowACT:   lACT,
				highACT:  hACT,
				lowSAT:   lSAT,
				highSAT:  hSAT,
				lowRate:  lRate,
				highRate: hRate,
			}
			safety = append(safety, safetySelectivity)
		}
	}

	var tempTarget info
	targetInfo.DataTo(&tempTarget)
	for _, info := range tempTarget.Info {
		str := strings.Split(info, ",")
		lACT, err := strconv.Atoi(str[0])
		hACT, err := strconv.Atoi(str[1])
		lSAT, err := strconv.Atoi(str[2])
		hSAT, err := strconv.Atoi(str[3])
		lRate, err := strconv.ParseFloat(str[4], 32)
		hRate, err := strconv.ParseFloat(str[5], 32)
		if err != nil {

		}
		targetSelectivity := CollegeSelectivityInfo{
			Score:    score - 1,
			lowACT:   lACT,
			highACT:  hACT,
			lowSAT:   lSAT,
			highSAT:  hSAT,
			lowRate:  lRate,
			highRate: hRate,
		}
		target = append(target, targetSelectivity)
	}

	var tempReach info
	reachInfo.DataTo(&tempReach)
	for _, info := range tempReach.Info {
		str := strings.Split(info, ",")
		lACT, err := strconv.Atoi(str[0])
		hACT, err := strconv.Atoi(str[1])
		lSAT, err := strconv.Atoi(str[2])
		hSAT, err := strconv.Atoi(str[3])
		lRate, err := strconv.ParseFloat(str[4], 32)
		hRate, err := strconv.ParseFloat(str[5], 32)
		if err != nil {

		}
		reachSelectivity := CollegeSelectivityInfo{
			Score:    score - 1,
			lowACT:   lACT,
			highACT:  hACT,
			lowSAT:   lSAT,
			highSAT:  hSAT,
			lowRate:  lRate,
			highRate: hRate,
		}
		reach = append(reach, reachSelectivity)
	}

	info := [][]CollegeSelectivityInfo{safety, target, reach}

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
	lowAct := strconv.Itoa(selectivityInfo.lowACT)
	highAct := strconv.Itoa(selectivityInfo.highACT)
	lowSat := strconv.Itoa(selectivityInfo.lowSAT)
	highSat := strconv.Itoa(selectivityInfo.highSAT)
	lowRate := strconv.FormatFloat(selectivityInfo.lowRate, 'f', 1, 64)
	highRate := strconv.FormatFloat(selectivityInfo.highRate, 'f', 1, 64)
	log.Println("XXX", lowRate, highRate)

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	params.Add("school.region_id", queryParams.Region)
	params.Add("school.degrees_awarded.highest__range", "3..")
	params.Add("fields", "school.name,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall,latest.student.size,school.locale,school.ownership,school.state_fips"+majorString)
	params.Add("per_page", "100")
	params.Add("latest.admissions.act_scores.midpoint.cumulative__range", lowAct+".."+highAct)
	params.Add("latest.admissions.admission_rate.overall__range", lowRate+".."+highRate)
	params.Add("latest.admissions.sat_scores.average.overall__range", lowSat+".."+highSat)
	// for _, v := range queryParams.Majors {
	// 	params.Add("latest.academics.program_percentage."+v+"__range", "0.00000001..")
	// }
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
		majors := parseMajors(queryParams, c)
		temp := college{
			c.SchoolName,
			c.AvgACT,
			c.AvgSAT,
			c.AdmissionsRate,
			c.Size,
			c.Location,
			c.State,
			c.Ownership,
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

func parseMajors(queryParams collegeParams, college Result) map[string]float32 {
	var temp map[string]float32
	temp = make(map[string]float32)
	switch queryParams.Majors[0] {
	case "ariculture":
		temp["agricuture"] = college.Agriculture
	case "resources":
		temp["resources"] = college.Resources
	case "architecture":
		temp["architecture"] = college.Architecture
	case "ethnicCulturalGender":
		temp["ethnicCulturalGender"] = college.EthnicCulturalGender
	case "communication":
		temp["communication"] = college.Communication
	case "communicationsTechnology":
		temp["communicationsTechnology"] = college.CommunicationsTechnology
	case "computer":
		temp["computer"] = college.Computer
	case "personalCulinary":
		temp["personalCulinary"] = college.PersonalCulinary
	case "education":
		temp["education"] = college.Education
	case "engineering":
		temp["engineering"] = college.Engineering
	case "engineeringTechnology":
		temp["engineeringTechnology"] = college.EngineeringTechnology
	case "language":
		temp["language"] = college.Language
	case "familyConsumerScience":
		temp["familyConsumerScience"] = college.FamilyConsumerScience
	case "legal":
		temp["legal"] = college.Legal
	case "english":
		temp["english"] = college.English
	case "humanities":
		temp["humanities"] = college.Humanities
	case "library":
		temp["library"] = college.Library
	case "biological":
		temp["biological"] = college.Biological
	case "mathematics":
		temp["mathematics"] = college.Mathematics
	case "military":
		temp["military"] = college.Military
	case "multidiscipline":
		temp["multidiscipline"] = college.Multidiscipline
	case "parksRecreationFitness":
		temp["parksRecreationFitness"] = college.ParksRecreationFitness
	case "philosophyReligious":
		temp["philosophyReligious"] = college.PhilosophyReligious
	case "theologyReligiousVocation":
		temp["theologyReligiousVocation"] = college.TheologyReligiousVocation
	case "physicalScience":
		temp["physicalScience"] = college.PhysicalScience
	case "scienceTechnology":
		temp["scienceTechnology"] = college.ScienceTechnology
	case "psychology":
		temp["psychology"] = college.Psychology
	case "securityLawEnforcement":
		temp["securityLawEnforcement"] = college.SecurityLawEnforcement
	case "publicAdministrationSocialService":
		temp["publicAdministrationSocialService"] = college.PublicAdministrationSocialService
	case "socialScience":
		temp["socialScience"] = college.SocialScience
	case "construction":
		temp["construction"] = college.Construction
	case "mechanicRepairTechnology":
		temp["mechanicRepairTechnology"] = college.MechanicRepairTechnology
	case "precisionProduction":
		temp["precisionProduction"] = college.PrecisionProduction
	case "transportation":
		temp["transportation"] = college.Transportation
	case "visualPerforming":
		temp["visualPerforming"] = college.VisualPerforming
	case "health":
		temp["health"] = college.Health
	case "businessMarketing":
		temp["businessMarketing"] = college.BusinessMarketing
	case "history":
		temp["history"] = college.History
	}

	if len(queryParams.Majors) == 2 {
		switch queryParams.Majors[1] {
		case "ariculture":
			temp["agricuture"] = college.Agriculture
		case "resources":
			temp["resources"] = college.Resources
		case "architecture":
			temp["architecture"] = college.Architecture
		case "ethnicCulturalGender":
			temp["ethnicCulturalGender"] = college.EthnicCulturalGender
		case "communication":
			temp["communication"] = college.Communication
		case "communicationsTechnology":
			temp["communicationsTechnology"] = college.CommunicationsTechnology
		case "computer":
			temp["computer"] = college.Computer
		case "personalCulinary":
			temp["personalCulinary"] = college.PersonalCulinary
		case "education":
			temp["education"] = college.Education
		case "engineering":
			temp["engineering"] = college.Engineering
		case "engineeringTechnology":
			temp["engineeringTechnology"] = college.EngineeringTechnology
		case "language":
			temp["language"] = college.Language
		case "familyConsumerScience":
			temp["familyConsumerScience"] = college.FamilyConsumerScience
		case "legal":
			temp["legal"] = college.Legal
		case "english":
			temp["english"] = college.English
		case "humanities":
			temp["humanities"] = college.Humanities
		case "library":
			temp["library"] = college.Library
		case "biological":
			temp["biological"] = college.Biological
		case "mathematics":
			temp["mathematics"] = college.Mathematics
		case "military":
			temp["military"] = college.Military
		case "multidiscipline":
			temp["multidiscipline"] = college.Multidiscipline
		case "parksRecreationFitness":
			temp["parksRecreationFitness"] = college.ParksRecreationFitness
		case "philosophyReligious":
			temp["philosophyReligious"] = college.PhilosophyReligious
		case "theologyReligiousVocation":
			temp["theologyReligiousVocation"] = college.TheologyReligiousVocation
		case "physicalScience":
			temp["physicalScience"] = college.PhysicalScience
		case "scienceTechnology":
			temp["scienceTechnology"] = college.ScienceTechnology
		case "psychology":
			temp["psychology"] = college.Psychology
		case "securityLawEnforcement":
			temp["securityLawEnforcement"] = college.SecurityLawEnforcement
		case "publicAdministrationSocialService":
			temp["publicAdministrationSocialService"] = college.PublicAdministrationSocialService
		case "socialScience":
			temp["socialScience"] = college.SocialScience
		case "construction":
			temp["construction"] = college.Construction
		case "mechanicRepairTechnology":
			temp["mechanicRepairTechnology"] = college.MechanicRepairTechnology
		case "precisionProduction":
			temp["precisionProduction"] = college.PrecisionProduction
		case "transportation":
			temp["transportation"] = college.Transportation
		case "visualPerforming":
			temp["visualPerforming"] = college.VisualPerforming
		case "health":
			temp["health"] = college.Health
		case "businessMarketing":
			temp["businessMarketing"] = college.BusinessMarketing
		case "history":
			temp["history"] = college.History
		}
	}
	return temp
}

func (h *Handler) getMatches(w http.ResponseWriter, r *http.Request) {
	// ctx := context.Background()
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
	for _, v := range selectivityInfo[0] {
		temp, err := queryColleges(&v, queryParams, nil)
		if err != nil {

		}
		safety = append(safety, temp...)
	}
	var target []college
	for _, v := range selectivityInfo[1] {
		temp, err := queryColleges(&v, queryParams, nil)
		if err != nil {

		}
		target = append(target, temp...)
	}
	var reach []college
	for _, v := range selectivityInfo[2] {
		temp, err := queryColleges(&v, queryParams, nil)
		if err != nil {

		}
		reach = append(reach, temp...)
	}

	// c1 := make(chan []college)
	// c2 := make(chan []college)
	// log.Println("Target")
	// wg.Add(1)
	// go queryColleges(&selectivityInfo[1][0], queryParams, c1)
	// log.Println("Reach")
	// wg.Add(1)
	// go queryColleges(&selectivityInfo[2][0], queryParams, c2)
	// target := <-c1
	// reach := <-c2
	// println(target)

	// println("Waiting")
	// wg.Wait()
	// println("Done")
	safetyResults := sortColleges(safety, queryParams)
	targetResults := sortColleges(target, queryParams)
	reachResults := sortColleges(reach, queryParams)

	results := SafetyTargetReach{
		Safety: safetyResults,
		Target: targetResults,
		Reach:  reachResults,
	}

	output, err := json.Marshal(results)
	if err != nil {
		log.Fatalln(err)
	}

	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func sortColleges(colleges []college, queryParams collegeParams) []college {
	//name to college
	// used to look up college based on name from ranking
	var collegeDict map[string]college
	collegeDict = make(map[string]college)

	//name to sorted rank
	var rankColleges map[string]int
	rankColleges = make(map[string]int)

	//Has Major
	var majorColleges map[string]int
	majorColleges = make(map[string]int)

	if len(needMap) == 0 {
		needMap = getSchoolNeedMet()
	}

	if len(statesMap) == 0 {
		statesMap = getStateCodes()
	}

	//major and affordability
	//requires majors and only shows schools based on affordability algorithm
	for _, c := range colleges {

		//Checks if the school has wanted majors
		switch len(queryParams.Majors) {
		case 1:
			if c.Majors[queryParams.Majors[0]] != 0 {
				majorColleges[c.SchoolName] = majorColleges[c.SchoolName] + 1
				collegeDict[c.SchoolName] = c
			}
		case 2:
			if c.Majors[queryParams.Majors[0]] != 0 && c.Majors[queryParams.Majors[1]] != 0 {
				majorColleges[c.SchoolName] = majorColleges[c.SchoolName] + 1
				collegeDict[c.SchoolName] = c
			}
		}
		_, exists := majorColleges[c.SchoolName]

		//Checks if the school exists in the list of schools that has the wanted majors then sorts
		if exists {

			//Affordability sort
			switch c.Ownership {
			//Public
			case 1:
				//if in-state
				if c.State == statesMap[user.State] {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				} else {
					//if out-of-state
					if user.AbilityToPay < 25000 {
						if strings.Contains(c.SchoolName, "University of North Carolina at Chapel Hill") || strings.Contains(c.SchoolName, "University of Michigan-Ann Arbor") || strings.Contains(c.SchoolName, "University of Virginia-Main Campus") {
							rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
						}
						//add regional colleges here
					} else {
						rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
					}
				}
			//Private
			default:
				if user.AbilityToPay <= 6000 {
					if needMap[c.SchoolName] >= 90 {
						rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
					}
				} else if user.AbilityToPay >= 6000 && user.AbilityToPay <= 10000 {
					if needMap[c.SchoolName] >= 87 {
						rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
					}
				} else if user.AbilityToPay >= 10000 && user.AbilityToPay <= 15000 {
					if needMap[c.SchoolName] >= 85 {
						rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
					}
				} else {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}

			}

			//Size Preference
			switch strings.ToLower(user.Size) {
			case "small":
				if c.Size < 2000 {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			case "medium":
				if c.Size > 2000 && c.Size < 10000 {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			case "large":
				if c.Size > 10000 && c.Size < 15000 {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			case "xlarge":
				if c.Size > 15000 {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			}
		}
	}

	for i, v := range rankColleges {
		log.Println(i, v)
	}

	//Location: City/large

	//in/out of state

	//use sortedColleges to look up list of actual colleges
	type kv struct {
		Key   string
		Value int
	}

	var sortedColleges []kv
	for k, v := range rankColleges {
		sortedColleges = append(sortedColleges, kv{k, v})
	}

	sort.Slice(sortedColleges, func(i, j int) bool {
		return sortedColleges[i].Value > sortedColleges[j].Value
	})

	var finalSort []college
	finalSort = make([]college, len(rankColleges))
	for i, kv := range sortedColleges {
		finalSort[i] = collegeDict[kv.Key]
	}

	return finalSort
}

func getSchoolNeedMet() map[string]int {
	file, err := os.Open("handler/school.csv")
	if err != nil {

	}
	m := make(map[string]int)
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
		m[schoolName] = needMet
	}
	return m
}

func getStateCodes() map[string]int {
	file, err := os.Open("handler/stateCodes.csv")
	if err != nil {

	}
	m := make(map[string]int)
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
		state := strings.TrimSpace(record[1])
		code, err := strconv.Atoi(record[0])
		m[state] = code
	}
	return m
}

type need struct {
	NeedMet int `json:"NeedMet"`
}

type rank struct {
	School college
	rank   int
}

//SafetyTargetReach structure
type SafetyTargetReach struct {
	Safety []college
	Target []college
	Reach  []college
}

// //CollegeSelectivityInfo structure
// type CollegeSelectivityInfo struct {
// 	ACT  []int `json:"act"`
// 	SAT  []int `json:"sat"`
// 	Rate int   `json:"rate"`
// }

//CollegeSelectivityInfo structure
type CollegeSelectivityInfo struct {
	Score    int
	lowACT   int
	highACT  int
	lowSAT   int
	highSAT  int
	lowRate  float64
	highRate float64
}

type info struct {
	Info []string `json:"info"`
}

// ScoreCardResponse structure
type ScoreCardResponse struct {
	Metadata Metadata `json:"metadata"`
	Results  []Result `json:"results"`
}

// Result structure
type Result struct {
	SchoolName                        string  `json:"school.name"`
	AvgACT                            float32 `json:"latest.admissions.act_scores.midpoint.cumulative"`
	AvgSAT                            float32 `json:"latest.admissions.sat_scores.average.overall"`
	AdmissionsRate                    float32 `json:"latest.admissions.admission_rate.overall"`
	Size                              int     `json:"latest.student.size"`
	Location                          int     `json:"school.locale"`
	State                             int     `json:"school.state_fips"`
	Ownership                         int     `json:"school.ownership"`
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
	AvgSAT         float32
	AdmissionsRate float32
	Size           int
	Location       int
	State          int
	Ownership      int
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
