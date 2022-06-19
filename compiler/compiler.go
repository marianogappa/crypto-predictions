package compiler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marianogappa/predictions/metadatafetcher"
	"github.com/marianogappa/predictions/types"
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

// Compile compiles a JSON string representing a prediction into a prediction that can be used throughout the engine.
// If a MetadataFetcher is supplied on construction, it may also return an Account representing the post author's
// social media account (but otherwise the Account will be nil).
func (c PredictionCompiler) Compile(rawPredictionBs []byte) (types.Prediction, *types.Account, error) {
	var (
		account    *types.Account
		prediction = types.Prediction{}
		raw        = Prediction{}
	)

	if err := json.Unmarshal([]byte(rawPredictionBs), &raw); err != nil {
		return prediction, nil, fmt.Errorf("%w: %v", types.ErrInvalidJSON, err)
	}

	type partiallyCompile = func(Prediction, *types.Prediction, *types.Account, *metadatafetcher.MetadataFetcher, func() time.Time) error

	partialCompilers := []partiallyCompile{
		compileVersion,
		compileUUID,
		compileReporter,
		compilePostURL,
		compileCreatedAt,
		compileMetadata,
		compileGiven,
		compileInnerPrediction,
		compilePredictionType,
		compilePredictionState,
	}
	for _, partiallyCompile := range partialCompilers {
		if err := partiallyCompile(raw, &prediction, account, c.metadataFetcher, c.timeNow); err != nil {
			return prediction, account, err
		}
	}

	return prediction, account, nil
}

func compileReporter(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.Reporter == "" {
		return types.ErrEmptyReporter
	}
	prediction.Reporter = raw.Reporter
	return nil
}

func compileVersion(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.Version == "" {
		raw.Version = "1.0.0"
	}
	prediction.Version = raw.Version
	return nil
}

func compileUUID(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.UUID = raw.UUID
	return nil
}

func compileCreatedAt(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if timeNow != nil && raw.CreatedAt == "" {
		raw.CreatedAt = types.ISO8601(timeNow().Format(time.RFC3339))
	}
	prediction.CreatedAt = raw.CreatedAt
	return nil
}

func compilePostURL(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	if raw.PostURL == "" {
		return types.ErrEmptyPostURL
	}
	prediction.PostUrl = raw.PostURL
	return nil
}

func compileMetadata(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	// N.B. This is not necessary but staticcheck complains that account is overwritten before first use.
	if account != nil {
		return errors.New("predictionCompiler.compile: received a non-nil account when about to compile metadata")
	}
	if mf == nil && (raw.PostAuthor == "" || raw.PostAuthorURL == "") {
		return types.ErrEmptyPostAuthor
	}
	if mf == nil && (raw.PostedAt == "") {
		return types.ErrEmptyPostedAt
	}
	if mf != nil && (raw.PostAuthor == "" || raw.PostedAt == "" || raw.PostAuthorURL == "") {
		log.Info().Msgf("Fetching metadata for %v\n", raw.PostURL)
		metadata, err := mf.Fetch(raw.PostURL)
		if err != nil {
			return err
		}
		account = &metadata.Author
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
		return types.ErrEmptyPostAuthor
	}
	if raw.PostedAt == "" {
		return types.ErrEmptyPostedAt
	}
	if _, err := raw.PostedAt.Seconds(); err != nil {
		return types.ErrInvalidPostedAt
	}

	prediction.PostAuthor = raw.PostAuthor
	prediction.PostedAt = raw.PostedAt
	prediction.PostAuthorURL = raw.PostAuthorURL

	return nil
}

func compileGiven(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.Given = map[string]*types.Condition{}
	for name, condition := range raw.Given {
		c, err := mapCondition(condition, name, prediction.PostedAt)
		if err != nil {
			return err
		}
		prediction.Given[name] = &c
	}
	return nil
}

func compilePredictionType(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	prediction.Type = types.PredictionTypeFromString(raw.Type)
	if prediction.Type == types.PREDICTION_TYPE_UNSUPPORTED {
		prediction.Type = CalculatePredictionType(*prediction)
	}
	return nil
}

func compilePredictionState(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	status, err := types.ConditionStatusFromString(raw.PredictionState.Status)
	if err != nil {
		return err
	}
	value, err := types.PredictionStateValueFromString(raw.PredictionState.Value)
	if err != nil {
		return err
	}
	prediction.State = types.PredictionState{
		Status: status,
		LastTs: raw.PredictionState.LastTs,
		Value:  value,
	}
	return nil
}

func compileInnerPrediction(raw Prediction, prediction *types.Prediction, account *types.Account, mf *metadatafetcher.MetadataFetcher, timeNow func() time.Time) error {
	var (
		b   *types.BoolExpr
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
			return types.ErrMissingRequiredPrePredictPredictIf
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
		return types.ErrEmptyPredict
	}
	b, err = mapBoolExpr(&raw.Predict.Predict, prediction.Given)
	if err != nil {
		return err
	}
	prediction.Predict.Predict = *b
	prediction.Predict.IgnoreUndecidedIfPredictIsDefined = raw.Predict.IgnoreUndecidedIfPredictIsDefined
	return nil
}
