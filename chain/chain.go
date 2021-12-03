package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/SparkNFT/key_server/abi"
	"github.com/SparkNFT/key_server/config"
	"github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum"
	ethabi "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/xerrors"
)

const (
	EventPublishName  = "Publish"
	EventTransferName = "Transfer"
	EventMintName     = "Mint"
)

var (
	ContractAddress map[string]common.Address

	EventPublishHash  = common.HexToHash("0x072ee21d81ebd9fc5f68a2c36d04cbbd9eff1e2567a48dd7ecce61d5af159fad")
	EventTransferHash = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

// Init generates ethclient and contract instance.
func Init(chainName string) (contract *abi.SparkLink, client *ethclient.Client, err error) {
	client, err = ethclient.Dial(config.C.Chain[chainName].RPCUrl)
	if err != nil {
		return nil, nil, xerrors.Errorf("error when dialing client: %w", err)
	}
	contract, err = abi.NewSparkLink(common.HexToAddress(config.C.Chain[chainName].ContractAddress), client)
	if err != nil {
		return nil, nil, xerrors.Errorf("error when initlizing client: %w", err)
	}

	if ContractAddress == nil {
		ContractAddress = make(map[string]common.Address, 0)
	}
	ContractAddress[chainName] = common.HexToAddress(config.C.Chain[chainName].ContractAddress)

	return contract, client, nil
}

// GenerateTrasactOpts returns a transact option to be used when
// interacting with contract.
func GenerateTransactOps(client *ethclient.Client, private_hex string) (auth *bind.TransactOpts, err error) {
	chain_id, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	private_key, err := crypto.HexToECDSA(private_hex)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	auth, err = bind.NewKeyedTransactorWithChainID(private_key, chain_id)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	nonce, err := client.PendingNonceAt(context.Background(), auth.From)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))

	gas_tip_cap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	auth.GasTipCap = gas_tip_cap
	auth.GasLimit = uint64(10000000)

	return auth, nil
}

// BlockHashOf returns block hash of a given block number.
func BlockHashOf(client *ethclient.Client, block_number *big.Int) (block_hash common.Hash, err error) {
	block_header, err := client.HeaderByNumber(context.TODO(), block_number)
	if err != nil {
		return common.HexToHash("0x"), xerrors.Errorf("error when fetching block [%v] header: %w", block_number, err)
	}
	return block_header.Hash(), nil
}

// IsOwnerOfNFT returns if a nft_id is owned by an address.
func IsOwnerOfNFT(contract *abi.SparkLink, nft_id *big.Int, address common.Address) (bool, error) {
	found_address, err := contract.OwnerOf(nil, nft_id)
	if err != nil {
		return false, xerrors.Errorf("error when calling ownerOf: %w", err)
	}

	return (found_address.Hex() == address.Hex()), nil
}

// ParseNFTId partitions NFTId into IssueID and EditionID
func ParseNFTId(nft_id uint64) (issue_id, edition_id uint32) {
	issue_id = uint32(nft_id >> 32)
	edition_id = uint32(nft_id)
	return issue_id, edition_id
}

// RootNFTIdOf returns the root NFT of the given nft_id (off-chain).
func RootNFTIdOf(nft_id uint64) (root uint64) {
	issue_id, _ := ParseNFTId(nft_id)
	root = (uint64(issue_id) << 32) + uint64(1)
	return root
}

// GetParentOf returns the parent of given NFT ID
func GetParentOf(contract *abi.SparkLink, nft_id uint64) (parent uint64, err error) {
	parent, err = contract.GetFatherByNFTId(nil, nft_id)

	return parent, err
}

// GetAllLogsOf returns all logs of this contract in a given block,
func GetAllLogsOf(client *ethclient.Client, chainName string, block_number *big.Int) (logs []types.Log, err error) {
	logrus.WithFields(logrus.Fields{"chain": chainName, "block_number": block_number.Uint64()}).Debugf("Fetching block headers")
	block_header, err := client.HeaderByNumber(context.TODO(), block_number)
	if err != nil {
		// typical value:
		// not found
		// unknown block
		return nil, xerrors.Errorf("error when heading block %d: %w", block_number.Uint64(), err)
	}
	// block_hash := block_header.Hash()
	// logrus.WithField("block_hash", block_hash.Hex()).Debugf("Block hash fetched")

	filter_query := ethereum.FilterQuery{
		// BlockHash: &block_hash,
		// FromBlock will not return err if block is not found
		FromBlock: block_header.Number,
		ToBlock: block_header.Number,
		Addresses: []common.Address{ContractAddress[chainName]},
		Topics: [][]common.Hash{{
			EventTransferHash,
			EventPublishHash,
		}},
	}

	logs, err = client.FilterLogs(context.TODO(), filter_query)
	if err != nil {
		return nil, xerrors.Errorf("error when getting all logs: %w", err)
	}
	return logs, nil
}

// FilterEventPublish
func FilterEventPublish(contract *abi.SparkLink, logs []types.Log) ([]abi.SparkLinkPublish, error) {
	contract_abi, err := ethabi.JSON(strings.NewReader(string(abi.SparkLinkABI)))
	result := make([]abi.SparkLinkPublish, 0, 1)

	if err != nil {
		return nil, xerrors.Errorf("error when parsing contract ABI: %w", err)
	}

	for _, log := range logs {
		if log.Topics[0] != EventPublishHash {
			continue
		}

		event := new(abi.SparkLinkPublish)
		err := contract_abi.UnpackIntoInterface(event, EventPublishName, log.Data)
		if err != nil {
			return nil, xerrors.Errorf("error when unpacking AcceptShill event: %w", err)
		}
		//    event Publish(
		//         address indexed publisher,
		//         uint64  indexed rootNFTId,
		//         address token_addr
		//    );
		event.Publisher = common.HexToAddress(log.Topics[1].Hex())
		event.RootNFTId = log.Topics[2].Big().Uint64()
		event.TokenAddr = common.BytesToAddress(log.Data)

		event.Raw = log
		result = append(result, *event)
	}

	return result, nil
}

// FilterEventTransfer
func FilterEventTransfer(contract *abi.SparkLink, logs []types.Log) ([]abi.SparkLinkTransfer, error) {
	result := make([]abi.SparkLinkTransfer, 0, 1)

	for _, log := range logs {
		if log.Topics[0] != EventTransferHash {
			continue
		}

		event := new(abi.SparkLinkTransfer)
		// event Transfer(
		//   address indexed from,
		//   address indexed to,
		//   uint256 indexed TokenID,
		// );
		event.From = common.HexToAddress(log.Topics[1].Hex())
		event.To = common.HexToAddress(log.Topics[2].Hex())
		event.TokenId = log.Topics[3].Big()
		event.Raw = log
		result = append(result, *event)
	}

	return result, nil
}

// ValidateSignature validates if a string is signed by given address
func ValidateSignature(pl, sig string, address common.Address) (result bool, err error) {
	payload := []byte(pl)
	signature := hexutil.MustDecode(sig)
	if signature[64] != 27 && signature[64] != 28 {
		return false, fmt.Errorf("Signature Recovery id not supported")
	}
	signature[64] -= 27

	hashed_payload := sign_hash(payload)
	pubkey, err := crypto.SigToPub(hashed_payload, signature)
	if err != nil {
		return false, xerrors.Errorf("error when validating signature: %w", err)
	}
	address_recovered := crypto.PubkeyToAddress(*pubkey)
	logrus.Infof("address: %v", address)
	logrus.Infof("recover: %v", address_recovered)
	return (address == address_recovered), nil
}

func sign_hash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}
