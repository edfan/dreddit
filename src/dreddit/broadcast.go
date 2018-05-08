package dreddit

import (
	"labrpc"
)

type BroadcastNetwork struct {
	sv    *Server
	me    *labrpc.ClientEnd
	peers []*labrpc.ClientEnd
}

func RandomKeep(sp SignedPost) bool {
	return true
}

func (n *BroadcastNetwork) NewPost(sp SignedPost) {
	for i := 0; i < len(n.peers); i++ {
		if n.peers[i] != n.me {
			go makeReceivePost(n.peers[i], sp)
		}
	}
}

func (n *BroadcastNetwork) GetPost(hash [32]byte) (SignedPost, bool) {
	for i := 0; i < len(n.peers); i++ {
		if n.peers[i] != n.me {
			args := RequestPostArgs{Hash: hash}
			var reply RequestPostReply

			status := sendRequestPost(n.peers[i], &args, &reply)

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

type ReceivePostArgs struct {
	Sp SignedPost
}

type ReceivePostReply struct {
	Success bool
}

func (n *BroadcastNetwork) ReceivePost(args *ReceivePostArgs, reply *ReceivePostReply) {
	_, ok := verifyPost(args.Sp, args.Sp.Hash)
	if ok {
		n.sv.mu.Lock()
		n.sv.Seeds[args.Sp.Hash] = n.sv.id
		n.sv.Posts[args.Sp.Hash] = args.Sp
		n.sv.mu.Unlock()

		reply.Success = true
	}
}

func sendReceivePost(server *labrpc.ClientEnd, args *ReceivePostArgs, reply *ReceivePostReply) bool {
	ok := server.Call("BroadcastNetwork.ReceivePost", args, reply)
	return ok
}

func makeReceivePost(server *labrpc.ClientEnd, sp SignedPost) {
	retry := true

	for retry {
		args := ReceivePostArgs{Sp: sp}
		var reply ReceivePostReply
		
		status := sendReceivePost(server, &args, &reply)
		
		if status {
			retry = false
		}
	}
}

type RequestPostArgs struct {
	Hash [32]byte
}

type RequestPostReply struct {
	Sp      SignedPost
	Success bool
}

func (n *BroadcastNetwork) RequestPost(args *RequestPostArgs, reply *RequestPostReply) {
	n.sv.mu.Lock()
	sp, ok := n.sv.Posts[args.Hash]
	n.sv.mu.Unlock()
	if ok {
		reply.Sp = sp
		reply.Success = true
	}
}

func sendRequestPost(server *labrpc.ClientEnd, args *RequestPostArgs, reply *RequestPostReply) bool {
	ok := server.Call("BroadcastNetwork.RequestPost", args, reply)
	return ok
}

func MakeBroadcastNetwork(sv *Server) *BroadcastNetwork {
	n := &BroadcastNetwork{}

	n.sv = sv
	n.peers = sv.initialPeers
	if n.sv.me >= 0 {
		n.me = n.peers[n.sv.me]
	}

	return n
}
