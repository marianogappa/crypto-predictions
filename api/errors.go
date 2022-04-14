package api

import (
	"errors"

	"github.com/marianogappa/predictions/types"
)

type ErrorContent struct {
	StatusCode int    `json:"status"`
	Message    string `json:"message,omitempty"`
	ErrorCode  string `json:"errorCode,omitempty"`
}

func (c ErrorContent) Status() int    { return c.StatusCode }
func (c ErrorContent) String() string { return c.Message }
func (c ErrorContent) Error() string  { return c.Message }
func (c ErrorContent) Unwrap() error  { return errors.New(c.Message) }

var (
	ErrInvalidRequestBody                = errors.New("invalid request body")
	ErrInvalidRequestJSON                = errors.New("invalid request JSON")
	ErrStorageErrorRetrievingPredictions = errors.New("storage had error retrieving predictions")
	ErrStorageErrorRetrievingAccounts    = errors.New("storage had error retrieving accounts")
	ErrStorageErrorStoringPrediction     = errors.New("storage had error storing predictions")
	ErrStorageErrorStoringAccount        = errors.New("storage had error storing accounts")
	ErrFailedToSerializePredictions      = errors.New("failed to serialize predictions")
	ErrFailedToSerializeAccount          = errors.New("failed to serialize account")
	ErrFailedToCompilePrediction         = errors.New("failed to compile prediction")
	ErrPredictionNotFound                = errors.New("prediction not found")

	errToResponse = map[error]ErrorContent{
		types.ErrUnknownOperandType:                 {StatusCode: 400, ErrorCode: "ErrUnknownOperandType", Message: "unknown value for operandType"},
		types.ErrUnknownBoolOperator:                {StatusCode: 400, ErrorCode: "ErrUnknownBoolOperator", Message: "unknown value for BoolOperator"},
		types.ErrUnknownConditionStatus:             {StatusCode: 400, ErrorCode: "ErrUnknownConditionStatus", Message: "invalid state: unknown value for ConditionStatus"},
		types.ErrUnknownPredictionStateValue:        {StatusCode: 400, ErrorCode: "ErrUnknownPredictionStateValue", Message: "invalid state: unknown value for PredictionStateValue"},
		types.ErrInvalidOperand:                     {StatusCode: 400, ErrorCode: "ErrInvalidOperand", Message: "invalid operand"},
		types.ErrEmptyQuoteAsset:                    {StatusCode: 400, ErrorCode: "ErrEmptyQuoteAsset", Message: "quote asset cannot be empty"},
		types.ErrNonEmptyQuoteAssetOnNonCoin:        {StatusCode: 400, ErrorCode: "ErrNonEmptyQuoteAssetOnNonCoin", Message: "quote asset must be empty for non-coin operand types"},
		types.ErrEqualBaseQuoteAssets:               {StatusCode: 400, ErrorCode: "ErrEqualBaseQuoteAssets", Message: "base asset cannot be equal to quote asset"},
		types.ErrInvalidDuration:                    {StatusCode: 400, ErrorCode: "ErrInvalidDuration", Message: "invalid duration"},
		types.ErrInvalidFromISO8601:                 {StatusCode: 400, ErrorCode: "ErrInvalidFromISO8601", Message: "invalid FromISO8601"},
		types.ErrInvalidToISO8601:                   {StatusCode: 400, ErrorCode: "ErrInvalidToISO8601", Message: "invalid ToISO8601"},
		types.ErrOneOfToISO8601ToDurationRequired:   {StatusCode: 400, ErrorCode: "ErrOneOfToISO8601ToDurationRequired", Message: "one of ToISO8601 or ToDuration is required"},
		types.ErrInvalidConditionSyntax:             {StatusCode: 400, ErrorCode: "ErrInvalidConditionSyntax", Message: "invalid condition syntax"},
		types.ErrUnknownConditionOperator:           {StatusCode: 400, ErrorCode: "ErrUnknownConditionOperator", Message: "unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]"},
		types.ErrErrorMarginRatioAbove30:            {StatusCode: 400, ErrorCode: "ErrErrorMarginRatioAbove30", Message: "error margin ratio above 30%% is not allowed"},
		types.ErrInvalidJSON:                        {StatusCode: 400, ErrorCode: "ErrInvalidJSON", Message: "invalid JSON"},
		types.ErrEmptyReporter:                      {StatusCode: 400, ErrorCode: "ErrEmptyReporter", Message: "reporter cannot be empty"},
		types.ErrEmptyPostURL:                       {StatusCode: 400, ErrorCode: "ErrEmptyPostURL", Message: "postUrl cannot be empty"},
		types.ErrEmptyPostAuthor:                    {StatusCode: 400, ErrorCode: "ErrEmptyPostAuthor", Message: "postAuthor cannot be empty"},
		types.ErrEmptyPostedAt:                      {StatusCode: 400, ErrorCode: "ErrEmptyPostedAt", Message: "postedAt cannot be empty"},
		types.ErrInvalidPostedAt:                    {StatusCode: 400, ErrorCode: "ErrInvalidPostedAt", Message: "postedAt must be a valid ISO8601 timestamp"},
		types.ErrEmptyPredict:                       {StatusCode: 400, ErrorCode: "ErrEmptyPredict", Message: "main predict clause cannot be empty"},
		types.ErrMissingRequiredPrePredictPredictIf: {StatusCode: 400, ErrorCode: "ErrMissingRequiredPrePredictPredictIf", Message: "pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause"},
		types.ErrBoolExprSyntaxError:                {StatusCode: 400, ErrorCode: "ErrBoolExprSyntaxError", Message: "syntax error in bool expression"},
		types.ErrPredictionFinishedAtStartTime:      {StatusCode: 400, ErrorCode: "ErrPredictionFinishedAtStartTime", Message: "prediction is finished at start time"},
		types.ErrInvalidMarketPair:                  {StatusCode: 400, ErrorCode: "ErrInvalidMarketPair", Message: "market pair does not exist on exchange"},
		ErrInvalidRequestBody:                       {StatusCode: 400, ErrorCode: "ErrInvalidRequestBody", Message: "invalid request body"},
		ErrInvalidRequestJSON:                       {StatusCode: 400, ErrorCode: "ErrInvalidRequestJSON", Message: "invalid request JSON"},
		ErrStorageErrorRetrievingPredictions:        {StatusCode: 500, ErrorCode: "ErrStorageErrorRetrievingPredictions", Message: "storage had error retrieving predictions"},
		ErrStorageErrorRetrievingAccounts:           {StatusCode: 500, ErrorCode: "ErrStorageErrorRetrievingAccounts", Message: "storage had error retrieving accounts"},
		ErrStorageErrorStoringPrediction:            {StatusCode: 500, ErrorCode: "ErrStorageErrorStoringPrediction", Message: "storage had error storing predictions"},
		ErrStorageErrorStoringAccount:               {StatusCode: 500, ErrorCode: "ErrStorageErrorStoringAccount", Message: "storage had error storing accounts"},
		ErrFailedToSerializePredictions:             {StatusCode: 500, ErrorCode: "ErrFailedToSerializePredictions", Message: "failed to serialize predictions"},
		ErrFailedToSerializeAccount:                 {StatusCode: 500, ErrorCode: "ErrFailedToSerializeAccount", Message: "failed to serialize account"},
		ErrFailedToCompilePrediction:                {StatusCode: 500, ErrorCode: "ErrFailedToCompilePrediction", Message: "failed to compile prediction"},
		ErrPredictionNotFound:                       {StatusCode: 404, ErrorCode: "ErrPredictionNotFound", Message: "prediction not found"},
	}
)
