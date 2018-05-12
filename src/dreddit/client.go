package dreddit

import (
	"crypto/rsa"
//	"fmt"
)

// Implements a simple front-end that contains a server (+ network node).
// Reads posts; forwards headers via HeaderCh.
// Can create a post by calling NewPost.
// Can read a post by calling GetPost.
// Can get all headers via PostQueue.

type Header struct {
	Username  string
	Title     string
	PublicKey rsa.PublicKey
	Seed      HashTriple
}

type Client struct {
	PostQueue []Header
	PostCount int
	Sv        *Server

	HeaderCh  chan Header
}

func createHeader(sp SignedPost) Header {
	p, _ := verifyPost(sp, sp.Seed)
	return Header{Username: p.Username, Title: p.Title,
		PublicKey: sp.PublicKey, Seed: sp.Seed}
}

func (c *Client) PostReader() {
	// Long-running function that reads in posts from PostCh.
	for {
		select {
		case sp := <- c.Sv.PostsCh:
			h := createHeader(sp)
			c.HeaderCh <- h
			c.PostQueue = append(c.PostQueue, h)
			c.PostCount++
			
			// TODO: prune queue if over threshold.
		}
	}
}

func (c *Client) NewPost(p Post) {
	sp := c.Sv.NewPost(p)
	h := createHeader(sp)
	c.HeaderCh <- h
	c.PostQueue = append(c.PostQueue, h)
	c.PostCount++
}

func (c *Client) GetPost(seed HashTriple) (Post, bool) {
	sp, ok := c.Sv.GetPost(seed)
	if ok {
		return verifyPost(sp, seed)
	} else {
		return Post{}, false
	}
}

func MakeClient(sv *Server) *Client {
	c := &Client{}
	c.Sv = sv
	c.HeaderCh = make(chan Header)

	go c.PostReader()
	
	return c
}
