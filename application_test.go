package main

import (
	"reflect"
	"testing"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/kylegrantlucas/speedtest"
	"github.com/urfave/cli"
)

// func Test_main(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			main()
// 		})
// 	}
// }

func Test_runSpeedtest(t *testing.T) {
	type args struct {
		c      *cli.Context
		client *speedtest.Client
	}
	tests := []struct {
		name    string
		args    args
		want    results
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runSpeedtest(tt.args.c, tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("runSpeedtest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runSpeedtest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_influxDBClient(t *testing.T) {
	type args struct {
		username string
		password string
	}
	tests := []struct {
		name    string
		args    args
		want    client.Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := influxDBClient(tt.args.username, tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("influxDBClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("influxDBClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeMetrics(t *testing.T) {
	type args struct {
		c        client.Client
		database string
		res      results
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := writeMetrics(tt.args.c, tt.args.database, tt.args.res); (err != nil) != tt.wantErr {
				t.Errorf("writeMetrics() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
