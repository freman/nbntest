package telnet

import (
	"errors"
	"time"

	ztelnet "github.com/ziutek/telnet"
)

func Expect(t *ztelnet.Conn, timeout time.Duration, d ...string) error {
	if err := t.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	if err := t.SkipUntil(d...); err != nil {
		return err
	}
	return nil
}

func Sendln(t *ztelnet.Conn, timeout time.Duration, s string) error {
	if err := t.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	if _, err := t.Write([]byte(s + "\n")); err != nil {
		return err
	}
	return nil
}

func SkipBytes(t *ztelnet.Conn, timeout time.Duration, n int) error {
	if err := t.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return err
	}
	for i := 0; i < n; i++ {
		if _, err := t.ReadByte(); err != nil {
			return err
		}
	}
	return nil
}

func Reply(t *ztelnet.Conn, timeout time.Duration, expect, reply string) error {
	if err := Expect(t, timeout, expect); err != nil {
		return err
	}
	if err := Sendln(t, timeout, reply); err != nil {
		return err
	}
	return nil
}

func Chat(t *ztelnet.Conn, timeout time.Duration, chatter ...string) error {
	if len(chatter)%2 != 0 {
		return errors.New("chatter should be in the form of expect reply")
	}
	for i := 0; i < len(chatter); i += 2 {
		if err := Reply(t, timeout, chatter[i], chatter[i+1]); err != nil {
			return err
		}
	}
	return nil
}
