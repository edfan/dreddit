package dreddit
import "math/rand"
import "labrpc"

type GetRandomArgs struct{
	T int //0, 1 - regular peer, storage peer
	Direction int //0, 1 - only used for random storage peer
}

type GetRandomResponse struct{
	Node int
	NodeM byte
}


func GetRandomKey(m map[int]int) (int){
	chosen_index := rand.Intn(len(m))
	var k int
	for k = range m{
		if chosen_index == 0{
			return k
		}
		chosen_index--
	}
	return k
}

func (dn *DredditNode) SendGetRandom(network []*labrpc.ClientEnd, server int, args *GetRandomArgs, reply *GetRandomResponse) (bool){
	ok := network[server].Call("DredditNode.GetRandom", args, reply)
	return ok
}

func (dn *DredditNode) GetRandom(args *GetRandomArgs, resp *GetRandomResponse){

	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	var chosen_peer int
	for true{
		if args.T == 0{
			chosen_peer = GetRandomKey(dn.peers)
		}
		if args.T == 1{
			if args.Direction == 0{
				chosen_peer = GetRandomKey(dn.storage_peers_same)
			}
			if args.Direction == 1{
				chosen_peer = GetRandomKey(dn.storage_peers_above)
			}
		}
		
		ok := dn.SendPing(dn.network, chosen_peer)

		if ok{
			resp.Node = chosen_peer
			resp.NodeM = dn.M
			return
		}else{
			if args.T == 0{
				delete(dn.peers, chosen_peer)
			}
			if args.T == 1{
				if args.Direction == 0{
					delete(dn.storage_peers_same, chosen_peer)
				}
				if args.Direction == 1{
					delete(dn.storage_peers_above, chosen_peer)
				}
			}
		}
	}
}

func (dn *DredditNode) HandlePing(args, reply int){
	return
}

func (dn *DredditNode) SendPing(network []*labrpc.ClientEnd, server int) (bool){
	ok := network[server].Call("DredditNode.HandlePing", 0, 0)
	return ok
}