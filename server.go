package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"text/template"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/statestorage"
	"github.com/marianogappa/predictions/types"
)

const (
	templatesDir = "public"
)

var (
	//go:embed public/*
	files     embed.FS
	templates map[string]*template.Template
)

func loadTemplates() error {
	if templates == nil {
		templates = make(map[string]*template.Template)
	}
	tmplFiles, err := fs.ReadDir(files, templatesDir)
	if err != nil {
		return err
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(files, templatesDir+"/"+tmpl.Name())
		if err != nil {
			return err
		}

		templates[tmpl.Name()] = pt
	}
	return nil
}

type server struct {
	store  statestorage.StateStorage
	market market.Market
}

func newServer(s statestorage.StateStorage, m market.Market) server {
	return server{store: s, market: m}
}

func (s server) mustBlockinglyServeAll(port int) {
	err := loadTemplates()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/add", s.putHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func (s server) indexHandler(w http.ResponseWriter, r *http.Request) {
	t, ok := templates["index.html"]
	if !ok {
		log.Fatal("Couldn't find index.hmtl")
	}

	predictions, err := s.store.GetPredictions([]types.PredictionStateValue{
		types.ONGOING_PRE_PREDICTION,
		types.ONGOING_PREDICTION,
		types.CORRECT,
		types.INCORRECT,
		types.ANNULLED,
	})

	data := make(map[string]interface{})
	data["GetPredictionsErr"] = err

	pDatas := []map[string]string{}
	for _, pred := range predictions {
		pData := map[string]string{}
		pData["predictionUrl"] = pred.PostUrl
		pData["predictionText"] = printer.NewPredictionPrettyPrinter(pred).Default()
		pData["predictionAuthor"] = pred.PostAuthor
		pData["predictionStatus"] = pred.State.Status.String()
		pData["predictionValue"] = pred.State.Value.String()
		pDatas = append(pDatas, pData)
	}
	data["Predictions"] = pDatas

	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func (s server) putHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	rawString := r.FormValue("prediction_input")
	var prediction types.Prediction
	if err := json.Unmarshal([]byte(rawString), &prediction); err != nil {
		log.Fatal(err)
	}
	log.Println(prediction)
	if err := s.store.UpsertPredictions(map[string]types.Prediction{"unused": prediction}); err != nil {
		log.Fatal(err)
	}
	predictions, err := s.store.GetPredictions([]types.PredictionStateValue{
		types.ONGOING_PRE_PREDICTION,
		types.ONGOING_PREDICTION,
		types.CORRECT,
		types.INCORRECT,
		types.ANNULLED,
	})
	log.Println(predictions, err)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
