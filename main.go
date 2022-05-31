package main

import (
	"os"
	"strings"

	"github.com/chanhos/go-jobcrawler/scrapper"
	"github.com/labstack/echo/v4"
)

func main() {
	//scrapper.ScrapJob("spring")
	e := echo.New()
	e.GET("/", handleIndex)
	e.POST("/scraper", handleScrpe)

	e.Logger.Fatal((e.Start(":8081")))
}

func handleIndex(c echo.Context) error {
	//return c.String(http.StatusOK, "It's Simple server")
	return c.File("home.html")
}

func handleScrpe(c echo.Context) error {
	defer os.Remove("jobs.csv")
	term := strings.ToLower(scrapper.CleanString(c.FormValue("term")))
	scrapper.ScrapJob(term)
	return c.Attachment("jobs.csv", term+"_find_job.csv")
}
