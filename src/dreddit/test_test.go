package dreddit

import (
	"fmt"
	"testing"
	"time"
)

func TestSignMessage(t *testing.T) {
	fmt.Println("\nStarting TestSignMessage...")

	sv := Make(nil, -1)
	p := Post{"ezfn", "Test post", "test post please ignore"}
	fmt.Println("Input post:", p)

	sp := sv.signPost(p)
	dp, _ := verifyPost(sp, sp.Hash)
	fmt.Println("Output post:", dp)
}

func TestNetworkSimple(t *testing.T) {
	fmt.Println("\nStarting TestNetworkSimple...")
	
	cfg := make_config(2)
	defer cfg.cleanup()

	p := Post{"ezfn", "Test post", "test post please ignore"}
	fmt.Println("Input post:", p)
	hash :=	cfg.servers[0].NewPost(p)
	
	time.Sleep(100 * time.Millisecond)
	
	op, _ := cfg.servers[1].GetPost(hash)
	dp, _ := verifyPost(op, hash)
	fmt.Println("Output post:", dp)
}

func TestNetworkConcurrentNewPosts(t *testing.T) {
	fmt.Println("\nStarting TestNetworkConcurrentNewPosts...")

	n := 100
	cfg := make_config(n)
	defer cfg.cleanup()
	hashes := make([][32]byte, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			p := Post{"ezfn", "Test post",
				fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.servers[i].NewPost(p)
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			op, _ := cfg.servers[i].GetPost(hashes[j])
			p, ok := verifyPost(op, hashes[j])
			if ok {
				// fmt.Printf("Server %d has post from %d\n", i, j)
			} else {
				fmt.Printf("Server %d missing post from %d, post received %v\n", i, j, p)
			}
		}
	}
}
