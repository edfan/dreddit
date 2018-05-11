package dreddit
import "labrpc"
import "time"


type PostGossipArgs struct{
	PostsList []Posts
	Sender int
}

type PostGossipResp struct{
	Success bool
	PostsList []Posts
}

func (dn *DredditNode) PostGossipHandling(args *GossipArgs, resp *GossipResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	var add bool

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
		ok, _ := dn.sv.Posts[args.Posts[i].Seed]
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

		if GOSSIP_SIZE <= len(dn.Seeds){
			args := GossipArgs{PostGossipArgs: dn.Posts[len(dn.Seeds) - GOSSIP_SIZE:], Sender: dn.me}
		}else{
			args := GossipArgs{PostGossipArgs: dn.Seeds[:], Sender: dn.me}
		}
		var resp PostGossipResp

		ok := dn.SendPostGossipHandling(dn.network, chosen_peer, &args, &resp)

		if ok{
			for i := range resp.Posts{
				ok2, _ := dn.sv.Posts[resp.Posts[i].Seed]
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

