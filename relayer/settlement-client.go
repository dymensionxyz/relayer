package relayer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dymensionxyz/cosmosclient/cosmosclient"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/ignite/cli/ignite/pkg/cosmosaccount"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	addressPrefix      = "dym"
	defaultNodeAddress = "http://localhost:26657"
)

// SettlementClient is the client for the Dymension Hub.
type SettlementClient struct {
	config             *Config
	client             cosmosclient.Client
	ctx                context.Context
	rollappQueryClient rollapptypes.QueryClient
}

// Config for the DymensionLayerClient
type Config struct {
	KeyringBackend cosmosaccount.KeyringBackend `json:"keyring_backend"`
	NodeAddress    string                       `json:"node_address"`
	KeyRingHomeDir string                       `json:"keyring_home_dir"`
	DymAccountName string                       `json:"dym_account_name"`
	RollappID      string                       `json:"rollapp_id"`
}

func NewSettlementClient(config []byte) (*SettlementClient, error) {
	ctx := context.Background()

	conf, err := getConfig(config)
	if err != nil {
		return nil, err
	}

	cosmosClient, err := cosmosclient.New(ctx, getCosmosClientOptions(conf)...)
	if err != nil {
		return nil, err
	}

	return &SettlementClient{
		ctx:                ctx,
		config:             conf,
		client:             cosmosClient,
		rollappQueryClient: rollapptypes.NewQueryClient(cosmosClient.Context()),
	}, nil
}

// GetLatestFinalizedStateHeight returns the latest-finalized-state height of the active rollapp
func (sc *SettlementClient) GetLatestFinalizedStateHeight(rollapID string) (int64, error) {
	latestFinalizedStateInfoResponse, err := sc.rollappQueryClient.LatestFinalizedStateInfo(sc.ctx,
		&rollapptypes.QueryGetLatestFinalizedStateInfoRequest{RollappId: rollapID})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return -1, nil
		}
		return -1, err
	}
	if latestFinalizedStateInfoResponse == nil {
		return -1, fmt.Errorf("can't get latest-finalized-state info")
	}
	return int64(latestFinalizedStateInfoResponse.StateInfo.StartHeight), nil
}

func decodeConfig(config []byte) (*Config, error) {
	var c Config
	err := json.Unmarshal(config, &c)
	return &c, err
}

func getConfig(config []byte) (*Config, error) {
	var c *Config
	if len(config) > 0 {
		var err error
		c, err = decodeConfig(config)
		if err != nil {
			return nil, err
		}
	} else {
		c = &Config{
			KeyringBackend: cosmosaccount.KeyringTest,
			NodeAddress:    defaultNodeAddress,
		}
	}
	return c, nil
}

func getCosmosClientOptions(config *Config) []cosmosclient.Option {
	options := []cosmosclient.Option{
		cosmosclient.WithAddressPrefix(addressPrefix),
		cosmosclient.WithNodeAddress(config.NodeAddress),
	}
	if config.KeyringBackend != "" {
		options = append(options,
			cosmosclient.WithKeyringBackend(config.KeyringBackend),
			cosmosclient.WithHome(config.KeyRingHomeDir))
	}
	return options
}
