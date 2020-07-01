package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Patient definition
type Patient struct {
	CurrentStatus    string `json:"currentstatus"`
	StatusChangeDate string `json:"statuschangedate"`
	DetectedState    string `json:"detectedstate"`
}

// RawData definition
type RawData struct {
	Data []Patient `json:"raw_data"`
}

// Delta definition
type Delta struct {
	Active       int
	Deaths       int
	Recovered    int
	Hospitalized int
	Migrated     int
}

func main1() {
	//https://api.covid19india.org/raw_data.json
	resp, err := http.Get("https://api.covid19india.org/raw_data.json")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	rd := &RawData{}

	err = json.Unmarshal(d, rd)
	if err != nil {
		panic(err)
	}
	status := make(map[string]*Delta)
	today := time.Now().Format("02/01/2006")
	for _, p := range rd.Data {
		//	_date, _ := time.Parse("02/01/2006", p.StatusChangeDate)
		if p.StatusChangeDate == today {
			if _, ok := status[p.DetectedState]; !ok {
				status[p.DetectedState] = &Delta{}
			} else {
				delta := status[p.DetectedState]
				delta.Active = delta.Active + 1
				switch p.CurrentStatus {
				case "Recovered":
					delta.Recovered = delta.Recovered + 1
				case "Hospitalized":
					delta.Hospitalized = delta.Hospitalized + 1
				case "Deceased":
					delta.Deaths = delta.Deaths + 1
				case "Migrated":
					delta.Migrated = delta.Migrated + 1
				}
				//status[p.DetectedState] = delta
			}
		}

	}
	fmt.Println("State\t\t\tA\tH\tR\tD\tM")
	for k, d := range status {
		fmt.Printf("%s\t\t\t%d\t%d\t%d\t%d\n", k, d.Active, d.Hospitalized, d.Deaths, d.Migrated)
	}
}

func main() {

	http.HandleFunc("/alerts", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			fmt.Print(err)
		}
		fmt.Println(string(b))
	})

	http.ListenAndServe(":9099", nil)
}
