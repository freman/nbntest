package tg789vac

import (
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
)

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

func (m *modem) getCSRF() (string, error) {
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

func (m *modem) postAuthenticate(data map[string]string, v interface{}) error {
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
