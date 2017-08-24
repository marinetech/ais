package main

import (
	"fmt"
	// "golang.org/x/net/html"
	// "io/ioutil"
	"github.com/PuerkitoBio/goquery"
	"log"
	// "net/http"
	// "os"
	"strings"
	"time"
)

func scrape() (string, string) {
	url := "https://www.marinetraffic.com/en/ais/details/ships/shipid:1336912/mmsi:244010235/vessel:BOEI%202"
	var coordinates string
	var lastUpdate string

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("body span strong a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		linkText := linkTag.Text()

		if strings.Contains(link, "centerx") {
			fmt.Printf("%s\n", linkText)
			coordinates = linkText
		}
	})

	doc.Find(".table-cell.cell-full.collapse-768 .group-ib strong").Each(func(index int, item *goquery.Selection) {
		divTag := item
		divText := divTag.Contents().Text()
		if strings.Contains(divText, "minutes ago") {
			fmt.Printf("%s\n", divText)
			lastUpdate = divText
		}
	})

	t := time.Now()
	fmt.Println(t.Format("Mon Jan _2 15:04:05 2006"))
	t2 := time.Now().Add(time.Minute * -5)
	fmt.Println(t2.Format("Mon Jan _2 15:04:05 2006"))

	return coordinates, lastUpdate

}

func main() {
	scrape()
}
