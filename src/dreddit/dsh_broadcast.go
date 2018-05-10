import "math/rand"

var NUM_PEERS int
var NUM_LAYERS int
var RANDOM_WALK_LENGTH int
var NUM_DOWNLOADS int

NUM_PEERS = 8
NUM_LAYERS = 8
RANDOM_WALK_LENGTH = 10
NUM_DOWNLOADS = 4

func (dn *DredditNode) NewPost(sp SignedPost) (bool){

	// two steps - append hash to Seed list, and push message to M layer

	// appending hash is handled by Reddit layer

	// FindStorageLayer on ParentHash

	// when we get to layer M

	// while duplicates <= 20 or 8 or whatever:
	// "please download" message onto current_layer_node
	// "random walk storage" from current_layer_node
	// current_layer_node is new node
	// duplicates += 1 if download returns true
	// return false
}

func (dn *DredditNode) FindStorageLayer(M int) (bool){
	// pick from "storage" list the closest storage node to M

	// while current_layer != M:
	// Ask current storage node for up/down layer node
	// set current_layer_node to that node

	// returns true if you find the appropriate layer, false otherwise
}

func (dn *DredditNode) GetPost(sd []Seed) (SignedPost, bool) { //modify to be a list of messages

	// FindStorageLayer on ParentHash

	// while no message:
	// ping node you're on for messages with "GiveMessage"
	// if no message -> random walk storage to new node

	// end: return post chain and true if you got a response, false otherwise

}


func (dn *DredditNode) BackgroundGossip(){
	// randomly pick peer and send them gossip
	// incorporate their gossip back
}

func GossipHandling(gossip []Seed) ([]Seed){
	// handle gossip, return your own
}

func FullGossipHandling(){
	// get a full copy of gossip
}

func PeerRefresh(){ // We might not implement or use this
	// Refresh your peers that haven't updated recently, or no longer respond
}

func (dn *DredditNode) BackgroundDownload(){
	// backtrack through your seeds until you find one you don't have and in your M
	// pick a random Storage Peer, try and download from it
	// repeat till you get the info
}

func DownloadHandling(){
	// return the requested message
}

func StoragePeerRefresh{ // We might not have implement or use this
	// Refresh your storage peers that haven't updated recently, or no longer respond
}

func PleaseDownload(message string) (bool){
	// copy in the comment sent
	// return true when copied in
}



// Cycle through peers several times to get a new random peer - these are called by other locked functions, so should not lock

func (dn *DredditNode) RandomWalk(t int) (bool, *labrpc.ClientEnd){// t = 0 random peer, t = 1 random same level storage peer
	// cycles through a bunch of GetRandomStoragePeer calls
	args := GetRandomArgs{T: t, Direction: 0}
	var resp GetRandomResponse

	ok := false

	while !ok{
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

	while counter < RANDOM_WALK_LENGTH{
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
}

func SendGetRandom(server *labrpc.ClientEnd, args *GetRandomArgs, reply *GetRandomResponse){
	ok := server.Call("DredditNode.GetRandom", args, reply)
	return ok
}

func (dn *DredditNode) GetRandom(args *GetRandomArgs, resp *GetRandomResponse){

	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	while true{
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
	Type int //0, 1 - regular peer and storage peer
}

type GetRandomWalkResponse struct{
	Node *labrpc.ClientEnd
}

func SendGetRandomWalk(server *labrpc.ClientEnd, args *GetRandomArgs, reply *GetRandomResponse){
	ok := server.Call("DredditNode.GetRandomWalk", args, reply)
	return ok
}

func (dn *DredditNode) GetRandomWalk(args *GetRandomArgs, resp *GetRandomResponse){
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
	storage_index       [NUM_PEERS]int
	storage_peers_same  [NUM_PEERS]*labrpc.ClientEnd
	storage_peers_above [NUM_PEERS]*labrpc.ClientEnd
	storage_peers_below [NUM_PEERS]*labrpc.ClientEnd
	M int
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

