package etcd_alloc

import (
	cs_alloc "etcd_test/cs-alloc"

	"github.com/google/uuid"
	etcd "go.etcd.io/etcd/client/v3"
)

type EtcdAllocOp struct {
	config         etcd.Config
	basePath       string
	nodeId         string
	allocStrategy  cs_alloc.AllocStrategy
	keyListenerMap *map[string]func(watchChan *etcd.WatchChan)
}

type EtcdAllocOption func(*EtcdAllocOp)

func (etcdAllocOp *EtcdAllocOp) applyOpts(opts []EtcdAllocOption) {
	for _, opt := range opts {
		opt(etcdAllocOp)
	}
}

func WithConfig(cfg etcd.Config) EtcdAllocOption {
	return func(op *EtcdAllocOp) {
		op.config = cfg
	}
}

func WithBasePath(basePath string) EtcdAllocOption {

	//  TODO: trim down trailing slash '/'

	return func(op *EtcdAllocOp) {
		op.basePath = basePath
	}
}

func WithMyNodeId(nodeId string) EtcdAllocOption {
	return func(op *EtcdAllocOp) {
		op.nodeId = nodeId
	}
}

func WithRandomNodeId() EtcdAllocOption {
	return func(op *EtcdAllocOp) {
		op.nodeId = uuid.New().String()
	}
}

func WithAllocStrategy(allocStrategy cs_alloc.AllocStrategy) EtcdAllocOption {
	return func(op *EtcdAllocOp) {
		op.allocStrategy = allocStrategy
	}
}

func WithKeyListenerMap(keyListenerMap *map[string]func(watchChan *etcd.WatchChan)) EtcdAllocOption {
	return func(op *EtcdAllocOp) {
		op.keyListenerMap = keyListenerMap
	}
}
