package dreddit

import (
	crand "crypto/rand"
	"encoding/base64"
	"labrpc"
)

/* 
  Most of this file is adapted from 6.824 Lab 2 (Raft)'s config.go.
*/

func randstring(n int) string {
	b := make([]byte, 2*n)
	crand.Read(b)
	s := base64.URLEncoding.EncodeToString(b)
	return s[0:n]
}
	
type Config struct {
	n         int
	net       *labrpc.Network
	Servers   []*Server
	connected []bool
	endnames  [][]string
	
}

func Make_config(n int, b Backend, o interface{}) *Config {
	cfg := &Config{}
	cfg.n = n
	cfg.net = labrpc.MakeNetwork()
	cfg.Servers = make([]*Server, cfg.n)
	cfg.connected = make([]bool, cfg.n)
	cfg.endnames = make([][]string, cfg.n)

	// create a full set of servers.
	for i := 0; i < cfg.n; i++ {
		// o should either be nil, or a slice of size n of options.
		if o == nil {
			cfg.start1(i, b, nil)
		} else {
			// Note: should rework this to allow options for other backends.
			cfg.start1(i, b, o.([]dshOptions)[i])
		}
	}

	// connect everyone.
	for i := 0; i < cfg.n; i++ {
		cfg.connect(i)
	}

	return cfg
}

func (cfg *Config) crash1(i int) {
	cfg.disconnect(i)
	cfg.net.DeleteServer(i)
}

func (cfg *Config) start1(i int, b Backend, o interface{}) {
	cfg.crash1(i)

	// a fresh set of outgoing ClientEnd names.
	// so that old crashed instance's ClientEnds can't send.
	cfg.endnames[i] = make([]string, cfg.n)
	for j := 0; j < cfg.n; j++ {
		cfg.endnames[i][j] = randstring(20)
	}

	// a fresh set of ClientEnds.
	ends := make([]*labrpc.ClientEnd, cfg.n)
	for j := 0; j < cfg.n; j++ {
		ends[j] = cfg.net.MakeEnd(cfg.endnames[i][j])
		cfg.net.Connect(cfg.endnames[i][j], j)
	}

	sv := MakeServer(ends, i, b, o)
	cfg.Servers[i] = sv

	svc := labrpc.MakeService(sv.net)
	srv := labrpc.MakeServer()
	srv.AddService(svc)
	cfg.net.AddServer(i, srv)
}

func (cfg *Config) connect(i int) {
	// fmt.Printf("connect(%d)\n", i)

	cfg.connected[i] = true

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, true)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.connected[j] {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, true)
		}
	}
}

func (cfg *Config) disconnect(i int) {
	// fmt.Printf("disconnect(%d)\n", i)

	cfg.connected[i] = false

	// outgoing ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[i] != nil {
			endname := cfg.endnames[i][j]
			cfg.net.Enable(endname, false)
		}
	}

	// incoming ClientEnds
	for j := 0; j < cfg.n; j++ {
		if cfg.endnames[j] != nil {
			endname := cfg.endnames[j][i]
			cfg.net.Enable(endname, false)
		}
	}
}

func (cfg *Config) cleanup() {
	cfg.net.Cleanup()
}


