package backoffice

import (
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"strings"
	"syscall"
	"text/template"

	"github.com/rs/zerolog/log"

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

// UI is the main struct for the BackOffice component.
type UI struct {
	apiClient     *apiClient
	files         embed.FS
	debug         bool
	basicAuthUser string
	basicAuthPass string
}

// NewBackOfficeUI is the constructor for the BackOffice component.
func NewBackOfficeUI(files embed.FS, basicAuthUser, basicAuthPass string) *UI {
	return &UI{files: files, basicAuthUser: basicAuthUser, basicAuthPass: basicAuthPass}
}

// SetDebug enables/disables debugging logs across the BackOffice component.
func (s *UI) SetDebug(b bool) {
	s.debug = b
}

// MustBlockinglyServe serves the BackOffice component.
func (s UI) MustBlockinglyServe(port int, apiURL string) {
	s.apiClient = newAPIClient(apiURL, s.basicAuthUser, s.basicAuthPass)

	if s.debug {
		s.apiClient.setDebug(s.debug)
	}

	err := loadTemplates(s.files)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	s.handleFunc("/", s.indexHandler)
	s.handleFunc("/add", s.putHandler)
	s.handleFunc("/prediction", s.predictionHandler)
	s.handleFunc("/predictionPage", s.predictionPageHandler)
	// s.handleFunc("/reRunAll", s.reRunAllHandler)

	addr := fmt.Sprintf(":%v", port)
	log.Info().Msgf("BackOffice listening on %v", addr)
	log.Fatal().Err(http.ListenAndServe(addr, nil)).Msg("")
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

func (s UI) indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	rawUUIDs := r.FormValue("UUIDs")
	rawAuthors := r.FormValue("authors")
	rawStatuses := r.FormValue("statuses")
	rawStateValues := r.FormValue("stateValues")

	t, ok := templates["index.html"]
	if !ok {
		log.Fatal().Msg("Couldn't find index.hmtl")
	}

	res := s.apiClient.get(getBody{Filters: types.APIFilters{
		AuthorHandles:         trim(strings.Split(rawAuthors, ",")),
		UUIDs:                 trim(strings.Split(rawUUIDs, ",")),
		PredictionStateValues: trim(strings.Split(rawStatuses, ",")),
		PredictionStateStatus: trim(strings.Split(rawStateValues, ",")),
		IncludeUIUnsupported:  true,
	}})

	data := make(map[string]interface{})
	data["GetPredictionsErr"] = res.ErrorMessage
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalErrorMessage"] = res.InternalErrorMessage

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

func (s UI) predictionHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	uuid := r.FormValue("uuid")
	if uuid == "" {
		uuid = r.URL.Query().Get("uuid")
	}
	action := r.FormValue("action")

	t, ok := templates["prediction.html"]
	if !ok {
		log.Fatal().Msg("Couldn't find prediction.html")
	}

	var res parsedResponse
	switch action {
	case "pause":
		res = s.apiClient.pausePrediction(uuid)
	case "unpause":
		res = s.apiClient.unpausePrediction(uuid)
	case "hide":
		res = s.apiClient.hidePrediction(uuid)
	case "unhide":
		res = s.apiClient.unhidePrediction(uuid)
	case "delete":
		res = s.apiClient.deletePrediction(uuid)
	case "undelete":
		res = s.apiClient.undeletePrediction(uuid)
	case "refreshAccount":
		res = s.apiClient.refreshAccount(uuid)
	case "clearState":
		res = s.apiClient.clearState(uuid)
	default:
		res.Status = 200
	}

	if res.Status == 200 {
		res = s.apiClient.get(getBody{Filters: types.APIFilters{
			UUIDs: []string{uuid},
		}})
	}

	base64ImageRes := s.apiClient.predictionImage(predictionImageBody{
		UUID: uuid,
	})

	data := make(map[string]interface{})
	if res.Predictions != nil && len(*res.Predictions) == 1 {
		pred := (*res.Predictions)[0]
		data["prediction"] = predictionToMap(pred)
	}

	data["GetPredictionsErr"] = res.ErrorMessage
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalErrorMessage"] = res.InternalErrorMessage
	data["Base64Image"] = base64ImageRes.Base64Image

	if err := t.Execute(w, data); err != nil {
		if nErr, ok := err.(*net.OpError); !ok || nErr.Err != syscall.EPIPE {
			log.Fatal().Err(err).Msg("")
		}
	}
}

func (s UI) predictionPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	url := r.FormValue("url")

	t, ok := templates["prediction_page.html"]
	if !ok {
		log.Fatal().Msg("Couldn't find prediction.html")
	}

	res := s.apiClient.predictionPage(getBody{Filters: types.APIFilters{
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

	data["GetPredictionsErr"] = res.ErrorMessage
	data["GetPredictionsStatus"] = res.Status
	data["GetPredictionsErrCode"] = res.ErrorCode
	data["GetPredictionsInternalMessage"] = res.InternalErrorMessage

	if err := t.Execute(w, data); err != nil {
		if nErr, ok := err.(*net.OpError); !ok || nErr.Err != syscall.EPIPE {
			log.Fatal().Err(err).Msg("")
		}
	}
}

func (s UI) putHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
	rawString := r.FormValue("prediction_input")
	storeStr := r.FormValue("store")

	t, ok := templates["add.html"]
	if !ok {
		log.Fatal().Msg("Couldn't find add.hmtl")
	}

	data := make(map[string]interface{})
	data["Err"] = ""
	data["Predictions"] = []map[string]string{}

	if rawString != "" {

		store := false
		if strings.ToLower(storeStr) == "on" {
			store = true
		}

		resp := s.apiClient.new([]byte(rawString), store)

		data["Err"] = resp.ErrorMessage
		data["Status"] = resp.Status
		data["ErrCode"] = resp.ErrorCode
		data["InternalErrorMessage"] = resp.InternalErrorMessage

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
			log.Fatal().Err(err).Msg("")
		}
	}
}

func (s UI) handleFunc(pattern string, fn http.HandlerFunc) {
	http.HandleFunc(pattern, basicAuthHandler(s.basicAuthUser, s.basicAuthPass, "BackOffice", fn))
}

func basicAuthHandler(user, pass, realm string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if checkBasicAuth(r, user, pass) {
			next(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm="%s"`, realm))
		w.WriteHeader(401)
		w.Write([]byte("401 Unauthorized\n"))
	}
}

func checkBasicAuth(r *http.Request, user, pass string) bool {
	u, p, ok := r.BasicAuth()
	if !ok {
		return false
	}
	return u == user && p == pass
}
