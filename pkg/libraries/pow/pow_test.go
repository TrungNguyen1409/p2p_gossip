package pow

import (
	"github.com/stretchr/testify/assert"
	"gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/enum"
	pb "gitlab.lrz.de/netintum/teaching/p2psec_projects_2024/Gossip-7/pkg/proto"
	"testing"
	"time"
)

// TestCalculate checks that the Calculate function correctly finds a nonce.
func TestCalculate(t *testing.T) {
	// Set up a test message
	testMessage := &pb.GossipMessage{
		Type:      1,
		From:      "node1",
		Payload:   []byte("test payload"),
		Ttl:       64,
		MessageId: "testMessageID",
	}

	enum.Difficulty = "0000"

	startTime := time.Now()
	nonce := Calculate(testMessage)
	duration := time.Since(startTime).Seconds()

	// Output the time taken to calculate
	t.Logf("Time taken to calculate PoW: %f seconds", duration)

	assert.True(t, Validate(testMessage, nonce))
}
