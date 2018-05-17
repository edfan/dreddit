package main

import (
	"dreddit"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

type middleware struct {
	cfg        dreddit.Config
	dClients   []*dreddit.Client
	wClients   map[int]*WClient
	inbound    chan []byte
	outbound   chan outMessage
	register   chan *WClient
	unregister chan *WClient
}

type outMessage struct {
	Node int
	Kind string
	Data interface{}
}

func makeMiddleware(n int) *middleware {
	m := &middleware{}
	m.wClients = make(map[int]*WClient)
	m.inbound = make(chan []byte)
	m.outbound = make(chan outMessage)
	m.register = make(chan *WClient)
	m.unregister = make(chan *WClient)

	// Set up clients and servers.
	cfg := dreddit.Make_config(n, dreddit.BFS, nil)
	for i := 0; i < n; i++ {
		c := dreddit.MakeClient(cfg.Servers[i])
		m.dClients = append(m.dClients, c)

		// Forward messages into one channel.
		go func(i int) {
			for {
				select {
				case h := <-c.HeaderCh:
					fmt.Println("Outbound message")
					m.outbound <- outMessage{Node: i, Kind: "Header", Data: h}
				}
			}
		}(i)
	}

	fmt.Println("Nodes have been set up")
	
	return m
}

func (m *middleware) run() {
	for {
		select {
		case client := <-m.register:
			fmt.Println("Registered node", client.node)
			m.wClients[client.node] = client
		case client := <-m.unregister:
			fmt.Println("Unregistered node", client.node)
			if _, ok := m.wClients[client.node]; ok {
				delete(m.wClients, client.node)
				close(client.send)
			}
		case msg := <-m.outbound:
//			fmt.Println("Outbound message in run")
			b, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("json marshal err: %v", err)
				continue
			}
			node, ok := m.wClients[msg.Node]
			if ok { 
				select {
				case node.send <- b:
				default:
				}
			}
		case msg := <-m.inbound:
//			fmt.Println("inbound message")
			var dat map[string]interface{}
			if err := json.Unmarshal(msg, &dat); err != nil {
				continue
			}
			
			switch dat["kind"].(string) {
			case "NewPost":
				go func(dat map[string]interface{}) {
					ph, _ := base64.StdEncoding.DecodeString(dat["parentHash"].(string))
					rth, _ := base64.StdEncoding.DecodeString(dat["replyToHash"].(string))
					var phc, rthc [32]byte
					copy(phc[:], ph[:32])
					copy(rthc[:], rth[:32])
					p := dreddit.Post{Username: dat["username"].(string),
						Title: dat["title"].(string),
						Body: dat["body"].(string),
						ParentHash: phc,
						ReplyToHash: rthc}
					i := int(dat["node"].(float64))
					m.dClients[i].NewPost(p)
					fmt.Printf("Node %d created new post\n", i)
				}(dat)
			case "GetPost":
				go func(dat map[string]interface{}) {
					h, _ := base64.StdEncoding.DecodeString(dat["hash"].(string))
					ph, _ := base64.StdEncoding.DecodeString(dat["parentHash"].(string))
					rth, _ := base64.StdEncoding.DecodeString(dat["replyToHash"].(string))
					var hc, phc, rthc [32]byte
					copy(hc[:], h[:32])
					copy(phc[:], ph[:32])
					copy(rthc[:], rth[:32])
					
					i := int(dat["node"].(float64))
					ht := dreddit.HashTriple{Hash: hc, ParentHash: phc, ReplyToHash: rthc}
					p, ok := m.dClients[i].GetPost(ht)
					fmt.Printf("Node %d getting post\n", i)
					var omsg outMessage
					if ok {
						omsg = outMessage{Node: i, Kind: "Post", Data: p}
					} else {
						omsg = outMessage{Node: i, Kind: "Error", Data: nil}
					}
					b, err := json.Marshal(omsg)
					if err != nil {
						return
					}
					select {
					case m.wClients[omsg.Node].send <- b:
					default:
					}
				}(dat)
			}
		}
	}	
}

