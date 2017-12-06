package td9970

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/freman/nbntest"
	"github.com/freman/nbntest/modem/generic/telnet"

	gct "github.com/freman/go-commontypes"
	ztelnet "github.com/ziutek/telnet"
)

// ModemName is the name of this modem module
const ModemName = `TPLink-TD-9970`

type modem struct {
	config modemConfiguration
}

type modemConfiguration struct {
	IP       gct.IP       `toml:"ip"`
	Port     int          `toml:"port"`
	Timeout  gct.Duration `toml:"timeout"`
	Username string       `toml:"username"`
	Password string       `toml:"password"`
}

func init() {
	nbntest.RegisterModems(ModemName, func() nbntest.Modem {
		return &modem{}
	})
}

func (m *modem) Init(c *nbntest.Configuration) (err error) {
	m.config = modemConfiguration{
		IP:       gct.IP{IP: net.IP{0xc0, 0xa8, 0x0a, 0x01}},
		Port:     23,
		Timeout:  gct.Duration{Duration: 10 * time.Second},
		Username: "admin",
		Password: "admin",
	}

	if err = c.UnifyModemConfiguration(ModemName, &m.config); err != nil {
		return
	}

	return nil
}

func setStreamStat(stat *nbntest.StreamStatistics, l, stream string) (r bool) {
	if r = strings.HasPrefix(l, stream); r {
		l = strings.TrimPrefix(l, stream)
		split := strings.Split(l, "=")
		v, _ := strconv.ParseFloat(split[1], 64)
		switch split[0] {
		case "CurrRate":
			stat.CurrRate = v
		case "MaxRate":
			stat.MaxRate = v
		case "NoiseMargin":
			stat.NoiseMargin = append(stat.NoiseMargin, v)
		case "Attenuation":
			stat.Attenuation = append(stat.Attenuation, v)
		case "Power":
			stat.Power = append(stat.Power, v)
		}
	}
	return
}

func setDurationStat(stat *time.Duration, l, name string) (r bool) {
	if r = strings.HasPrefix(l, name); r {
		d, _ := strconv.Atoi(strings.TrimPrefix(l, name+"="))
		*stat = time.Duration(d) * time.Second
	}
	return
}

func (m *modem) Gather() (*nbntest.ModemStatistics, error) {
	t, err := ztelnet.Dial("tcp", fmt.Sprintf("%s:%d", m.config.IP.String(), m.config.Port))
	if err != nil {
		return nil, err
	}
	defer t.Close()

	t.SetUnixWriteMode(true)
	t.SetEcho(false)

	//TODO: handle 'Login incorrect. Try again.'

	err = telnet.Chat(t, m.config.Timeout.Duration,
		"username:", m.config.Username,
		"password:", m.config.Password,
		"#", "adsl show info",
	)
	if err != nil {
		return nil, err
	}

	if err = telnet.Expect(t, m.config.Timeout.Duration, "{"); err != nil {
		return nil, err
	}

	if err = telnet.SkipBytes(t, m.config.Timeout.Duration, 2); err != nil {
		return nil, err
	}

	t.SetReadDeadline(time.Now().Add(m.config.Timeout.Duration))
	data, err := t.ReadBytes('}')
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data[:len(data)-2])

	stats := nbntest.ModemStatistics{}

	s := bufio.NewScanner(buf)
	for s.Scan() {
		l := s.Text()

		if setStreamStat(&stats.Upstream, l, "upstream") {
			continue
		}
		if setStreamStat(&stats.Downstream, l, "downstream") {
			continue
		}
		if setDurationStat(&stats.ModemUptime, l, "totalStart") {
			continue
		}
		setDurationStat(&stats.ShowtimeUptime, l, "showtimeStart")

	}

	return &stats, nil
}
