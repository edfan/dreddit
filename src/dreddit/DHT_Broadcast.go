package dreddit
import "labrpc"

const(
	MAX_STORAGE_PEERS = 8
	LOG_NUM_LAYERS = 2
	GOSSIP_SIZE = 100
)

type dshOptions struct {
	initialPeers map[int]int
	initialStorage map[int]int
	isStorage bool
	M byte
	initialStoragePeerSame map[int]int
	initialStoragePeerAbove map[int]int
}



type PleaseDownloadArgs struct{
	Post SignedPost
	Seed HashTriple
	SuggestAbove int
}

type PleaseDownloadResp struct{
	Success bool
}

func (dn *DredditNode) PleaseDownload(args *PleaseDownloadArgs, resp *PleaseDownloadResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	if args.SuggestAbove != -1{
		dn.storage_peers_above[args.SuggestAbove] = 0
	}

	dn.sv.Posts[args.Seed] = args.Post
	dn.Posts = append(dn.Posts, args.Post)

	resp.Success = true
}

func SendPleaseDownload(network []*labrpc.ClientEnd, server int, args *PleaseDownloadArgs, reply *PleaseDownloadResp) (bool){
	ok := network[server].Call("DredditNode.PleaseDownload", args, reply)
	return ok
}


func (dn *DredditNode) NewPost(sp SignedPost) bool {

	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	nodeM := int(sp.Seed.ParentHash[0] >> (8-LOG_NUM_LAYERS))

	ok, current_node := dn.FindStorageLayer(nodeM)
	if !ok{
		return false
	}

	a := -1
	ok , r := dn.storage_peers_above[(nodeM+1)%NUM_LAYERS]
	if ok{
		a = r
	}

	args := PleaseDownloadArgs{Post: sp, Seed: sp.Seed, SuggestAbove: a}
	var resp PleaseDownloadResp
	ok := dn.SendPleaseDownload(dn.network, current_node, &args, &resp)

	if ok{
		return true
	} else{
		delete(dn.storage, nodeM) //questionable
		return false
	}
}





type PleaseSendArgs struct{
	Seed HashTriple
	SuggestAbove int
}

type PleaseSendResp struct{
	Success bool
	Post SignedPost
}

func (dn *DredditNode) PleaseSend(args *PleaseSendArgs, resp *PleaseSendResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	post, ok := dn.sv.Posts[args.Seed]
	if args.SuggestAbove != -1{
		dn.storage_peers_above[args.SuggestAbove] = 0
	}
	if ok{
		resp.Success = true
		resp.Post = post
	} else{
		resp.Success = false
	}
}

func SendPleaseSend(network []*labrpc.ClientEnd, server int, args *PleaseSendArgs, reply *PleaseSendResp) (bool){
	ok := network[server].Call("DredditNode.PleaseSend", args, reply)
	return ok
}

func (dn *DredditNode) GetPost(sd HashTriple) (SignedPost, bool) {

	// FindStorageLayer on ParentHash
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	nodeM := int(sd.ParentHash[0] >> (8-LOG_NUM_LAYERS))

	ok, current_node := dn.FindStorageLayer(nodeM)

	var return_post SignedPost
	
	if !ok{
		return return_post, false
	}

	a := -1
	ok, r := dn.storage_peers_above[(nodeM+1)%NUM_LAYERS]
	if ok{
		a = r
	}

	args := PleaseSendArgs{Seed: sd, SuggestAbove: a}
	var resp PleaseSendResp 
	ok = dn.SendPleaseSend(dn.network, current_node, &args, &resp)

	if !ok{
		delete(dn.storage, nodeM)
		return return_post, false
	}

	if resp.Success{
			return_post = resp.Post
			return return_post, true
	}
}






func (dn *DredditNode) FindStorageLayer(M int) (bool, int){

	ok, node := storage[M]
	if ok{
		return true, node
	}

	ok, node = storage[(M+1)%NUM_LAYERS]
	if !ok{
		ok, node = FindStorageLayer((M+1)%NUM_LAYERS)
		if !ok{
			return false, node
		}
	}
	var resp GetRandomResponse
	var args GetRandomArgs
	args = GetRandomArgs{T: 1, Direction: 1}
	ok = dn.SendGetRandom(dn.network, node, &args, &resp)
	if !ok{
		delete(dn.storage, (M+1)%NUM_LAYERS)
		return false, node
	}else{
		storage[M] = resp.Node
		return true, resp.Node
	}
}


type DredditNode struct {
	sv                  *Server
	me                  int
	peers               map[int]int
	storage             map[int]int
	storage_peers_same  map[int]int
	storage_peers_above map[int]int
	M                   byte
	network             []*labrpc.ClientEnd
	isStorage	        bool
	Seeds               []HashTriple
	Posts               []SignedPost
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
		dn.storage_peers_above = o.initialStoragePeerAbove
		dn.storage_peers_same = o.initialStoragePeerSame
	}
	return dn
}

func (dn *DredditNode) StartDredditNode(){
	go dn.BackgroundGossip()
	if dn.isStorage{
		go dn.BackgroundPostGossip()
	}
}