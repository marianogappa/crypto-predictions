package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marianogappa/predictions/core"
	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/rs/zerolog/log"
)

// PredictionCompiler is the component that takes a JSON string representing a prediction and compiles it into a
// prediction that can be used throughout the engine. It may use a MetadataFetcher to fetch metadata from the
// Twitter / Youtube post, e.g. when it was posted & by whom.
type PredictionCompiler struct {
	metadataFetcher *metadatafetcher.MetadataFetcher
	timeNow         func() time.Time
}

// NewPredictionCompiler constructs a PredictionCompiler.
func NewPredictionCompiler(fetcher *metadatafetcher.MetadataFetcher, timeNow func() time.Time) PredictionCompiler {
	return PredictionCompiler{metadataFetcher: fetcher, timeNow: timeNow}
}

type partiallyCompile = func(Prediction, *core.Prediction, *core.Account, *metadatafetcher.MetadataFetcher, func() time.Time) error

var (
	partialCompilerNames = []string{
		"compileVersion",
		"compileUUID",
		"compileReporter",
		"compilePostURL",
		"compileCreatedAt",
		"compileMetadata",
		"compileGiven",
		"compileInnerPrediction",
		"compilePredictionType",
		"compilePredictionState",
		"compileFlags",
	}
	partialCompilerFns = map[string]partiallyCompile{
		partialCompilerNames[0]:  compileVersion,
		partialCompilerNames[1]:  compileUUID,
		partialCompilerNames[2]:  compileReporter,
		partialCompilerNames[3]:  compilePostURL,
		partialCompilerNames[4]:  compileCreatedAt,
		partialCompilerNames[5]:  compileMetadata,
		partialCompilerNames[6]:  compileGiven,
		partialCompilerNames[7]:  compileInnerPrediction,
		partialCompilerNames[8]:  compilePredictionType,
		partialCompilerNames[9]:  compilePredictionState,
		partialCompilerNames[10]: compileFlags,
	}
)

// Compile compiles a JSON string representing a prediction into a prediction that can be used throughout the engine.
// If a MetadataFetcher is supplied on construction, it may also return an Account representing the post author's
// social media account (but otherwise the Account will be nil).
func (c PredictionCompiler) Compile(rawPredictionBs []byte) (core.Prediction, *core.Account, error) {
	var (
		account    *core.Account
		prediction = core.Prediction{}
		raw        = Prediction{}
	)

	if err := json.Unmarshal([]byte(rawPredictionBs), &raw); err != nil {
		return prediction, nil, fmt.Errorf("while unmarshalling raw incoming JSON for compilation: %w: %v", core.ErrInvalidJSON, err)
	}

	for _, name := range partialCompilerNames {
		var maybeAccount core.Account
		if err := partialCompilerFns[name](raw, &prediction, &maybeAccount, c.metadataFetcher, c.timeNow); err != nil {
			return prediction, &maybeAccount, fmt.Errorf("while running compiler step %v: %w", name, err)
		}
		if maybeAccount.URL != nil {
			account = &maybeAccount
		}
	}

	return prediction, account, nil
}

func compileReporter(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.Reporter == "" {
		return core.ErrEmptyReporter
	}
	prediction.Reporter = raw.Reporter
	return nil
}

func compileVersion(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.Version == "" {
		raw.Version = "1.0.0"
	}
	prediction.Version = raw.Version
	return nil
}

func compileUUID(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.UUID = raw.UUID
	return nil
}

func compileFlags(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.Paused = raw.Paused
	prediction.Hidden = raw.Hidden
	prediction.Deleted = raw.Deleted
	return nil
}

func compileCreatedAt(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if timeNow != nil && raw.CreatedAt == "" {
		raw.CreatedAt = core.ISO8601(timeNow().Format(time.RFC3339))
	}
	prediction.CreatedAt = raw.CreatedAt
	return nil
}

func compilePostURL(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.PostURL == "" {
		return core.ErrEmptyPostURL
	}
	prediction.PostURL = raw.PostURL
	return nil
}

func compileMetadata(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	// N.B. This is not necessary but staticcheck complains that account is overwritten before first use.
	if account != nil && account.URL != nil {
		return errors.New("predictionCompiler.compile: received a non-nil account when about to compile metadata")
	}
	if mf == nil && (raw.PostAuthor == "" || raw.PostAuthorURL == "") {
		return core.ErrEmptyPostAuthor
	}
	if mf == nil && (raw.PostedAt == "") {
		return core.ErrEmptyPostedAt
	}
	if mf != nil && (raw.PostAuthor == "" || raw.PostedAt == "" || raw.PostAuthorURL == "") {
		log.Info().Msgf("Fetching metadata for %v\n", raw.PostURL)
		metadata, err := mf.Fetch(raw.PostURL)
		if err != nil {
			return fmt.Errorf("while fetching metadata: %w", err)
		}
		*account = metadata.Author
		// N.B. This is not necessary but staticcheck complains that account is never used.
		if account.Handle == "" {
			return errors.New("predictionCompiler.compile: failed to fetch metadata, but without an error")
		}
		if raw.PostAuthor == "" {
			raw.PostAuthor = metadata.Author.Handle
		}
		if raw.PostAuthor == "" {
			raw.PostAuthor = metadata.Author.Name
		}
		if raw.PostAuthorURL == "" && metadata.Author.URL != nil {
			raw.PostAuthorURL = metadata.Author.URL.String()
		}
		if raw.PostedAt == "" {
			raw.PostedAt = metadata.PostCreatedAt
		}
	}
	if raw.PostAuthor == "" {
		return core.ErrEmptyPostAuthor
	}
	if raw.PostedAt == "" {
		return core.ErrEmptyPostedAt
	}
	if _, err := raw.PostedAt.Seconds(); err != nil {
		return core.ErrInvalidPostedAt
	}

	prediction.PostAuthor = raw.PostAuthor
	prediction.PostedAt = raw.PostedAt
	prediction.PostAuthorURL = raw.PostAuthorURL

	return nil
}

func compileGiven(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.Given = map[string]*core.Condition{}
	for name, condition := range raw.Given {
		c, err := mapCondition(condition, name, prediction.PostedAt)
		if err != nil {
			return err
		}
		prediction.Given[name] = &c
	}
	return nil
}

func compilePredictionType(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.Type = core.PredictionTypeFromString(raw.Type)
	if prediction.Type == core.PredictionTypeUnsupported {
		prediction.Type = CalculatePredictionType(*prediction)
	}
	return nil
}

func compilePredictionState(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	status, err := core.ConditionStatusFromString(raw.PredictionState.Status)
	if err != nil {
		return err
	}
	value, err := core.PredictionStateValueFromString(raw.PredictionState.Value)
	if err != nil {
		return err
	}
	prediction.State = core.PredictionState{
		Status: status,
		LastTs: raw.PredictionState.LastTs,
		Value:  value,
	}
	return nil
}

func compileInnerPrediction(raw Prediction, prediction *core.Prediction, account *core.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	var (
		b   *core.BoolExpr
		err error
	)
	if raw.PrePredict != nil {
		b, err = mapBoolExpr(raw.PrePredict.WrongIf, prediction.Given)
		if err != nil {
			return err
		}
		prediction.PrePredict.WrongIf = b

		b, err = mapBoolExpr(raw.PrePredict.AnnulledIf, prediction.Given)
		if err != nil {
			return err
		}
		prediction.PrePredict.AnnulledIf = b

		b, err = mapBoolExpr(raw.PrePredict.Predict, prediction.Given)
		if err != nil {
			return err
		}
		prediction.PrePredict.Predict = b

		prediction.PrePredict.IgnoreUndecidedIfPredictIsDefined = raw.PrePredict.IgnoreUndecidedIfPredictIsDefined
		prediction.PrePredict.AnnulledIfPredictIsFalse = raw.PrePredict.AnnulledIfPredictIsFalse

		if prediction.PrePredict.Predict == nil && (prediction.PrePredict.WrongIf != nil || prediction.PrePredict.AnnulledIf != nil) {
			return core.ErrMissingRequiredPrePredictPredictIf
		}
	}

	if raw.Predict.WrongIf != nil {
		b, err = mapBoolExpr(raw.Predict.WrongIf, prediction.Given)
		if err != nil {
			return err
		}
		prediction.Predict.WrongIf = b
	}

	if raw.Predict.AnnulledIf != nil {
		b, err = mapBoolExpr(raw.Predict.AnnulledIf, prediction.Given)
		if err != nil {
			return err
		}
		prediction.Predict.AnnulledIf = b
	}

	if raw.Predict.Predict == "" {
		return core.ErrEmptyPredict
	}
	b, err = mapBoolExpr(&raw.Predict.Predict, prediction.Given)
	if err != nil {
		return err
	}
	prediction.Predict.Predict = *b
	prediction.Predict.IgnoreUndecidedIfPredictIsDefined = raw.Predict.IgnoreUndecidedIfPredictIsDefined
	return nil
}
