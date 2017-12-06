package influxdb

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	gct "github.com/freman/go-commontypes"
	"github.com/freman/nbntest"
	"github.com/influxdata/influxdb/client/v2"
)

const OutputName = "influxdb"

type influxFields map[string]interface{}

type InfluxDB struct {
	// Influx parameters
	Address            gct.URL      `toml:"address"`
	Username           string       `toml:"username"`
	Password           string       `toml:"password"`
	Timeout            gct.Duration `toml:"timeout"`
	InsecureSkipVerify bool         `toml:"insecure"`

	// My parameters
	Database string `toml:"database"`

	client client.Client
}

func (f *influxFields) arrayStat(n string, s []float64) {
	l := len(s)
	if l == 0 {
		return
	}

	// Record the first as primary
	(*f)[n] = s[0]

	// Record all the stats
	for i, v := range s {
		(*f)[fmt.Sprintf("%s_%d", n, i+1)] = v
	}
}

func (o *InfluxDB) Init(c *nbntest.Configuration) {
	c.UnifyOutputConfiguration(OutputName, o)

	var err error
	o.client, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:               o.Address.String(),
		Username:           o.Username,
		Password:           o.Password,
		Timeout:            o.Timeout.Duration,
		InsecureSkipVerify: o.InsecureSkipVerify,
	})

	if err != nil {
		panic(err)
	}

	_, err = o.client.Query(client.NewQuery(fmt.Sprintf("CREATE DATABASE %s", o.Database), "", ""))
	if err != nil {
		panic(err)
	}
}

func (o *InfluxDB) write(name string, tags map[string]string, fields map[string]interface{}, t time.Time) {
	point, err := client.NewPoint(name, tags, fields, t)
	if err != nil {
		panic(err)
	}

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  o.Database,
		Precision: "s",
	})

	bp.AddPoint(point)
	err = o.client.Write(bp)
	if err != nil {
		panic(err)
	}
}

func (o *InfluxDB) RecordError(t time.Time, err error) {
	fields := map[string]interface{}{
		"error": err.Error(),
	}

	o.write("errors", nil, fields, t)
}

func (o *InfluxDB) RecordModem(t time.Time, stat nbntest.ModemStatistics) {
	fields := influxFields{
		"down_current_rate": stat.Downstream.CurrRate,
		"down_maximum_rate": stat.Downstream.MaxRate,

		"up_current_rate": stat.Upstream.CurrRate,
		"up_maximum_rate": stat.Upstream.MaxRate,

		"modem_uptime":    stat.ModemUptime,
		"showtime_uptime": stat.ShowtimeUptime,
	}

	fields.arrayStat("down_attenuation", stat.Downstream.Attenuation)
	fields.arrayStat("down_noise_margin", stat.Downstream.NoiseMargin)
	fields.arrayStat("down_power", stat.Downstream.Power)

	fields.arrayStat("up_attenuation", stat.Upstream.Attenuation)
	fields.arrayStat("up_noise_margin", stat.Upstream.NoiseMargin)
	fields.arrayStat("up_power", stat.Upstream.Power)

	o.write("modem", nil, fields, t)
}

func (o *InfluxDB) RecordSpeedtest(t time.Time, id int, latency time.Duration, up, down float64) {
	tags := map[string]string{
		"site": strconv.Itoa(id),
	}
	fields := map[string]interface{}{
		"latency":  latency.Nanoseconds(),
		"upload":   up,
		"download": down,
	}

	o.write("speedtest", tags, fields, t)
}

func init() {
	nbntest.RegisterOutputter(OutputName, func() nbntest.Outputter {
		return &InfluxDB{
			Address: gct.URL{
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:8086",
				},
			},
			Timeout:  gct.Duration{Duration: 5 * time.Second},
			Database: "nbntest",
		}
	})
}
