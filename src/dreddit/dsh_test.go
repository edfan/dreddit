package dreddit

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func mod(d, m int) int {
	var res int = d % m
	if ((res < 0 && m > 0) || (res > 0 && m < 0)) {
		return res + m
	}
	return res
}

func TestDSH(t *testing.T) {
	fmt.Println("Starting TestDSH...")

	n := 64	
	perLayer := 16
	layers := n / perLayer
	var dshConfig []dshOptions
	var r int

	for i := 0; i < n; i++ {
		var o dshOptions
		layer := i / perLayer

		// Choose random peers.
		o.initialPeers = make(map[int]int)
		for j := 0; j < 4; j++ {
			r = rand.Intn(n)
			for r == i {
				r = rand.Intn(n)
			}
			//o.initialPeers = append(o.initialPeers, r)
			o.initialPeers[r] = 0
		}

		// Choose random storage peers (one per layer).
		o.initialStorage = make(map[int]int)
		for j := 0; j < layers; j++ {
			r := rand.Intn(perLayer) + j * perLayer
			for r == i {
				r = rand.Intn(perLayer) + j * perLayer
			}
			//o.initialStorage = append(o.initialStorage, r)
			o.initialStorage[j] = r
		}

		o.isStorage = true
		o.M = layer

		// Choose random storage peer from layer below.
		//r := rand.Intn(perLayer) + perLayer * mod(layer - 1, layers)
		//o.initialStoragePeerBelow = append(o.initialStoragePeerBelow, r)

		// Choose random storage peer from same layer.
		o.initialStoragePeerSame = make(map[int]int)
		counter := 0
		for counter < perLayer/2{
			r = rand.Intn(perLayer) + perLayer * layer
			_, ok := o.initialStoragePeerAbove[r]
			if !ok && r != i{
				o.initialStoragePeerSame[r] = 0
				counter++
			}
		}
		//o.initialStoragePeerSame = append(o.initialStoragePeerSame, r)

		// Choose random storage peer from layer below.
		o.initialStoragePeerAbove = make(map[int]int)
		//r = rand.Intn(perLayer) + perLayer * mod(layer + 1, layers)
		counter = 0
		for counter < perLayer/2{
			r = rand.Intn(perLayer) + perLayer * mod(layer + 1, layers)
			_, ok := o.initialStoragePeerAbove[r]
			if !ok{
				o.initialStoragePeerAbove[r] = 0
				counter++
			}
		}
		//o.initialStoragePeerAbove = append(o.initialStoragePeerAbove, r)
		//o.initialStoragePeerAbove[r] = 0 

		dshConfig = append(dshConfig, o)
	}

	cfg := Make_config(n, dshConfig)
	defer cfg.cleanup()
	hashes := make([]HashTriple, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			p := Post{Username: "ezfn", Title: "Test post",
				Body: fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.Servers[i].NewPost(p).Seed
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			fmt.Printf("Server %d looking for post from %d\n", i, j)
			op, found := cfg.Servers[i].GetPost(hashes[j])
			for !found{
				op, found = cfg.Servers[i].GetPost(hashes[j])
			}
			p, ok := verifyPost(op, hashes[j])
			if ok {
//				fmt.Printf("Server %d has post from %d\n", i, j)
			} else {
				fmt.Printf("Server %d missing post from %d, post received %v\n", i, j, p)
				t.Fail()
			}
		}
	}
}


