package imagebuilder

import (
	"context"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/serializer"
	"github.com/rs/zerolog/log"
)

// PredictionImageBuilder builds images based on Predictions, including candlestick charts, post author image, etc.
type PredictionImageBuilder struct {
	market     core.IMarket
	templates  map[string]*template.Template
	chromePath string
}

// NewPredictionImageBuilder constructs a PredictionImageBuilder.
func NewPredictionImageBuilder(m core.IMarket, files embed.FS, chromePath string) PredictionImageBuilder {
	templates, _ := loadTemplates(files)
	return PredictionImageBuilder{market: m, templates: templates, chromePath: chromePath}
}

// BuildImageBase64 builds the prediction image and returns it as a base64-encoded-bytestream string.
//
// Successfully building depends on having a valid Chrome installation, filesystem write access, being able to
// execute the Chrome binary. Prediction must be well-formed. Building candlestick chart depends on Internet access,
// and the market communicates with an exchange's API that might be offline, might decide to 429, etc. A lot can
// go wrong unfortunately.
func (r PredictionImageBuilder) BuildImageBase64(prediction core.Prediction, account core.Account) (string, error) {
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

// BuildImage builds the prediction image and returns the filesystem's path to the built image. It is your
// responsibility to delete it once you're done using it, or otherwise you might fill your disk eventually!
//
// Successfully building depends on having a valid Chrome installation, filesystem write access, being able to
// execute the Chrome binary. Prediction must be well-formed. Building candlestick chart depends on Internet access,
// and the market communicates with an exchange's API that might be offline, might decide to 429, etc. A lot can
// go wrong unfortunately.
func (r PredictionImageBuilder) BuildImage(prediction core.Prediction, account core.Account) (string, error) {
	if r.chromePath == "" {
		return "", errors.New("daemon.buildImageAction: PREDICTIONS_CHROME_PATH env not set")
	}

	// Inputs
	name := account.Handle
	if name == "" {
		name = account.Name
	}
	if name == "" {
		return "", errors.New("daemon.buildImageAction: account has no handle nor name")
	}
	if len(account.Thumbnails) == 0 {
		return "", errors.New("daemon.buildImageAction: account has no thumbnails")
	}
	serializedAccount, err := serializer.NewAccountSerializer().Serialize(&account)
	if err != nil {
		return "", err
	}

	serializer := serializer.NewPredictionSerializer(&r.market)
	serializedPrediction, err := serializer.SerializeForAPI(&prediction, true)
	if err != nil {
		return "", err
	}

	t, ok := r.templates["index.tpl"]
	if !ok {
		return "", fmt.Errorf("Daemon.buildImageAction: template file not found: index.tpl")
	}

	randomHTMLPath := fmt.Sprintf("deleteme_%v.html", uuid.NewString())

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
		return "", fmt.Errorf("Daemon.buildImageAction: templated html file was never created")
	}

	randomImagePath := fmt.Sprintf("deleteme_%v.png", uuid.NewString())

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()

	cmd := exec.CommandContext(
		ctx,
		r.chromePath,
		"--headless",
		"--disable-gpu",

		// Otherwise on Ubuntu 20 (Heroku) it complains when generating the file (writing to filesystem)
		"--no-sandbox",

		// Otherwise on Ubuntu 20 (Heroku) the image comes out blurry
		"--force-device-scale-factor=2",

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
		return "", fmt.Errorf("Daemon.buildImageAction: image was never created")
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

func loadTemplates(files embed.FS) (map[string]*template.Template, error) {
	var (
		templatesDir = "public"
		templates    map[string]*template.Template
	)
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		return nil, err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name())
		if err != nil {
			return nil, err
		}

		templates[tmpl.Name()] = pt
	}
	return templates, nil
}
