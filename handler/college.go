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

	firestore "cloud.google.com/go/firestore"
	"github.com/mitchellh/mapstructure"
)

var wg sync.WaitGroup
var needMap map[string]int
var statesMap map[string]int
var numToState map[int]string
var regionMap map[string]int
var majorsMap map[string][]string
var majorCategories map[string][]string

func (h *Handler) testOtherFunc(w http.ResponseWriter, r *http.Request) {

}

//Automate this when CollegeScoreCard updates to allow for querying program_percentage
//No need for security checks, doesnt access firestore
func (h *Handler) collegeMajors(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("handler/MajorCategories.csv")
	if err != nil {
		log.Println(err.Error())
	}
	csvfile := csv.NewReader(file)
	var majors map[string][]string
	majors = make(map[string][]string)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Println(err.Error())
		}
		if _, ok := majors[record[0]]; ok {
			majors[record[0]] = append(majors[record[0]], record[1])
		} else {
			majors[record[0]] = []string{record[1]}
		}
	}
	i := 0
	keys := make([]string, len(majors))
	for k := range majors {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	output, err := json.Marshal(keys)
	if err != nil {
		log.Println(err.Error())
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

//Takes in query params based on student perferences, delegates tasks to query and sort STR schools
func (h *Handler) getMatches(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idToken := r.Header.Get("Authorization")
	token, err := Verify(idToken)
	if err != nil {
		log.Println("error verifying ID token: ", err.Error())
		http.Error(w, err.Error(), 401)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	var queryParams collegeParams
	err = json.Unmarshal(body, &queryParams)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	//scores the student from 1-5
	score, test, err := scoreStudent(token.UID)

	selectivityInfo, err := getCollegeRanges(score)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}

	var ccTarget []college
	var ccSafety []college

	//set up schools with wanted major
	schoolsWithMajor, err := setUpMajors(queryParams.Majors)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	//TODO: switch to go routines
	//This section is for Community College searchs based on grades and test scores
	if score == 1 {
		//no selectivity info since its a CC, given same queryParams as other searches,
		// given a Channel to post the response back to when go routine is complete,
		// gives that is a CC search and it wants all results
		temp, err := queryColleges(nil, queryParams, nil, true, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ccTarget = append(ccTarget, temp...)
	} else if score == 2 {
		//Same idea as above but puts the results into safety rather than target schools
		temp, err := queryColleges(nil, queryParams, nil, true, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ccSafety = append(ccSafety, temp...)
	} else if queryParams.UnweightedGPA <= 2.5 || (strings.ToLower(test) == "act" && queryParams.ACT <= 17) || (strings.ToLower(test) == "sat" && queryParams.SAT <= 880) {
		//This only gets colleges within 25 miles from the persons zipcode
		// getAllResults bool = false and within queryColleges it uses that bool
		// to add a parameter to collegescorecard request stating only schools within 25 miles
		temp, err := queryColleges(nil, queryParams, nil, true, false)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ccSafety = append(ccSafety, temp...)
	} else if queryParams.UnweightedGPA <= 3.25 || (strings.ToLower(test) == "act" && queryParams.ACT <= 18) || (strings.ToLower(test) == "sat" && queryParams.SAT <= 950) {
		//Same as above
		temp, err := queryColleges(nil, queryParams, nil, true, false)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		ccSafety = append(ccSafety, temp...)
	}

	//takes each of the STR info ranges and querys college scorecard API
	var safety []college
	for _, v := range selectivityInfo[0] {
		temp, err := queryColleges(&v, queryParams, nil, false, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		safety = append(safety, temp...)
	}
	var target []college
	for _, v := range selectivityInfo[1] {
		temp, err := queryColleges(&v, queryParams, nil, false, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		target = append(target, temp...)
	}
	var reach []college
	for _, v := range selectivityInfo[2] {
		temp, err := queryColleges(&v, queryParams, nil, false, true)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		reach = append(reach, temp...)
	}

	if score <= 2 {
		safety = append(safety, ccSafety...)
		target = append(target, ccTarget...)
	} else {
		safety = append(safety, ccSafety...)
	}

	cSafe := make(chan chanResult)
	cTarget := make(chan chanResult)
	cReach := make(chan chanResult)

	//sorts each of the resulting queries based on student preferences
	wg.Add(1)
	go sortColleges(safety, queryParams, "safety", schoolsWithMajor, cSafe)
	wg.Add(1)
	go sortColleges(target, queryParams, "target", schoolsWithMajor, cTarget)
	wg.Add(1)
	go sortColleges(reach, queryParams, "reach", schoolsWithMajor, cReach)

	safetyResults := <-cSafe
	targetResults := <-cTarget
	reachResults := <-cReach

	wg.Wait()

	results := SafetyTargetReach{
		Safety: safetyResults.results,
		Target: targetResults.results,
		Reach:  reachResults.results,
	}
	resultIDs := SafetyTargetReachIDs{
		Safety: safetyResults.ids,
		Target: targetResults.ids,
		Reach:  reachResults.ids,
	}

	_, err = client.Collection("userMatches").Doc(token.UID).Set(ctx, map[string]interface{}{
		"results": resultIDs,
		"majors":  queryParams.Majors,
	}, firestore.MergeAll)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 404)
		return
	}

	output, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

//Identifies the categories of student score and the scores above and below
//Gets the highest and lowest possible score for each range to cut down on colleges returned from Scorecard
func getCollegeRanges(score int) ([][]CollegeSelectivityInfo, error) {
	ctx := context.Background()
	reachScore := strconv.Itoa(score + 1)
	reachInfo, err := client.Collection("Selectivity").Doc(reachScore).Get(ctx)
	if err != nil {
		return nil, err
	}
	var target []CollegeSelectivityInfo
	var reach []CollegeSelectivityInfo
	var safety []CollegeSelectivityInfo
	var targetInfo *firestore.DocumentSnapshot
	if score > 1 {
		targetScore := strconv.Itoa(score)
		targetInfo, err = client.Collection("Selectivity").Doc(targetScore).Get(ctx)
		if err != nil {
			return nil, err
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
				return nil, err
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
	}
	if score > 2 {
		var temp info
		safetyScore := strconv.Itoa(score - 1)
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
				return nil, err
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
			return nil, err
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

func getRegionParams(region []string, state string) (string, string) {
	switch strings.Trim(strings.ToLower(region[0]), " ") {
	case "home":
		return "distance", "25mi"
	case "state":
		return "school.state_fips", strconv.Itoa(statesMap[state])
	case "national":
		return "", ""
	default:
		var regionString = ""
		for _, v := range region {
			if regionString == "" {
				regionString = v
			} else {
				regionString = regionString + "," + v
			}
		}
		return "school.region_id", regionString
	}
}

//Querys college scorecard API for each STR
func queryColleges(selectivityInfo *CollegeSelectivityInfo, queryParams collegeParams, c chan []college, ccSearch bool, getAllResults bool) ([]college, error) {
	var err error
	if len(statesMap) == 0 {
		statesMap, numToState, err = getStateCodes()
		if err != nil {
			return nil, err
		}
	}
	if len(regionMap) == 0 {
		regionMap, err = getStatesByRegion()
		if err != nil {
			return nil, err
		}
	}

	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		return nil, err
	}

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	if ccSearch {
		params.Add("school.carnegie_basic__range", "..14")
		params.Add("school.state_fips", strconv.Itoa(statesMap[queryParams.State]))
		if !getAllResults {
			params.Add("zip", queryParams.Zip)
			params.Add("distance", "25mi")
		}
	} else {
		key, value := getRegionParams(queryParams.Region, queryParams.State)
		if len(key) != 0 {
			if key == "distance" {
				params.Add("zip", queryParams.Zip)
			}
			params.Add(key, value)
		}
		//sets ranges of possible scores to limit query, Takes lowest and highest of each data point from STR
		lowAct := strconv.Itoa(selectivityInfo.lowACT)
		highAct := strconv.Itoa(selectivityInfo.highACT)
		lowSat := strconv.Itoa(selectivityInfo.lowSAT)
		highSat := strconv.Itoa(selectivityInfo.highSAT)
		lowRate := strconv.FormatFloat(selectivityInfo.lowRate, 'f', 1, 64)
		highRate := strconv.FormatFloat(selectivityInfo.highRate, 'f', 1, 64)
		params.Add("school.carnegie_basic__range", "14..")
		params.Add("latest.admissions.act_scores.midpoint.cumulative__range", lowAct+".."+highAct)
		params.Add("latest.admissions.admission_rate.overall__range", lowRate+".."+highRate)
		params.Add("latest.admissions.sat_scores.average.overall__range", lowSat+".."+highSat)
	}
	params.Add("fields", "id,school.name,school.carnegie_basic,latest.student.demographics.race_ethnicity.white,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall,latest.student.size,school.locale,school.ownership,school.state_fips")
	//Limited to 100 per page max
	params.Add("per_page", "100")

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
	//converts results into type: College for rest of algorithm
	var colleges []college
	for _, c := range scorecardColleges.Results {
		temp := college{
			c.ID,
			c.SchoolName,
			c.CIPCode,
			c.AvgACT,
			c.AvgSAT,
			c.AdmissionsRate,
			c.Size,
			c.Location,
			c.Diversity,
			numToState[c.State],
			c.Ownership,
			queryParams.Majors,
		}
		colleges = append(colleges, temp)
	}
	if c != nil {
		c <- colleges
		wg.Done()
	}
	return colleges, nil
}

func setUpMajors(majors []string) (map[string]bool, error) {
	codes, err := GetMajorParams(majors)
	if err != nil {
		return nil, err
	}
	var schoolsWithMajor map[string]bool
	schoolsWithMajor = make(map[string]bool)
	if len(majors) == 1 {
		schoolsWithMajor, err = listCollegesWithMajors(codes[majors[0]])
		if err != nil {
			return nil, err
		}
	} else if len(majors) == 2 {
		tempA, err := listCollegesWithMajors(codes[majors[0]])
		if err != nil {
			return nil, err
		}
		tempB, err := listCollegesWithMajors(codes[majors[1]])
		if err != nil {
			return nil, err
		}
		for school := range tempA {
			if tempB[school] {
				schoolsWithMajor[school] = true
			}
		}
	}
	return schoolsWithMajor, nil
}

//GetMajorParams not using currently since collegescorecard has bugs but hopefully can use in future
func GetMajorParams(majors []string) (map[string][]string, error) {
	var err error
	if len(majorsMap) == 0 {
		majorsMap, err = getMajorsByCipCode()
		if err != nil {
			return nil, err
		}
	}
	if len(majorCategories) == 0 {
		majorCategories, err = getMajorCategories()
		if err != nil {
			return nil, err
		}
	}
	var codes map[string][]string
	codes = make(map[string][]string)
	for _, m := range majorCategories[majors[0]] {
		codes[m] = append(codes[m], majorsMap[m]...)
	}
	if len(majors) == 2 {
		for _, m := range majorCategories[majors[1]] {
			codes[m] = append(codes[m], majorsMap[m]...)
		}
	}
	return codes, nil
}

func listCollegesWithMajors(codes []string) (map[string]bool, error) {
	ctx := context.Background()
	var schools map[string]bool
	schools = make(map[string]bool)
	for _, c := range codes {
		schoolsData, err := client.Collection("majors").Doc(c).Get(ctx)
		if err != nil {
			return nil, err
		}
		var MajorSchools majorSchools
		schoolsData.DataTo(&MajorSchools)
		for _, school := range MajorSchools.Schools {
			schools[school] = true
		}
	}
	return schools, nil
}

type majorSchools struct {
	Schools []string
}

func checkAffordability(c college, AbilityToPay int, state string) bool {
	switch c.Ownership {
	//Public
	case 1:
		//if in-state
		if c.State == state {
			return true
		}
		//if out-of-state
		if AbilityToPay < 25000 {
			if strings.Contains(c.SchoolName, "University of North Carolina at Chapel Hill") || strings.Contains(c.SchoolName, "University of Michigan-Ann Arbor") || strings.Contains(c.SchoolName, "University of Virginia-Main Campus") {
				return true
			}
		} else {
			return true
		}
	//Private
	case 2:
		if AbilityToPay <= 6000 {
			if needMap[c.SchoolName] >= 90 {
				return true
			}
		} else if AbilityToPay >= 6000 && AbilityToPay <= 10000 {
			if needMap[c.SchoolName] >= 87 {
				return true
			}
		} else if AbilityToPay >= 10000 && AbilityToPay <= 15000 {
			if needMap[c.SchoolName] >= 85 {
				return true
			}
		} else {
			return true
		}
	}
	return false
}

func sizePreference(sizes []string, collegeSize int) int {
	for index, v := range sizes {
		sizes[index] = strings.Trim(strings.ToLower(v), " ")
	}
	if collegeSize < 2000 && contains(sizes, "small") {
		return 1
	} else if collegeSize > 2000 && collegeSize < 10000 && contains(sizes, "medium") {
		return 1
	} else if collegeSize > 10000 && collegeSize < 15000 && contains(sizes, "large") {
		return 1
	} else if collegeSize > 15000 && contains(sizes, "xlarge") {
		return 1
	}
	return 0
}

func locationPreference(locations []int, loc int) int {
	if loc == 1 && (containsInt(locations, 11) || containsInt(locations, 12) || containsInt(locations, 13)) {
		return 1
	}
	if loc == 2 && (containsInt(locations, 21) || containsInt(locations, 22) || containsInt(locations, 23)) {
		return 1
	}
	if loc == 3 && (containsInt(locations, 31) || containsInt(locations, 32) || containsInt(locations, 33) || containsInt(locations, 41) || containsInt(locations, 42) || containsInt(locations, 43)) {
		return 1
	}
	return 0
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func containsInt(arr []int, val int) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

//Sorts the three categories STR into a ranked list based on preferences
func sortColleges(colleges []college, queryParams collegeParams, rank string, schoolsWithMajor map[string]bool, c chan chanResult) ([]college, []int32, error) {
	//maps "name" to all of the info on that specific college
	// used to look up college based on name from ranking
	var collegeDict map[string]college
	collegeDict = make(map[string]college)

	//maps "name" to sorted rank
	var rankColleges map[string]int
	rankColleges = make(map[string]int)
	var err error
	//in order to limit firestore queries and time
	//we save needMet of each private school (get this info from client) in a csv file
	//Will need to automate process for client to be able to upload a new file every year
	if len(needMap) == 0 {
		needMap, err = getSchoolNeedMet()
		if err != nil {
			return nil, nil, err
		}
	}

	//Maps states to specific code from ScoreCard API
	if len(statesMap) == 0 {
		statesMap, numToState, err = getStateCodes()
		if err != nil {
			return nil, nil, err
		}
	}

	//major and affordability
	//requires majors and only shows schools based on affordability algorithm
	for _, c := range colleges {

		var hasMajors = true
		var canAfford = true

		if c.CIPCode >= 14 {
			if strings.Contains(c.SchoolName, "/") {
				c.SchoolName = strings.ReplaceAll(c.SchoolName, "/", " ")
			}
			if schoolsWithMajor[c.SchoolName] == true {
				hasMajors = true
			} else {
				hasMajors = false
			}
			canAfford = checkAffordability(c, queryParams.AbilityToPay, queryParams.State)
		}

		//Checks if the school exists in the list of schools that has the wanted majors then sorts
		if hasMajors && canAfford {
			collegeDict[c.SchoolName] = c
			rankColleges[c.SchoolName] = 0
			//Size Preference
			if score := sizePreference(queryParams.Size, c.Size); score == 1 {
				rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
			}

			//Location Preference
			if score := locationPreference(queryParams.Location, c.Location); score == 1 {
				rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
			}

			//Diversity latest.student.demographics.race_ethnicity.white
			c.Diversity = 1 - c.Diversity
			switch {
			case c.Diversity >= 0.20:
				if queryParams.Diversity[0] == "some" {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			case c.Diversity > 0.30:
				if queryParams.Diversity[0] == "more" {
					rankColleges[c.SchoolName] = rankColleges[c.SchoolName] + 1
				}
			}
		}
	}

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

	var finalIDs []int32
	var finalSort []college
	finalSort = make([]college, len(rankColleges))
	for i, kv := range sortedColleges {
		finalIDs = append(finalIDs, collegeDict[kv.Key].ID)
		finalSort[i] = collegeDict[kv.Key]
	}
	if c != nil {
		var temp chanResult
		temp = chanResult{
			results: finalSort,
			ids:     finalIDs,
		}
		c <- temp
		wg.Done()
	}
	return finalSort, finalIDs, nil
}

func getMajorCategories() (map[string][]string, error) {
	file, err := os.Open("handler/MajorCategories.csv")
	if err != nil {
		return nil, err
	}
	csvfile := csv.NewReader(file)
	var majors map[string][]string
	majors = make(map[string][]string)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, ok := majors[record[0]]; ok {
			majors[record[0]] = append(majors[record[0]], record[1])
		} else {
			majors[record[0]] = []string{record[1]}
		}
	}
	return majors, nil
}

func getMajorsByCipCode() (map[string][]string, error) {
	file, err := os.Open("handler/majors.csv")
	if err != nil {
		return nil, err
	}
	csvfile := csv.NewReader(file)
	var majors map[string][]string
	majors = make(map[string][]string)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if _, ok := majors[record[0]]; ok {
			majors[record[0]] = append(majors[record[0]], record[1])
		} else {
			majors[record[0]] = []string{record[1]}
		}
	}
	return majors, nil
}

func getSchoolNeedMet() (map[string]int, error) {
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
			return nil, err
		}
		schoolName := strings.TrimSpace(record[0])
		needMet, err := strconv.Atoi(record[1])
		m[schoolName] = needMet
	}
	return m, nil
}

func getStatesByRegion() (map[string]int, error) {
	file, err := os.Open("handler/region.csv")
	if err != nil {
		return nil, err
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
			return nil, err
		}
		state := strings.TrimSpace(record[1])
		code, err := strconv.Atoi(record[0])
		m[state] = code
	}
	return m, nil
}

func getStateCodes() (map[string]int, map[int]string, error) {
	file, err := os.Open("handler/stateCodes.csv")
	if err != nil {

	}
	m := make(map[string]int)
	n := make(map[int]string)
	csvfile := csv.NewReader(file)
	for {
		// Read each record from csv
		record, err := csvfile.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		state := strings.TrimSpace(record[1])
		code, err := strconv.Atoi(record[0])
		m[state] = code
		n[code] = state
	}
	return m, n, nil
}

func (h *Handler) getPastMatches(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	idToken := r.Header.Get("Authorization")
	token, err := Verify(idToken)
	if err != nil {
		log.Println("error verifying ID token: ", err.Error())
		http.Error(w, err.Error(), 401)
		return
	}
	docsnap, err := client.Collection("userMatches").Doc(token.UID).Get(ctx)
	if !docsnap.Exists() {
		temp := SafetyTargetReach{
			Safety: nil,
			Target: nil,
			Reach:  nil,
		}
		output, err := json.Marshal(temp)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("content-type", "application/json")
		w.Write(output)
		return
	}

	matchesData, err := docsnap.DataAt("results")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var matches SafetyTargetReachIDs
	mapstructure.Decode(matchesData, &matches)

	majorsData, err := docsnap.DataAt("majors")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	var majors []string
	mapstructure.Decode(majorsData, &majors)

	//TODO GO ROUTINES HERE
	c1 := make(chan []college)
	c2 := make(chan []college)
	c3 := make(chan []college)

	wg.Add(1)
	go queryCollegesByID(matches.Safety, majors, c1)

	wg.Add(1)
	go queryCollegesByID(matches.Target, majors, c2)

	wg.Add(1)
	go queryCollegesByID(matches.Reach, majors, c3)

	safety := <-c1
	target := <-c2
	reach := <-c3

	wg.Wait()

	Temp := SafetyTargetReach{
		Safety: safety,
		Target: target,
		Reach:  reach,
	}
	output, err := json.Marshal(Temp)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Write(output)
	return
}

func queryCollegesByID(ids []int32, majors []string, c chan []college) ([]college, error) {

	baseURL, err := url.Parse("https://api.data.gov/ed/collegescorecard/v1/schools?")
	if err != nil {
		return nil, err
	}
	stringIDs := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(ids)), ","), "[]")

	majorString := ""
	for _, major := range majors {
		majorString = majorString + ",latest.academics.program_percentage." + major
	}

	// Prepare Query Parameters
	params := url.Values{}
	params.Add("api_key", os.Getenv("SCORECARDAPIKEY"))
	params.Add("id", stringIDs)
	params.Add("fields", "id,school.name,school.carnegie_basic,latest.student.demographics.race_ethnicity.white,latest.admissions.act_scores.midpoint.cumulative,latest.admissions.sat_scores.average.overall,latest.admissions.admission_rate.overall,latest.student.size,school.locale,school.ownership,school.state_fips")
	//Limited to 100 per page max
	params.Add("per_page", "100")

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters
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

	//gets total amount of pages from metadata
	totalPages := math.Ceil(float64(scorecardColleges.Metadata.Total) / float64(scorecardColleges.Metadata.PerPage))
	//loops through remaining pages and takes in results and addes them to our array of colleges
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
	//converts results into type: College for rest of algorithm
	var colleges []college
	for _, c := range scorecardColleges.Results {
		temp := college{
			c.ID,
			c.SchoolName,
			c.CIPCode,
			c.AvgACT,
			c.AvgSAT,
			c.AdmissionsRate,
			c.Size,
			c.Location,
			c.Diversity,
			numToState[c.State],
			c.Ownership,
			majors,
		}
		colleges = append(colleges, temp)
	}
	if c != nil {
		c <- colleges
		wg.Done()
	}
	return colleges, nil
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

//SafetyTargetReachIDs structure
type SafetyTargetReachIDs struct {
	Safety []int32
	Target []int32
	Reach  []int32
}

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
	ID                                int32   `json:"id"`
	SchoolName                        string  `json:"school.name"`
	CIPCode                           int32   `json:"school.carnegie_basic"`
	AvgACT                            float32 `json:"latest.admissions.act_scores.midpoint.cumulative"`
	AvgSAT                            float32 `json:"latest.admissions.sat_scores.average.overall"`
	AdmissionsRate                    float32 `json:"latest.admissions.admission_rate.overall"`
	Size                              int     `json:"latest.student.size"`
	Location                          int     `json:"school.locale"`
	Diversity                         float32 `json:"latest.student.demographics.race_ethnicity.white"`
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
	ID             int32
	SchoolName     string
	CIPCode        int32
	AvgACT         float32
	AvgSAT         float32
	AdmissionsRate float32
	Size           int
	Location       int
	Diversity      float32
	State          string
	Ownership      int
	Majors         []string
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
	State         string
	UnweightedGPA float32
	ACT           int
	SAT           int
	Region        []string
	Majors        []string
	AbilityToPay  int
	Size          []string
	Location      []int
	Diversity     []string
	Zip           string
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

type chanResult struct {
	results []college
	ids     []int32
}
