type GetRandomWalkArgs struct{
	T int //0, 1 - regular peer and storage peer
}

type GetRandomWalkResponse struct{
	Node int
	Success bool
}

func SendGetRandomWalk(network []*labrpc.ClientEnd, server int, args *GetRandomWalkArgs, reply *GetRandomWalkResponse)(bool){
	ok := network[server].Call("DredditNode.GetRandomWalk", args, reply)
	return ok
}

func (dn *DredditNode) GetRandomWalk(args *GetRandomWalkArgs, resp *GetRandomWalkResponse){
	dn.sv.mu.Lock()
	defer dn.sv.mu.Unlock()

	if args.T == 0{
		resp.Success, resp.Node = dn.RandomWalk(0)
	}
	if args.T == 1{
		resp.Success, resp.Node = dn.RandomWalk(1)
	}
}

func (dn *DredditNode) RandomWalk(t int) (bool, int){// t = 0 random peer, t = 1 random same level storage peer
	// cycles through a bunch of GetRandomStoragePeer calls
	args := GetRandomArgs{T: t, Direction: 0}
	var resp GetRandomResponse

	ok := false
	var chosen_peer int

	for !ok{
		chosen_peer_index := rand.Intn(NUM_PEERS)
		if t == 0{
			chosen_peer = dn.peers[chosen_peer_index]
		}
		if t == 1{
			chosen_peer = dn.storage_peers_same[chosen_peer_index]
		}
		ok = SendPing(dn.network, chosen_peer)
	}

	counter := 0

	for counter < RANDOM_WALK_LENGTH{
		ok = SendGetRandom(dn.network, chosen_peer, &args, &resp)
		if !ok{
			return false, dn.peers[0]
		}
		chosen_peer = resp.Node
		counter += 1
	}

	return true, chosen_peer
}

