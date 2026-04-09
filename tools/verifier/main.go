package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	auditv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
	"github.com/andrlikjirka/dp-teals/pkg/hash"
	"github.com/andrlikjirka/dp-teals/pkg/mmr"
	"github.com/gowebpki/jcs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	eventID := flag.String("event-id", "", "UUID of the audit event to verify (required)")
	payloadFile := flag.String("payload-file", "", "Path to JSON file; use '-' for stdin")
	addr := flag.String("addr", "localhost:50051", "gRPC server address")
	flag.Parse()

	if *eventID == "" || *payloadFile == "" {
		fmt.Fprintln(os.Stderr, "error: -event-id and -payload-file are required")
		flag.Usage()
		os.Exit(1)
	}

	rawJSON, err := os.ReadFile(*payloadFile)
	if err != nil {
		log.Fatalf("read payload file: %v", err)
	}
	canonical, err := jcs.Transform(rawJSON)
	if err != nil {
		log.Fatalf("jcs canonicalize: %v", err)
	}

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := auditv1.NewProofServiceClient(conn).GetInclusionProof(ctx,
		&auditv1.GetInclusionProofRequest{EventId: *eventID},
	)
	if err != nil {
		log.Fatalf("GetInclusionProof: %v", err)
	}

	proof := &mmr.InclusionProof{
		Siblings: resp.Proof.Siblings,
		Left:     resp.Proof.Left,
	}

	valid := mmr.VerifyInclusionProof(
		canonical,
		proof,
		resp.RootHash,
		hash.SHA3HashFunc,
	)

	//valid := mmr.VerifyInclusionProofByHash(resp.LeafHash, proof, resp.RootHash, hash.SHA3HashFunc)

	fmt.Printf("event_id:    %s\n", resp.EventId)
	fmt.Printf("leaf_index:  %d\n", resp.LeafIndex)
	fmt.Printf("ledger_size: %d\n", resp.LedgerSize)
	fmt.Printf("leaf_hash:   %s\n", hex.EncodeToString(resp.LeafHash))
	fmt.Printf("root_hash:   %s\n", hex.EncodeToString(resp.RootHash))
	fmt.Printf("siblings:    %d\n", len(proof.Siblings))
	fmt.Println()
	if valid {
		fmt.Println("PROOF VALID — payload is committed to this root")
	} else {
		fmt.Printf("PROOF INVALID — payload may have been tampered or proof is corrupt")
		os.Exit(1)
	}
}
