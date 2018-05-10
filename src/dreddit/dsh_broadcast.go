package dreddit

/*

import "math/rand"
import "time"

var NUM_PEERS int
var LOG_NUM_LAYERS int
var RANDOM_WALK_LENGTH int
var NUM_DOWNLOADS int

NUM_PEERS = 8
LOG_NUM_LAYERS = 8
RANDOM_WALK_LENGTH = 10
NUM_DOWNLOADS = 4
GOSSIP_SIZE = 10

func (dn *DredditNode) NewPost(sp SignedPost) bool {

	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	nodeM = sp.Seed.ParentHash[0]

	ok, current_node := dn.FindStorageLayer(nodeM)
	if !ok{
		return false
	}

	duplicates := 0

	for duplicates <= NUM_DOWNLOADS{
		args := PleaseDownloadArgs{Post: sp, Seed: sp.Seed}
		var resp PleaseDownloadResp
		ok := SendPleaseDownload(current_node, args, resp)
		if ok{
			duplicates += 1
		}else{
			return false
		}

		args := GetRandomWalkArgs{T: 1}
		var resp GetRandomResponse
		ok = SendGetRandomWalk(current_node, args, resp)
		if ok{
			current_node = resp.Node
		}
	}
	return true
}


// no locks, only called by other funcitons 
func (dn *DredditNode) FindStorageLayer(M byte) (bool, *labrpc.ClientEnd){
	current_diff := 256
	current_closest := -1
	for i := range dn.storage_index{
		diff := int(M) - int(dn.storage_index[i])
		if diff < 0{
			diff = -1*diff
		}
		if diff < current_diff{
			current_diff = diff
			current_closest = i
		}
	}

	current_layer := dn.storage_index[i]
	current_node := dn.storage[i]
	for current_layer != M{
		var resp GetRandomResponse
		if M < current_layer{
			args := GetRandomArgs{t: 1, Direction: -1}
		}else{
			args := GetRandomArgs{t: 1, Direction: 1}
		}
		ok := SendGetRandom(current_node, args, resp)
		if !ok{
			return false, current_node
		}
		current_layer = resp.NodeM
		current_node = resp.Node
	}
	return true, current_node
}

func (dn *DredditNode) GetPost(sd []Seed) ([]SignedPost, bool) {

	// FindStorageLayer on ParentHash
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	nodeM = sd[0].ParentHash[0]

	ok, current_node := dn.FindStorageLayer(nodeM)
	
	if !ok{
		return false
	}

	return_posts := make([]SignedPost)

	counter := 0

	for counter < len(sd){
		args := PleaseSendArgs{Seed: sd[i]}
		var resp PleaseSendResp 
		ok = SendPleaseSend(current_node, args, resp)
		if !ok{
			return return_posts, false
		}
		if resp.Success{
			return_posts = append(return_posts, resp.Post)
			counter += 1
		}else{
			args := GetRandomWalkArgs{T: 1}
			var resp GetRandomResponse
			ok = SendGetRandomWalk(current_node, args, resp)
			if ok{
				current_node = resp.Node
			}
		}
	}

	return return_posts, true
}


func (dn *DredditNode) BackgroundGossip(){
	for true{
		dn.sv.mu.Lock()
		chosen_peer_index := rand.Intn(NUM_PEERS)
		chosen_peer := dn.peers[chosen_peer_index]

		args := GossipArgs{Seeds: dn.sv.Seeds[len(dn.sv.Seeds) - GOSSIP_SIZE:], FullReply: false}
		var resp GossipResp

		ok := SendGossipHandling(chosen_peer, args, resp)

		if ok{
			for i := resp.Seeds{
				add = true
				for j := dn.sv.Seeds{
					if resp.Seeds[i] == dn.sv.Seeds[j]{
						add = false
					}
				}
				if add{
					dn.sv.Seeds = append(dn.sv.Seeds, resp.Seeds[i])
				}
			}
		}

		dn.sv.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}


// doesn't need locks because, again, called by other functions
func (dn *DredditNode) FullGossip(){

	args := GossipArgs{Seeds: make([]HashTriple), FullReply: true}
	var resp GossipResp

	for true{
		chosen_peer_index := rand.Intn(NUM_PEERS)
		chosen_peer := dn.peers[chosen_peer_index]

		ok := SendGossipHandling(chosen_peer, args, resp)
		if ok{
			dn.sv.Seeds = resp.Seeds
			break
		}
	}
}


type GossipArgs struct{
	Seeds []HashTriple
	FullReply bool
}

type GossipResp struct{
	Success bool
	Seeds []HashTriple
}

func (dn *DredditNode) GossipHandling(args GossipArgs, resp GossipResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	var add bool

	if args.FullReply{
		resp.Seeds = dn.sv.Seeds
	} else {
		resp.Seeds = dn.sv.Seeds[len(dn.sv.Seeds) - GOSSIP_SIZE:]
	}

	for i := args.Seeds{
		add = true
		for j := dn.sv.Seeds{
			if args.Seeds[i] == dn.sv.Seeds[j]{
				add = false
			}
		}
		if add{
			dn.sv.Seeds = append(dn.sv.Seeds, args.Seeds[i])
		}
	}

	resp.Success = true
}

func SendGossipHandling(server *labrpc.ClientEnd, args *GossipArgs, reply *GossipResp){
	ok := server.Call("DredditNode.GossipHandling", args, reply)
	return ok
}

func PeerRefresh(){ // We might not implement or use this
	// Refresh your peers that haven't updated recently, or no longer respond
}

func (dn *DredditNode) BackgroundDownload(){
	// backtrack through your seeds until you find one you don't have and in your M
	// pick a random Storage Peer, try and download from it
	// repeat till you get the info
	for true{
		dn.sv.mu.Lock()

		for i := range dn.sv.Seeds{
			sd := dn.sv.Seeds[len(dn.sv.Seeds) - i - 1]
			if M != sd.Hash[0]{
				continue
			}
			_, ok := dn.sv.Posts[sd]
			if !ok{
				args := PleaseSendArgs{Seed: sd}
				var resp PleaseSendResp
				for true{
					chosen_peer_index := rand.Intn(NUM_PEERS)
					chosen_peer := dn.storage[chosen_peer_index]
					ok2 := SendPleaseSend(chosen_peer, args, resp)
					if ok2{
						dn.sv.Posts[sd] = resp.Post
						break
					} else {
						dn.sv.mu.Unlock()
						time.Sleep(100 * time.Millisecond)
						dn.sv.mu.Lock()
					}
				}
			}
		}

		dn.sv.mu.Unlock()

		time.Sleep(100 * time.Millisecond)
	}
}

// Sends a request to download the message associated with a specific seed

type PleaseSendArgs struct{
	Seed HashTriple
}

type PleaseSendResp struct{
	Success bool
	Post SignedPost
}

func (dn *DredditNode) PleaseSend(args PleaseSendArgs, resp PleaseSendResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	post, ok := dn.sv.Posts[args.Seed]
	if ok{
		resp.Success = true
		resp.Post = post
	} else{
		resp.Success = false
	}
}

func SendPleaseSend(server *labrpc.ClientEnd, args *PleaseSendArgs, reply *PleaseSendResp){
	ok := server.Call("DredditNode.PleaseSend", args, reply)
	return ok
}

// We might not have implement or use this

func StoragePeerRefresh{
	// Refresh your storage peers that haven't been useful recently, or no longer respond
}

// Sends a message to be downloaded and stored by storage node 

type PleaseDownloadArgs struct{
	Post SignedPost
	Seed HashTriple
}

type PleaseDownloadResp struct{
	Success bool
}

func (dn *DredditNode) PleaseDownload(args PleaseDownloadArgs, resp PleaseDownloadResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	dn.sv.Posts[args.Seed] = args.Posts
	dn.sv.Seeds = append(dn.sv.Seeds, args.Seed)

	resp.Success = true
}

func SendPleaseDownload(server *labrpc.ClientEnd, args *PleaseDownloadArgs, reply *PleaseDownloadResp){
	ok := server.Call("DredditNode.PleaseDownload", args, reply)
	return ok
}

// Cycle through peers several times to get a new random peer - these are called by other locked functions, so should not lock

func (dn *DredditNode) RandomWalk(t int) (bool, *labrpc.ClientEnd){// t = 0 random peer, t = 1 random same level storage peer
	// cycles through a bunch of GetRandomStoragePeer calls
	args := GetRandomArgs{T: t, Direction: 0}
	var resp GetRandomResponse

	ok := false

	for !ok{
		chosen_peer_index := rand.Intn(NUM_PEERS)
		if t == 0{
			chosen_peer := dn.peers[chosen_peer_index]
		}
		if t == 1{
			chosen_peer := dn.storage_peers_same[chosen_peer_index]
		}
		ok = SendPing(chosen_peer)
	}

	counter := 0

	for counter < RANDOM_WALK_LENGTH{
		ok = SendGetRandom(chosen_peer, args, resp)
		if !ok{
			return false, dn.peers[0]
		}
		chosen_peer = resp.Node
		counter += 1
	}

	return true, chosen_peer
}

// Getting random values from another server's peer or storage peer lists

type GetRandomArgs struct{
	T int //0, 1, 2 - regular peer, storage peer, and just storage
	Direction int //0, 1, or -1 - only used for random storage peer
}

type GetRandomResponse struct{
	Node *labrpc.ClientEnd
	NodeM byte
}

func SendGetRandom(server *labrpc.ClientEnd, args *GetRandomArgs, reply *GetRandomResponse){
	ok := server.Call("DredditNode.GetRandom", args, reply)
	return ok
}

func (dn *DredditNode) GetRandom(args *GetRandomArgs, resp *GetRandomResponse){

	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	for true{
		chosen_peer_index := rand.Intn(NUM_PEERS)
		if args.Type == 0{
			chosen_peer := dn.peers[chosen_peer_index]
		}
		if args.Type == 1{
			if args.Direction == 0{
			chosen_peer := dn.storage_peers_same[chosen_peer_index]
			}
			if args.Direction == 1{
				chosen_peer := dn.storage_peers_above[chosen_peer_index]
			}
			if args.Direction == -1{
				chosen_peer := dn.storage_peers_below[chosen_peer_index]
			}
		}
		if args.Type == 2{
			chosen_peer := dn.storage[chosen_peer_index]
		}
		
		ok := SendPing(chosen_peer)

		if ok{
			resp.Node = chosen_peer
			resp.NodeM = dn.M
			return
		}
	}
}

// Pinging another server to verify its existence

func HandlePing(){
	return
}

func SendPing(server *labrpc.ClientEnd){
	ok := server.Call("DredditNode.HandlePing")
	return ok
}

// Getting the output of another server's random walk

type GetRandomWalkArgs struct{
	T int //0, 1 - regular peer and storage peer
}

type GetRandomWalkResponse struct{
	Node *labrpc.ClientEnd
}

func SendGetRandomWalk(server *labrpc.ClientEnd, args *GetRandomWalkArgs, reply *GetRandomWalkResponse){
	ok := server.Call("DredditNode.GetRandomWalk", args, reply)
	return ok
}

func (dn *DredditNode) GetRandomWalk(args *GetRandomWalkArgs, resp *GetRandomWalkResponse){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	if args.Type == 0{
		resp.Node = dn.RandomWalkPeer()
	}
	if args.Type == 1{
		resp.Node = dn.RandomWalkStoragePeer()
	}
}

type DredditNode struct {
	sv                  *Server
	me                  *labrpc.ClientEnd
	peers               [NUM_PEERS]*labrpc.ClientEnd
	storage             [NUM_PEERS]*labrpc.ClientEnd
	storage_index       [NUM_PEERS]byte
	storage_peers_same  [NUM_PEERS]*labrpc.ClientEnd
	storage_peers_above [NUM_PEERS]*labrpc.ClientEnd
	storage_peers_below [NUM_PEERS]*labrpc.ClientEnd
	M                   byte
}

func MakeDredditNode(sv *Server, bool isStorageNode) *DredditNode {
	dn := &DredditNode{}
	dn.sv = sv

	// find 8 peers
	// find 8 storage

	// if isStorageNode:
	// pick an M
	// find 8 storage peers on each relevant level
}


*/
