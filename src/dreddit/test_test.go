package dreddit

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestSignMessage(t *testing.T) {
	fmt.Println("\nStarting TestSignMessage...")

	sv := MakeServer(nil, -1, nil)
 	p := Post{Username: "ezfn", Title: "Test post", Body: "test post please ignore"}
	fmt.Println("Input post:", p)

	sp := sv.signPost(p)
	dp, _ := verifyPost(sp, sp.Seed)
	fmt.Println("Output post:", dp)
}

func TestNetworkSimple(t *testing.T) {
	fmt.Println("\nStarting TestNetworkSimple...")
	
	cfg := make_config(2)
	defer cfg.cleanup()

 	p := Post{Username: "ezfn", Title: "Test post", Body: "test post please ignore"}
	fmt.Println("Input post:", p)
	hash :=	cfg.servers[0].NewPost(p).Seed
	
	time.Sleep(100 * time.Millisecond)
	
	op, _ := cfg.servers[1].GetPost(hash)
	dp, _ := verifyPost(op, hash)
	fmt.Println("Output post:", dp)
}

func TestNetworkConcurrentNewPosts(t *testing.T) {
	fmt.Println("\nStarting TestNetworkConcurrentNewPosts...")

	n := 25
	cfg := make_config(n)
	defer cfg.cleanup()
	hashes := make([]HashTriple, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			p := Post{Username: "ezfn", Title: "Test post",
				Body: fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.servers[i].NewPost(p).Seed
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			fmt.Printf("Server %d looking for post from %d\n", i, j)
			op, _ := cfg.servers[i].GetPost(hashes[j])
			p, ok := verifyPost(op, hashes[j])
			if ok {
				// fmt.Printf("Server %d has post from %d\n", i, j)
			} else {
				fmt.Printf("Server %d missing post from %d, post received %v\n", i, j, p)
				t.Fail()
			}
		}
	}
}

func TestNetworkDisconnect(t *testing.T) {
	// This test only makes sure posts are reachable. It does not check seeds.

	fmt.Println("\nStarting TestNetworkDisconnect...")

	n := 25
	s := 15
	cfg := make_config(n)
	defer cfg.cleanup()
	hashes := make([]HashTriple, n)

	// Disconnect the last n-s nodes.
	for i := s; i < n; i++ {
		cfg.disconnect(i)
	}

	for i := 0; i < s; i++ {
		go func(i int) {
			p := Post{Username: "ezfn", Title: "Test post",
				Body: fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.servers[i].NewPost(p).Seed
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	// Reconnect the last n-s nodes.
	for i := s; i < n; i++ {
		cfg.connect(i)
	}
	

	for i := 0; i < n; i++ {
		for j := 0; j < s; j++ {
			fmt.Printf("Server %d looking for post from %d\n", i, j)
			op, _ := cfg.servers[i].GetPost(hashes[j])
			p, ok := verifyPost(op, hashes[j])
			if ok {
				// fmt.Printf("Server %d has post from %d\n", i, j)
			} else {
				fmt.Printf("Server %d missing post from %d, post received %v\n", i, j, p)
				t.Fail()
			}
		}
	}
}

func TestDeletePosts(t *testing.T) {
	// This test empties the post cache on some servers, then checks for reachability.
	
	fmt.Println("\nStarting TestDeletePosts...")

	n := 25
	s := 10
	cfg := make_config(n)
	defer cfg.cleanup()
	hashes := make([]HashTriple, n)

	for i := 0; i < s; i++ {
		go func(i int) {
			p := Post{Username: "ezfn", Title: "Test post",
				Body: fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.servers[i].NewPost(p).Seed
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	// On s random nodes, delete all posts except those authored by the node.
	for i := s; i < n; i++ {
		r := rand.Intn(n)
		rp := cfg.servers[r].Posts[hashes[r]]
		cfg.servers[r].Posts = make(map[HashTriple]SignedPost)
		cfg.servers[r].Posts[hashes[r]] = rp
	}

	for i := 0; i < n; i++ {
		for j := 0; j < s; j++ {
			fmt.Printf("Server %d looking for post from %d\n", i, j)
			op, _ := cfg.servers[i].GetPost(hashes[j])
			p, ok := verifyPost(op, hashes[j])
			if ok {
				// fmt.Printf("Server %d has post from %d\n", i, j)
			} else {
				fmt.Printf("Server %d missing post from %d, post received %v\n", i, j, p)
				t.Fail()
			}
		}
	}
}
