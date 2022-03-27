package types

import (
	"errors"

	"github.com/marianogappa/predictions/types"
)

type PostType int

const (
	OTHER PostType = iota
	TWITTER
	YOUTUBE
)

var (
	ErrNoMetadataFound = errors.New("none of the metadata fetchers could resolve metadata for this url")
)

type PostMetadata struct {
	Author             types.Account
	PostTitle          string
	PostText           string
	PostCreatedAt      types.ISO8601
	PostType           PostType
	ThumbnailImgSmall  string
	ThumbnailImgMedium string
}
