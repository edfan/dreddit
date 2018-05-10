func (dn *DredditNode) NewPost(sp SignedPost) (bool){

	// two steps - append hash to Seed list, and push message to M layer

	// appending hash is handled by Reddit layer

	// FindStorageLayer on ParentHash

	// when we get to layer M

	// while duplicates <= 20 or 8 or whatever:
	// "please download" message onto current_layer_node
	// "random walk storage" from current_layer_node
	// current_layer_node is new node
	// duplicates += 1
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
	// ping node you're on for messages
	// if no message -> random walk storage to new node

	// end: return post chain and true if you got a response, false otherwise

}


func (dn *DredditNode) BackgroundGossip() (bool){
	// randmoly pick peer and send them gossip
	// incorporate their gossip back
	// true if you get a response, false otherwise
}

func GossipHandling(gossip []Seed) ([]Seed){
	// handle gossip, return your own
}

func PleaseDownload(message string) (bool){
	// copy in the comment sent
	// return true when copied in
}

func RandomWalkStorage(){

}

func RandomWalkPeer(){

}

func GetRandomPeer(){
	// picks a random peer, pings it for existence, and then returns it
	// if no response, random walk storage to a new node and return that
}

func GetRandomStorage(){
	// picks a random storage node, pings it for existence, and then returns it
	// if no response, random walk storage to a new node and return that
}

func GetRandomStoragePeer(direction int){
	// picks a random storage peer in the right direction, pings it for existence, and then returns it
	// if no response, random walk storage to a new node and return that
}

func Ping(){
	//just return true
}

type DredditNode struct {
	sv            *Server
	me            *labrpc.ClientEnd

	// TODO: any attempt to access these 
	peers         []*labrpc.ClientEnd
	storage       map[int]labrpc.ClientEnd
	storage_peers_same []*labrpc.ClientEnd
	torage_peers_above []*labrpc.ClientEnd
	torage_peers_below []*labrpc.ClientEnd
}

func MakeDredditNode(sv *Server, bool isStorageNode) *DredditNode {
	dn := &DredditNode{}

	dn.sv = sv
}

