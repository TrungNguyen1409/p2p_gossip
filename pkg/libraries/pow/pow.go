package pow

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
)

func ConcatMembers(gm *pb.GossipMessage) string {
	return fmt.Sprintf("%d%s%s%d%s", gm.Type, gm.From, gm.Payload, gm.Ttl, gm.MessageId)
}

// Calculate runs the Proof of Work algorithm to find a valid hash.
func Calculate(msg *pb.GossipMessage) uint64 {
	data := ConcatMembers(msg)
	var hash string
	var nonce uint64
	for {
		// Combine the data with the nonce
		combinedData := fmt.Sprintf("%s%d", data, nonce)

		// Generate the SHA-256 hash of the combined data
		h := sha256.New()
		h.Write([]byte(combinedData))
		hash = hex.EncodeToString(h.Sum(nil))

		// Check if the hash meets the target difficulty
		if strings.HasPrefix(hash, enum.Difficulty) {
			break
		}
		nonce++
	}
	return nonce
}

// Validate checks if the provided hash is valid for the given data and nonce.
func Validate(msg *pb.GossipMessage, nonce uint64) bool {
	h := sha256.New()
	data := ConcatMembers(msg)
	combinedData := fmt.Sprintf("%s%d", data, nonce)
	h.Write([]byte(combinedData))
	calculatedHash := hex.EncodeToString(h.Sum(nil))
	return strings.HasPrefix(calculatedHash, enum.Difficulty)
}
