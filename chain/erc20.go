package chain

import (
	"github.com/SparkNFT/key_server/abi"
	"github.com/SparkNFT/key_server/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/xerrors"
)

func GetERC20Info(erc20addr string) (name, symbol string, err error) {
	// FIXME: chain-switchable
	client, err := ethclient.Dial(config.C.Chain["etereum"].RPCUrl)
	if err != nil {
		return "", "", xerrors.Errorf("%w", err)
	}
	erc20, err := abi.NewERC20(common.HexToAddress(erc20addr), client)
	if err != nil {
		return "", "", xerrors.Errorf("%w", err)
	}
	name, err = erc20.Name(nil)
	if err != nil {
		return "", "", xerrors.Errorf("%w", err)
	}
	symbol, err = erc20.Symbol(nil)
	if err != nil {
		return "", "", xerrors.Errorf("%w", err)
	}
	return name, symbol, nil
}
