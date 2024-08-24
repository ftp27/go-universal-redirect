package analytics

import (
	"context"
	"time"

	"github.com/InfluxCommunity/influxdb3-go/influxdb3"
)

type InfluxAnalytics struct {
	database string
}

type Data struct {
	Measurement string    `lp:"measurement"`
	Type        string    `lp:"tag,type"`
	Platform    string    `lp:"tag,platform"`
	Value       int64     `lp:"field,value"`
	Time        time.Time `lp:"timestamp"`
}

func NewInfluxAnalytics(database string) *InfluxAnalytics {
	return &InfluxAnalytics{database: database}
}

func (a *InfluxAnalytics) LogClick(meta string, platform string, ctx context.Context) error {
	client, err := influxdb3.NewFromEnv()
	if err != nil {
		return err
	}
	defer client.Close()
	point := Data{
		Measurement: "link",
		Type:        "click",
		Platform:    platform,
		Value:       1,
		Time:        time.Now(),
	}
	points := []any{&point}
	return client.WriteData(
		ctx,
		points,
		influxdb3.WithDatabase(a.database),
	)
}

func (a *InfluxAnalytics) LogInstall(meta string, platform string, ctx context.Context) error {
	client, err := influxdb3.NewFromEnv()
	if err != nil {
		return err
	}
	defer client.Close()
	point := Data{
		Measurement: "link",
		Type:        "install",
		Platform:    platform,
		Value:       1,
		Time:        time.Now(),
	}
	points := []any{&point}
	return client.WriteData(
		ctx,
		points,
		influxdb3.WithDatabase(a.database),
	)
}
