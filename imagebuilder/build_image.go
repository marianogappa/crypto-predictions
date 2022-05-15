package imagebuilder

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type PredictionImageBuilder struct {
	market market.IMarket
}

func NewPredictionImageBuilder(m market.IMarket) PredictionImageBuilder {
	return PredictionImageBuilder{m}
}

func (r PredictionImageBuilder) BuildImageBase64(prediction types.Prediction, account types.Account) (string, error) {
	fmt.Println("Building image for ", prediction, account)
	url, err := r.BuildImage(prediction, account)
	if err != nil {
		return "", err
	}
	defer os.Remove(url)

	bytes, err := ioutil.ReadFile(url)
	if err != nil {
		return "", err
	}

	var base64Encoding string

	mimeType := http.DetectContentType(bytes)
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	default:
		return "", fmt.Errorf("unsupported image mime type: %v", mimeType)
	}

	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	return base64Encoding, nil
}

func (r PredictionImageBuilder) BuildImage(prediction types.Prediction, account types.Account) (string, error) {
	fmt.Println("Building image for ", prediction, account)
	if os.Getenv("PREDICTIONS_CHROME_PATH") == "" {
		return "", errors.New("Daemon.buildImageAction: PREDICTIONS_CHROME_PATH env not set.")
	}
	if os.Getenv("PREDICTIONS_POST_IMAGE_PATH") == "" {
		return "", errors.New("Daemon.buildImageAction: PREDICTIONS_POST_IMAGE_PATH env not set.")
	}
	chromePath := os.Getenv("PREDICTIONS_CHROME_PATH")
	postImagePath := os.Getenv("PREDICTIONS_POST_IMAGE_PATH")
	indexTPLPath := fmt.Sprintf("%v/index.tpl", postImagePath)

	// Inputs
	name := account.Handle
	if name == "" {
		name = account.Name
	}
	if name == "" {
		return "", errors.New("Daemon.buildImageAction: account has no handle nor name!")
	}
	if len(account.Thumbnails) == 0 {
		return "", errors.New("Daemon.buildImageAction: account has no thumbnails!")
	}
	serializedAccount, err := compiler.NewAccountSerializer().Serialize(&account)
	if err != nil {
		return "", err
	}

	serializer := compiler.NewPredictionSerializer(&r.market)
	serializedPrediction, err := serializer.SerializeForAPI(&prediction, true)
	if err != nil {
		return "", err
	}

	t, err := template.ParseFiles(indexTPLPath)
	if err != nil {
		return "", fmt.Errorf("Daemon.buildImageAction: error parsing template file: %v", err)
	}

	randomHTMLPath := fmt.Sprintf("%v/deleteme_%v.html", postImagePath, uuid.NewString())

	f, err := os.Create(randomHTMLPath)
	if err != nil {
		return "", fmt.Errorf("Daemon.buildImageAction: error creating templated file: %v", err)
	}
	defer os.Remove(randomHTMLPath)

	config := map[string]string{
		"Prediction": string(serializedPrediction),
		"Account":    string(serializedAccount),
	}

	if err := t.Execute(f, config); err != nil {
		return "", fmt.Errorf("Daemon.buildImageAction: error executing template: %v", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("Daemon.buildImageAction: error closing templated file: %v", err)
	}

	htmlExists := waitFor(fileExists(randomHTMLPath), 3, 1*time.Second)
	if !htmlExists {
		return "", fmt.Errorf("Daemon.buildImageAction: templated html file was never created!")
	}

	randomImagePath := fmt.Sprintf("deleteme_%v.png", uuid.NewString())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		chromePath,
		"--headless",
		"--disable-gpu",
		fmt.Sprintf(`--screenshot=%v`, randomImagePath),
		randomHTMLPath,
		"--window-size=1203,678",
		"--hide-scrollbars",
		// Note: these logs are very important for debug, e.g. I found it was failing due to folder permissions.
		// "--enable-logging=stderr",
		// "--v=1",
		// `--crash-dumps-dir="/tmp"`,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("Daemon.buildImageAction: error running chrome on templated file: %v", err)
	}

	imageExists := waitFor(fileExists(randomImagePath), 3, 1*time.Second)
	if !imageExists {
		return "", fmt.Errorf("Daemon.buildImageAction: image was never created!")
	}

	return randomImagePath, nil
}

func fileExists(path string) func() bool {
	return func() bool {
		_, err := os.Stat(path)
		log.Info().Msgf("Daemon.buildImageAction: checking if %v exists: %v!", path, err == nil)
		return err == nil
	}
}

func waitFor(f func() bool, attempts int, interval time.Duration) bool {
	result := false
	for a := attempts; a > 0 && !result; a-- {
		result = f()
		if result {
			return true
		}
		log.Info().Msg("Daemon.buildImageAction: checking again shortly...")
		time.Sleep(interval)
	}
	return false
}
