package messari

import "github.com/marianogappa/predictions/types"

type Messari struct {
	apiURL string
	debug  bool
	apiKey string
}

func NewMessari() *Messari {
	return &Messari{apiURL: "https://data.messari.io/api/v1/", apiKey: "1ec22c58-744e-4453-93c6-ad73e2641054"}
}

func (m *Messari) SetDebug(debug bool) {
	m.debug = debug
}

func (m Messari) BuildTickIterator(asset, metricID string, initialISO8601 types.ISO8601) *messariTickIterator {
	return m.newTickIterator(asset, metricID, initialISO8601)
}
