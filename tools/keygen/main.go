package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/andrlikjirka/dp-teals/pkg/jws"
)

func main() {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("failed to generate key pair: %v\n", err)
		return
	}

	kid, err := jws.Thumbprint(pub)
	if err != nil {
		fmt.Printf("failed to compute thumbprint: %v\n", err)
		return
	}

	fmt.Printf("PRIVATE_KEY_B64: %s\n", base64.StdEncoding.EncodeToString(priv))
	fmt.Printf("PUBLIC_KEY_B64: %s\n", base64.StdEncoding.EncodeToString(pub))
	fmt.Printf("KID: %s\n", kid)
}
