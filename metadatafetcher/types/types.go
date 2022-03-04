package types

import (
	"errors"

	"github.com/marianogappa/signal-checker/common"
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
	Author        string
	AuthorURL     string
	AuthorImgUrl  string
	PostTitle     string
	PostText      string
	PostCreatedAt common.ISO8601
	PostType      PostType
}
