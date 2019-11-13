# EasyAccessServer

Welcome to our EasyAccessServer!
Easy Access is a simple and easy to use, all-in-one app for college applicants and high school counselors to find the perfect college for each student and guide them through the process, thus promoting equity in higher education.  Personalized to each student’s criteria for their future and includes integrated education tips, matching software, connection to college applications and admissions staff, and data collection.

# Getting Started

Download and set up golang/GOPATH

Here is a good step by step example

https://www.callicoder.com/golang-installation-setup-gopath-workspace/ 
	
Now clone repo into go/src/github.com/YOURUSERNAME
	
EasyAccess uses dep to manage dependencies.

# To start go server locally

First you will need access to the Firebase EasyAccessServer Console.

Go to Project Settings -> Service Accounts and Click "Generate New Private Key"

This will save a key file to your local machine. MAKE SURE TO REMEMBER WHERE THIS IS SAVED AND WHAT IT'S NAME IT.

Inside of the top level of the project create a .env file

Add `export GOOGLE_APPLICATION_CREDENTIALS="PATH TO YOUR PRIVATE KEY"`, `PORT="3001"`, and `SCORECARDAPIKEY="Insert Secret Key"`

Now you can run the following to start up a local server

```bash
go run main.go
```

Once your terminal says "INFO[0001] listening on port 3001" you're good

# Testing

Within the handler package you will find file names ending with `_test.go`

Once in those files, at the very top there is a button to `run package tests | run file tests`

Package tests will run all of the unit tests within handler and file test will only run the unit test within that file

You can also run individual unit test by clicking `run test` right above each function.

Or you can run `go test -v` in the terminal to run all tests `-v` gives log output.

# Deployment

Heroku: No addons. Set up through git. Git push heroku master will publish master branch to production.
	
	Make sure dynos are upgraded. heroku ps: web=1 will start one dyno. 0=no dynos on, 

Firebase: Firestore live at all times. Make sure it is upgraded to one of the following plans when in production

	Flame Plan: Fixed rate

	Blaze Plan: Pay for what you use 
	
# Architecture

 See adr-back-end.md file

# Contributing

A developer will need access to...

	Github repos
	
		1. EasyAccessServer
		
		2. EasyAccessReactNative
		
	Heroku - easy-access-server
	
	Firebase - easyaccess-9ffaa
	
		firestore
		
		Authentication
		
		service account
		
	College Score Card API key - heroku -> settings -> reveal config vars - SCORECARDAPIKEY
	
# Authors
	Bailey Frederick
	Ashley Smith
	Zach Glontz
	
# Acknowledgements
	Rocky Moon
	Vitaly Radsky
	Dennis Brown
	

# Http Requests
```bash
/user - POST request.body(idToken from firebase.auth.signinwithEmailandPasword)
	you will get back a student struct
/colleges - GET no body
  	you will get back all colleges
/majors - GET no body
	get back an array of strings
/collegeMatches - POST request.body
		type collegeParams struct {
			ZIP      string `json:"ZIP"`
			State    string
			Region   string
			Majors   []string
			Location int
		}
  	you will get back
		type SafetyTargetReach struct {
			Safety []Result
			Target []Result
			Reach  []Result
		}
/updateUser - PATCH request
	Body = 
	{
		"uid":"xLwd4c1WjKaxG3Vf3GDVMXMTLFE3",
		"info": [{
			"Path":"Name",
			"Value":"Bailey"
		},
		{
			"Path":"SAT",
			"Value":"2400"
		},
		{
			"Path":"ACT",
			"Value":"35"
		}]
	}
		   
```

# structs
```bash
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
```

```bash
type college struct {
	AcceptanceRate float64 `json:"Acceptance Rate"`
	AverageGPA	float64 `json:"Average GPA"`
	AverageSAT int64 `json:"Average SAT"`
	Diversity float32 `json:"Diversity"`
	Name string `json:"Name"`
	Size int64 `json:"Size"`
	Zip int64 `json:"Zip Code"`
}
```

#College Info
```bash
Region ID
0	U.S. Service Schools
1	New England (CT, ME, MA, NH, RI, VT)
2	Mid East (DE, DC, MD, NJ, NY, PA)
3	Great Lakes (IL, IN, MI, OH, WI)
4	Plains (IA, KS, MN, MO, NE, ND, SD)
5	Southeast (AL, AR, FL, GA, KY, LA, MS, NC, SC, TN, VA, WV)
6	Southwest (AZ, NM, OK, TX)
7	Rocky Mountains (CO, ID, MT, UT, WY)
8	Far West (AK, CA, HI, NV, OR, WA)
9	Outlying Areas (AS, FM, GU, MH, MP, PR, PW, VI)

Type of Degree
0	Non-degree-granting
1	Certificate degree
2	Associate degree
3	Bachelor's degree
4	Graduate degree
```
