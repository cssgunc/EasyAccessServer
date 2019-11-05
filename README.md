# EasyAccessServer

welcome to our EasyAccessServer

# To start go server

run in server/src/cmd directory

```bash
cd server/src/cmd
go run main.go
```

Once your terminal says "INFO[0001] listening on port 3000" you're good


# Http Requests
```bash
/user - POST request.body(idToken from firebase.auth.signinwithEmailandPasword)
	you will get back a student struct
/colleges - GET no body
  	you will get back all colleges
/majors - GET no body
	get back an array of strings
/matches - GET request.body(User UID)
  	you will get back array of colleges 
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
