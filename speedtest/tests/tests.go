package tests

import (
	"fmt"
	"log"
	"strings"

	"github.com/kylegrantlucas/speedtest-to-influxdb/speedtest/misc"
	"github.com/kylegrantlucas/speedtest-to-influxdb/speedtest/sthttp"
)

var (
	// DefaultDLSizes defines the default download sizes
	DefaultDLSizes = []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}
	// DefaultULSizes defines the default upload sizes
	DefaultULSizes = []int{int(0.25 * 1024 * 1024), int(0.5 * 1024 * 1024), int(1.0 * 1024 * 1024), int(1.5 * 1024 * 1024), int(2.0 * 1024 * 1024)}
)

// Tester defines a Speedtester client tester
type Tester struct {
	Client   *sthttp.Client
	DLSizes  []int
	ULSizes  []int
	Quiet    bool
	Report   bool
	Debug    bool
	AlgoType string
}

func NewTester(client *sthttp.Client, dlsizes []int, ulsizes []int) *Tester {
	return &Tester{
		Client:  client,
		DLSizes: dlsizes,
		ULSizes: ulsizes,
	}
}

// Download will perform the "normal" speedtest download test
func (tester *Tester) Download(server sthttp.Server) (float64, error) {
	var urls []string
	var maxSpeed float64
	var avgSpeed float64

	// http://speedtest1.newbreakcommunications.net/speedtest/speedtest/
	for size := range tester.DLSizes {
		url := server.URL
		splits := strings.Split(url, "/")
		baseURL := strings.Join(splits[1:len(splits)-1], "/")
		randomImage := fmt.Sprintf("random%dx%d.jpg", tester.DLSizes[size], tester.DLSizes[size])
		downloadURL := "http:/" + baseURL + "/" + randomImage
		urls = append(urls, downloadURL)
	}

	for u := range urls {
		dlSpeed, err := tester.Client.DownloadSpeed(urls[u])
		if err != nil {
			return 0, err
		}

		if tester.AlgoType == "max" {
			if dlSpeed > maxSpeed {
				maxSpeed = dlSpeed
			}
		} else {
			avgSpeed = avgSpeed + dlSpeed
		}
	}

	if tester.AlgoType != "max" {
		return avgSpeed / float64(len(urls)), nil
	}
	return maxSpeed, nil

}

// Upload runs a "normal" speedtest upload test
func (tester *Tester) Upload(server sthttp.Server) (float64, error) {
	// https://github.com/sivel/speedtest-cli/blob/master/speedtest-cli
	var ulsize []int
	var maxSpeed float64
	var avgSpeed float64

	for size := range tester.ULSizes {
		ulsize = append(ulsize, tester.ULSizes[size])
	}

	for i := 0; i < len(ulsize); i++ {
		r := misc.Urandom(ulsize[i])
		ulSpeed, err := tester.Client.UploadSpeed(server.URL, "text/xml", r)
		if err != nil {
			return 0, err
		}

		if tester.AlgoType == "max" {
			if ulSpeed > maxSpeed {
				maxSpeed = ulSpeed
			}
		} else {
			avgSpeed = avgSpeed + ulSpeed
		}

	}

	if tester.AlgoType != "max" {
		return avgSpeed / float64(len(ulsize)), nil
	}
	return maxSpeed, nil
}

// FindServer will find a specific server in the servers list
func (tester *Tester) FindServer(id string, serversList []sthttp.Server) sthttp.Server {
	var foundServer sthttp.Server
	for s := range serversList {
		if serversList[s].ID == id {
			foundServer = serversList[s]
		}
	}
	if foundServer.ID == "" {
		log.Printf("cannot locate server id '%s' in our list of speedtest servers", id)
	}
	return foundServer
}
