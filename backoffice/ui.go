package backoffice

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sort"
	"text/template"

	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
)

const (
	templatesDir = "public"
)

var (
	templates map[string]*template.Template
)

func loadTemplates(files embed.FS) error {
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

type backOfficeUI struct {
	apiClient APIClient
	files     embed.FS
}

func NewBackOfficeUI(files embed.FS) backOfficeUI {
	return backOfficeUI{files: files}
}

func (s backOfficeUI) MustBlockinglyServe(port int, apiUrl string) {
	s.apiClient = NewAPIClient(apiUrl)

	err := loadTemplates(s.files)
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/add", s.putHandler)
	// http.HandleFunc("/reRunAll", s.reRunAllHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func (s backOfficeUI) indexHandler(w http.ResponseWriter, r *http.Request) {
	t, ok := templates["index.html"]
	if !ok {
		log.Fatal("Couldn't find index.hmtl")
	}

	res := s.apiClient.Get()

	data := make(map[string]interface{})
	data["GetPredictionsErr"] = res.Message
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalMessage"] = res.InternalMessage

	preds := []types.Prediction{}
	if res.Predictions != nil {
		preds = *res.Predictions
	}

	pDatas := []map[string]string{}
	for _, pred := range preds {
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

func (s backOfficeUI) putHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	rawString := r.FormValue("prediction_input")

	resp := s.apiClient.New([]byte(rawString))

	t, ok := templates["add.html"]
	if !ok {
		log.Fatal("Couldn't find add.hmtl")
	}

	data := make(map[string]interface{})
	data["Err"] = resp.Message
	data["Status"] = resp.Status
	data["ErrCode"] = resp.ErrorCode
	data["InternalMessage"] = resp.InternalMessage

	pDatas := []map[string]string{}

	if resp.Status == 200 {
		pred := *resp.Prediction
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

// func (s backOfficeUI) reRunAllHandler(w http.ResponseWriter, r *http.Request) {
// 	predictions, err := s.store.GetPredictions([]types.PredictionStateValue{
// 		types.ONGOING_PRE_PREDICTION,
// 		types.ONGOING_PREDICTION,
// 		types.CORRECT,
// 		types.INCORRECT,
// 		types.ANNULLED,
// 	})
// 	if err != nil {
// 		log.Println(err)
// 		http.Redirect(w, r, "/", http.StatusSeeOther)
// 		return
// 	}
// 	for key := range predictions {
// 		pred := predictions[key]
// 		(&pred).ClearState()
// 		predictions[key] = pred
// 	}
// 	if err := s.store.UpsertPredictions(predictions); err != nil {
// 		log.Fatal(err)
// 	}
// 	http.Redirect(w, r, "/", http.StatusSeeOther)
// }
