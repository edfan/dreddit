package dreddit
import "labrpc"
import "time"


type PostGossipArgs struct{
	Posts []SignedPost
	Sender int
}

type PostGossipResp struct{
	Success bool
	Posts []SignedPost
}

func (dn *DredditNode) PostGossipHandling(args *PostGossipArgs, resp *PostGossipResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	if len(dn.storage_peers_same) > MAX_STORAGE_PEERS{
		resp.Success = false
		return
	}

	dn.storage_peers_same[args.Sender] = 0

	if GOSSIP_SIZE <= len(dn.Posts){
		resp.Posts = dn.Posts[len(dn.Posts) - GOSSIP_SIZE:]
	}else{
		resp.Posts = dn.Posts[:]
	}

	for i := range args.Posts{
		_, ok := dn.sv.Posts[args.Posts[i].Seed]
		if !ok{
			dn.sv.Posts[args.Posts[i].Seed] = args.Posts[i]
		}
	}

	resp.Success = true
}

func (dn *DredditNode) SendPostGossipHandling(network []*labrpc.ClientEnd, server int, args *PostGossipArgs, reply *PostGossipResp) (bool){
	ok := network[server].Call("DredditNode.PostGossipHandling", args, reply)
	return ok
}

func (dn *DredditNode) BackgroundPostGossip(){
	for true{
		dn.sv.mu.Lock()
		chosen_peer := GetRandomKey(dn.peers)

		var args PostGossipArgs

		if GOSSIP_SIZE <= len(dn.Seeds){
			args = PostGossipArgs{Posts: dn.Posts[len(dn.Seeds) - GOSSIP_SIZE:], Sender: dn.me}
		}else{
			args = PostGossipArgs{Posts: dn.Posts[:], Sender: dn.me}
		}
		var resp PostGossipResp

		ok := dn.SendPostGossipHandling(dn.network, chosen_peer, &args, &resp)

		if ok{
			for i := range resp.Posts{
				_, ok2 := dn.sv.Posts[resp.Posts[i].Seed]
				if !ok2{
					dn.sv.Posts[resp.Posts[i].Seed] = resp.Posts[i]
				}
			}
		} else{
			delete(dn.peers, chosen_peer)
		}

		dn.sv.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

