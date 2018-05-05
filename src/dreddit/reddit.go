package dreddit

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"labgob"
	"github.com/rs/xid"
)

type Message struct {
	Username string
	Title    string
	Body     string
}

type SignedMessage struct {
	Message   []byte
	Hash      [32]byte
	PublicKey rsa.PublicKey
	Signature []byte
}

type Server struct {
	id      string
	key     rsa.PrivateKey
}

func encodeMessage(m Message) []byte {
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	e.Encode(m)
	return w.Bytes()
}

func decodeMessage(data []byte) Message {
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	
	var m Message
	if d.Decode(&m) != nil {
		fmt.Println("Error in decoding message")
	}
	return m
}

func (sv *Server) signMessage(m Message) SignedMessage {
	encoded := encodeMessage(m)
	hash := sha256.Sum256(encoded)

	sig, err := rsa.SignPKCS1v15(rand.Reader, &sv.key, crypto.SHA256, hash[:])
	if (err != nil) {
		fmt.Println("Error in signing message")
	}

	return SignedMessage{Message: encoded, Hash: hash, PublicKey: sv.key.PublicKey, Signature: sig}
}

func (sv *Server) verifyMessage(sm SignedMessage) (Message, bool) {
	hash := sha256.Sum256(sm.Message)
	message := decodeMessage(sm.Message)
	
	if hash != sm.Hash {
		fmt.Println("Message hash does not match")
		return message, false
	}

	if rsa.VerifyPKCS1v15(&sm.PublicKey, crypto.SHA256, hash[:], sm.Signature) != nil {
		fmt.Println("Message signature does not match")
		return message, false
	}

	return message, true
}

func make() *Server {
	sv := &Server{}
	sv.id = xid.New().String()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating RSA keypair")
	}
	sv.key = *key

	return sv
}
