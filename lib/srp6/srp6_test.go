package srp6_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/freman/nbntest/lib/srp6"
)

func TestClient(t *testing.T) {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i % 256)
	}
	c, err := srp6.NewClient("fred", "blogs", srp6.Reader(bytes.NewBuffer(b)))
	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	I, A := c.StartAuthentication()
	if I != "fred" {
		t.Errorf("Expected I to be 'fred', got '%s'", I)
	}
	if A != `47118709be0de1160a63b14b0524ae22270f78b0120855c88f285f2eb193e4727a79e25e136ecfc7cf3ecd0cdaaafefcf6a93c29dae8a76a9b9e6c044b3af05a47fe7122f579ac62bd7c8263b832139ae29deb5258b91a7f64fcf073e1800c3819e83161b3dcdbbd116680cbfeff2f25e06be14f654e0e274b71eb3dce95207367c319f02f0562a8fd26db4837da5be55c6e367a6642bdb2a82decab0c4e768671fbd8d924774729172f848e72913d0e842fb5abb25bdea4d43ac7502cfb3db6eb980c3f469475d707e5adf40137e74bc9d1ac072c764108ffba19dd917a018994fc967e840c4a6665a04933927589cb3e7761513d41b8a50e2314684364d465` {
		t.Error("Expected A to be something long and unreasonable to print, and it wasn't")
	}

	M, AMK := c.ProcessChallenge(fmt.Sprintf("%0x", b), fmt.Sprintf("%0x", b))
	if e := `1c1dcdcb50d5e318250f53c67d5ba067a08357208914d22963a9e38ea7df56e7`; M != e {
		t.Errorf(`Expected M to be '%s' got '%s'`, e, M)
	}
	if e := `020dfb5753b6f5e8c79d59b01bb16591c4d260c95304a281370737d6b6fcaaf4`; AMK != e {
		t.Errorf(`Expecting AMK to be '%s' got '%s'`, e, AMK)
	}
}
