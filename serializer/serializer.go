package serializer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

// PredictionSerializer is the component that serializes a Prediction to a string representation, to be persisted or
// returned in an API call.
type PredictionSerializer struct {
	mkt *market.IMarket
}

// NewPredictionSerializer constructs a PredictionSerializer.
func NewPredictionSerializer(market *market.IMarket) PredictionSerializer {
	return PredictionSerializer{mkt: market}
}

// PreSerialize serializes a Prediction to a compiler.Prediction, but doesn't take the extra step of serializing it to a
// JSON []byte.
func (s PredictionSerializer) PreSerialize(p *types.Prediction) (compiler.Prediction, error) {
	pp, err := marshalPrePredict(p.PrePredict)
	if err != nil {
		return compiler.Prediction{}, err
	}
	pd, err := marshalPredict(p.Predict)
	if err != nil {
		return compiler.Prediction{}, err
	}
	return compiler.Prediction{
		UUID:            p.UUID,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		Reporter:        p.Reporter,
		PostAuthor:      p.PostAuthor,
		PostAuthorURL:   p.PostAuthorURL,
		PostedAt:        p.PostedAt,
		PostURL:         p.PostUrl,
		Given:           marshalGiven(p.Given),
		PrePredict:      pp,
		Predict:         pd,
		PredictionState: marshalPredictionState(p.State),
		Type:            p.Type.String(),
	}, nil
}

// Serialize serializes a Prediction to a JSON []byte. It is meant to be used for persisting, but must not be used
// by the API. There's a separate PreSerializeForAPI method for that purpose.
func (s PredictionSerializer) Serialize(p *types.Prediction) ([]byte, error) {
	pre, err := s.PreSerialize(p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(pre)
}

// PreSerializeForAPI serializes a Prediction to a compiler.Prediction, but doesn't take the extra step of serializing
// it to a JSON []byte. It is meant to be used only by the API.
func (s PredictionSerializer) PreSerializeForAPI(p *types.Prediction, includeSummary bool) (compiler.Prediction, error) {
	pred := compiler.Prediction{
		UUID:            p.UUID,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		Reporter:        p.Reporter,
		PostAuthor:      p.PostAuthor,
		PostAuthorURL:   p.PostAuthorURL,
		PostedAt:        p.PostedAt,
		PostURL:         p.PostUrl,
		PredictionState: marshalPredictionState(p.State),
		Type:            p.Type.String(),
		PredictionText:  printer.NewPredictionPrettyPrinter(*p).Default(),
		Summary:         compiler.PredictionSummary{},
	}

	if includeSummary && s.mkt != nil {
		var err error
		pred.Summary, err = s.BuildPredictionMarketSummary(*p)
		if err != nil {
			log.Error().Err(err).Msg("compiler.PreSerializeForAPI: writing summary failed")
			return pred, err
		}
	}

	return pred, nil
}

// SerializeForAPI serializes a Prediction to a JSON []byte. It is meant to be used only by the API.
func (s PredictionSerializer) SerializeForAPI(p *types.Prediction, includeSummary bool) ([]byte, error) {
	pred, err := s.PreSerializeForAPI(p, includeSummary)
	if err != nil {
		return nil, err
	}

	return json.Marshal(pred)
}

func marshalInnerCondition(c *types.Condition) string {
	if c.Operator == "BETWEEN" {
		return fmt.Sprintf(`%v BETWEEN %v AND %v`, c.Operands[0].Str, c.Operands[1].Str, c.Operands[2].Str)
	}
	return fmt.Sprintf(`%v %v %v`, c.Operands[0].Str, c.Operator, c.Operands[1].Str)
}

func marshalGiven(given map[string]*types.Condition) map[string]compiler.Condition {
	result := map[string]compiler.Condition{}
	for key, cond := range given {
		c := compiler.Condition{
			Condition:        marshalInnerCondition(cond),
			FromISO8601:      types.ISO8601(time.Unix(int64(cond.FromTs), 0).Format(time.RFC3339)),
			ToISO8601:        types.ISO8601(time.Unix(int64(cond.ToTs), 0).Format(time.RFC3339)),
			ToDuration:       cond.ToDuration,
			Assumed:          cond.Assumed,
			ErrorMarginRatio: cond.ErrorMarginRatio,
			State: compiler.ConditionState{
				Status:    cond.State.Status.String(),
				LastTs:    cond.State.LastTs,
				LastTicks: cond.State.LastTicks,
				Value:     cond.State.Value.String(),
			},
		}
		result[key] = c
	}
	return result
}

func marshalBoolExpr(b *types.BoolExpr, nestLevel int) (*string, error) {
	if b == nil {
		return nil, nil
	}
	var prefix, suffix string
	if nestLevel > 0 {
		prefix = "("
		suffix = ")"
	}
	switch b.Operator {
	case types.LITERAL:
		return &b.Literal.Name, nil
	case types.AND, types.OR:
		operator := " and "
		if b.Operator == types.OR {
			operator = " or "
		}
		operands := []string{}
		for _, operand := range b.Operands {
			s, err := marshalBoolExpr(operand, nestLevel+1)
			if err != nil {
				return nil, err
			}
			if s == nil {
				continue
			}
			operands = append(operands, *s)
		}
		s := fmt.Sprintf("%v%v%v", prefix, strings.Join(operands, operator), suffix)
		return &s, nil
	case types.NOT:
		operand, err := marshalBoolExpr(b.Operands[0], nestLevel+1)
		if err != nil {
			return nil, err
		}
		s := fmt.Sprintf("%vnot %v%v", prefix, *operand, suffix)
		return &s, nil
	}
	return nil, fmt.Errorf("marshalBoolExpr: unknown operator '%v'", b.Operator)
}

func marshalPrePredict(pp types.PrePredict) (*compiler.PrePredict, error) {
	wrongIf, err := marshalBoolExpr(pp.WrongIf, 0)
	if err != nil {
		return nil, err
	}
	annulledIf, err := marshalBoolExpr(pp.AnnulledIf, 0)
	if err != nil {
		return nil, err
	}
	predictIf, err := marshalBoolExpr(pp.Predict, 0)
	if err != nil {
		return nil, err
	}
	result := &compiler.PrePredict{
		WrongIf:                           wrongIf,
		AnnulledIf:                        annulledIf,
		Predict:                           predictIf,
		AnnulledIfPredictIsFalse:          pp.AnnulledIfPredictIsFalse,
		IgnoreUndecidedIfPredictIsDefined: pp.IgnoreUndecidedIfPredictIsDefined,
	}
	return result, nil
}

func marshalPredict(p types.Predict) (compiler.Predict, error) {
	wrongIf, err := marshalBoolExpr(p.WrongIf, 0)
	if err != nil {
		return compiler.Predict{}, err
	}
	annulledIf, err := marshalBoolExpr(p.AnnulledIf, 0)
	if err != nil {
		return compiler.Predict{}, err
	}
	predictIf, err := marshalBoolExpr(&p.Predict, 0)
	if err != nil {
		return compiler.Predict{}, err
	}
	result := compiler.Predict{
		WrongIf:                           wrongIf,
		AnnulledIf:                        annulledIf,
		Predict:                           *predictIf,
		IgnoreUndecidedIfPredictIsDefined: p.IgnoreUndecidedIfPredictIsDefined,
	}
	return result, nil
}

func marshalPredictionState(ps types.PredictionState) compiler.PredictionState {
	return compiler.PredictionState{
		Status: ps.Status.String(),
		LastTs: ps.LastTs,
		Value:  ps.Value.String(),
	}
}

// AccountSerializer is the component that serializes an Account to a string representation, to be persisted or returned
// in an API call.
type AccountSerializer struct{}

// NewAccountSerializer constructs an AccountSerializer.
func NewAccountSerializer() AccountSerializer {
	return AccountSerializer{}
}

// Account is the struct that represents a post author's social media account.
type Account struct {
	URL           string   `json:"url"`
	AccountType   string   `json:"accountType"`
	Handle        string   `json:"handle"`
	FollowerCount int      `json:"followerCount"`
	Thumbnails    []string `json:"thumbnails"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	CreatedAt     string   `json:"createdAt,omitempty"`
}

// PreSerialize serializes an Account to a compiler.Account, but doesn't take the extra step of serializing it to a
// JSON []byte.
func (s AccountSerializer) PreSerialize(p *types.Account) (Account, error) {
	thumbs := []string{}
	for _, thumb := range p.Thumbnails {
		thumbs = append(thumbs, thumb.String())
	}

	createdAt := ""
	if p.CreatedAt != nil {
		createdAt = p.CreatedAt.Format(time.RFC3339)
	}

	return Account{
		URL:           p.URL.String(),
		AccountType:   p.AccountType,
		Handle:        p.Handle,
		FollowerCount: p.FollowerCount,
		Thumbnails:    thumbs,
		Name:          p.Name,
		Description:   p.Description,
		CreatedAt:     createdAt,
	}, nil
}

// Serialize serializes an Account to a JSON []byte.
func (s AccountSerializer) Serialize(p *types.Account) ([]byte, error) {
	acc, err := s.PreSerialize(p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(acc)
}
