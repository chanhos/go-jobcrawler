package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)

type extractedJob struct {
	id       string
	location string
	title    string
	salary   string
	summary  string
}

func (e extractedJob) contentArr() []string {
	var content []string
	content = append(content, baseLink+e.id)
	content = append(content, " "+e.title)
	content = append(content, e.location)
	content = append(content, e.salary)
	content = append(content, e.summary)
	return content
}

var baseLink string = "https://kr.indeed.com/viewjob?jk="

// Scrap Indeed
func ScrapJob(term string) {

	baseURL := "https://kr.indeed.com/jobs?q=" + term + "&limit=50"

	var jobs []extractedJob
	totalpages := getPages(baseURL)
	c := make(chan []extractedJob)

	for i := 0; i < totalpages; i++ {
		go getPage(baseURL, i, c)
	}

	for i := 0; i < totalpages; i++ {
		extractedJob := <-c
		jobs = append(jobs, extractedJob...)
	}

	writeJobs_goRoutine(jobs)
	//writeJobs(jobs)
	fmt.Println("extracted Jobs Done : ", len(jobs), " jobs are here!!")
}

func writeJobs_goRoutine(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")

	isSuccess := make(chan bool)

	checkErr(err)

	//w := csv.NewWriter(file)
	w, cerr := ccsv.NewCsvWriter(file.Name())
	checkErr(cerr)

	header := []string{"Link", "제목", "지역", "급여", "공고요약"}

	werr := w.Write(header)
	checkErr(werr)

	for _, job := range jobs {
		go writeCsv(w, job, isSuccess)
	}

	for i := 0; i < len(jobs); i++ {
		<-isSuccess
	}

	defer w.Flush()
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")

	checkErr(err)

	w := csv.NewWriter(file)

	header := []string{"Link", "제목", "지역", "급여", "공고요약"}
 
	werr := w.Write(header)
	checkErr(werr)

	for _, job := range jobs {
		err := w.Write(job.contentArr())
		checkErr(err)
	}

	defer w.Flush()
}

func writeCsv(w *ccsv.CsvWriter, job extractedJob, c chan<- bool) {
	err := w.Write(job.contentArr())

	if err != nil {
		c <- false
	} else {
		c <- true
	}
}

func getPage(url string, pageNo int, oc chan<- []extractedJob) {
	var jobs []extractedJob

	c := make(chan extractedJob)

	pageUrl := url + "&start=" + strconv.Itoa(10*pageNo)
	resp, err := http.Get(pageUrl)
	checkErr(err)
	checkCode(resp)

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	checkErr(err)

	searchCard := doc.Find(".fs-unmask")

	searchCard.Each(func(i int, card *goquery.Selection) {
		go extractJob(card, c)
		//jobs = append(jobs, job)
	})

	for i := 0; i < searchCard.Size(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	oc <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	jobid, _ := card.Find(".jcs-JobTitle").Attr("data-jk")

	title := CleanString(card.Find(".jobTitle>a").Text())

	location := CleanString(card.Find(".companyLocation").Text())

	salary := CleanString(card.Find(".salary-snippet").Text())

	summary := CleanString(card.Find(".job-snippet").Text())

	c <- extractedJob{id: jobid, title: title, location: location, salary: salary, summary: summary}
}

func getPages(url string) int {
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)
	pages := 0
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages

}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
}

func CleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}
