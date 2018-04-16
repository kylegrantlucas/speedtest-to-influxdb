package main

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/dchest/uniuri"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/kylegrantlucas/speedtest-to-influxdb/speedtest/sthttp"
	"github.com/kylegrantlucas/speedtest-to-influxdb/speedtest/tests"
	"github.com/urfave/cli"
)

// Version placeholder, injected in Makefile
var Version string
var DB client.Client
var conf config

type config struct {
	numclosest      int
	numlatencytests int
	algotype        string
	httptimeout     time.Duration
	dlsizes         []int
	ulsizes         []int
	configurl       string
	serverurl       string
	useragent       string
	ipInterface     string
	blacklist       []string
}

type results struct {
	server   sthttp.Server
	latency  *float64
	download *float64
	upload   *float64
}

func init() {
	conf = config{
		numclosest:      3,
		numlatencytests: 5,
		algotype:        "avg",
		httptimeout:     120,
		dlsizes:         []int{350, 500, 750, 1000, 1500, 2000},
		ulsizes:         []int{int(0.25 * 1024 * 1024), int(0.5 * 1024 * 1024), int(1.0 * 1024 * 1024)},
		configurl:       "http://c.speedtest.net/speedtest-config.php?x=" + uniuri.New(),
		serverurl:       "http://c.speedtest.net/speedtest-servers-static.php?x=" + uniuri.New(),
		useragent:       "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.21 Safari/537.36",
	}
}

func main() {
	// seeding randomness
	rand.Seed(time.Now().UTC().UnixNano())

	// set logging to stdout for global logger
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
		cli.StringSliceFlag{
			Name:  "blacklist, b",
			Usage: "Blacklist a server.  Use this multiple times for more than one server",
		},
		cli.StringFlag{
			Name:  "useragent, ua",
			Usage: "Specify a useragent string",
		},
		cli.IntFlag{
			Name:  "numclosest, nc",
			Value: conf.numclosest,
			Usage: "Number of 'closest' servers to find",
		},
		cli.IntFlag{
			Name:  "numlatency, nl",
			Value: conf.numlatencytests,
			Usage: "Number of latency tests to perform",
		},
		cli.StringFlag{
			Name:  "interface, I",
			Usage: "Source IP address or name of an interface",
		},
		cli.StringFlag{
			Name:  "influxUsername, iu",
			Usage: "The username for the influxDB instance",
		},
		cli.StringFlag{
			Name:  "influxPasword, ip",
			Usage: "The password for the influxDB instance",
		},
		cli.StringFlag{
			Name:  "influxDB, idb",
			Usage: "The name for the influxDB database",
		},
	}

	// toggle our switches and setup variables
	app.Action = func(c *cli.Context) {
		var err error
		conf.numclosest = c.Int("numclosest")
		conf.numlatencytests = c.Int("numlatency")
		conf.ipInterface = c.String("interface")
		if len(c.StringSlice("blacklist")) > 0 {
			conf.blacklist = c.StringSlice("blacklist")
		}

		DB, err = influxDBClient(c.String("influxUsername"), c.String("influxPassword"))
		if err != nil {
			log.Printf("error connecting to influxdb: %v", err)
		}

		stClient := sthttp.NewClient(
			&sthttp.SpeedtestConfig{
				ConfigURL:       conf.configurl,
				ServersURL:      conf.serverurl,
				AlgoType:        conf.algotype,
				NumClosest:      conf.numclosest,
				NumLatencyTests: conf.numlatencytests,
				Interface:       conf.ipInterface,
				Blacklist:       conf.blacklist,
				UserAgent:       conf.useragent,
			},
			&sthttp.HTTPConfig{
				HTTPTimeout: conf.httptimeout * time.Second,
			},
		)

		tester := tests.NewTester(
			stClient,
			conf.dlsizes,
			conf.ulsizes,
		)

		res, err := runSpeedtest(c, stClient, tester)
		if err != nil {
			log.Printf("error running speedtest: %v", err)
		}

		if res.latency != nil && res.upload != nil && res.download != nil {
			err := writeMetrics(DB, c.String("influxDB"), res)
			if err != nil {
				log.Printf("error writing to influxdb: %v", err)
			}

			log.Printf("writing speedtest results {server: %s, ping: %3.2fms, download: %3.2fMbps, upload: %3.2fMbps} to influxdb", res.server.Name, *res.latency, *res.download, *res.upload)
		} else {
			log.Printf("speedtest results have no values, skipping writing to influxdb")
		}

		// exit nicely
		os.Exit(0)
	}

	// run the app
	err := app.Run(os.Args)
	if err != nil {
		log.Printf("error running application: %v", err)
	}
}

func runSpeedtest(c *cli.Context, stClient *sthttp.Client, tester *tests.Tester) (results, error) {
	var testServer sthttp.Server
	var err error

	var allServers []sthttp.Server
	allServers, err = stClient.GetServers()
	if err != nil {
		return results{}, err
	}

	if c.String("server") != "" {
		testServer = tester.FindServer(c.String("server"), allServers)
		testServer.Latency, err = stClient.GetLatency(testServer, stClient.GetLatencyURL(testServer))
		if err != nil {
			return results{latency: &testServer.Latency}, err
		}
	} else {
		closestServers := stClient.GetClosestServers(allServers)
		testServer, err = stClient.GetFastestServer(closestServers)
		if err != nil {
			return results{}, err
		}
	}

	dmbps, err := tester.Download(testServer)
	if err != nil {
		return results{}, err
	}

	umbps, err := tester.Upload(testServer)
	if err != nil {
		return results{}, err
	}

	return results{latency: &testServer.Latency, download: &dmbps, upload: &umbps, server: testServer}, nil
}

func influxDBClient(username string, password string) (client.Client, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
	if err != nil {
		return c, err
	}
	return c, nil
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
		"server_url":     res.server.Name,
	}

	fields := map[string]interface{}{
		"latency":         *res.latency,
		"download":        *res.download,
		"upload":          *res.upload,
		"server_distance": res.server.Distance,
	}

	point, err := client.NewPoint(
		"speedtest",
		tags,
		fields,
		time.Now(),
	)
	if err != nil {
		return err
	}

	bp.AddPoint(point)

	err = c.Write(bp)
	if err != nil {
		return err
	}

	return nil
}
