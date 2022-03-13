package backoffice

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/marianogappa/predictions/printer"
	"github.com/marianogappa/predictions/types"
)

func predictionToMap(p types.Prediction) map[string]interface{} {
	urlType := "UNKNOWN"
	urlSiteSpecificId := ""
	u, err := url.Parse(p.PostUrl)
	if err == nil {
		hostname := u.Hostname()
		switch {
		case strings.Contains(hostname, "youtube"):
			urlType = "YOUTUBE"
			urlSiteSpecificId = u.Query().Get("v")
		case strings.Contains(hostname, "twitter"):
			urlType = "TWITTER"
		}
	}

	postedAtTime, _ := p.PostedAt.Time()
	postedAt := postedAtTime.Format(time.RFC850)

	m := map[string]interface{}{
		"UUID":              p.UUID,
		"Version":           p.Version,
		"CreatedAt":         p.CreatedAt,
		"PostAuthor":        p.PostAuthor,
		"PostText":          p.PostText,
		"PostedAt":          postedAt,
		"PostUrl":           p.PostUrl,
		"Given":             mapifyGiven(p.Given),
		"PrePredict":        mapifyPrePredict(p.PrePredict),
		"Predict":           mapifyPredict(p.Predict),
		"State":             mapifyState(p.State),
		"Reporter":          p.Reporter,
		"PrettyPrint":       printer.NewPredictionPrettyPrinter(p).Default(),
		"URLType":           urlType,
		"URLSiteSpecificId": urlSiteSpecificId,
	}
	return m
}

func mapifyGiven(given map[string]*types.Condition) map[string]interface{} {
	if given == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range given {
		m[k] = mapifyCondition(v)
	}
	return m
}

func mapifyCondition(c *types.Condition) map[string]interface{} {
	if c == nil {
		return nil
	}

	operands := []string{}
	for _, op := range c.Operands {
		operands = append(operands, op.Str)
	}

	lastTicks := map[string]map[string]interface{}{}
	for k, lt := range c.State.LastTicks {
		lastTicks[k] = mapifyLastTick(lt)
	}

	state := map[string]interface{}{
		"Status":    c.State.Status.String(),
		"LastTs":    c.State.LastTs,
		"LastTicks": lastTicks,
		"Value":     c.State.Value.String(),
	}

	return map[string]interface{}{
		"Name":             c.Name,
		"Operator":         c.Operator,
		"Operands":         operands,
		"FromTs":           c.FromTs,
		"ToTs":             c.ToTs,
		"ToDuration":       c.ToDuration,
		"Assumed":          c.Assumed,
		"State":            state,
		"ErrorMarginRatio": fmt.Sprintf("%v", c.ErrorMarginRatio),
	}
}

func mapifyLastTick(t types.Tick) map[string]interface{} {
	timestamp := time.Unix(int64(t.Timestamp), 0).Format(time.RFC850)
	return map[string]interface{}{
		"Timestamp": timestamp,
		"Value":     t.Value,
	}
}

func mapifyPrePredict(prePredict types.PrePredict) map[string]interface{} {
	return map[string]interface{}{
		"WrongIf":                           mapifyBoolExpr(prePredict.WrongIf),
		"AnnulledIf":                        mapifyBoolExpr(prePredict.AnnulledIf),
		"Predict":                           mapifyBoolExpr(prePredict.Predict),
		"AnnulledIfPredictIsFalse":          prePredict.AnnulledIfPredictIsFalse,
		"IgnoreUndecidedIfPredictIsDefined": prePredict.IgnoreUndecidedIfPredictIsDefined,
	}
}

func mapifyBoolExpr(b *types.BoolExpr) map[string]interface{} {
	if b == nil {
		return nil
	}

	operands := []map[string]interface{}{}
	for _, op := range b.Operands {
		operands = append(operands, mapifyBoolExpr(op))
	}

	return map[string]interface{}{
		"Operator": b.Operator,
		"Operands": operands,
		"Literal":  mapifyCondition(b.Literal),
	}
}

func mapifyPredict(predict types.Predict) map[string]interface{} {
	return map[string]interface{}{
		"WrongIf":                           mapifyBoolExpr(predict.WrongIf),
		"AnnulledIf":                        mapifyBoolExpr(predict.AnnulledIf),
		"Predict":                           mapifyBoolExpr(&predict.Predict),
		"IgnoreUndecidedIfPredictIsDefined": predict.IgnoreUndecidedIfPredictIsDefined,
	}
}

func mapifyState(state types.PredictionState) map[string]interface{} {
	return map[string]interface{}{
		"Status": state.Status.String(),
		"LastTs": state.LastTs,
		"Value":  state.Value.String(),
	}
}
