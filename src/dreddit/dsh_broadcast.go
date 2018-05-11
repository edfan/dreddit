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

type dshOptions struct {
	initialPeers []int
	initialStorage []int
	isStorage bool
	M byte
	initialStoragePeerSame []int
	initialStoragePeerAbove []int
	initialStoragePeerBelow []int
}

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
		ok := SendPleaseDownload(dn.network, current_node, args, resp)
		if ok{
			duplicates += 1
		}else{
			return false
		}

		args := GetRandomWalkArgs{T: 1}
		var resp GetRandomResponse
		ok = SendGetRandomWalk(dn.network, current_node, args, resp)
		if ok{
			current_node = resp.Node
		}
	}
	return true
}


// no locks, only called by other funcitons 
func (dn *DredditNode) FindStorageLayer(M byte) (bool, int){
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
		ok := SendGetRandom(dn.network, current_node, args, resp)
		if !ok{
			return false, current_node
		}
		current_layer = resp.NodeM
		current_node = resp.Node
	}
	return true, current_node
}

func (dn *DredditNode) GetPost(sd Seed) (SignedPost, bool) {

	// FindStorageLayer on ParentHash
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	nodeM = sd.ParentHash[0]

	ok, current_node := dn.FindStorageLayer(nodeM)

	var return_post []SignedPost
	
	if !ok{
		return return_post, false
	}

	for true{
		args := PleaseSendArgs{Seed: sd[i]}
		var resp PleaseSendResp 
		ok = SendPleaseSend(dn.network, current_node, args, resp)

		if !ok{
			return return_post, false
		}

		if resp.Success{
			return_post = resp.Post
			chosen_peer_index := rand.Intn(NUM_PEERS)
			dn.storage[chosen_peer_index] = current_node
			return return_posts, true
		}else{
			args := GetRandomWalkArgs{T: 1}
			var resp GetRandomResponse
			ok = SendGetRandomWalk(dn.network, current_node, args, resp)
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

		args := GossipArgs{Seeds: dn.Seeds[len(dn.Seeds) - GOSSIP_SIZE:], FullReply: false}
		var resp GossipResp

		ok := SendGossipHandling(dn.network, chosen_peer, args, resp)

		if ok{
			for i := resp.Seeds{
				add = true
				for j := dn.Seeds{
					if resp.Seeds[i] == dn.Seeds[j]{
						add = false
					}
				}
				if add{
					dn.Seeds = append(dn.Seeds, resp.Seeds[i])
				}
			}
		}

		dn.sv.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}


// doesn't need locks because, again, called by other functions, or before any other functions are running
func (dn *DredditNode) FullGossip(){

	args := GossipArgs{Seeds: make([]HashTriple), FullReply: true}
	var resp GossipResp

	for true{
		chosen_peer_index := rand.Intn(NUM_PEERS)
		chosen_peer := dn.peers[chosen_peer_index]

		ok := SendGossipHandling(dn.network, chosen_peer, args, resp)
		if ok{
			dn.Seeds = resp.Seeds
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
		resp.Seeds = dn.Seeds
	} else {
		resp.Seeds = dn.Seeds[len(dn.Seeds) - GOSSIP_SIZE:]
	}

	for i := args.Seeds{
		add = true
		for j := dn.Seeds{
			if args.Seeds[i] == dn.Seeds[j]{
				add = false
			}
		}
		if add{
			dn.Seeds = append(dn.Seeds, args.Seeds[i])
		}
	}

	resp.Success = true
}

func SendGossipHandling(network []*labrpc.ClientEnd, server int, args *GossipArgs, reply *GossipResp){
	ok := network[server].Call("DredditNode.GossipHandling", args, reply)
	return ok
}

func PeerRefresh(){ // We might not implement or use this
	// Refresh your peers that haven't updated recently, or no longer respond
	var ok bool
	for true{
		dn.sv.mu.Lock()
		chosen_peer_index := rand.Intn(NUM_PEERS)
		chosen_peer := dn.peers[chosen_peer_index]
		ok = SendPing(dn.network, chosen_peer)
		if !ok{
			ok2, new_peer := dn.RandomWalk(0)
			if ok2 && new_peer != dn.me{
				dn.peers[chosen_peer_index] = new_peer
			}
		}

		dn.sv.mu.Unlock()

		time.Sleep(1000* time.Millisecond)
	}

}

func (dn *DredditNode) BackgroundDownload(){
	// backtrack through your seeds until you find one you don't have and in your M
	// pick a random Storage Peer, try and download from it
	// repeat till you get the info
	for true{
		dn.sv.mu.Lock()

		for i := range dn.Seeds{
			sd := dn.Seeds[len(dn.Seeds) - i - 1]
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
					ok2 := SendPleaseSend(dn.network, chosen_peer, args, resp)
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

func SendPleaseSend(network []*labrpc.ClientEnd, server int, args *PleaseSendArgs, reply *PleaseSendResp){
	ok := network[server].Call("DredditNode.PleaseSend", args, reply)
	return ok
}

// We might not have implement or use this

func StoragePeerRefresh{
	// Refresh your storage peers that haven't been useful recently, or no longer respond
	var ok bool
	for true{
		dn.sv.mu.Lock()
		chosen_peer_index := rand.Intn(NUM_PEERS)
		chosen_peer_index_type := rand.Intn(3)
		if chosen_peer_index_type == 0{
			chosen_peer := dn.storage_peers_same[chosen_peer_index]
		}
		if chosen_peer_index_type == 1{
			chosen_peer := dn.storage_peers_above[chosen_peer_index]
		}
		if chosen_peer_index_type == 2{
			chosen_peer := dn.storage_peers_below[chosen_peer_index]
		}
		ok = SendPing(dn.network, chosen_peer)
		if !ok{
			if chosen_peer_index_type == 0{
				ok2, new_peer := dn.RandomWalk(1)
				if ok2 && new_peer != dn.me{
					dn.storage_peers_same[chosen_peer_index] = new_peer
				}
			}
			if chosen_peer_index == 1{
				another_chosen_peer_index := rand.Intn(NUM_PEERS)
				start_peer := dn.storage_peers_above[another_chosen_peer_index]
				args := GetRandomWalkArgs{T: 1}
				var resp GetRandomWalkResponse
				ok2 := SendGetRandomWalk(dn.network, start_peer, args, resp)
				if ok2{
					dn.storage_peers_above[chosen_peer_index] = resp.Node
				}
			}
			if chosen_peer_index == 2{
				another_chosen_peer_index := rand.Intn(NUM_PEERS)
				start_peer := dn.storage_peers_below[another_chosen_peer_index]
				args := GetRandomWalkArgs{T: 1}
				var resp GetRandomWalkResponse
				ok2 := SendGetRandomWalk(dn.network, start_peer, args, resp)
				if ok2{
					dn.storage_peers_below[chosen_peer_index] = resp.Node
				}
			}
		}
		dn.sv.mu.Unlock()

		time.Sleep(333* time.Millisecond)

	}
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
	dn.Seeds = append(dn.Seeds, args.Seed)

	resp.Success = true
}

func SendPleaseDownload(network []*labrpc.ClientEnd, server int, args *PleaseDownloadArgs, reply *PleaseDownloadResp){
	ok := network[server].Call("DredditNode.PleaseDownload", args, reply)
	return ok
}

// Cycle through peers several times to get a new random peer - these are called by other locked functions, so should not lock

func (dn *DredditNode) RandomWalk(t int) (bool, int){// t = 0 random peer, t = 1 random same level storage peer
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
		ok = SendPing(dn.network, chosen_peer)
	}

	counter := 0

	for counter < RANDOM_WALK_LENGTH{
		ok = SendGetRandom(dn.network, chosen_peer, args, resp)
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
	Node int
	NodeM byte
}

func SendGetRandom(network []*labrpc.ClientEnd, server int, args *GetRandomArgs, reply *GetRandomResponse){
	ok := network[server].Call("DredditNode.GetRandom", args, reply)
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
		
		ok := SendPing(dn.network, chosen_peer)

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

func SendPing(network []*labrpc.ClientEnd, server int){
	ok := network[server].Call("DredditNode.HandlePing")
	return ok
}

// Getting the output of another server's random walk

type GetRandomWalkArgs struct{
	T int //0, 1 - regular peer and storage peer
}

type GetRandomWalkResponse struct{
	Node int
}

func SendGetRandomWalk(network []*labrpc.ClientEnd, server int, args *GetRandomWalkArgs, reply *GetRandomWalkResponse){
	ok := network[server].Call("DredditNode.GetRandomWalk", args, reply)
	return ok
}

func (dn *DredditNode) GetRandomWalk(args *GetRandomWalkArgs, resp *GetRandomWalkResponse){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	if args.Type == 0{
		resp.Node = dn.RandomWalk(0)
	}
	if args.Type == 1{
		resp.Node = dn.RandomWalk(1)
	}
}

type DredditNode struct {
	sv                  *Server
	me                  int
	peers               [NUM_PEERS]int
	storage             [NUM_PEERS]int
	storage_index       [NUM_PEERS]byte
	storage_peers_same  [NUM_PEERS]int
	storage_peers_above [NUM_PEERS]int
	storage_peers_below [NUM_PEERS]int
	M                   byte
	network             []*labrpc.ClientEnd
	isStorage	    bool
	Seeds               []HashTriple
}


func MakeDredditNode(sv *Server, o dshOptions) *DredditNode {
	dn := &DredditNode{}
	dn.sv = sv
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()
	dn.me = sv.me

	dn.peers = o.initialPeers
	dn.storage = o.initialStorage
	dn.network = sv.network
	dn.isStorage = o.isStorage
	dn.FullGossip()

	if o.isStorage {
		dn.M = o.M
		dn.storage_peers_below = o.initialStoragePeerBelow
		dn.storage_peers_above = o.initialStoragePeerAbove
		dn.storage_peers_same = o.initialStoragePeerSame
	}
}

func (dn *DredditNode) StartDredditNode(){
	go BackgroundGossip()
	go PeerRefresh()
	if dn.isStorage{
		go BackgroundDownload()
		go StoragePeerRefresh()
	}
}

*/
