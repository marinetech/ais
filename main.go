package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	// "github.com/xidram/ais/geo"
	"github.com/im7mortal/UTM"
	"log"
	"os"
	"path/filepath"
	"math"
	"strconv"
	"strings"
	"time"
	// "encoding/json"
	"net/smtp"
)


// ------  Global Scope ------ //
const Base_lat = "33.0350685"
const Base_lon = "34.9447517"
const MAX_DISTANCE_ALLOWED = 200
const SMTP_SERVER = "mr1.haifa.ac.il"
const BUOY_NAME = "tabs225m09"
var logFile *os.File

// Sends mail using Haifa university SMTP_SERVER
// subject, body, recipient are all hards coded as well as the auth details
func send_mail() {

	log.Println("-I- sending boundaries alert through " + SMTP_SERVER)

	body := "According to the most recent AIS info, the buoy is " +
					strconv.FormatInt(MAX_DISTANCE_ALLOWED, 32) +
					" meters away from its base location\r\n"
	auth := smtp.PlainAuth("", "themo@univ.haifa.ac.il", "TexasA&M1", SMTP_SERVER)
	to := []string{"imardix@univ.haifa.ac.il","sdahan3@univ.haifa.ac.il"}
	// to := []string{"imardix@univ.haifa.ac.il"}
	msg := []byte("To: Themo\r\n" +
		"Subject: Alert: " + BUOY_NAME + " is out of boundaries\r\n" +
		"\r\n" +
		body)
	err := smtp.SendMail(SMTP_SERVER + ":25", auth, "themo@univ.haifa.ac.il", to, msg)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Alert was sent successfully!")
	}
}

func send_mail_no_report() {

	log.Println("-I- sending no_report alert through " + SMTP_SERVER)

	body := "No AIS info was recived during the last hour"
	auth := smtp.PlainAuth("", "themo@univ.haifa.ac.il", "TexasA&M1", SMTP_SERVER)
	to := []string{"imardix@univ.haifa.ac.il","sdahan3@univ.haifa.ac.il"}
	// to := []string{"imardix@univ.haifa.ac.il"}
	msg := []byte("To: Themo\r\n" +
		"Subject: Alert: " + BUOY_NAME + " AIS info is missing\r\n" +
		"\r\n" +
		body)
	err := smtp.SendMail(SMTP_SERVER + ":25", auth, "themo@univ.haifa.ac.il", to, msg)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Alert was sent successfully!")
	}
}

func init_log() {
	current_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	log_path := current_dir + string(os.PathSeparator) + BUOY_NAME + ".log"
	logFile, _ := os.OpenFile(log_path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	log.SetOutput(logFile)
	log.Print("------------------------------------------")
}

func scrape() (string, string) {

	log.Println("Verifying AIS information for " + BUOY_NAME)

	url := "https://www.marinetraffic.com/en/ais/details/ships/shipid:1336912/mmsi:244010235/vessel:BOEI%202"
	var coordinates string
	var lastUpdate string

	log.Println("Scraping: " + url)

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	// find the coordinates
	log.Println("extracting last location")
	doc.Find("body span strong a").Each(func(index int, item *goquery.Selection) {
		linkTag := item
		link, _ := linkTag.Attr("href")
		linkText := linkTag.Text()

		if strings.Contains(link, "centerx") {
			coordinates = linkText
		}
	})

	// find the time of the last update
	log.Println("extracting last update time")
	doc.Find(".table-cell.cell-full.collapse-768 .group-ib strong").Each(func(index int, item *goquery.Selection) {
		divTag := item
		divText := divTag.Contents().Text()
		if strings.Contains(divText, "minutes ago") {
			lastUpdate = divText
		}
	})

	return coordinates, lastUpdate
}

func last_upadte_time(lastUpdate string) {
	minutes_ago := strings.Split(lastUpdate, " ")[0]

	i, _ := strconv.Atoi(minutes_ago)
	i = i * -1

	t2 := time.Now().Add(time.Minute * time.Duration(i))
	log.Println("last update: " + minutes_ago + " minutes ago [" + t2.Format("Mon Jan _2 15:04:05 2006") + "]")

}

// removes special charcters from string
func stripCtlAndExtFromBytes(str string) string {
	b := make([]byte, len(str))
	var bl int
	for i := 0; i < len(str); i++ {
		c := str[i]
		if c >= 32 && c < 127 {
			b[bl] = c
			bl++
		}
	}
	return string(b[:bl])
}

func getUTMtuple(str_lat string, str_lon string) UTM.LatLon {
	lat := stripCtlAndExtFromBytes(str_lat)
	lat = strings.Trim(lat, " ")
	flt_lat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Println("flt_lat: %f ", flt_lat)

	lon := stripCtlAndExtFromBytes(str_lon) //remove degree symbol °
	lon = strings.Trim(lon, " ")
	flt_lon, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		panic(err.Error())
	}
	// fmt.Println("flt_lon: %f ", flt_lon)


	return UTM.LatLon{flt_lat, flt_lon}

}

func main() {

	init_log(); 	defer logFile.Close()

	coordinates, lastUpdate := scrape()
	if lastUpdate == "" {
		log.Fatal("When was the last update?")
		send_mail_no_report()
	}
	if coordinates == "" {
		log.Fatal("Can't find coordinates")
	}
	log.Println(coordinates)
	last_upadte_time(lastUpdate)

	// get base location in UTM
	base_latlon := getUTMtuple(Base_lat, Base_lon)
	base_utm, _ := base_latlon.FromLatLon()

	// get current location in UTM
	str_lat := strings.Split(coordinates, "/")[0]
	str_lon := strings.Split(coordinates, "/")[1]
	current_latLon := getUTMtuple(str_lat, str_lon)
	current_utm , _:= current_latLon.FromLatLon()

	// calculate the distance - math.Sqrt((N1-N2)²+(E1-E2)²)/1000 - results in km
	Northing := math.Pow((base_utm.Northing - current_utm.Northing),2)
	Easting :=  math.Pow((base_utm.Easting - current_utm.Easting),2)
	distance := math.Sqrt(Northing + Easting)

	log.Println(fmt.Sprintf("Distance: %f meters", distance,))
	if distance > MAX_DISTANCE_ALLOWED {
		send_mail()
	}


	// result, _ := current_latLon.FromLatLon()
	// fmt.Println(
	// 	fmt.Sprintf(
	// 		"Easting: %f; Northing: %f; ZoneNumber: %d; ZoneLetter: %s;",
	// 		result.Easting,
	// 		result.Northing,
	// 		result.ZoneNumber,
	// 		result.ZoneLetter,
	// ))


}
