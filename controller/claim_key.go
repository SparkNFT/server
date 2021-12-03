package controller

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SparkNFT/key_server/chain"
	"github.com/SparkNFT/key_server/model"
	"github.com/SparkNFT/key_server/pinata"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type ClaimKeyRequest struct {
	Chain     string `json:"chain"`
	NFTId     string `json:"nft_id"`
	Account   string `json:"account"`
	Signature string `json:"signature"`
}

type ClaimKeyResponse struct {
	Key    string                 `json:"key"`
	Pinata ClaimKeyPinataResponse `json:"pinata"`
}

type ClaimKeyPinataResponse struct {
	Key    string `json:"api_key"`
	Secret string `json:"api_secret"`
}

func claim_key(c *gin.Context) {
	// JSON Unmarshal req
	req := ClaimKeyRequest{}
	err := c.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: err.Error(),
		})
		return
	}

	// Check sig
	if !claim_key_check_signature(&req) {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: "signature invalid",
		})
		return
	}

	// Param validation
	if claim_key_param_invalid(&req) {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: "param invalid",
		})
		return
	}
	nft_id, err := strconv.ParseUint(req.NFTId, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: "param invalid",
		})
		return
	}
	root_nft_id := chain.RootNFTIdOf(nft_id)

	// NFT ownership
	if err = claim_key_check_nft(req.Chain, req.Account, nft_id); err != nil {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: err.Error(),
		})
		return
	}

	// Get key
	key, err := model.GetKey(req.Chain, root_nft_id)
	if err == nil {
		// Success. Just return the key.
		c.JSON(http.StatusOK, ClaimKeyResponse{
			Key: key,
		})
		return
	}

	if !strings.Contains(err.Error(), "NFT not found") {
		// Something unexped happens
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: err.Error(),
		})
		return
	}

	// GetKey returns Not Found. So we create one.
	// Before creating, make sure only Root NFT can create this key.
	if root_nft_id != nft_id {
		c.JSON(http.StatusNotFound, ErrorMessage{
			Message: "Key haven't generated. Please wait for root owner to create this.",
		})
		return
	}

	// All set. Create this.
	key_instance, err := model.CreateKey(req.Chain, req.Account, root_nft_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: err.Error(),
		})
		return
	}

	// Generate Pinata upload key
	pinata_key, err := pinata.GenerateAPIKey(req.Chain, nft_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: "Pinata error: " + err.Error(),
		})
		return
	}
	// Revoke this key after 5min
	go func(pinata_api_key string) {
		time.Sleep(5 * time.Minute)
		pinata.RevokeAPIKey(pinata_api_key)
	}(pinata_key.PinataAPIKey)

	c.JSON(http.StatusCreated, ClaimKeyResponse{
		Key: key_instance.Key,
		Pinata: ClaimKeyPinataResponse{
			Key: pinata_key.PinataAPIKey,
			Secret: pinata_key.PinataAPISecret,
		},
	})
}

func claim_key_param_invalid(req *ClaimKeyRequest) bool {
	return (req.NFTId == "0" || req.Chain == "" || req.NFTId == "" || req.Signature == "" || req.Account == "")
}

func claim_key_check_signature(req *ClaimKeyRequest) bool {
	signature_raw_payload := gin.H{
		"account": req.Account,
		"chain":   req.Chain,
		"nft_id":  req.NFTId,
	}
	signature_payload, _ := json.Marshal(signature_raw_payload)
	result, _ := chain.ValidateSignature(
		string(signature_payload),
		req.Signature,
		common.HexToAddress(req.Account),
	)
	return result
}

// claim_key_check_nft checks if nft_id is exists and is owned by this account
func claim_key_check_nft(chainName string, account string, nft_id uint64) error {
	contract, _, err := chain.Init(chainName)

	result, err := chain.IsOwnerOfNFT(contract, big.NewInt(int64(nft_id)), common.HexToAddress(account))

	if err != nil {
		if strings.Contains(err.Error(), "nonexistent") {
			return fmt.Errorf("not found")
		}
		return err
	}
	if !result {
		return fmt.Errorf("not owned")
	}
	return nil
}
