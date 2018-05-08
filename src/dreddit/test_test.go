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
