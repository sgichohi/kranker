package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/puerkitobio/goquery"
)

const (
	//POSTDATA is a magic suffix for post reqs to the page
	POSTDATA = "&ctl00%24cphMain%24TabContainer1%24Marks%24ddlYear=2014&ctl00%24cphMain%24TabContainer1%24Marks%24btnFind=Find"
	//KNECURL is the URL for the Kenya National Examination Council
	KNECURL = "http://www.knec-portal.ac.ke/RESULTS/ResultKCPE.aspx"
	//DEBUG is determines whether we should print debug info
	DEBUG = false
	//MAXATTEMPTS for getting data per candidate
	MAXATTEMPTS = 3
)

var studentinfo = map[string]string{
	"total":      "#ctl00_cphMain_TabContainer1_Marks_txtTotal",
	"eng":        "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl02_MKS",
	"kis":        "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl03_MKS",
	"mat":        "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl04_MKS",
	"sci":        "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl05_MKS",
	"ssr":        "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl06_MKS",
	"schoolName": "#ctl00_cphMain_TabContainer1_Marks_txtSchool",
	"gender":     "#ctl00_cphMain_TabContainer1_Marks_txtGender",
	"engGrade":   "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl02_GRADE",
	"kisGrade":   "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl03_GRADE",
	"matGrade":   "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl04_GRADE",
	"sciGrade":   "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl05_GRADE",
	"ssrGrade":   "#ctl00_cphMain_TabContainer1_Marks_Gridview1_ctl06_GRADE",
}

var hiddenfields = getHiddenField()

type pageResult struct {
	Page  string
	Index string
}

//This pattern is a poor man's way of having a toggleable log
func debug(stuff string) {
	if DEBUG {
		fmt.Println(stuff)
	}
}

func getCandidateResults(index string, client *http.Client) (htmlPage string, err error) {

	data := hiddenfields + index + POSTDATA
	bb := bytes.NewBuffer([]byte(data))
	req, err := http.NewRequest("POST", KNECURL, bb)
	if err != nil {
		return "", errors.New("Couldn't make post request")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-agent", "Mozilla/5.0")
	attempt := 0
	debug("making request")
	for {
		time.Sleep(defaultBackoff.Duration(attempt))
		resp, err := client.Do(req)

		if err != nil {
			if attempt < MAXATTEMPTS {
				attempt++
				debug("failed" + string(attempt))
				continue
			}
			log.Fatal(err)
			return "", errors.New("Making request error")
		}
		restr, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return "", errors.New("Body parse error")
		}
		return string(restr), nil
	}

}

func worker(schools <-chan string, client *http.Client, students chan<- map[string]string) {

	var wg sync.WaitGroup
	debug("working")
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for schoolIndex := range schools {
				studs := getStudentDetails(schoolIndex, client)
				debug(fmt.Sprintf("student: %v", studs))
				for _, stud := range studs {
					students <- stud
				}
			}

		}()

	}
	wg.Wait()

}

func getHiddenField() string {

	hfield := "ctl00_cphMain_ToolkitScriptManager1_HiddenField=%3B%3BAjaxControlToolkit"
	hfield += "%2C+Version%3D3.5.40412.0%2C+Culture%3Dneutral%2C+PublicKeyToken%3D28f01b0e84b6d53e"
	hfield += "%3Aen-GB%3A1547e793-5b7e-48fe-8490-03a375b13a"
	hfield += "33%3A475a4ef5%3Aeffe2a26%3A8e94f951%3A1d3ed089&ctl00_cphMain_TabContainer1_ClientState=%"
	hfield += "7B%22ActiveTabIndex%22%3A0%2C%22TabState%22%3A%5Btrue%5D%7D&__EVENTTARGET=&__EVENTARGUMENT=&__VIEWSTATE"
	hfield += "=%2FwEPDwUJNzIwODUwMzE2D2QWAmYPZBYCAgQPZBYEAgEPDxYCHgRUZXh0BQ8yMSBKYW51YXJ5IDIwMTVkZAILD2QWAgIBD2QWAmYPZBY"
	hfield += "CAgEPZBYCAhsPPCsADQBkGAMFHl9fQ29udHJvbHNSZXF1aXJlUG9zdEJhY2tLZXlfXxYBBRtjdGwwMCRjcGhNYWluJFRhYkNvbnRhaW5lcjEFG2N0bD"
	hfield += "AwJGNwaE1haW4kVGFiQ29udGFpbmVyMQ8PZGZkBStjdGwwMCRjcGhNYWluJFRhYkNvbnRhaW5lcjEkTWFya3MkR3JpZHZpZXcxD2dkCcktLvt%2FXIMC"
	hfield += "IziOu3%2Bqi8MlsNs%3D&__VIEWSTATEGENERATOR=8A3A71A8&__EVENTVALIDATION=%2FwEWBAKgps3rDQLS1aG9AgKX%2F9f8CwK2lIrpBn45EiR"
	hfield += "ZIqaQ%2FpmQ8A5Ae9qmxnMY&ctl00%24cphMain%24TabContainer1%24Marks%24txtIndex="

	return hfield

}

func parsePage(pageRes *pageResult) (stud map[string]string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(pageRes.Page))
	student := make(map[string]string)
	if err != nil {
		return student, err
	}

	for subj, param := range studentinfo {

		f, ok := doc.Find(param).Attr("value")
		if !ok {
			return student, errors.New("Bad page")
		}
		student[subj] = strings.TrimSpace(f)
	}
	student["index"] = pageRes.Index
	return student, nil
}

func main() {

	client := &http.Client{}
	countysch := getCountySchools()
	fields := []string{"total", "eng",
		"kis", "mat", "sci", "ssr",
		"schoolName", "gender", "engGrade", "kisGrade",
		"matGrade", "sciGrade", "ssrGrade"}

	fmt.Println("total,eng,kis,mat,sci,ssr,schoolName,gender,engGrade,kisGrade,matGrade,sciGrade,ssrGrade")
	csch := make(chan string)
	go func() {
		for _, cs := range countysch {
			csch <- cs
		}
		close(csch)
	}()
	students := make(chan map[string]string)
	go func() {
		worker(csch, client, students)

		close(students)
	}()
	//print out students

	for student := range students {
		out := ""
		for i, details := range fields {

			if i < len(fields)-1 {
				out += student[details] + ","
			} else {

				out += student[details]
			}
		}
		fmt.Println(out)
	}

}
