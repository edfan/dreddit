package dreddit

import (
	"fmt"
	"testing"
)

func TestSignMessage(t *testing.T) {
	sv := make()
	m := Message{"ezfn", "Test post", "test post please ignore"}
	fmt.Println("Input message:", m)

	sm := sv.signMessage(m)
	dm, _ := sv.verifyMessage(sm)
	fmt.Println("Output message:", dm)
}
