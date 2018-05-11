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
	
type config struct {
	n         int
	net       *labrpc.Network
	servers   []*Server
	connected []bool
	endnames  [][]string
	
}

func make_config(n int) *config {
	cfg := &config{}
	cfg.n = n
	cfg.net = labrpc.MakeNetwork()
	cfg.servers = make([]*Server, cfg.n)
	cfg.connected = make([]bool, cfg.n)
	cfg.endnames = make([][]string, cfg.n)

	// create a full set of servers.
	for i := 0; i < cfg.n; i++ {
		cfg.start1(i)
	}

	// connect everyone.
	for i := 0; i < cfg.n; i++ {
		cfg.connect(i)
	}

	return cfg
}

func (cfg *config) crash1(i int) {
	cfg.disconnect(i)
	cfg.net.DeleteServer(i)
}

func (cfg *config) start1(i int) {
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

	sv := MakeServer(ends, i, nil)
	cfg.servers[i] = sv

	svc := labrpc.MakeService(sv.net)
	srv := labrpc.MakeServer()
	srv.AddService(svc)
	cfg.net.AddServer(i, srv)
}

func (cfg *config) connect(i int) {
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

func (cfg *config) disconnect(i int) {
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

func (cfg *config) cleanup() {
	cfg.net.Cleanup()
}


