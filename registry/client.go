package registry

import (
	"context"
	"crypto/ecdsa"
	"time"

	"github.com/crypto4people/zoro/registry/pb"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type Client struct {
	client          pb.RegistryClient
	publicKey       *ecdsa.PublicKey
	publicKeyBytes  []byte
	serverPublicKey *ecdsa.PublicKey
}

func NewClient(ctx context.Context, cc grpc.ClientConnInterface, publicKey *ecdsa.PublicKey) (*Client, error) {
	client := pb.NewRegistryClient(cc)

	info, err := client.Info(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	serverPublicKey, err := ethCrypto.DecompressPubkey(info.PublicKey)
	if err != nil {
		return nil, err
	}

	publicKeyBytes, err := publicKey.Bytes()
	if err != nil {
		return nil, err
	}

	return &Client{
		client:          client,
		publicKey:       publicKey,
		publicKeyBytes:  publicKeyBytes,
		serverPublicKey: serverPublicKey,
	}, nil
}

func (client *Client) Info(ctx context.Context) (*pb.InfoResult, error) {
	return client.client.Info(ctx, nil)
}

func (client *Client) List(ctx context.Context) (*pb.ListResult, error) {
	return client.client.List(ctx, &pb.ListParams{
		PublicKey: client.publicKeyBytes,
		Timestamp: time.Now().Unix(),
	})
}

func (client *Client) Vote(ctx context.Context, server string, vote bool) (*empty.Empty, error) {
	return client.client.Vote(ctx, &pb.VoteParams{
		PublicKey: client.publicKeyBytes,
		Server:    server,
		Vote:      vote,
	})
}
