package dreddit

import (
	"labrpc"
	"math/rand"
	"sync"
)

type BroadcastNetwork struct {
	mu    sync.Mutex
	sv    *Server
	me    *labrpc.ClientEnd
	peers []*labrpc.ClientEnd

	seeds map[HashTriple]int
}

const BroadcastKeepPercent = 0.2

func (n *BroadcastNetwork) keep(sp SignedPost) bool {
	return rand.Float32() < BroadcastKeepPercent
}

func (n *BroadcastNetwork) NewPost(sp SignedPost) bool {
	n.mu.Lock()
	n.seeds[sp.Seed] = n.sv.me
	n.mu.Unlock()
	
	for i := 0; i < len(n.peers); i++ {
		if n.peers[i] != n.me {
			go n.makeReceivePost(n.peers[i], sp, n.sv.me)
		}
	}

	return true
}

func (n *BroadcastNetwork) GetPost(hash HashTriple) (SignedPost, bool) {
	// Try saved origin first.
	origin, ok := n.seeds[hash]
	if ok {
		args := BroadcastRequestPostArgs{Hash: hash}
		var reply BroadcastRequestPostReply
		
		status := n.sendRequestPost(n.peers[origin], &args, &reply)

		if status && reply.Success {
			_, ok := verifyPost(reply.Sp, hash)
			if ok {
				return reply.Sp, true
			}
		}
	}
	
	for i := 0; i < len(n.peers); i++ {
		if n.peers[i] != n.me {
			args := BroadcastRequestPostArgs{Hash: hash}
			var reply BroadcastRequestPostReply

			status := n.sendRequestPost(n.peers[i], &args, &reply)

			if status && reply.Success {
				_, ok := verifyPost(reply.Sp, hash)
				if ok {
					return reply.Sp, true
				}
			}
		}
	}
	
	return SignedPost{}, false
}

type BroadcastReceivePostArgs struct {
	Origin int
	Sp     SignedPost
}

type BroadcastReceivePostReply struct {
	Success bool
}

func (n *BroadcastNetwork) ReceivePost(args *BroadcastReceivePostArgs, reply *BroadcastReceivePostReply) {
	_, ok := verifyPost(args.Sp, args.Sp.Seed)
	if ok {
		n.mu.Lock()
		n.seeds[args.Sp.Seed] = args.Origin
		n.mu.Unlock()
		
		if n.keep(args.Sp) {
			n.sv.mu.Lock()
			n.sv.Posts[args.Sp.Seed] = args.Sp
			n.sv.mu.Unlock()
		}
		
		reply.Success = true
	}
}

func (n *BroadcastNetwork) sendReceivePost(server *labrpc.ClientEnd, args *BroadcastReceivePostArgs, reply *BroadcastReceivePostReply) bool {
	ok := server.Call("BroadcastNetwork.ReceivePost", args, reply)
	return ok
}

func (n *BroadcastNetwork) makeReceivePost(server *labrpc.ClientEnd, sp SignedPost, origin int) {
	retry := true

	for retry {
		args := BroadcastReceivePostArgs{Sp: sp, Origin: origin}
		var reply BroadcastReceivePostReply
		
		status := n.sendReceivePost(server, &args, &reply)
		
		if status {
			retry = false
		}
	}
}

type BroadcastRequestPostArgs struct {
	Hash HashTriple
}

type BroadcastRequestPostReply struct {
	Sp      SignedPost
	Success bool
}

func (n *BroadcastNetwork) RequestPost(args *BroadcastRequestPostArgs, reply *BroadcastRequestPostReply) {
	n.sv.mu.Lock()
	sp, ok := n.sv.Posts[args.Hash]
	n.sv.mu.Unlock()
	if ok {
		reply.Sp = sp
		reply.Success = true
	}
}

func (n *BroadcastNetwork) sendRequestPost(server *labrpc.ClientEnd, args *BroadcastRequestPostArgs, reply *BroadcastRequestPostReply) bool {
	ok := server.Call("BroadcastNetwork.RequestPost", args, reply)
	return ok
}

func MakeBroadcastNetwork(sv *Server) *BroadcastNetwork {
	n := &BroadcastNetwork{}

	n.sv = sv
	n.peers = sv.network
	if n.sv.me >= 0 {
		n.me = n.peers[n.sv.me]
	}
	n.seeds = make(map[HashTriple]int)

	return n
}
