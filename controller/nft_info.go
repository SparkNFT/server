package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SparkNFT/key_server/model"
	"github.com/gin-gonic/gin"
)

type NFTInfoRequest struct {
	Chain string `form:"chain"`
	NFTId uint64 `form:"nft_id"`
}

type NFTInfoResponse struct {
	ChildrenCount int            `json:"children_count"`
	Tree          *model.NFTTree `json:"tree"`
	Suggest       string         `json:"suggest_next_nft"`
	ShillTimes    int            `json:"shill_times"`
	MaxShillTimes int            `json:"max_shill_times"`
}

func nft_info(c *gin.Context) {
	var req NFTInfoRequest
	err := c.ShouldBindQuery(&req)
	if err != nil || req.NFTId == 0 {
		c.JSON(http.StatusBadRequest, ErrorMessage{
			Message: "Parse param error",
		})
		return
	}

	nft, err := model.FindNFT(req.Chain, req.NFTId)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorMessage{
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: fmt.Sprintf("Error when getting NFT: %s", err.Error()),
		})
		return
	}

	count, err := model.ChildrenCount(req.Chain, nft.NFTID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: fmt.Sprintf("Erorr when counting NFT children: %s", err.Error()),
		})
		return
	}

	suggest, err := nft.Suggest(nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: fmt.Sprintf("Erorr when suggesting next NFT: %s", err.Error()),
		})
		return
	}
	if suggest == nil {
		suggest = &model.NFT{NFTID: uint64(0)}
	}

	tree, err := model.ChildrenTree(req.Chain, nft.NFTID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorMessage{
			Message: fmt.Sprintf("Erorr when fetching NFT Tree: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, NFTInfoResponse{
		ChildrenCount: count,
		Tree:          tree,
		Suggest:       strconv.FormatUint(suggest.NFTID, 10),
		MaxShillTimes: int(nft.MaxShillCount),
		ShillTimes:    int(nft.ShillCount),
	})
}
