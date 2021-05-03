package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/senseyeio/spaniel"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func FilterTimeSpans(mainSpan spaniel.Span, spans spaniel.Spans) spaniel.Spans {
	filtered := make(spaniel.Spans, 0, len(spans))

	for _, span := range spans {
		if (mainSpan.Start().Before(span.Start()) || mainSpan.Start().Equal(span.Start())) &&
			(mainSpan.End().After(span.End()) || mainSpan.End().Equal(span.End())) {
			filtered = append(filtered, span)
		}
	}

	return filtered
}

func FlatComplementOfSpanSet(mainSpan spaniel.Span, flatSpanSet spaniel.Spans) spaniel.Spans {
	startMainInterval := mainSpan.Start()
	endMainInterval := mainSpan.End()

	complementsOfTimeSpans := make([]spaniel.Spans, 0, len(flatSpanSet))
	for _, span := range flatSpanSet {
		beforePart := spaniel.New(startMainInterval, span.Start())
		afterPart := spaniel.New(span.End(), endMainInterval)

		spanComplement := spaniel.Spans{beforePart, afterPart}

		complementsOfTimeSpans = append(complementsOfTimeSpans, spanComplement)
	}

	flatComplementOfSpanSet := make(spaniel.Spans, 0, len(flatSpanSet)+1)
	for _, spanComplement := range complementsOfTimeSpans {
		if len(flatComplementOfSpanSet) == 0 {
			flatComplementOfSpanSet = spanComplement
			continue
		}
		flatComplementOfSpanSet = flatComplementOfSpanSet.IntersectionBetween(spanComplement)
	}

	return flatComplementOfSpanSet
}

func TestFreeBusySingle(t *testing.T) {
	var uc EventUseCase
	now := time.Now()

	year, month, day := now.Date()
	loc := now.Location()

	hoursBefore := time.Date(year, month, day, 0, 0, 0, 0, loc)
	hoursAfter := hoursBefore.AddDate(0, 0, 1)

	mainSpan := spaniel.New(hoursBefore, hoursAfter)

	freeBusy := types.FreeBusy{
		Users: []string{"alersh@internet.ru", "mr.eskov1@yandex.ru"},
		From:  mainSpan.Start(),
		To:    mainSpan.End(),
	}

	response, err := uc.GetUsersBusyIntervals("691a5723d5277133991893df990cdb4561138b2337363830", freeBusy)
	assert.NoError(t, err)

	var busyTimeSpan spaniel.Spans
	for _, userSpans := range response.Data.FreeBusy {
		for _, span := range userSpans.FreeBusy {
			busyTimeSpan = append(busyTimeSpan, spaniel.New(span.From, span.To))
		}
	}
	busyFlatTimeSpan := busyTimeSpan.Union()
	fmt.Printf("%+v\n", busyFlatTimeSpan)

	busyFlatFiltered := FilterTimeSpans(mainSpan, busyFlatTimeSpan)
	fmt.Printf("%+v\n", busyFlatFiltered)

	freeTimeSpans := FlatComplementOfSpanSet(mainSpan, busyFlatFiltered)
	fmt.Printf("%+v\n", freeTimeSpans)
}
