package service

import (
	"context"
	"crypto/ed25519"
	"github.com/hashicorp/go-plugin"
	"github.com/samber/lo"
	"go.lumeweb.com/portal-plugin-sync-grpc/gen/proto"
	"go.lumeweb.com/portal-plugin-sync/internal/metadata"
	"google.golang.org/grpc"
)

var _ Sync = (*SyncGRPC)(nil)

type Sync interface {
	Init(logPublicKey ed25519.PublicKey, nodePrivateKey ed25519.PrivateKey, dataDir string) error
	Update(meta metadata.FileMeta) error
	Query(keys []string) ([]*metadata.FileMeta, error)
	UpdateNodes(nodes []ed25519.PublicKey) error
	RemoveNode(node ed25519.PublicKey) error
}

type SyncGrpcPlugin struct {
	plugin.Plugin
}

func (p *SyncGrpcPlugin) GRPCServer(_ *plugin.GRPCBroker, _ *grpc.Server) error {
	return nil
}

func (p *SyncGrpcPlugin) GRPCClient(_ context.Context, _ *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &SyncGRPC{client: proto.NewSyncClient(c)}, nil
}

type Result struct {
	Hash   []byte
	Proof  []byte
	Length uint
}
type SyncGRPC struct {
	client proto.SyncClient
}

func (b *SyncGRPC) Init(logPublicKey ed25519.PublicKey, nodePrivateKey ed25519.PrivateKey, dataDir string) error {
	_, err := b.client.Init(context.Background(), &proto.InitRequest{LogPublicKey: logPublicKey, NodePrivateKey: nodePrivateKey, DataDir: dataDir})

	if err != nil {
		return err
	}

	return nil
}
func (b *SyncGRPC) Update(meta metadata.FileMeta) error {
	_, err := b.client.Update(context.Background(), &proto.UpdateRequest{Data: meta.ToProtobuf()})

	if err != nil {
		return err
	}

	return nil
}

func (b *SyncGRPC) Query(keys []string) ([]*metadata.FileMeta, error) {
	ret, err := b.client.Query(context.Background(), &proto.QueryRequest{Keys: keys})

	if err != nil {
		return nil, err
	}

	if ret == nil || len(ret.Data) == 0 {
		return nil, nil
	}

	meta := make([]*metadata.FileMeta, 0)

	for _, data := range ret.Data {
		fileMeta, err := metadata.FileMetaFromProtobuf(data)
		if err != nil {
			return nil, err
		}
		meta = append(meta, fileMeta)
	}

	return meta, nil
}

func (b *SyncGRPC) UpdateNodes(nodes []ed25519.PublicKey) error {
	nodeList := lo.Map[ed25519.PublicKey, []byte](nodes, func(node ed25519.PublicKey, _ int) []byte {
		return node
	})

	ret, err := b.client.UpdateNodes(context.Background(), &proto.UpdateNodesRequest{Nodes: nodeList})

	if err != nil {
		return err
	}

	if ret == nil {
		return nil
	}

	return nil
}

func (b *SyncGRPC) RemoveNode(node ed25519.PublicKey) error {
	_, err := b.client.RemoveNode(context.Background(), &proto.RemoveNodeRequest{Node: node})

	if err != nil {
		return err
	}

	return nil
}
