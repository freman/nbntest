package tg789vac

import (
	"github.com/freman/nbntest"
	"github.com/freman/nbntest/modem/technicolor"
)

const ModemName = `Technicolor-TG789vac`

type modem struct {
	technicolor.MediaAccess
}

func init() {
	nbntest.RegisterModems(ModemName, func() nbntest.Modem {
		return &modem{}
	})
}

func (m *modem) Init(c *nbntest.Configuration) (err error) {
	return m.MediaAccess.Init(ModemName, c)
}

func (m *modem) Gather() (*nbntest.ModemStatistics, error) {
	return m.MediaAccess.Gather(
		"span[id='DSL Uptime']",
		"span[id='Maximum Line rate']",
		"span[id='Line Rate']",
		"span[id='Output Power']",
		"span[id='Line Attenuation']",
		"span[id='Noise Margin']",
		"#Uptime",
	)
}
