package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/freman/nbntest"
	_ "github.com/freman/nbntest/modem"
	_ "github.com/freman/nbntest/output"
)

func envStr(name, def string) string {
	if tmp := os.Getenv(name); tmp != "" {
		return tmp
	}
	return def
}

func main() {
	configFile := flag.String("config", envStr("NBNTEST_CONFIG", "config.toml"), "Path to the configuration file (env:NBNTEST_CONFIG)")
	listModems := flag.Bool("modems", false, "List the modems supported")
	flag.Parse()

	if *listModems {
		for _, v := range nbntest.ListModems() {
			fmt.Printf("\t%s\n", v)
		}
		return
	}

	cfg, err := nbntest.LoadConfiguration(*configFile)
	if err != nil {
		panic(err)
	}

	st := &nbntest.NBNTest{
		Config: cfg,
	}

	st.Outputs.Init(cfg)

	spew.Dump(st.Outputs)

	st.Run()
}
