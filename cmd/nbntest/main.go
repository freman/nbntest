package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/freman/nbntest"
	_ "github.com/freman/nbntest/modem"
	_ "github.com/freman/nbntest/output"
)

var (
	version = "Undefined"
	commit  = "Undefined"
)

func envStr(name, def string) string {
	if tmp := os.Getenv(name); tmp != "" {
		return tmp
	}
	return def
}

func main() {
	showVersion := flag.Bool("version", false, "Show the current version")
	configFile := flag.String("config", envStr("NBNTEST_CONFIG", "config.toml"), "Path to the configuration file (env:NBNTEST_CONFIG)")
	listModems := flag.Bool("modems", false, "List the modems supported")
	flag.Parse()

	if *showVersion {
		fmt.Printf("nbntest - %s (%s)\n", version, commit)
		fmt.Println("https://github.com/freman/nbntest")
		return
	}

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

	st.Run()
}
