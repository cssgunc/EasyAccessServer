# EasyAccessServer

welcome to our EasyAccess repo!!

# To start the react app

run in client directory

```bash
npm start
```

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
/matches - GET request.body(User UID)
  you will get back array of colleges 
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
