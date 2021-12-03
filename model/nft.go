package model

import (
	"strconv"
	"time"

	"golang.org/x/xerrors"
	"xorm.io/builder"
)

type NFT struct {
	Id            uint64 `xorm:"BIGINT pk autoincr"`
	Chain         string `xorm:"'chain' index notnull"`
	NFTID         uint64 `xorm:"'nft_id' index notnull"`
	Parent        uint64 `xorm:"'parent' index"`
	ShillCount    uint16 `xorm:"'shill_count' default(0)"`
	MaxShillCount uint16 `xorm:"'max_shill_count' notnull"`
	Owner         string `xorm:"'owner' index notnull"`
	TokenAddr     string `xorm:"'token_addr' index notnull default('0x0')"`

	CreatedAt time.Time `xorm:"created 'created_at'"`
	UpdatedAt time.Time `xorm:"updated 'updated_at'"`
}

type NFTTree struct {
	NFTID    string     `json:"nft_id"`
	Children []*NFTTree `json:"children"`
}

func (NFT) TableName() string {
	return "nft"
}

// `NFTId = (IssueId(32bit) << 32) | EditionId(32bit)`
func (nft NFT) IssueId() uint32 {
	return uint32(nft.NFTID >> 32)
}

// `NFTId = (IssueId(32bit) << 32) | EditionId(32bit)`
func (nft NFT) EditionId() uint32 {
	return uint32(nft.NFTID)
}

func (nft NFT) IsRoot() bool {
	return nft.EditionId() == uint32(1)
}

func (nft NFT) Children() (children []*NFT, err error) {
	children = make([]*NFT, 0, 10)
	err = Engine.Where(builder.Eq{"parent": nft.NFTID, "chain": nft.Chain}).Find(&children)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}
	return children, nil
}

func (nft NFT) ChildrenCount() (count int64) {
	count, _ = Engine.Where(builder.Eq{"chain": nft.Chain, "parent": nft.NFTID}).Count(&NFT{})
	return count
}

// CanShill detects if a NFT can be shilled.
func (nft NFT) CanShill() (result bool) {
	return nft.MaxShillCount > nft.ShillCount
}

// FindNFT returns a NFT instance by nft_id.
func FindNFT(chainName string, nft_id uint64) (nft *NFT, err error) {
	nft = &NFT{Chain: chainName, NFTID: nft_id}
	found, err := Engine.Get(nft)
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	if !found {
		return nil, xerrors.Errorf("NFT not found for nft_id %d", nft_id)
	}

	return nft, nil
}

func ChildrenCount(chainName string, nft_id uint64) (count int, err error) {
	// children := make([]uint64, 0, 10)
	children := []uint64{nft_id}
	count = 0

	for len(children) != 0 {
		children_instances := make([]NFT, 0, 10)
		err := Engine.Where(builder.In("parent", children).And(builder.Eq{"chain": chainName})).Select("nft_id").Find(&children_instances)
		if err != nil {
			return 0, xerrors.Errorf("%w", err)
		}
		children = make([]uint64, 0, 10)
		for _, nft := range children_instances {
			children = append(children, nft.NFTID)
		}
		count += len(children)
	}

	return count, nil
}

func ChildrenTree(chainName string, nftId uint64) (tree *NFTTree, err error) {
	root := &NFTTree{
		NFTID:    strconv.FormatUint(nftId, 10),
		Children: []*NFTTree{},
	}
	traverseList := make([]*NFTTree, 0, 20)
	traverseList = append(traverseList, root)

	for i := 0; i < len(traverseList); i++ {
		children := make([]NFT, 0, 10)
		current := traverseList[i]
		nft_id, err := strconv.ParseUint(current.NFTID, 10, 0)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		err = Engine.Where(builder.Eq{"parent": nft_id, "chain": chainName}).Select("nft_id").Find(&children)
		if err != nil {
			return nil, xerrors.Errorf("%w", err)
		}
		for _, nft := range children {
			leaf := &NFTTree{
				NFTID:    strconv.FormatUint(nft.NFTID, 10),
				Children: []*NFTTree{},
			}
			current.Children = append(current.Children, leaf)
			traverseList = append(traverseList, leaf)
		}
	}

	return root, nil
}

// Suggest suggests next buyable NFT if current NFT is full-shilled (recursively), returns an NFT w/ same owner in children.
func (nft *NFT) Suggest(base_nft *NFT) (next_nft *NFT, err error) {
	// Beginning of recursive
	if base_nft == nil {
		base_nft = nft
	}

	// Current NFT has room to shill. Return itself.
	if nft.CanShill() && nft.Owner == base_nft.Owner {
		return nft, nil
	}

	children, err := nft.Children()
	if err != nil {
		return nil, xerrors.Errorf("%w", err)
	}

	// Breadth-first search. Strict check Owner for maximum owner profit.
	for _, child := range children {
		if child.CanShill() && child.Owner == base_nft.Owner {
			return child, nil
		}
	}

	// Search again, emit owner check.
	for _, child := range children {
		if child.CanShill() {
			return child, nil
		}
	}

	// Current depth not found. Start recursion.
	for _, child := range children {
		next_nft, err = child.Suggest(base_nft)
		if err != nil {
			return nil, err
		}
		if next_nft != nil {
			return next_nft, nil
		}
	}

	// Not found.
	return nil, nil
}
