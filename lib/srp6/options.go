package srp6

import (
	cryptoRand "crypto/rand"
	"crypto/sha256"
	"hash"
	"io"
)

// Options for customising the SRP6 client
type Options struct {
	Hasher  func() hash.Hash
	Reader  io.Reader
	KeySize int
	K       string
}

type option func(o *Options)

func Hasher(h func() hash.Hash) option {
	return func(o *Options) {
		o.Hasher = h
	}
}

func Reader(r io.Reader) option {
	return func(o *Options) {
		o.Reader = r
	}
}

func KeySize(sz int) option {
	return func(o *Options) {
		o.KeySize = sz
	}
}

func K(k string) option {
	return func(o *Options) {
		o.K = k
	}
}

func newOptions(options []option) *Options {
	opts := &Options{
		Hasher:  sha256.New,
		Reader:  cryptoRand.Reader,
		KeySize: 2048,
		K:       `0x05b9e8ef059c6b32ea59fc1d322d37f04aa30bae5aa9003b8321e21ddb04e300`,
	}

	for _, o := range options {
		o(opts)
	}

	return opts
}
