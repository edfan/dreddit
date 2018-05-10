package dreddit

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"labrpc"
	"sync"
//	"github.com/rs/xid"
)

type Post struct {
	Username  string
	Title     string
	Body      string
	ParentHash [32]byte
	ReplyToHash [32]byte
}

type SignedPost struct {
	Post      []byte
	Seed      HashTriple
	PublicKey rsa.PublicKey
	Signature []byte
}

type HashTriple struct {
	Hash [32]byte
	ParentHash [32]byte
	ReplyToHash [32]byte
}

type Network interface {
	NewPost(sp SignedPost) bool
	GetPost(seed HashTriple) (SignedPost, bool)
}

type Server struct {
	mu           sync.RWMutex
	
	key          rsa.PrivateKey
	
	net          Network
	me           int
	initialPeers []*labrpc.ClientEnd

//	Seeds        []HashTriple
	Posts        map[HashTriple]SignedPost
}

func encodePost(p Post, key rsa.PublicKey) []byte {
	w := new(bytes.Buffer)
	e := gob.NewEncoder(w)
	e.Encode(p)
	e.Encode(key)
	return w.Bytes()
}

func decodePost(data []byte) (Post, rsa.PublicKey) {
	r := bytes.NewBuffer(data)
	d := gob.NewDecoder(r)
	
	var p Post
	var key rsa.PublicKey
	if d.Decode(&p) != nil || d.Decode(&key) != nil {
		fmt.Println("Error in decoding post")
	}
	return p, key
}

func (sv *Server) signPost(p Post) SignedPost {
	encoded := encodePost(p, sv.key.PublicKey)
	hash := sha256.Sum256(encoded)
	sig, err := rsa.SignPKCS1v15(rand.Reader, &sv.key, crypto.SHA256, hash[:])
	if (err != nil) {
		fmt.Println("Error in signing post")
	}
	
	ht := HashTriple{Hash: hash, ParentHash: p.ParentHash, ReplyToHash: p.ReplyToHash}
	return SignedPost{Post: encoded, Seed: ht, PublicKey: sv.key.PublicKey, Signature: sig}
}

func verifyPost(sp SignedPost, sd HashTriple) (Post, bool) {
	intHash := sha256.Sum256(sp.Post)
	post, key := decodePost(sp.Post)
	
	if sd != sp.Seed {
		fmt.Println("Post hash does not match provided hash")
		return post, false
	}
	if intHash != sp.Seed.Hash {
		fmt.Println("Post hash does not match internally")
		return post, false
	}

	if key.N.Cmp(sp.PublicKey.N) != 0 || key.E != sp.PublicKey.E {
		fmt.Println("Post key does not match signed key")
		return post, false
	}

	if rsa.VerifyPKCS1v15(&sp.PublicKey, crypto.SHA256, sd.Hash[:], sp.Signature) != nil {
		fmt.Println("Post signature does not match")
		return post, false
	}
	
	return post, true
}

func (sv *Server) NewPost(p Post) HashTriple {
	sp := sv.signPost(p)
	
	sv.mu.Lock()
//	sv.Seeds = append(sv.Seeds, sp.Seed)
	sv.Posts[sp.Seed] = sp
	sv.mu.Unlock()

	ok := false
	for !ok {
		ok = sv.net.NewPost(sp)
	}
	
	return sp.Seed
}


func (sv *Server) GetPost(seed HashTriple) (SignedPost, bool) {
	sp, ok := sv.Posts[seed]
	if ok {
		return sp, true
	} else {
		sp, ok = sv.net.GetPost(seed)
		if ok {
			_, good := verifyPost(sp, seed)
			if good {
				sv.mu.Lock()
				sv.Posts[seed] = sp
				sv.mu.Unlock()
				return sp, true
			}
		}
		return sp, false
	}
}

func Make(initialPeers []*labrpc.ClientEnd, me int) *Server {
	sv := &Server{}
	// sv.id = xid.New().String()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating RSA keypair")
	}
	sv.key = *key

	sv.me = me
	sv.initialPeers = initialPeers

	sv.Posts = make(map[HashTriple]SignedPost)
//	sv.Seeds = make([]HashTriple)

	// Change this to change the network type.
        // sv.net = MakeBroadcastNetwork(sv)
	sv.net = MakeBFSNetwork(sv)

	return sv
}
