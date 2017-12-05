package nbntest

import (
	"time"
)

var modemInterfaces = map[string]func() Modem{}

type Modem interface {
	Init(*Configuration) error
	Gather() (*ModemStatistics, error)
}

func RegisterModems(name string, f func() Modem) {
	modemInterfaces[name] = f
}

func GetModem(name string) (modemInterface Modem) {
	if modemInterfacef, ok := modemInterfaces[name]; ok {
		modemInterface = modemInterfacef()
	}
	return
}

func ListModems() []string {
	res := make([]string, len(modemInterfaces))
	var c int
	for n := range modemInterfaces {
		res[c] = n
		c++
	}
	return res
}

type ModemStatistics struct {
	Upstream       StreamStatistics
	Downstream     StreamStatistics
	ModemUptime    time.Duration
	ShowtimeUptime time.Duration
}

type StreamStatistics struct {
	CurrRate    float64
	MaxRate     float64
	NoiseMargin float64
	Attenuation float64
	Power       float64
}
