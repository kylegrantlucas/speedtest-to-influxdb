package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/kylegrantlucas/speedtest"
	"github.com/kylegrantlucas/speedtest/http"
	"github.com/urfave/cli"
)

// Version placeholder, injected in Makefile
var Version string

type results struct {
	server   http.Server
	latency  *float64
	download *float64
	upload   *float64
}

func main() {
	// seeding randomness
	rand.Seed(time.Now().UTC().UnixNano())

	log.SetOutput(os.Stdout)

	// setting up cli settings
	app := cli.NewApp()
	app.Name = "speedtest-to-influxdb"
	app.Usage = "Speedtest -> InfluxDB ingestion daemon"
	app.Author = "Kyle Lucas"
	app.Version = Version

	// setup cli flags
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "server, s",
			Usage: "Use a specific server",
		},
		cli.StringFlag{
			Name:  "influxUsername, u",
			Usage: "The username for the influxDB instance",
		},
		cli.StringFlag{
			Name:  "influxPasword, p",
			Usage: "The password for the influxDB instance",
		},
		cli.StringFlag{
			Name:  "influxDB, db",
			Usage: "The name for the influxDB database",
		},
		cli.StringFlag{
			Name:  "influxURL, url",
			Value: "http://localhost:8086",
			Usage: "The name for the influxDB database",
		},
		cli.IntFlag{
			Name:  "interval, i",
			Value: 20,
			Usage: "The amount of time in minutes to wait between speedtest runs",
		},
	}

	// toggle our switches and setup variables
	app.Action = func(c *cli.Context) {
		db, err := influxDBClient(c.String("influxURL"), c.String("influxUsername"), c.String("influxPassword"))
		if err != nil {
			log.Printf("error connecting to influxdb: %v", err)
		}

		speedtestClient, err := speedtest.NewDefaultClient()
		if err != nil {
			log.Printf("couldn't create speedtest client: %v", err)
		}

		// Run speedtest indefinitely
		for {
			res, err := runSpeedtest(c, speedtestClient)
			if err != nil {
				log.Printf("error running speedtest: %v", err)
			}

			if res.latency != nil && res.upload != nil && res.download != nil {
				err := writeMetrics(db, c.String("influxDB"), res)
				if err != nil {
					log.Printf("error writing to influxdb: %v", err)
				}

				log.Printf("writing speedtest results {server: %s, ping: %3.2fms, download: %3.2fMbps, upload: %3.2fMbps} to influxdb", res.server.Sponsor, *res.latency, *res.download, *res.upload)
			} else {
				log.Printf("speedtest results have no values, skipping writing to influxdb")
			}

			<-time.After(time.Duration(c.Int("interval")) * time.Minute)
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Printf("error running application: %v", err)
	}
}

func runSpeedtest(c *cli.Context, client *speedtest.Client) (results, error) {
	server, err := client.GetServer(c.String("server"))
	if err != nil {
		return results{}, err
	}

	dmbps, err := client.Download(server)
	if err != nil {
		return results{}, err
	}

	umbps, err := client.Upload(server)
	if err != nil {
		return results{}, err
	}

	return results{
		latency:  &server.Latency,
		download: &dmbps,
		upload:   &umbps,
		server:   server,
	}, nil
}

func influxDBClient(url string, username string, password string) (client.Client, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     url,
		Username: username,
		Password: password,
	})

	return c, err
}

func writeMetrics(c client.Client, database string, res results) error {
	bp, err := client.NewBatchPoints(
		client.BatchPointsConfig{
			Database:  database,
			Precision: "s",
		},
	)
	if err != nil {
		return err
	}

	tags := map[string]string{
		"server_name":    res.server.Name,
		"server_id":      res.server.ID,
		"server_sponsor": res.server.Sponsor,
		"server_url":     res.server.URL,
		"server_country": res.server.Country,
	}

	fields := map[string]interface{}{
		"latency":         *res.latency,
		"download":        *res.download,
		"upload":          *res.upload,
		"server_distance": res.server.Distance,
	}

	point, err := client.NewPoint("speedtest", tags, fields, time.Now())
	if err != nil {
		return err
	}

	bp.AddPoint(point)

	err = c.Write(bp)
	return err
}
