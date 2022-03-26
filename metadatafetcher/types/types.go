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
	Author             PostAuthor
	PostTitle          string
	PostText           string
	PostCreatedAt      types.ISO8601
	PostType           PostType
	ThumbnailImgSmall  string
	ThumbnailImgMedium string
}

type PostAuthor struct {
	URL               string
	AuthorImgSmall    string
	AuthorImgMedium   string
	AuthorName        string
	AuthorHandle      string
	AuthorDescription string
	IsVerified        bool // Youtube API currently does not return this field so it's always false
	FollowerCount     int
}
