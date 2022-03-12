package api

import (
	"errors"

	"github.com/marianogappa/predictions/types"
)

var (
	ErrInvalidRequestBody                = errors.New("invalid request body")
	ErrInvalidRequestJSON                = errors.New("invalid request JSON")
	ErrStorageErrorRetrievingPredictions = errors.New("storage had error retrieving predictions")
	ErrStorageErrorStoringPrediction     = errors.New("storage had error storing predictions=")
	ErrFailedToSerializePredictions      = errors.New("failed to serialize predictions")
	ErrFailedToCompilePrediction         = errors.New("failed to compile prediction")

	errToResponse = map[error]APIResponse{
		types.ErrUnknownOperandType:                 {Status: 400, ErrorCode: "ErrUnknownOperandType", Message: "unknown value for operandType"},
		types.ErrUnknownBoolOperator:                {Status: 400, ErrorCode: "ErrUnknownBoolOperator", Message: "unknown value for BoolOperator"},
		types.ErrUnknownConditionStatus:             {Status: 400, ErrorCode: "ErrUnknownConditionStatus", Message: "invalid state: unknown value for ConditionStatus"},
		types.ErrUnknownPredictionStateValue:        {Status: 400, ErrorCode: "ErrUnknownPredictionStateValue", Message: "invalid state: unknown value for PredictionStateValue"},
		types.ErrInvalidOperand:                     {Status: 400, ErrorCode: "ErrInvalidOperand", Message: "invalid operand"},
		types.ErrEmptyQuoteAsset:                    {Status: 400, ErrorCode: "ErrEmptyQuoteAsset", Message: "quote asset cannot be empty"},
		types.ErrNonEmptyQuoteAssetOnNonCoin:        {Status: 400, ErrorCode: "ErrNonEmptyQuoteAssetOnNonCoin", Message: "quote asset must be empty for non-coin operand types"},
		types.ErrEqualBaseQuoteAssets:               {Status: 400, ErrorCode: "ErrEqualBaseQuoteAssets", Message: "base asset cannot be equal to quote asset"},
		types.ErrInvalidDuration:                    {Status: 400, ErrorCode: "ErrInvalidDuration", Message: "invalid duration"},
		types.ErrInvalidFromISO8601:                 {Status: 400, ErrorCode: "ErrInvalidFromISO8601", Message: "invalid FromISO8601"},
		types.ErrInvalidToISO8601:                   {Status: 400, ErrorCode: "ErrInvalidToISO8601", Message: "invalid ToISO8601"},
		types.ErrOneOfToISO8601ToDurationRequired:   {Status: 400, ErrorCode: "ErrOneOfToISO8601ToDurationRequired", Message: "one of ToISO8601 or ToDuration is required"},
		types.ErrInvalidConditionSyntax:             {Status: 400, ErrorCode: "ErrInvalidConditionSyntax", Message: "invalid condition syntax"},
		types.ErrUnknownConditionOperator:           {Status: 400, ErrorCode: "ErrUnknownConditionOperator", Message: "unknown condition operator: supported are [>|<|>=|<=|BETWEEN...AND]"},
		types.ErrErrorMarginRatioAbove30:            {Status: 400, ErrorCode: "ErrErrorMarginRatioAbove30", Message: "error margin ratio above 30%% is not allowed"},
		types.ErrInvalidJSON:                        {Status: 400, ErrorCode: "ErrInvalidJSON", Message: "invalid JSON"},
		types.ErrEmptyReporter:                      {Status: 400, ErrorCode: "ErrEmptyReporter", Message: "reporter cannot be empty"},
		types.ErrEmptyPostURL:                       {Status: 400, ErrorCode: "ErrEmptyPostURL", Message: "postUrl cannot be empty"},
		types.ErrEmptyPostAuthor:                    {Status: 400, ErrorCode: "ErrEmptyPostAuthor", Message: "postAuthor cannot be empty"},
		types.ErrEmptyPostedAt:                      {Status: 400, ErrorCode: "ErrEmptyPostedAt", Message: "postedAt cannot be empty"},
		types.ErrInvalidPostedAt:                    {Status: 400, ErrorCode: "ErrInvalidPostedAt", Message: "postedAt must be a valid ISO8601 timestamp"},
		types.ErrMissingRequiredPrePredictPredictIf: {Status: 400, ErrorCode: "ErrMissingRequiredPrePredictPredictIf", Message: "pre-predict clause must have predictIf if it has either wrongIf or annuledIf. Otherwise, add them directly on predict clause"},
		types.ErrBoolExprSyntaxError:                {Status: 400, ErrorCode: "ErrBoolExprSyntaxError", Message: "syntax error in bool expression"},
		types.ErrPredictionFinishedAtStartTime:      {Status: 400, ErrorCode: "ErrPredictionFinishedAtStartTime", Message: "prediction is finished at start time"},
		types.ErrInvalidMarketPair:                  {Status: 400, ErrorCode: "ErrInvalidMarketPair", Message: "market pair does not exist on exchange"},
		ErrInvalidRequestBody:                       {Status: 400, ErrorCode: "ErrInvalidRequestBody", Message: "invalid request body"},
		ErrInvalidRequestJSON:                       {Status: 400, ErrorCode: "ErrInvalidRequestJSON", Message: "invalid request JSON"},
		ErrStorageErrorRetrievingPredictions:        {Status: 500, ErrorCode: "ErrStorageErrorRetrievingPredictions", Message: "storage had error retrieving predictions"},
		ErrStorageErrorStoringPrediction:            {Status: 500, ErrorCode: "ErrStorageErrorStoringPrediction", Message: "storage had error storing predictions"},
		ErrFailedToSerializePredictions:             {Status: 500, ErrorCode: "ErrFailedToSerializePredictions", Message: "failed to serialize predictions"},
		ErrFailedToCompilePrediction:                {Status: 500, ErrorCode: "ErrFailedToCompilePrediction", Message: "failed to compile prediction"},
	}
)
