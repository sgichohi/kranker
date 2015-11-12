package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	//CANDIDATEGROUPTHRESHOLD is number of tolerated failures before hopping to the next candidate group
	CANDIDATEGROUPTHRESHOLD = 2
)

var candidates map[int][]int

func init() {
	candidates = genCandidateIndex()
}

func getStudentDetails(schoolIndex string, client *http.Client) []map[string]string {

	lst := []map[string]string{}
	for j := 0; j < len(candidates); j++ {
		var errCount uint64
		for n := 0; n < len(candidates[j]); n++ {
			candidateIndex := "000" + strconv.Itoa(candidates[j][n])
			candidateIndex = candidateIndex[(len(candidateIndex) - 3):]

			stud := schoolIndex + string(candidateIndex)
			debug(stud)
			res, err := getCandidateResults(stud, client)
			if err != nil {
				log.Fatal(err)
			}
			debug("got results")
			p := &pageResult{Page: res, Index: stud}
			student, err := parsePage(p)
			debug(fmt.Sprintf("parsed student: %v", student))
			if err != nil {
				debug("error")
				if errCount > CANDIDATEGROUPTHRESHOLD {
					break
				}
				errCount++

			} else {
				lst = append(lst, student)
				debug("appended a student")
				errCount = 0

			}
		}
	}

	return lst

}

/*Within a school there are bands of candidates representing
the type of admission: For instance repeat students and so on*/
func genCandidateIndex() map[int][]int {

	candidates := make(map[int][]int)

	for i := 0; i < 300; i++ {
		candidates[0] = append(candidates[0], i+1)
	}
	for i := 300; i < 600; i++ {
		candidates[1] = append(candidates[1], i+1)
	}
	for i := 700; i < 800; i++ {
		candidates[2] = append(candidates[2], i+1)
	}
	for i := 800; i < 900; i++ {
		candidates[3] = append(candidates[3], i+1)
	}
	return candidates
}

/*We have a list of schools in every county*/
func getCountySchools() []string {
	dat, err := ioutil.ReadFile("index_nums")
	if err != nil {
		panic(err)
	}
	indexes := strings.Split(string(dat), "\n")

	chk := make(map[string]struct{})
	countySchools := []string{}
	for _, index := range indexes {
		if len(index) != 11 {
			continue
		}

		_, ok := chk[string(index[0:8])]

		if !ok {
			countySchools = append(countySchools, string(index[0:8]))
		}
	}
	return countySchools
}
