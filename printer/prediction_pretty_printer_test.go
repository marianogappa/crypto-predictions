package printer

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/stretchr/testify/require"
)

func TestPrettyPrinter(t *testing.T) {
	var (
		rawPredictions = readGoldenPredictions(t)
		prettyPrints   = readGoldenPrettyPrints(t)
		compiler       = compiler.NewPredictionCompiler(nil, time.Now)
	)

	for i, rawPrediction := range rawPredictions {
		if rawPrediction == "" {
			continue
		}
		prediction, _, err := compiler.Compile([]byte(rawPrediction))
		if err != nil {
			continue
		}
		require.Equal(t, prettyPrints[i], NewPredictionPrettyPrinter(prediction).String(), "on %v", i)
	}

}

func readGoldenPredictions(t *testing.T) []string {
	content, err := ioutil.ReadFile("testdata/prediction_input.golden")
	if err != nil {
		t.Fatalf("Error loading golden file: %s", err)
	}
	return strings.Split(string(content), "\n")
}

func readGoldenPrettyPrints(t *testing.T) []string {
	content, err := ioutil.ReadFile("testdata/prediction_pretty_printed.golden")
	if err != nil {
		t.Fatalf("Error loading golden file: %s", err)
	}
	return strings.Split(string(content), "\n")
}
