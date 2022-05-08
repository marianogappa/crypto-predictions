package compiler

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/marianogappa/predictions/market"
	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

type PredictionSerializer struct {
	mkt *market.IMarket
}

func NewPredictionSerializer(market *market.IMarket) PredictionSerializer {
	return PredictionSerializer{mkt: market}
}

func (s PredictionSerializer) PreSerialize(p *types.Prediction) (Prediction, error) {
	pp, err := marshalPrePredict(p.PrePredict)
	if err != nil {
		return Prediction{}, err
	}
	pd, err := marshalPredict(p.Predict)
	if err != nil {
		return Prediction{}, err
	}
	return Prediction{
		UUID:            p.UUID,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		Reporter:        p.Reporter,
		PostAuthor:      p.PostAuthor,
		PostAuthorURL:   p.PostAuthorURL,
		PostedAt:        p.PostedAt,
		PostUrl:         p.PostUrl,
		Given:           marshalGiven(p.Given),
		PrePredict:      pp,
		Predict:         pd,
		PredictionState: marshalPredictionState(p.State),
		Type:            p.Type.String(),
	}, nil
}

func (s PredictionSerializer) Serialize(p *types.Prediction) ([]byte, error) {
	pre, err := s.PreSerialize(p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(pre)
}

func (s PredictionSerializer) PreSerializeForAPI(p *types.Prediction, includeSummary bool) (Prediction, error) {
	pred := Prediction{
		UUID:            p.UUID,
		Version:         p.Version,
		CreatedAt:       p.CreatedAt,
		Reporter:        p.Reporter,
		PostAuthor:      p.PostAuthor,
		PostAuthorURL:   p.PostAuthorURL,
		PostedAt:        p.PostedAt,
		PostUrl:         p.PostUrl,
		PredictionState: marshalPredictionState(p.State),
		Type:            p.Type.String(),
		PredictionText:  printer.NewPredictionPrettyPrinter(*p).Default(),
		Summary:         PredictionSummary{},
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

func marshalGiven(given map[string]*types.Condition) map[string]Condition {
	result := map[string]Condition{}
	for key, cond := range given {
		c := Condition{
			Condition:        marshalInnerCondition(cond),
			FromISO8601:      types.ISO8601(time.Unix(int64(cond.FromTs), 0).Format(time.RFC3339)),
			ToISO8601:        types.ISO8601(time.Unix(int64(cond.ToTs), 0).Format(time.RFC3339)),
			ToDuration:       cond.ToDuration,
			Assumed:          cond.Assumed,
			ErrorMarginRatio: cond.ErrorMarginRatio,
			State: ConditionState{
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

func marshalPrePredict(pp types.PrePredict) (*PrePredict, error) {
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
	result := &PrePredict{
		WrongIf:                           wrongIf,
		AnnulledIf:                        annulledIf,
		Predict:                           predictIf,
		AnnulledIfPredictIsFalse:          pp.AnnulledIfPredictIsFalse,
		IgnoreUndecidedIfPredictIsDefined: pp.IgnoreUndecidedIfPredictIsDefined,
	}
	return result, nil
}

func marshalPredict(p types.Predict) (Predict, error) {
	wrongIf, err := marshalBoolExpr(p.WrongIf, 0)
	if err != nil {
		return Predict{}, err
	}
	annulledIf, err := marshalBoolExpr(p.AnnulledIf, 0)
	if err != nil {
		return Predict{}, err
	}
	predictIf, err := marshalBoolExpr(&p.Predict, 0)
	if err != nil {
		return Predict{}, err
	}
	result := Predict{
		WrongIf:                           wrongIf,
		AnnulledIf:                        annulledIf,
		Predict:                           *predictIf,
		IgnoreUndecidedIfPredictIsDefined: p.IgnoreUndecidedIfPredictIsDefined,
	}
	return result, nil
}

func marshalPredictionState(ps types.PredictionState) PredictionState {
	return PredictionState{
		Status: ps.Status.String(),
		LastTs: ps.LastTs,
		Value:  ps.Value.String(),
	}
}

type AccountSerializer struct {
}

func NewAccountSerializer() AccountSerializer {
	return AccountSerializer{}
}

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
func (s AccountSerializer) Serialize(p *types.Account) ([]byte, error) {
	acc, err := s.PreSerialize(p)
	if err != nil {
		return nil, err
	}
	return json.Marshal(acc)
}
