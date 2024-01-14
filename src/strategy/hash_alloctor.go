package strategy

import (
	"errors"
	"hash/fnv"
	"sort"
)

var HashAllocator = func(jobKey string, myNodeId string, allNodeIds []string) (bool, error) {
	sort.Strings(allNodeIds)
	nodeCount := uint32(len(allNodeIds))

	myNodeIndex := -1
	for i, node := range allNodeIds {
		if myNodeId == node {
			myNodeIndex = i
		}
	}

	if myNodeIndex == -1 {
		return false, errors.New("given nodeId doesnt exist in the allNodeIds")
	}

	var jobHash = int32(hash(jobKey) % nodeCount)

	if jobHash == int32(myNodeIndex) {
		return true, nil
	}

	return false, nil
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
