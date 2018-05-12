package dreddit

import (
	"fmt"
	"labrpc"
	"math/rand"
	"sync"
)

const NUM_PEERS = 6

func min(x, y int) int {
    if x < y {
        return x
    }
    return y
}

type BFSNetwork struct {
	sv    *Server
	
	net   []*labrpc.ClientEnd
	peers []int
	seeds sync.Map
}

const BFSKeepPercent = 0.2

func (n *BFSNetwork) keep(sp SignedPost) bool {
	return rand.Float32() < BFSKeepPercent
}

func (n *BFSNetwork) NewPost(sp SignedPost) bool {
	n.seeds.Store(sp.Seed, n.sv.me)

	fmt.Println("BFS NewPost called")
	
	for i := 0; i < len(n.peers); i++ {
		go n.makeReceivePost(n.peers[i], sp, n.sv.me)
	}

	return true
}

func (n *BFSNetwork) GetPostRecursive(hash HashTriple) (SignedPost, bool) {
	// Using seeds as a map to known server that had post at some point, recurse.

	// Avoid cycles.
	seen := make(map[int]bool)
	seen[-1] = true
	
	lastStored, ok := n.seeds.Load(hash)
	for ok {
		seen[lastStored.(int)] = true
		fmt.Printf("Server %d asking %d\n", n.sv.me, lastStored)
		args := BFSRequestPostArgs{Hash: hash}
		var reply BFSRequestPostReply
		
		status := n.sendRequestPost(lastStored.(int), &args, &reply)
		
		if status {
			if reply.Success {
				fmt.Printf("Server %d found post on %d\n", n.sv.me, lastStored)
				_, ok := verifyPost(reply.Sp, hash)
				if ok {
					return reply.Sp, true
				}
			} else {
				if seen[reply.Redirect] {
					fmt.Printf("Server %d got dead redirect from %d\n", n.sv.me, lastStored)
					n.seeds.Delete(hash)
				} else {
					fmt.Printf("Server %d got redirect %d from %d\n", n.sv.me, reply.Redirect, lastStored)
					n.seeds.Store(hash, reply.Redirect)
				}
			}
		} else {
			n.seeds.Delete(hash)
		}

		lastStored, ok = n.seeds.Load(hash)
	}

	return SignedPost{}, false
}

func (n *BFSNetwork) GetPost(hash HashTriple) (SignedPost, bool) {
	fmt.Printf("Server %d calling GetPost\n", n.sv.me)
	
	// Try saved peer first.
	sp, ok := n.GetPostRecursive(hash)
	if ok {
		return sp, ok
	}
	
	// Ask peers for help.
	fmt.Printf("Server %d asking peers\n", n.sv.me)
	for i := 0; i < len(n.peers); i++ {
		n.seeds.Store(hash, n.peers[i])

		sp, ok = n.GetPostRecursive(hash)
		if ok {
			return sp, ok
		}
	}

	// Post is apparently unreachable.
	return SignedPost{}, false
}

type BFSReceivePostArgs struct {
	Sp         SignedPost
	LastStored int
}

type BFSReceivePostReply struct {
	Success bool
	Stored  bool
}

func (n *BFSNetwork) ReceivePost(args *BFSReceivePostArgs, reply *BFSReceivePostReply) {
	_, ok := n.seeds.Load(args.Sp.Seed)

	if !ok {
		_, ok := verifyPost(args.Sp, args.Sp.Seed)
		if ok {
			n.seeds.Store(args.Sp.Seed, args.LastStored)
			
			if n.keep(args.Sp) {
				n.sv.mu.Lock()
				n.sv.Posts[args.Sp.Seed] = args.Sp
				n.sv.mu.Unlock()
				
				reply.Stored = true
				args.LastStored = n.sv.me
			}

			n.sv.PostsCh <- args.Sp
			
			reply.Success = true
		}

		// Send out post to peers.
		for i := 0; i < len(n.peers); i++ {
			go n.makeReceivePost(n.peers[i], args.Sp, args.LastStored)
		}
	}
}

func (n *BFSNetwork) sendReceivePost(server int, args *BFSReceivePostArgs, reply *BFSReceivePostReply) bool {
	ok := n.net[server].Call("BFSNetwork.ReceivePost", args, reply)
	return ok	
}

func (n *BFSNetwork) makeReceivePost(server int, sp SignedPost, lastStored int) {
	retry := 0

	for retry < 10 {
		args := BFSReceivePostArgs{Sp: sp, LastStored: lastStored}
		var reply BFSReceivePostReply
		
		status := n.sendReceivePost(server, &args, &reply)
		
		if status {
			if reply.Stored {
				n.seeds.Store(args.Sp.Seed, server)
			}
			return
		}
		retry++
	}
}

type BFSRequestPostArgs struct {
	Hash HashTriple
}

type BFSRequestPostReply struct {
	Sp       SignedPost
	Success  bool
	Redirect int
}

func (n *BFSNetwork) RequestPost(args *BFSRequestPostArgs, reply *BFSRequestPostReply) {
	// If we have the post, just send it.
	n.sv.mu.RLock()
	sp, ok := n.sv.Posts[args.Hash]
	n.sv.mu.RUnlock()
	
	if ok {
		reply.Sp = sp
		reply.Success = true
	} else {
		// Send forwarding information, if we have it.
		lastStored, ok := n.seeds.Load(args.Hash)
		if ok {
			reply.Redirect = lastStored.(int)
		} else {
			reply.Redirect = -1
		}
	}
}

func (n *BFSNetwork) sendRequestPost(server int, args *BFSRequestPostArgs, reply *BFSRequestPostReply) bool {
	ok := n.net[server].Call("BFSNetwork.RequestPost", args, reply)
	return ok
}

type BFSSendSeedsArgs struct {
	Origin   int
	FullSync bool
}

type BFSSendSeedsReply struct {
	Seeds sync.Map
}

func (n *BFSNetwork) SendSeeds(args *BFSSendSeedsArgs, reply *BFSSendSeedsReply) {
	// Add as peer.
	n.peers = append(n.peers, args.Origin)
	
	if args.FullSync {
		reply.Seeds = n.seeds
	}
}

func (n *BFSNetwork) sendSendSeeds(server int, args *BFSSendSeedsArgs, reply *BFSSendSeedsReply) bool {
	ok := n.net[server].Call("BFSNetwork.SendSeeds", args, reply)
	return ok
}

func (n *BFSNetwork) setup() {
	// Set-up function to get seeds from peers (and add them as peers).
	seedsReady := false
	
	for i := 0; i < len(n.peers); i++ {
		var args BFSSendSeedsArgs
		if !seedsReady {
			args.FullSync = true
		}
		var reply BFSSendSeedsReply
		
		status := n.sendSendSeeds(n.peers[i], &args, &reply)

		if status {
			if !seedsReady {
				reply.Seeds.Range(func (seed, server interface{}) bool {
					_, ok := n.seeds.Load(seed.(HashTriple))
					if !ok {
						n.seeds.Store(seed.(HashTriple), server.(int))
					}
					return true
				})
				seedsReady = true

				
				go reply.Seeds.Range(func (seed, _ interface{}) bool {
					sp, ok := n.GetPost(seed.(HashTriple))
					if ok {
						if n.keep(sp) {
							n.sv.mu.Lock()
							_, ok := n.sv.Posts[sp.Seed]
							if !ok {
								n.sv.Posts[sp.Seed] = sp
							}
							n.sv.mu.Unlock()
						}

						n.sv.PostsCh <- sp
					}
					return true
				})
                                
			}
		}
	}
}

func (n *BFSNetwork) findRandomPeers(rootPeer int) {
	// Generates log(n) random peers.
	// TODO: actually use random walk to generate peers.

	var peers []int
	for i := 0; i < min(len(n.net) - 1, NUM_PEERS); i++ {
		r := rand.Intn(len(n.net))
		for r == n.sv.me {
			r = rand.Intn(len(n.net))
		}
		peers = append(peers, r)
	}
	n.peers = peers
	fmt.Printf("Peers on %d are %v\n", n.sv.me, n.peers)
}

func MakeBFSNetwork(sv *Server) *BFSNetwork {
	n := &BFSNetwork{}

	n.sv = sv
	n.net = sv.network

	n.findRandomPeers(0)

	n.setup()

	return n
}
