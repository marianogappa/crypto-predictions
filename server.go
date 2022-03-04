package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"text/template"

	"github.com/marianogappa/predictions/compiler"
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
	http.HandleFunc("/reRunAll", s.reRunAllHandler)
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
		pData["predictionCreatedAt"] = string(pred.CreatedAt)
		pData["predictionUrl"] = pred.PostUrl
		pData["predictionText"] = printer.NewPredictionPrettyPrinter(pred).Default()
		pData["predictionAuthor"] = pred.PostAuthor
		pData["predictionStatus"] = pred.State.Status.String()
		pData["predictionValue"] = pred.State.Value.String()
		pDatas = append(pDatas, pData)
	}
	sort.Slice(pDatas, func(i, j int) bool { return pDatas[i]["predictionCreatedAt"] > pDatas[j]["predictionCreatedAt"] })

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

	prediction, err := compiler.NewPredictionCompiler().Compile([]byte(rawString))
	if err == nil {
		err = s.store.UpsertPredictions(map[string]types.Prediction{"unused": prediction})
	}

	t, ok := templates["add.html"]
	if !ok {
		log.Fatal("Couldn't find add.hmtl")
	}

	data := make(map[string]interface{})
	data["Err"] = err

	pDatas := []map[string]string{}

	if err == nil {
		pred := prediction
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

func (s server) reRunAllHandler(w http.ResponseWriter, r *http.Request) {
	predictions, err := s.store.GetPredictions([]types.PredictionStateValue{
		types.ONGOING_PRE_PREDICTION,
		types.ONGOING_PREDICTION,
		types.CORRECT,
		types.INCORRECT,
		types.ANNULLED,
	})
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	for key := range predictions {
		pred := predictions[key]
		(&pred).ClearState()
		predictions[key] = pred
	}
	if err := s.store.UpsertPredictions(predictions); err != nil {
		log.Fatal(err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
