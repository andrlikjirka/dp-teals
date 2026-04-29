package main

import (
	"context"
	"encoding/base64"
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
	// shared flags
	mode := flag.String("mode", "inclusion", "verification mode: inclusion or consistency")
	addr := flag.String("addr", "localhost:50051", "gRPC server address")

	// inclusion flags
	eventID := flag.String("event-id", "", "UUID of the audit event to verify")
	ledgerSize := flag.Int64("ledger-size", 0, "ledger size to anchor inclusion proof against (0 = current)")
	trustedRoot := flag.String("trusted-root", "", "trusted root hash at ledger-size (base64); if omitted, verifies against server-returned root only")
	payloadFile := flag.String("payload-file", "", "path to JSON payload file; use '-' for stdin")

	// consistency flags
	fromSize := flag.Int64("from-size", 0, "old ledger size")
	toSize := flag.Int64("to-size", 0, "new ledger size")
	oldRootB64 := flag.String("old-root", "", "trusted root hash at from-size (base64)")
	newRootB64 := flag.String("new-root", "", "trusted root hash at to-size (base64)")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := auditv1.NewProofServiceClient(conn)

	switch *mode {
	case "inclusion":
		runInclusionProofVerification(ctx, client, *eventID, *payloadFile, *ledgerSize, *trustedRoot)
	case "consistency":
		runConsistencyProofVerification(ctx, client, *fromSize, *toSize, *oldRootB64, *newRootB64)
	default:
		fmt.Printf("unknown mode %q — use inclusion or consistency\n", *mode)
		os.Exit(1)
	}
}

func runInclusionProofVerification(ctx context.Context, client auditv1.ProofServiceClient, eventID, payloadFile string, size int64, rootB64 string) {
	if eventID == "" || payloadFile == "" {
		fmt.Printf("error: -event-id and -payload-file are required for inclusion mode")
		flag.Usage()
		return
	}

	rawJSON, err := os.ReadFile(payloadFile)
	if err != nil {
		log.Fatalf("read payload file: %v", err)
	}
	canonical, err := jcs.Transform(rawJSON)
	if err != nil {
		log.Fatalf("jcs canonicalize: %v", err)
	}

	resp, err := client.GetInclusionProof(ctx, &auditv1.GetInclusionProofRequest{EventId: eventID, LedgerSize: size})
	if err != nil {
		log.Fatalf("GetInclusionProof: %v", err)
	}

	proof := &mmr.InclusionProof{
		Siblings: resp.Proof.Siblings,
		Left:     resp.Proof.Left,
	}

	verifyRoot := resp.RootHash
	if rootB64 != "" {
		verifyRoot, err = base64.StdEncoding.DecodeString(rootB64)
		if err != nil {
			log.Fatalf("decode -trusted-root: %v", err)
		}
	}

	valid := mmr.VerifyInclusionProof(canonical, proof, verifyRoot, hash.SHA3HashFunc)

	fmt.Printf("event_id:    %s\n", resp.EventId)
	fmt.Printf("leaf_index:  %d\n", resp.LeafIndex)
	fmt.Printf("ledger_size: %d\n", resp.LedgerSize)
	fmt.Printf("leaf_hash:   %s\n", hex.EncodeToString(resp.LeafHash))
	fmt.Printf("root_hash:   %s\n", hex.EncodeToString(resp.RootHash))
	fmt.Printf("siblings:    %d\n", len(proof.Siblings))
	if rootB64 == "" {
		fmt.Println("(warning: no -trusted-root provided, verifying against server-returned root only)")
	}
	fmt.Println()
	if valid {
		fmt.Println("PROOF VALID — payload is committed to this root")
	} else {
		fmt.Println("PROOF INVALID — payload may have been tampered or proof is corrupt")
		return
	}
}

func runConsistencyProofVerification(ctx context.Context, client auditv1.ProofServiceClient, fromSize, toSize int64, oldRootB64, newRootB64 string) {
	if fromSize <= 0 || toSize <= 0 || oldRootB64 == "" || newRootB64 == "" {
		fmt.Printf("error: -from-size, -to-size, -old-root and -new-root are required for consistency mode")
		flag.Usage()
		return
	}

	oldRoot, err := base64.StdEncoding.DecodeString(oldRootB64)
	if err != nil {
		log.Fatalf("decode -old-root: %v", err)
	}
	newRoot, err := base64.StdEncoding.DecodeString(newRootB64)
	if err != nil {
		log.Fatalf("decode -new-root: %v", err)
	}

	resp, err := client.GetConsistencyProof(ctx, &auditv1.GetConsistencyProofRequest{
		FromSize: fromSize,
		ToSize:   toSize,
	})
	if err != nil {
		log.Fatalf("GetConsistencyProof: %v", err)
	}

	paths := make([]*mmr.ConsistencyPath, len(resp.Proof.ConsistencyPaths))
	for i, p := range resp.Proof.ConsistencyPaths {
		paths[i] = &mmr.ConsistencyPath{Siblings: p.Siblings, Left: p.Left}
	}
	proof := &mmr.ConsistencyProof{
		OldSize:          int(resp.Proof.OldSize),
		NewSize:          int(resp.Proof.NewSize),
		OldPeaksHashes:   resp.Proof.OldPeaksHashes,
		ConsistencyPaths: paths,
		RightPeaks:       resp.Proof.RightPeaks,
	}

	valid := mmr.VerifyConsistencyProof(proof, oldRoot, newRoot, hash.SHA3HashFunc)

	fmt.Printf("from_size:   %d\n", proof.OldSize)
	fmt.Printf("to_size:     %d\n", proof.NewSize)
	fmt.Printf("old_root:    %s\n", oldRootB64)
	fmt.Printf("new_root:    %s\n", newRootB64)
	fmt.Printf("old_peaks:   %d\n", len(proof.OldPeaksHashes))
	fmt.Printf("paths:       %d\n", len(proof.ConsistencyPaths))
	fmt.Printf("right_peaks: %d\n", len(proof.RightPeaks))
	fmt.Println()
	if valid {
		fmt.Println("PROOF VALID — audit log grew consistently")
	} else {
		fmt.Println("PROOF INVALID — audit log history may have been tampered")
		return
	}
}
