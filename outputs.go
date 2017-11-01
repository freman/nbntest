package nbntest

import (
	"fmt"
	"time"
)

type Outputter interface {
	Init(c *Configuration)
	RecordError(t time.Time, err error)
	RecordModem(t time.Time, stat ModemStatistics)
	RecordSpeedtest(t time.Time, id int, latency time.Duration, up, down float64)
}

type Outputs []Outputter

func (o Outputs) RecordError(t time.Time, err error) {
	for _, n := range o {
		n.RecordError(t, err)
	}
}

func (o Outputs) RecordModem(t time.Time, stat ModemStatistics) {
	for _, n := range o {
		n.RecordModem(t, stat)
	}
}

func (o Outputs) RecordSpeedtest(t time.Time, id int, latency time.Duration, up, down float64) {
	for _, n := range o {
		n.RecordSpeedtest(t, id, latency, up, down)
	}
}

func (o *Outputs) Init(c *Configuration) {
	for n, f := range outputterInterfaces {
		var test struct {
			Enabled bool `toml:"enabled"`
		}

		if c.UnifyOutputConfiguration(n, &test); test.Enabled {
			fmt.Println("it's enabled mang")
			i := f()
			i.Init(c)
			*o = append(*o, i)
		}
	}
}

var outputterInterfaces = map[string]func() Outputter{}

func RegisterOutputter(name string, f func() Outputter) {
	outputterInterfaces[name] = f
}
