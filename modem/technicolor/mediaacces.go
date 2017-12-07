package technicolor

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	gct "github.com/freman/go-commontypes"
	"github.com/freman/nbntest"

	"github.com/freman/nbntest/lib/srp6"
)

type modemConfiguration struct {
	IP       gct.IP       `toml:"ip"`
	Timeout  gct.Duration `toml:"timeout"`
	Username string       `toml:"username"`
	Password string       `toml:"password"`
}

type MediaAccess struct {
	httpClient *http.Client
	csrf       string
	config     modemConfiguration
}

var reDuration = regexp.MustCompile(`(\d+)(days|hours|min|sec)`)

func (m *MediaAccess) Init(n string, c *nbntest.Configuration) (err error) {
	jar, _ := cookiejar.New(nil)

	m.config = modemConfiguration{
		IP:       gct.IP{IP: net.IP{0xc0, 0xa8, 0x0a, 0x01}},
		Timeout:  gct.Duration{Duration: 10 * time.Second},
		Username: "admin",
		Password: "admin",
	}

	if err = c.UnifyModemConfiguration(n, &m.config); err != nil {
		return
	}

	m.httpClient = &http.Client{
		Jar:     jar,
		Timeout: m.config.Timeout.Duration,
	}

	return nil
}

func (m *MediaAccess) Login() error {
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
		return errors.New("unable to verify - please check your credentials")
	}

	return nil
}

func asDuration(s string) (t time.Duration, err error) {
	for _, s := range strings.Fields(s) {
		if v := reDuration.FindStringSubmatch(s); len(v) == 3 {
			var (
				i   int
				mul time.Duration
			)

			i, err = strconv.Atoi(v[1])
			if err != nil {
				return
			}

			switch v[2] {
			case "days":
				mul = 24 * time.Hour
			case "hours":
				mul = time.Hour
			case "min":
				mul = time.Minute
			case "sec":
				mul = time.Second
			}
			t += time.Duration(i) * mul
		}
	}
	return
}

func asSpeeds(s string) (up float64, down float64, err error) {
	speeds := strings.Fields(s)
	up, err = strconv.ParseFloat(speeds[0], 64)
	if err != nil {
		return
	}
	if strings.EqualFold(speeds[1], "Mbps") {
		up *= 1000
	}

	down, err = strconv.ParseFloat(speeds[2], 64)
	if err != nil {
		return
	}
	if strings.EqualFold(speeds[3], "Mbps") {
		down *= 1000
	}

	return
}

func as2power(s string) (up float64, down float64, err error) {
	p := strings.Fields(s)
	up, err = strconv.ParseFloat(p[0], 64)
	if err != nil {
		return
	}
	down, err = strconv.ParseFloat(p[2], 64)
	return
}

func (m *MediaAccess) getCSRF() (string, error) {
	if m.csrf == "" {
		resp, err := m.httpClient.Get("http://" + m.config.IP.String() + "/login.lp?action=getcsrf")
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		csrfToken := make([]byte, 64)
		n, err := resp.Body.Read(csrfToken)

		if err != nil && err != io.EOF {
			return "", err
		}

		if n != 64 {
			return "", errors.New("n not equal to 64")
		}

		m.csrf = string(csrfToken)
	}

	return m.csrf, nil
}

func (m *MediaAccess) postAuthenticate(data map[string]string, v interface{}) error {
	csrf, err := m.getCSRF()
	if err != nil {
		return err
	}

	form := url.Values{
		"CSRFtoken": {csrf},
	}

	for n, v := range data {
		form.Add(n, v)
	}

	resp, err := m.httpClient.PostForm("http://"+m.config.IP.String()+"/authenticate", form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(&v)
}

func (m *MediaAccess) Gather(dslUptime, maximumLineRate, lineRate, outputPower, lineAtttenuation, noiseMargin, uptime string) (*nbntest.ModemStatistics, error) {
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

	stats.ShowtimeUptime, err = asDuration(doc.Find(dslUptime).Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.MaxRate, stats.Downstream.MaxRate, err = asSpeeds(doc.Find(maximumLineRate).Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.CurrRate, stats.Downstream.CurrRate, err = asSpeeds(doc.Find(lineRate).Text())
	if err != nil {
		return nil, err
	}

	stats.Upstream.Power = make([]float64, 1)
	stats.Downstream.Power = make([]float64, 1)
	stats.Upstream.Power[0], stats.Downstream.Power[0], err = as2power(doc.Find(outputPower).Text())
	if err != nil {
		return nil, err
	}

	upstream := true
	for _, v := range strings.Fields(doc.Find(lineAtttenuation).Text()) {
		v = strings.Trim(v, ", ")
		if strings.EqualFold(v, "dB") {
			if upstream {
				upstream = false
				continue
			}
			break
		}

		vf, _ := strconv.ParseFloat(v, 64)
		if upstream {
			stats.Upstream.Attenuation = append(stats.Upstream.Attenuation, vf)
			continue
		}

		stats.Downstream.Attenuation = append(stats.Downstream.Attenuation, vf)
	}

	stats.Upstream.NoiseMargin = make([]float64, 1)
	stats.Downstream.NoiseMargin = make([]float64, 1)
	stats.Upstream.NoiseMargin[0], stats.Downstream.NoiseMargin[0], err = as2power(doc.Find(noiseMargin).Text())
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

	stats.ModemUptime, err = asDuration(doc.Find(uptime).Text())
	if err != nil {
		return nil, err
	}

	return stats, err
}
