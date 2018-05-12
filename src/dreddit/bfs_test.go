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

func TestBFSNetworkSimple(t *testing.T) {
	fmt.Println("\nStarting TestBFSNetworkSimple...")
	
	cfg := Make_config(2, nil)
	defer cfg.cleanup()

 	p := Post{Username: "ezfn", Title: "Test post", Body: "test post please ignore"}
	fmt.Println("Input post:", p)
	hash :=	cfg.Servers[0].NewPost(p).Seed
	
	time.Sleep(100 * time.Millisecond)
	
	op, _ := cfg.Servers[1].GetPost(hash)
	dp, _ := verifyPost(op, hash)
	fmt.Println("Output post:", dp)
}

func TestBFSNetworkConcurrentNewPosts(t *testing.T) {
	fmt.Println("\nStarting TestBFSNetworkConcurrentNewPosts...")

	n := 25
	cfg := Make_config(n, nil)
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
			op, _ := cfg.Servers[i].GetPost(hashes[j])
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

func TestBFSNetworkDisconnect(t *testing.T) {
	// This test only makes sure posts are reachable. It does not check seeds.

	fmt.Println("\nStarting TestBFSNetworkDisconnect...")

	n := 25
	s := 15
	cfg := Make_config(n, nil)
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
			hashes[i] = cfg.Servers[i].NewPost(p).Seed
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
			op, _ := cfg.Servers[i].GetPost(hashes[j])
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

func TestBFSDeletePosts(t *testing.T) {
	// This test empties the post cache on some servers, then checks for reachability.
	
	fmt.Println("\nStarting TestBFSDeletePosts...")

	n := 25
	s := 10
	cfg := Make_config(n, nil)
	defer cfg.cleanup()
	hashes := make([]HashTriple, n)

	for i := 0; i < s; i++ {
		go func(i int) {
			p := Post{Username: "ezfn", Title: "Test post",
				Body: fmt.Sprintf("test post from %d", i)}
			hashes[i] = cfg.Servers[i].NewPost(p).Seed
		}(i)
	}

	fmt.Println("Sends started")

	time.Sleep(5 * time.Second)

	// On s random nodes, delete all posts except those authored by the node.
	for i := s; i < n; i++ {
		r := rand.Intn(n)
		rp := cfg.Servers[r].Posts[hashes[r]]
		cfg.Servers[r].Posts = make(map[HashTriple]SignedPost)
		cfg.Servers[r].Posts[hashes[r]] = rp
	}

	for i := 0; i < n; i++ {
		for j := 0; j < s; j++ {
			fmt.Printf("Server %d looking for post from %d\n", i, j)
			op, _ := cfg.Servers[i].GetPost(hashes[j])
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
