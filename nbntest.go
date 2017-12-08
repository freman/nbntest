package nbntest

import (
	"errors"
	"time"

	"github.com/freman/speedtest"
)

type NBNTest struct {
	Config      *Configuration
	Outputs     Outputs
	muModem     chan struct{}
	muSpeedTest chan struct{}
}

func (t *NBNTest) Run() error {
	t.muModem = make(chan struct{}, 1)
	t.muSpeedTest = make(chan struct{}, 1)

	modemTimer := time.NewTicker(t.Config.Modem.Interval.Duration)
	speedtestTimer := time.NewTicker(t.Config.SpeedTest.Interval.Duration)

	now := time.Now()
	if err := t.CollectFromModem(now); err != nil {
		return err
	}
	t.CollectFromSpeedtest(now)

	for {
		select {
		case now := <-modemTimer.C:
			t.CollectFromModem(now)
		case now := <-speedtestTimer.C:
			t.CollectFromSpeedtest(now)
		}
	}
}

func (t *NBNTest) CollectFromModem(now time.Time) error {
	modem := GetModem(t.Config.Modem.Interface)
	if modem == nil {
		return errors.New("Unable to load modem driver " + t.Config.Modem.Interface)
	}

	select {
	case t.muModem <- struct{}{}:
		go func() {
			defer func() { <-t.muModem }()
			err := modem.Init(t.Config)
			if err != nil {
				t.Outputs.RecordError(now, err)
				return
			}

			stats, err := modem.Gather()
			if err != nil {
				t.Outputs.RecordError(now, err)
				return
			}

			t.Outputs.RecordModem(now, *stats)
		}()
	default:
		t.Outputs.RecordError(now, errors.New("skipped modem test, previous test still running"))
	}

	return nil
}

func (t *NBNTest) CollectFromSpeedtest(now time.Time) {
	select {
	case t.muSpeedTest <- struct{}{}:
		go func() {
			defer func() { <-t.muSpeedTest }()

			st := speedtest.NewClient()
			sl, err := st.GetServerList()
			if err != nil {
				t.Outputs.RecordError(now, err)
				return
			}

			var s *speedtest.Server

			if t.Config.SpeedTest.Site > 0 {
				s, err = sl.ByID(t.Config.SpeedTest.Site)
				if err != nil {
					t.Outputs.RecordError(now, err)
					return
				}
			}

			if s == nil {
				n := t.Config.SpeedTest.Fastest
				if n == 0 {
					n = 5
				}

				l := sl.Fastest(n)
				s = &l[0]
			}

			latency := s.TestLatency()
			download := s.TestDownload()
			upload := s.TestUpload()

			t.Outputs.RecordSpeedtest(now, s.ID, latency, upload, download)
		}()
	default:
		t.Outputs.RecordError(now, errors.New("skipped speedtest, previous test still running"))
	}
}
