package controller

import (
	"net/http"
	"strconv"

	"github.com/SparkNFT/key_server/model"
	"github.com/gin-gonic/gin"
	"xorm.io/builder"
)

type NFTListRequest struct {
	Owner string `form:"owner"`
	Chain string `form:"chain"`
}

type NFTListResponse struct {
	NFT []string `json:"nft"`
}

// nft_list returns all NFT owned by a wallet
func nft_list(c *gin.Context) {
	var req NFTListRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: "Parse param error",
		})
		return
	}

	nfts := make([]string, 0, 10)
	err = model.Engine.Where(builder.Eq{"owner": req.Owner, "chain": req.Chain}).Iterate(new(model.NFT), func(i int, bean interface{})error{
		nft := bean.(*model.NFT)
		nfts = append(nfts, strconv.FormatUint(nft.NFTID, 10))
		return nil
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: "Error when fetching user NFTs",
		})
		return
	}

	c.JSON(http.StatusOK, NFTListResponse{
		NFT: nfts,
	})
}
