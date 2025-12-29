package ring

import (
	"slices"
	"strconv"

	"github.com/buraksezer/consistent"
	"github.com/cespare/xxhash/v2"
)

func uniqueSorted(s []string) []string {
	result := slices.Clone(s)
	slices.Sort(result)
	return slices.Compact(result)
}

type xxhashHasher struct{}

func (h xxhashHasher) Sum64(data []byte) uint64 {
	return xxhash.Sum64(data)
}

type ShardMember string

func (m ShardMember) String() string {
	return string(m)
}

func consistentHashConfig() consistent.Config {
	return consistent.Config{
		PartitionCount:    997, // Prime number for better distribution
		ReplicationFactor: 50,  // Number of replicas per node
		Load:              1.1, // Load factor for bounded loads
		Hasher:            xxhashHasher{},
	}
}

func getShardForWorkflow(workflowID string, shardCount uint32) uint32 {
	if shardCount == 0 {
		return 0
	}

	// Create members for shards 0 to shardCount-1
	members := make([]consistent.Member, shardCount)
	for i := uint32(0); i < shardCount; i++ {
		members[i] = ShardMember(strconv.FormatUint(uint64(i), 10))
	}

	// Create consistent hash ring
	ring := consistent.New(members, consistentHashConfig())

	// Use consistent hashing to find the member for this workflow
	member := ring.LocateKey([]byte(workflowID))
	if member == nil {
		return 0
	}

	// Parse shard ID from member name
	shardID, err := strconv.ParseUint(member.String(), 10, 32)
	if err != nil {
		return 0
	}

	return uint32(shardID)
}
