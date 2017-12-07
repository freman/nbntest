package tgiinet1

import (
	"github.com/freman/nbntest"
	"github.com/freman/nbntest/modem/technicolor"
)

const ModemName = `Technicolor-TGiiNet-1`

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
		`div:hasChild(label:contains('DSL Uptime')) span`,
		`div:hasChild(label:contains('Maximum Line rate')) span`,
		`div:hasChild(label:contains('Line Rate'):not(:contains("Maximum"))) span`,
		`div:hasChild(label:contains('Output Power')) span`,
		`div:hasChild(label:contains('Line Attenuation')) span`,
		`div:hasChild(label:contains('Noise Margin')) span`,
		`div:hasChild(label:contains('Uptime')) span`,
	)
}
