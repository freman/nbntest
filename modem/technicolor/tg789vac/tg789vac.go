package tg789vac

import (
	"errors"
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	gct "github.com/freman/go-commontypes"
	"github.com/freman/nbntest"
	"github.com/freman/nbntest/lib/srp6"
)

const ModemName = `Technicolor-TG789vac`

var reDuration = regexp.MustCompile(`(\d+)(days|hours|min|sec)`)

type modemConfiguration struct {
	IP       gct.IP       `toml:"ip"`
	Timeout  gct.Duration `toml:"timeout"`
	Username string       `toml:"username"`
	Password string       `toml:"password"`
}

type modem struct {
	httpClient *http.Client
	csrf       string
	config     modemConfiguration
}

func init() {
	nbntest.RegisterModems(ModemName, func() nbntest.Modem {
		jar, _ := cookiejar.New(nil)
		return &modem{
			httpClient: &http.Client{
				Jar:     jar,
				Timeout: time.Second * 10,
			},
		}
	})
}

func (m *modem) Init(c *nbntest.Configuration) (err error) {
	m.config = modemConfiguration{
		IP:       gct.IP{IP: net.IP{0xc0, 0xa8, 0x0a, 0x01}},
		Timeout:  gct.Duration{Duration: 10 * time.Second},
		Username: "admin",
		Password: "admin",
	}

	if err = c.UnifyModemConfiguration(ModemName, &m.config); err != nil {
		return
	}

	return nil
}

func (m *modem) Login() error {
	c, err := srp6.NewClient(m.config.Username, m.config.Password)
	if err != nil {
		return err
	}

	i, a := c.StartAuthentication()

	challenge := struct {
		Salt string `json:"s"`
		B    string
	}{}
	if err := m.postAuthenticate(map[string]string{"I": i, "A": a}, &challenge); err != nil {
		return err
	}
	ma, mc := c.ProcessChallenge(challenge.Salt, challenge.B)

	verify := struct {
		M string
	}{}
	if err := m.postAuthenticate(map[string]string{"M": ma}, &verify); err != nil {
		return err
	}

	if !strings.EqualFold(verify.M, mc) {
		return errors.New("unable to verify")
	}

	return nil
}

func (m *modem) Gather() (*nbntest.ModemStatistics, error) {
	if err := m.Login(); err != nil {
		return nil, err
	}

	res, err := m.httpClient.Get("http://" + m.config.IP.String() + "/modals/broadband-modal.lp")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	stats := &nbntest.ModemStatistics{}

	stats.ShowtimeUptime, err = asDuration(doc.Find("span[id='DSL Uptime']").Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.MaxRate, stats.Downstream.MaxRate, err = asSpeeds(doc.Find("span[id='Maximum Line rate']").Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.CurrRate, stats.Downstream.CurrRate, err = asSpeeds(doc.Find("span[id='Line Rate']").Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.Power = make([]float64, 1)
	stats.Downstream.Power = make([]float64, 1)
	stats.Upstream.Power[0], stats.Downstream.Power[0], err = as2power(doc.Find("span[id='Output Power']").Text())
	if err != nil {
		return nil, err
	}

	for i, v := range strings.Fields(doc.Find("span[id='Line Attenuation']").Text()) {
		if i == 4 || i == 8 {
			continue
		}

		vf, _ := strconv.ParseFloat(strings.Trim(v, ","), 64)
		if i < 4 {
			stats.Upstream.Attenuation = append(stats.Upstream.Attenuation, vf)
			continue
		}

		stats.Downstream.Attenuation = append(stats.Downstream.Attenuation, vf)
	}

	stats.Upstream.NoiseMargin = make([]float64, 1)
	stats.Downstream.NoiseMargin = make([]float64, 1)
	stats.Upstream.NoiseMargin[0], stats.Downstream.NoiseMargin[0], err = as2power(doc.Find("span[id='Noise Margin']").Text())
	if err != nil {
		return nil, err
	}

	res, err = m.httpClient.Get("http://" + m.config.IP.String() + "/modals/gateway-modal.lp")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	stats.ModemUptime, err = asDuration(doc.Find("#Uptime").Text())
	if err != nil {
		return nil, err
	}

	return stats, err
}
