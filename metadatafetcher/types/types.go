package types

import (
	"errors"

	"github.com/marianogappa/predictions/types"
)

// PostType is the website used for the post, e.g. Twitter, Youtube.
type PostType int

const (
	// OTHER is a PostType
	OTHER PostType = iota
	// TWITTER is a PostType
	TWITTER
	// YOUTUBE is a PostType
	YOUTUBE
)

var (
	// ErrNoMetadataFound means: none of the metadata fetchers could resolve metadata for this url.
	ErrNoMetadataFound = errors.New("none of the metadata fetchers could resolve metadata for this url")
)

// PostMetadata contains the metadata gathered about a post by querying a website's API about it.
type PostMetadata struct {
	Author             types.Account
	PostTitle          string
	PostText           string
	PostCreatedAt      types.ISO8601
	PostType           PostType
	ThumbnailImgSmall  string
	ThumbnailImgMedium string
}
