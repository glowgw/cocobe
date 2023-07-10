package impl

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/dymensionxyz/rollapp/app"
	"github.com/glowgw/cocobe/internal/config"
	"github.com/glowgw/cocobe/internal/node"
	constypes "github.com/tendermint/tendermint/consensus/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	httpclient "github.com/tendermint/tendermint/rpc/client/http"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
	"golang.org/x/exp/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const MaxConnections = 20

type NodeImpl struct {
	ctx             context.Context
	cfg             *config.Config
	client          *httpclient.HTTP
	grpcConnection  *grpc.ClientConn
	txServiceClient tx.ServiceClient
	codec           codec.Codec
	logger          *slog.Logger
}

var _ node.Node = &NodeImpl{}

func NewNode(cfg *config.Config) (node.Node, error) {
	httpClient, err := client.DefaultHTTPClient(cfg.RpcUrl)
	if err != nil {
		return nil, err
	}

	httpTransport, ok := (httpClient.Transport).(*http.Transport)
	if !ok {
		return nil, fmt.Errorf("invalid HTTP Transport: %T", httpTransport)
	}
	httpTransport.MaxConnsPerHost = MaxConnections

	rpcClient, err := httpclient.NewWithClient(cfg.RpcUrl, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}
	err = rpcClient.Start()
	if err != nil {
		return nil, err
	}

	grpcConn, err := CreateGrpcConnection(cfg)
	if err != nil {
		return nil, err
	}

	return &NodeImpl{
			ctx:             context.Background(),
			cfg:             cfg,
			client:          rpcClient,
			grpcConnection:  grpcConn,
			txServiceClient: tx.NewServiceClient(grpcConn),
			codec:           app.MakeEncodingConfig().Codec,
			logger:          slog.Default(),
		},
		nil
}

// CreateGrpcConnection creates a new gRPC client connection from the given configuration
func CreateGrpcConnection(cfg *config.Config) (*grpc.ClientConn, error) {
	var grpcOpts []grpc.DialOption
	if cfg.Insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})))
	}
	HTTPProtocols := regexp.MustCompile("https?://")
	address := HTTPProtocols.ReplaceAllString(cfg.GrpcUrl, "")
	return grpc.Dial(address, grpcOpts...)
}

func (n *NodeImpl) Genesis() (*tmctypes.ResultGenesis, error) {
	res, err := n.client.Genesis(n.ctx)
	if err != nil && strings.Contains(err.Error(), "use the genesis_chunked API instead") {
		return n.getGenesisChunked()
	}
	return res, nil
}

func (n *NodeImpl) getGenesisChunksStartingFrom(id uint) ([]byte, error) {
	res, err := n.client.GenesisChunked(n.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error while getting genesis chunk %d out of %d", id, res.TotalChunks)
	}

	bz, err := base64.StdEncoding.DecodeString(res.Data)
	if err != nil {
		return nil, fmt.Errorf("error while decoding genesis chunk %d out of %d", id, res.TotalChunks)
	}

	if id == uint(res.TotalChunks-1) {
		return bz, nil
	}

	nextChunk, err := n.getGenesisChunksStartingFrom(id + 1)
	if err != nil {
		return nil, err
	}

	return append(bz, nextChunk...), nil
}

func (n *NodeImpl) getGenesisChunked() (*tmctypes.ResultGenesis, error) {
	bz, err := n.getGenesisChunksStartingFrom(0)
	if err != nil {
		return nil, err
	}

	var genDoc *tmtypes.GenesisDoc
	err = tmjson.Unmarshal(bz, &genDoc)
	if err != nil {
		return nil, err
	}

	return &tmctypes.ResultGenesis{Genesis: genDoc}, nil
}

func (n *NodeImpl) ConsensusState() (*constypes.RoundStateSimple, error) {
	state, err := n.client.ConsensusState(context.Background())
	if err != nil {
		return nil, err
	}

	var data constypes.RoundStateSimple
	err = tmjson.Unmarshal(state.RoundState, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (n *NodeImpl) LatestHeight() (int64, error) {
	status, err := n.client.Status(n.ctx)
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

func (n *NodeImpl) ChainID() (string, error) {
	status, err := n.client.Status(n.ctx)
	if err != nil {
		return "", err
	}

	chainID := status.NodeInfo.Network
	return chainID, err
}

func (n *NodeImpl) Validators(height int64) (*tmctypes.ResultValidators, error) {
	vals := &tmctypes.ResultValidators{
		BlockHeight: height,
	}

	page := 1
	perPage := 100 // maximum 100 entries per page
	stop := false
	for !stop {
		result, err := n.client.Validators(n.ctx, &height, &page, &perPage)
		if err != nil {
			return nil, err
		}
		vals.Validators = append(vals.Validators, result.Validators...)
		vals.Count += result.Count
		vals.Total = result.Total
		page++
		stop = vals.Count == vals.Total
	}

	return vals, nil
}

func (n *NodeImpl) Block(height int64) (*tmctypes.ResultBlock, error) {
	return n.client.Block(n.ctx, &height)
}

func (n *NodeImpl) BlockResults(height int64) (*tmctypes.ResultBlockResults, error) {
	return n.client.BlockResults(n.ctx, &height)
}

func (n *NodeImpl) Tx(hash string) (*node.Tx, error) {
	n.logger.Info("node tx", "hash", hash)
	res, err := n.txServiceClient.GetTx(context.Background(), &tx.GetTxRequest{Hash: hash})
	if err != nil {
		n.logger.Error("err get tx", "err", err)
		return nil, err
	}

	// Decode messages
	for _, msg := range res.Tx.Body.Messages {
		var stdMsg sdk.Msg
		err = n.codec.UnpackAny(msg, &stdMsg)
		if err != nil {
			return nil, fmt.Errorf("error while unpacking message: %s", err)
		}
	}

	convTx, err := node.NewTx(res.TxResponse, res.Tx)
	if err != nil {
		return nil, fmt.Errorf("error converting transaction: %s", err.Error())
	}

	return convTx, nil
}

func (n *NodeImpl) Txs(block *tmctypes.ResultBlock) ([]*node.Tx, error) {
	txResponses := make([]*node.Tx, len(block.Block.Txs))
	for i, tmTx := range block.Block.Txs {
		txResponse, err := n.Tx(fmt.Sprintf("%X", tmTx.Hash()))
		if err != nil {
			return nil, err
		}

		txResponses[i] = txResponse
	}

	return txResponses, nil
}

func (n *NodeImpl) TxSearch(query string, page *int, perPage *int, orderBy string) (*tmctypes.ResultTxSearch, error) {
	return n.client.TxSearch(n.ctx, query, false, page, perPage, orderBy)
}

func (n *NodeImpl) SubscribeEvents(subscriber, query string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	eventCh, err := n.client.Subscribe(ctx, subscriber, query)
	return eventCh, cancel, err
}

func (n *NodeImpl) SubscribeNewBlocks(subscriber string) (<-chan tmctypes.ResultEvent, context.CancelFunc, error) {
	return n.SubscribeEvents(subscriber, "tm.event = 'NewBlock'")
}

func (n *NodeImpl) Stop() {
	err := n.client.Stop()
	if err != nil {
		panic(fmt.Errorf("error while stopping proxy: %s", err))
	}

	err = n.grpcConnection.Close()
	if err != nil {
		panic(fmt.Errorf("error while closing gRPC connection: %s", err))
	}
}
