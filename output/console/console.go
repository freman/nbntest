package console

import (
	"fmt"
	"os"
	"time"

	"github.com/freman/nbntest"
)

const OutputName = "console"

type Console struct {
}

func (o *Console) Init(c *nbntest.Configuration) {
}

func (o *Console) RecordError(t time.Time, err error) {
	fmt.Fprintf(os.Stderr, "[%v] ERR %v\n", t, err)
}

func (o *Console) RecordModem(t time.Time, stat nbntest.ModemStatistics) {
	fmt.Printf("[%v] MDM %v\n", t, stat)
}

func (o *Console) RecordSpeedtest(t time.Time, id int, latency time.Duration, up, down float64) {
	fmt.Printf("[%v] STN %v %v %v %v\n", t, id, latency, up, down)
}

func init() {
	nbntest.RegisterOutputter(OutputName, func() nbntest.Outputter {
		return &Console{}
	})
}
