package dreddit
import "labrpc"
import "time"


type GossipArgs struct{
	Seeds []HashTriple
	FullReply bool
	Sender int
}

type GossipResp struct{
	Success bool
	Seeds []HashTriple
}

func (dn *DredditNode) GossipHandling(args *GossipArgs, resp *GossipResp){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	var add bool

	dn.peers[args.Sender] = 0

	if args.FullReply{
		resp.Seeds = dn.Seeds[:]
	} else {
		if GOSSIP_SIZE <= len(dn.Seeds){
			resp.Seeds = dn.Seeds[len(dn.Seeds) - GOSSIP_SIZE:]
		}else{
			resp.Seeds = dn.Seeds[:]
		}
	}

	for i := range args.Seeds{
		add = true
		for j := range dn.Seeds{
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

func (dn *DredditNode) SendGossipHandling(network []*labrpc.ClientEnd, server int, args *GossipArgs, reply *GossipResp) (bool){
	ok := network[server].Call("DredditNode.GossipHandling", args, reply)
	return ok
}


func (dn *DredditNode) BackgroundGossip(){
	for true{
		dn.sv.mu.Lock()
		chosen_peer := GetRandomKey(dn.peers)

		if GOSSIP_SIZE <= len(dn.Seeds){
			args := GossipArgs{Seeds: dn.Seeds[len(dn.Seeds) - GOSSIP_SIZE:], FullReply: false, Sender: dn.me}
		}else{
			args := GossipArgs{Seeds: dn.Seeds[:], FullReply: false, Sender: dn.me}
		}
		var resp GossipResp

		ok := dn.SendGossipHandling(dn.network, chosen_peer, &args, &resp)

		var add bool
		if ok{
			for i := range resp.Seeds{
				add = true
				for j := range dn.Seeds{
					if resp.Seeds[i] == dn.Seeds[j]{
						add = false
					}
				}
				if add{
					dn.Seeds = append(dn.Seeds, resp.Seeds[i])
				}
			}
		} else{
			delete(dn.peers, chosen_peer)
		}

		dn.sv.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

