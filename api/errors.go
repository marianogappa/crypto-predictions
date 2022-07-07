package api

import (
	"errors"

	"github.com/marianogappa/crypto-candles/candles/common"
	"github.com/marianogappa/predictions/core"
)

// ErrorContent struct contains necessary data to return an error compliant with the API spec.
type ErrorContent struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
}

// Status returns the error's status code.
func (c ErrorContent) Status() int { return c.StatusCode }

// String returns the error's message.
func (c ErrorContent) String() string { return c.Message }

// Error returns the error's message.
func (c ErrorContent) Error() string { return c.Message }

// Unwrap returns an error with the error's message.
func (c ErrorContent) Unwrap() error { return errors.New(c.Message) }

var (
	// ErrInvalidRequestBody is returned by the API when the received request has an invalid body.
	ErrInvalidRequestBody = errors.New("invalid request body")
	// ErrInvalidRequestJSON is returned by the API when the received request has a valid body containing invalid JSON.
	ErrInvalidRequestJSON = errors.New("invalid request JSON")
	// ErrStorageErrorRetrievingPredictions is returned by the API when something went wrong retrieving a prediction.
	ErrStorageErrorRetrievingPredictions = errors.New("storage had error retrieving predictions")
	// ErrStorageErrorStoringPrediction is returned by the API when something went wrong storing a prediction.
	ErrStorageErrorStoringPrediction = errors.New("storage had error storing predictions")
	// ErrStorageErrorStoringAccount is returned by the API when something went wrong storing an account.
	ErrStorageErrorStoringAccount = errors.New("storage had error storing accounts")
	// ErrFailedToSerializePredictions is returned by the API when something went wrong serializing a prediction.
	ErrFailedToSerializePredictions = errors.New("failed to serialize predictions")
	// ErrFailedToSerializeAccount is returned by the API when something went wrong serializing an account.
	ErrFailedToSerializeAccount = errors.New("failed to serialize account")
	// ErrFailedToCompilePrediction is returned by the API when something went wrong compiling a prediction from a string.
	ErrFailedToCompilePrediction = errors.New("failed to compile prediction")
	// ErrPredictionNotFound is returned by the API when an unknown prediction was requested.
	ErrPredictionNotFound = errors.New("prediction not found")

	errToResponse = map[error]ErrorContent{
		core.ErrUnknownOperandType:                 {StatusCode: 400, ErrorCode: "ErrUnknownOperandType", Message: "unknown value for operandType"},
		core.ErrUnknownBoolOperator:                {StatusCode: 400, ErrorCode: "ErrUnknownBoolOperator", Message: "unknown value for BoolOperator"},
		core.ErrUnknownConditionStatus:             {StatusCode: 400, ErrorCode: "ErrUnknownConditionStatus", Message: "invalid state: unknown value for ConditionStatus"},
		core.ErrUnknownPredictionStateValue:        {StatusCode: 400, ErrorCode: "ErrUnknownPredictionStateValue", Message: "invalid state: unknown value for PredictionStateValue"},
		core.ErrInvalidOperand:                     {StatusCode: 400, ErrorCode: "ErrInvalidOperand", Message: "invalid operand"},
		core.ErrEmptyQuoteAsset:                    {StatusCode: 400, ErrorCode: "ErrEmptyQuoteAsset", Message: "quote asset cannot be empty"},
		core.ErrNonEmptyQuoteAssetOnNonCoin:        {StatusCode: 400, ErrorCode: "ErrNonEmptyQuoteAssetOnNonCoin", Message: "quote asset must be empty for non-coin operand types"},
		core.ErrEqualBaseQuoteAssets:               {StatusCode: 400, ErrorCode: "ErrEqualBaseQuoteAssets", Message: "base asset cannot be equal to quote asset"},
		core.ErrInvalidDuration:                    {StatusCode: 400, ErrorCode: "ErrInvalidDuration", Message: "invalid duration"},
		core.ErrInvalidFromISO8601:                 {StatusCode: 400, ErrorCode: "ErrInvalidFromISO8601", Message: "invalid FromISO8601"},
		core.ErrInvalidToISO8601:                   {StatusCode: 400, ErrorCode: "ErrInvalidToISO8601", Message: "invalid ToISO8601"},
		core.ErrOneOfToISO8601ToDurationRequired:   {StatusCode: 400, ErrorCode: "ErrOneOfToISO8601ToDurationRequired", Message: "one of ToISO8601 or ToDuration is required"},
		core.ErrInvalidConditionSyntax:             {StatusCode: 400, ErrorCode: "ErrInvalidConditionSyntax", Message: "invalid condition syntax"},
		core.ErrUnknownConditionOperator:           {StatusCode: 400, ErrorCode: "ErrUnknownConditionOperator", Message: "unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]"},
		core.ErrErrorMarginRatioAbove30:            {StatusCode: 400, ErrorCode: "ErrErrorMarginRatioAbove30", Message: "error margin ratio above 30%% is not allowed"},
		core.ErrInvalidJSON:                        {StatusCode: 400, ErrorCode: "ErrInvalidJSON", Message: "invalid JSON"},
		core.ErrEmptyReporter:                      {StatusCode: 400, ErrorCode: "ErrEmptyReporter", Message: "reporter cannot be empty"},
		core.ErrEmptyPostURL:                       {StatusCode: 400, ErrorCode: "ErrEmptyPostURL", Message: "postUrl cannot be empty"},
		core.ErrEmptyPostAuthor:                    {StatusCode: 400, ErrorCode: "ErrEmptyPostAuthor", Message: "postAuthor cannot be empty"},
		core.ErrEmptyPostedAt:                      {StatusCode: 400, ErrorCode: "ErrEmptyPostedAt", Message: "postedAt cannot be empty"},
		core.ErrInvalidPostedAt:                    {StatusCode: 400, ErrorCode: "ErrInvalidPostedAt", Message: "postedAt must be a valid ISO8601 timestamp"},
		core.ErrEmptyPredict:                       {StatusCode: 400, ErrorCode: "ErrEmptyPredict", Message: "main predict clause cannot be empty"},
		core.ErrMissingRequiredPrePredictPredictIf: {StatusCode: 400, ErrorCode: "ErrMissingRequiredPrePredictPredictIf", Message: "pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause"},
		core.ErrBoolExprSyntaxError:                {StatusCode: 400, ErrorCode: "ErrBoolExprSyntaxError", Message: "syntax error in bool expression"},
		core.ErrPredictionFinishedAtStartTime:      {StatusCode: 400, ErrorCode: "ErrPredictionFinishedAtStartTime", Message: "prediction is finished at start time"},

		// From Market
		common.ErrInvalidMarketPair:               {StatusCode: 400, ErrorCode: "ErrInvalidMarketPair", Message: "market pair does not exist on exchange"},
		common.ErrOutOfTicks:                      {StatusCode: 500, ErrorCode: "ErrOutOfTicks", Message: "out of ticks"},
		common.ErrOutOfCandlesticks:               {StatusCode: 500, ErrorCode: "ErrOutOfCandlesticks", Message: "exchange ran out of candlesticks"},
		common.ErrOutOfTrades:                     {StatusCode: 500, ErrorCode: "ErrOutOfTrades", Message: "exchange ran out of trades"},
		common.ErrRateLimit:                       {StatusCode: 500, ErrorCode: "ErrRateLimit", Message: "exchange asked us to enhance our calm"},
		common.ErrNoNewTicksYet:                   {StatusCode: 500, ErrorCode: "ErrNoNewTicksYet", Message: "no new ticks yet"},
		common.ErrExchangeReturnedNoTicks:         {StatusCode: 500, ErrorCode: "ErrExchangeReturnedNoTicks", Message: "exchange returned no ticks"},
		common.ErrExchangeReturnedOutOfSyncTick:   {StatusCode: 500, ErrorCode: "ErrExchangeReturnedOutOfSyncTick", Message: "exchange returned out of sync tick"},
		common.ErrOutOfSyncTimestampPatchingHoles: {StatusCode: 500, ErrorCode: "ErrOutOfSyncTimestampPatchingHoles", Message: "out of sync timestamp found patching holes"},

		ErrInvalidRequestBody:                  {StatusCode: 400, ErrorCode: "ErrInvalidRequestBody", Message: "invalid request body"},
		ErrInvalidRequestJSON:                  {StatusCode: 400, ErrorCode: "ErrInvalidRequestJSON", Message: "invalid request JSON"},
		ErrStorageErrorRetrievingPredictions:   {StatusCode: 500, ErrorCode: "ErrStorageErrorRetrievingPredictions", Message: "storage had error retrieving predictions"},
		core.ErrStorageErrorRetrievingAccounts: {StatusCode: 500, ErrorCode: "ErrStorageErrorRetrievingAccounts", Message: "storage had error retrieving accounts"},
		ErrStorageErrorStoringPrediction:       {StatusCode: 500, ErrorCode: "ErrStorageErrorStoringPrediction", Message: "storage had error storing predictions"},
		ErrStorageErrorStoringAccount:          {StatusCode: 500, ErrorCode: "ErrStorageErrorStoringAccount", Message: "storage had error storing accounts"},
		ErrFailedToSerializePredictions:        {StatusCode: 500, ErrorCode: "ErrFailedToSerializePredictions", Message: "failed to serialize predictions"},
		ErrFailedToSerializeAccount:            {StatusCode: 500, ErrorCode: "ErrFailedToSerializeAccount", Message: "failed to serialize account"},
		ErrFailedToCompilePrediction:           {StatusCode: 500, ErrorCode: "ErrFailedToCompilePrediction", Message: "failed to compile prediction"},
		ErrPredictionNotFound:                  {StatusCode: 404, ErrorCode: "ErrPredictionNotFound", Message: "prediction not found"},
	}
)
