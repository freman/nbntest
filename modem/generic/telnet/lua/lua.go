package lua

import (
	"context"
	"errors"
	"fmt"
	"time"

	gct "github.com/freman/go-commontypes"
	"github.com/freman/nbntest"
	"github.com/freman/nbntest/modem/generic/telnet"
	glua "github.com/yuin/gopher-lua"

	ztelnet "github.com/ziutek/telnet"
)

// ModemName is the name of this modem module
const ModemName = `Telnet-LUA`

type modem struct {
	IP      gct.IP       `toml:"ip"`
	Port    int          `toml:"port"`
	Timeout gct.Duration `toml:"timeout"`
	Script  string       `toml:"script"`
}

func init() {
	nbntest.RegisterModems(ModemName, func() nbntest.Modem {
		return &modem{}
	})
}

func (m *modem) Init(c *nbntest.Configuration) (err error) {
	if err = c.UnifyModemConfiguration(ModemName, &m); err != nil {
		return
	}

	return nil
}

func (m *modem) Gather() (*nbntest.ModemStatistics, error) {
	t, err := ztelnet.Dial("tcp", fmt.Sprintf("%s:%d", m.IP.String(), m.Port))
	if err != nil {
		return nil, err
	}
	defer t.Close()

	t.SetUnixWriteMode(true)
	t.SetEcho(false)

	L := glua.NewState()
	defer L.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	L.SetContext(ctx)

	L.SetGlobal("telnet_chat", L.NewFunction(func(L *glua.LState) int {
		args := make([]string, L.GetTop())
		for i := range args {
			args[i] = L.ToString(i + 1)
		}
		err := telnet.Chat(t, m.Timeout.Duration, args...)

		if err == nil {
			L.Push(glua.LNil)
		} else {
			L.Push(glua.LString(err.Error()))
		}
		return 1
	}))

	L.SetGlobal("telnet_expect", L.NewFunction(func(L *glua.LState) int {
		args := make([]string, L.GetTop())
		for i := range args {
			args[i] = L.ToString(i + 1)
		}
		err := telnet.Expect(t, m.Timeout.Duration, args...)

		if err == nil {
			L.Push(glua.LNil)
		} else {
			L.Push(glua.LString(err.Error()))
		}
		return 1
	}))

	L.SetGlobal("telnet_skipbytes", L.NewFunction(func(L *glua.LState) int {
		b := L.ToInt(1)
		err := telnet.SkipBytes(t, m.Timeout.Duration, b)

		if err == nil {
			L.Push(glua.LNil)
		} else {
			L.Push(glua.LString(err.Error()))
		}
		return 1
	}))

	L.SetGlobal("telnet_sendln", L.NewFunction(func(L *glua.LState) int {
		s := L.ToString(1)
		err := telnet.Sendln(t, m.Timeout.Duration, s)

		if err == nil {
			L.Push(glua.LNil)
		} else {
			L.Push(glua.LString(err.Error()))
		}
		return 1
	}))

	L.SetGlobal("telnet_reply", L.NewFunction(func(L *glua.LState) int {
		e := L.ToString(1)
		r := L.ToString(2)
		err := telnet.Reply(t, m.Timeout.Duration, e, r)

		if err == nil {
			L.Push(glua.LNil)
		} else {
			L.Push(glua.LString(err.Error()))
		}
		return 1
	}))

	L.SetGlobal("telnet_readuntil", L.NewFunction(func(L *glua.LState) int {
		e := L.ToString(1)

		t.SetReadDeadline(time.Now().Add(m.Timeout.Duration))

		data, err := t.ReadBytes(e[0])
		if err != nil {
			L.Push(glua.LFalse)
			L.Push(glua.LString(err.Error()))
		} else {
			L.Push(glua.LTrue)
			L.Push(glua.LString(string(data)))
		}
		return 2
	}))

	stats := &nbntest.ModemStatistics{}

	L.SetGlobal("record_direction_stat", L.NewFunction(func(L *glua.LState) int {
		e := func(err error) int {
			L.Push(glua.LString(err.Error()))
			return 1
		}

		d := L.ToString(1)
		s := L.ToString(2)
		v := float64(L.ToNumber(3))

		var stat *nbntest.StreamStatistics
		switch d {
		case "up":
			stat = &stats.Upstream
		case "down":
			stat = &stats.Downstream
		default:
			return e(errors.New("unknown direction"))
		}

		switch s {
		case "currrate":
			stat.CurrRate = v
		case "maxrate":
			stat.MaxRate = v
		case "noisemargin":
			stat.NoiseMargin = v
		case "attenuation":
			stat.Attenuation = v
		case "power":
			stat.Power = v
		default:
			return e(errors.New("unknown statistic"))
		}

		return 0
	}))

	L.SetGlobal("record_duration_stat", L.NewFunction(func(L *glua.LState) int {
		e := func(err error) int {
			L.Push(glua.LString(err.Error()))
			return 1
		}

		s := L.ToString(1)
		v := float64(L.ToNumber(2))

		var stat *time.Duration
		switch s {
		case "modemuptime":
			stat = &stats.ModemUptime
		case "showtimeuptime":
			stat = &stats.ShowtimeUptime
		default:
			return e(errors.New("unknown statistic"))
		}

		*stat = time.Duration(v) * time.Second

		return 0
	}))

	err = L.DoString(m.Script)
	if err != nil {
		return nil, err
	}

	return stats, err
}
