package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type CreativeId int

type Creative struct {
	Id            CreativeId
	Height, Width int
	Path          string
}

var staticFilePath = filepath.Join("/srv", "www", "adserver", "static")

var adMap = make(map[CreativeId]*Creative)
var heightIndex = make(map[int][]CreativeId)
var widthIndex = make(map[int][]CreativeId)

type CreativeStat struct {
	Impressions   int
	LastServed    time.Time
	ServedPerHour map[int]int
}

var adStats = make(map[int]*CreativeStat)
var universalCreativeId CreativeId = 0

func getId() CreativeId {
	universalCreativeId += 1
	return universalCreativeId
}

func registerCreative(ad *Creative) {
	adMap[ad.Id] = ad
	heightIndex[ad.Height] = append(heightIndex[ad.Height], ad.Id)
	widthIndex[ad.Width] = append(widthIndex[ad.Width], ad.Id)
}

func NewCreative(path string, width, height int) *Creative {
	return &Creative{Id: getId(), Height: height, Width: width, Path: path}
}

func registerImpression(stat *CreativeStat, ad *Creative) {
	stat.Impressions += 1
	previousServed := stat.LastServed
	stat.LastServed = time.Now()
	stat.ServedPerHour[time.Now().Hour()] += 1
	log.Printf("Serving ad[%d]\n", ad.Id)
	log.Printf("Total Impressions: [%d]\n", stat.Impressions)
	log.Printf("Previously served at: [%s]\n", previousServed.UTC())
	log.Printf("Served this hour: [%d]\n", stat.ServedPerHour[time.Now().Hour()])
}

func getCreative(height, width int) *Creative {
	heightMatches := heightIndex[height]
	widthMatches := widthIndex[width]
	for _, heightMatch := range heightMatches {
		for _, widthMatch := range widthMatches {
			if heightMatch == widthMatch {
				return adMap[heightMatch]
			}
		}
	}
	return nil
}

func handleCreativeCall(w http.ResponseWriter, r *http.Request) {
	log.Println("Begin serve call")
	defer func() { log.Println("End serve call") }()
	height, _ := strconv.Atoi(r.URL.Query().Get("height"))
	width, _ := strconv.Atoi(r.URL.Query().Get("width"))
	ad := getCreative(height, width)
	if ad == nil {
		log.Printf("ERROR - Creative with matching width [%d] and height [%d] not found\n", width, height)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if _, exists := adStats[int(ad.Id)]; exists == false {
		adStats[int(ad.Id)] = &CreativeStat{0, time.Now(), make(map[int]int, 24)}
	}

	fullPath := filepath.Join(staticFilePath, "images", ad.Path)
	_, err := os.Stat(fullPath)
	if err != nil {
		log.Fatal(err)
	}

	http.ServeFile(w, r, fullPath)

	stat := adStats[int(ad.Id)]
	registerImpression(stat, ad)
}

func main() {
	log.Println("---starting adserver---")
	http.HandleFunc("/ad", handleCreativeCall)
	appNexusCreative := NewCreative("appnexus-logo.png", 640, 480)
	registerCreative(appNexusCreative)
	http.ListenAndServe(":8081", nil)
	log.Println("---stopping adserver---")
}
