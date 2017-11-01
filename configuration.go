package nbntest

import (
	"github.com/BurntSushi/toml"
	gct "github.com/freman/go-commontypes"
)

type SpeedtestConfiguration struct {
	Interval gct.Duration `toml:"interval"`
	Site     int          `toml:"site"`
	Fastest  int          `toml:"fastest"`
}

type ModemConfiguration struct {
	Interval      gct.Duration              `toml:"interval"`
	Interface     string                    `toml:"type"`
	Configuration map[string]toml.Primitive `toml:"config"`
}

type Configuration struct {
	md toml.MetaData `toml:"-"`

	Modem     *ModemConfiguration       `toml:"modem"`
	SpeedTest *SpeedtestConfiguration   `toml:"speedtest"`
	Output    map[string]toml.Primitive `toml:"output"`
}

func LoadConfiguration(file string) (*Configuration, error) {
	var err error
	config := Configuration{}
	config.md, err = toml.DecodeFile(file, &config)
	return &config, err
}

func (c *Configuration) UnifyModemConfiguration(name string, v interface{}) (err error) {
	if c.md.IsDefined("modem", "config", name) {
		err = c.md.PrimitiveDecode(c.Modem.Configuration[name], v)
	}
	return
}

func (c *Configuration) UnifyOutputConfiguration(name string, v interface{}) (err error) {
	if c.md.IsDefined("output", name) {
		err = c.md.PrimitiveDecode(c.Output[name], v)
	}
	return
}
