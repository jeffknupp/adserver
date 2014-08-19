package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jeffknupp/adserver/creative"
)

var staticFilePath = filepath.Join("/srv", "www", "adserver", "static")

var adMap = make(map[creative.CreativeId]*creative.Creative)
var heightIndex = make(map[int][]creative.CreativeId)
var widthIndex = make(map[int][]creative.CreativeId)

var adStats = make(map[int]*creative.CreativeStat)

func registerCreative(ad *creative.Creative) {
	adMap[ad.Id] = ad
	heightIndex[ad.Height] = append(heightIndex[ad.Height], ad.Id)
	widthIndex[ad.Width] = append(widthIndex[ad.Width], ad.Id)
}

func recordImpression(stat *creative.CreativeStat, ad *creative.Creative) {
	stat.Impressions += 1
	previousServed := stat.LastServed
	stat.LastServed = time.Now()
	stat.ServedPerHour[time.Now().Hour()] += 1
	log.Printf("Serving ad[%d]\n", ad.Id)
	log.Printf("Total Impressions: [%d]\n", stat.Impressions)
	log.Printf("Previously served at: [%s]\n", previousServed.UTC())
	log.Printf("Served this hour: [%d]\n", stat.ServedPerHour[time.Now().Hour()])
}

func getCreative(height, width int) *creative.Creative {
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
		adStats[int(ad.Id)] = creative.NewCreativeStat()
	}

	fullPath := filepath.Join(staticFilePath, "images", ad.Path)
	_, err := os.Stat(fullPath)
	if err != nil {
		log.Fatal(err)
	}

	http.ServeFile(w, r, fullPath)

	stat := adStats[int(ad.Id)]
	recordImpression(stat, ad)
}

func main() {
	log.Println("---starting adserver---")
	http.HandleFunc("/ad", handleCreativeCall)
	appNexusCreative := creative.NewCreative("appnexus-logo.png", 640, 480)
	registerCreative(appNexusCreative)
	http.ListenAndServe(":8081", nil)
	log.Println("---stopping adserver---")
}
