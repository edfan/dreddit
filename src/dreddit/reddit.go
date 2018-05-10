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
	// TODO: Support comments - ParentHash Field and ReplyToHash Field
	Username  string
	Title     string
	Body      string
}

type SignedPost struct {
	// TODO: Support comments - Seed struct object (self hash, ParentHash, and ReplyToHash)
	Post      []byte
	Hash      [32]byte
	PublicKey rsa.PublicKey
	Signature []byte
}

type Network interface {
	NewPost(sp SignedPost)
	GetPost(hash [32]byte) (SignedPost, bool)
}

type Server struct {
	mu           sync.RWMutex
	
	//id           string
	key          rsa.PrivateKey
	
	net          Network
	me           int
	initialPeers []*labrpc.ClientEnd

	//Seeds        map[[32]byte]string
	Seeds        [][32]byte
	Posts        map[[32]byte]SignedPost
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
	// set ParentHash, ReplyToHash fields
	sig, err := rsa.SignPKCS1v15(rand.Reader, &sv.key, crypto.SHA256, hash[:])
	if (err != nil) {
		fmt.Println("Error in signing post")
	}

	return SignedPost{Post: encoded, Hash: hash, PublicKey: sv.key.PublicKey, Signature: sig}
}

func verifyPost(sp SignedPost, hash [32]byte) (Post, bool) {
	intHash := sha256.Sum256(sp.Post)
	post, key := decodePost(sp.Post)
	
	if hash != sp.Hash {
		fmt.Println("Post hash does not match provided hash")
		return post, false
	}
	if intHash != sp.Hash {
		fmt.Println("Post hash does not match internally")
		return post, false
	}

	if key.N.Cmp(sp.PublicKey.N) != 0 || key.E != sp.PublicKey.E {
		fmt.Println("Post key does not match signed key")
		return post, false
	}

	if rsa.VerifyPKCS1v15(&sp.PublicKey, crypto.SHA256, hash[:], sp.Signature) != nil {
		fmt.Println("Post signature does not match")
		return post, false
	}

	return post, true
}

func (sv *Server) NewPost(p Post) [32]byte {
	sp := sv.signPost(p)
	
	sv.mu.Lock()
	sv.Seeds = append(sv.Seeds, sp.Hash)
	sv.Posts[sp.Hash] = sp
	sv.mu.Unlock()
	
	sv.net.NewPost(sp)
	return sp.Hash
}

func (sv *Server) GetPost(hash [32]byte) (SignedPost, bool) {
	sv.mu.RLock()
	sp, ok := sv.Posts[hash]
	sv.mu.RUnlock()
	if ok {
		return sp, true
	} else {
		sp, ok = sv.net.GetPost(hash)
		if ok {
			_, good := verifyPost(sp, hash)
			if good {
				sv.mu.Lock()
				sv.Posts[hash] = sp
				sv.mu.Unlock()
				return sp, true
			}
		}
		return sp, false
	}
}

func Make(initialPeers []*labrpc.ClientEnd, me int) *Server {
	sv := &Server{}
	//sv.id = xid.New().String()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating RSA keypair")
	}
	sv.key = *key

	sv.me = me
	sv.initialPeers = initialPeers

	sv.Posts = make(map[[32]byte]SignedPost)

	// Change this to change the network type.
        // sv.net = MakeBroadcastNetwork(sv)
	sv.net = MakeBFSNetwork(sv)

	return sv
}
