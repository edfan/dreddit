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
	"github.com/rs/xid"
)

type Post struct {
	// TODO: Support comments - ParentHash Field and ReplyToHash Field
	Username  string
	Title     string
	Body      string
	ParentHash [32]byte
	ReplyToHash [32]byte
}

type SignedPost struct {
	// TODO: Support comments - Seed struct object (self hash, ParentHash, and reply to hash)
	Post      []byte
	//Hash      [32]byte
	Seed     HashTriple
	PublicKey rsa.PublicKey
	Signature []byte
}

type HashTriple struct {
	Hash [32]byte
	ParentHash [32]byte
	ReplyToHash [32]byte
}

type Network interface {
	NewPost(sp SignedPost)
	GetPost(seed HashTriple) (SignedPost, bool)
}

type KeepRule func(SignedPost) bool

type Server struct {
	mu           sync.Mutex
	
	//id           string
	key          rsa.PrivateKey
	
	net          Network
	keepRule     KeepRule
	me           int
	initialPeers []*labrpc.ClientEnd

	//Seeds        map[[32]byte]string
	Seeds        []HashTriple
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
	
	//return SignedPost{Post: encoded, Hash: hash, PublicKey: sv.key.PublicKey, Signature: sig}
	return SignedPost{Post: encoded, Seed: {Hash: hash, ParentHash: p.ParentHash, ReplyToHash: p.ReplyToHash}, PublicKey: sv.key.PublicKey, Signature: sig}
}

func verifyPost(sp SignedPost, sd HashTriple) (Post, bool) { // This is changed
	intHash := sha256.Sum256(sp.Post)
	post, key := decodePost(sp.Post)
	
	if sd != sp.Seed { // This is changed
		fmt.Println("Post hash does not match provided hash")
		return post, false
	}
	if intHash != sp.Seed.Hash { // This is changed
		fmt.Println("Post hash does not match internally")
		return post, false
	}

	if key.N.Cmp(sp.PublicKey.N) != 0 || key.E != sp.PublicKey.E {
		fmt.Println("Post key does not match signed key")
		return post, false
	}

	if rsa.VerifyPKCS1v15(&sp.PublicKey, crypto.SHA256, sd.Hash[:], sp.Signature) != nil { // This is changed
		fmt.Println("Post signature does not match")
		return post, false
	}

	return post, true
}

func (sv *Server) NewPost(p Post) [32]byte {
	sp := sv.signPost(p)
	
	sv.mu.Lock()
	sv.Seeds.append(sv.Seeds, sp.Seed) // This is changed
	sv.Posts[sp.Seed] = sp // This is changed
	sv.mu.Unlock()
	
	sv.net.NewPost(sp)
	return sp.Seed // This is changed
}

func (sv *Server) GetPost(seeds []HashTriple) (SignedPost, bool) { // This is changed
	sp, ok := sv.Posts[seed] // This is changed
	if ok {
		return sp, true
	} else {
		sp, ok = sv.net.GetPost(seeds) // This is changed
		if ok {
			for i := range seeds{ // This is changed
				_, good := verifyPost(sp[i], seed[i]) // This is changed
				if good {
					sv.mu.Lock()
					sv.Posts[seed[i]] = sp[i] // This is changed
					sv.mu.Unlock()
					return sp, true
				}
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

	//sv.Seeds = make(map[[32]byte]string)
	sv.Seeds = make([]HashTriple) // This is changed
	sv.Posts = make(map[HashTriple]SignedPost) // This is changed

	// Change this to change the network type.
	sv.net = MakeBroadcastNetwork(sv)
	sv.keepRule = RandomKeep

	return sv
}
