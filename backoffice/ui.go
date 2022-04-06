package backoffice

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"syscall"
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
	http.HandleFunc("/prediction", s.predictionHandler)
	http.HandleFunc("/predictionPage", s.predictionPageHandler)
	// http.HandleFunc("/reRunAll", s.reRunAllHandler)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
}

func trim(ss []string) []string {
	ns := []string{}
	for _, s := range ss {
		if len(s) > 0 {
			ns = append(ns, s)
		}
	}
	return ns
}

func (s backOfficeUI) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	rawUUIDs := r.FormValue("UUIDs")
	rawAuthors := r.FormValue("authors")
	rawStatuses := r.FormValue("statuses")
	rawStateValues := r.FormValue("stateValues")

	t, ok := templates["index.html"]
	if !ok {
		log.Fatal("Couldn't find index.hmtl")
	}

	res := s.apiClient.Get(getBody{Filters: types.APIFilters{
		AuthorHandles:         trim(strings.Split(rawAuthors, ",")),
		UUIDs:                 trim(strings.Split(rawUUIDs, ",")),
		PredictionStateValues: trim(strings.Split(rawStatuses, ",")),
		PredictionStateStatus: trim(strings.Split(rawStateValues, ",")),
	}})

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
		pData["predictionUUID"] = pred.UUID
		pData["predictionUrl"] = pred.PostUrl
		pData["predictionText"] = printer.NewPredictionPrettyPrinter(pred).Default()
		pData["predictionAuthor"] = pred.PostAuthor
		pData["predictionStatus"] = pred.State.Status.String()
		pData["predictionValue"] = pred.State.Value.String()
		pDatas = append(pDatas, pData)
	}
	sort.Slice(pDatas, func(i, j int) bool { return pDatas[i]["predictionCreatedAt"] > pDatas[j]["predictionCreatedAt"] })

	data["Predictions"] = pDatas

	t.Execute(w, data)
}

func (s backOfficeUI) predictionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	uuid := r.FormValue("uuid")
	action := r.FormValue("action")

	t, ok := templates["prediction.html"]
	if !ok {
		log.Fatal("Couldn't find prediction.html")
	}

	var res parsedResponse
	switch action {
	case "pause":
		res = s.apiClient.PausePrediction(uuid)
	case "unpause":
		res = s.apiClient.UnpausePrediction(uuid)
	case "hide":
		res = s.apiClient.HidePrediction(uuid)
	case "unhide":
		res = s.apiClient.UnhidePrediction(uuid)
	case "delete":
		res = s.apiClient.DeletePrediction(uuid)
	case "undelete":
		res = s.apiClient.UndeletePrediction(uuid)
	case "refreshAccount":
		res = s.apiClient.RefreshAccount(uuid)
	default:
		res.Status = 200
	}

	if res.Status == 200 {
		res = s.apiClient.Get(getBody{Filters: types.APIFilters{
			UUIDs: []string{uuid},
		}})
	}

	data := make(map[string]interface{})
	if res.Predictions != nil && len(*res.Predictions) == 1 {
		pred := (*res.Predictions)[0]
		data["prediction"] = predictionToMap(pred)
	}

	data["GetPredictionsErr"] = res.Message
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalMessage"] = res.InternalMessage

	if err := t.Execute(w, data); err != nil {
		if nErr, ok := err.(*net.OpError); !ok || nErr.Err != syscall.EPIPE {
			log.Fatal(err)
		}
	}
}

func (s backOfficeUI) predictionPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	url := r.FormValue("url")

	t, ok := templates["prediction_page.html"]
	if !ok {
		log.Fatal("Couldn't find prediction.html")
	}

	res := s.apiClient.PredictionPage(getBody{Filters: types.APIFilters{
		URLs: []string{url},
	}})

	data := make(map[string]interface{})
	if res.Predictions != nil && len(*res.Predictions) == 1 {
		pred := (*res.Predictions)[0]
		data["prediction"] = predictionToMap(pred)
	}
	if res.PredictionSummary != nil {
		data["predictionSummary"] = predictionSummaryToMap(*res.PredictionSummary)
	}

	data["GetPredictionsErr"] = res.Message
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalMessage"] = res.InternalMessage

	if err := t.Execute(w, data); err != nil {
		if nErr, ok := err.(*net.OpError); !ok || nErr.Err != syscall.EPIPE {
			log.Fatal(err)
		}
	}
}

func (s backOfficeUI) putHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	rawString := r.FormValue("prediction_input")
	storeStr := r.FormValue("store")

	t, ok := templates["add.html"]
	if !ok {
		log.Fatal("Couldn't find add.hmtl")
	}

	data := make(map[string]interface{})
	data["Err"] = ""
	data["Predictions"] = []map[string]string{}

	if rawString != "" {

		store := false
		if strings.ToLower(storeStr) == "on" {
			store = true
		}

		resp := s.apiClient.New([]byte(rawString), store)

		data["Err"] = resp.Message
		data["Status"] = resp.Status
		data["ErrCode"] = resp.ErrorCode
		data["InternalMessage"] = resp.InternalMessage

		if resp.Stored != nil {
			data["Stored"] = *resp.Stored
		}

		if resp.Status == 200 {
			pred := *resp.Prediction
			data["prediction"] = predictionToMap(pred)
			data["predictionStr"] = rawString
		}
	}

	if err := t.Execute(w, data); err != nil {
		if nErr, ok := err.(*net.OpError); !ok || nErr.Err != syscall.EPIPE {
			log.Fatal(err)
		}
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
