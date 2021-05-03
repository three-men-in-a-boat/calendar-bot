package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/senseyeio/spaniel"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

// ------------filters------------

type FilterFunc func(span spaniel.Span) bool

func mergeSpanFilters(filters ...FilterFunc) FilterFunc {
	return func(span spaniel.Span) bool {
		for _, filter := range filters {
			if !filter(span) {
				return false
			}
		}
		return true
	}
}

func notInInterval(span, borders spaniel.Span) bool {
	return borders.Start().After(span.Start()) || borders.End().Before(span.End())
}

func notInDayPart(span, dayBorders spaniel.Span) bool {
	spanStartHour, spanStartMinute, spanStartSecond := span.Start().Clock()
	spanStartNanosecond := span.Start().Nanosecond()
	spanStartLoc := span.Start().Location()

	spanStart := time.Date(
		0, 0, 0,
		spanStartHour, spanStartMinute, spanStartSecond, spanStartNanosecond,
		spanStartLoc,
	)

	spanEndHour, spanEndMinute, spanEndSecond := span.End().Clock()
	spanEndNanosecond := span.End().Nanosecond()
	spanEndLoc := span.End().Location()

	spanEnd := time.Date(
		0, 0, 0,
		spanEndHour, spanEndMinute, spanEndSecond, spanEndNanosecond,
		spanEndLoc,
	)

	borderStartHour, borderStartMinute, borderStartSecond := dayBorders.Start().Clock()
	borderStartNanosecond := dayBorders.Start().Nanosecond()
	borderStartLoc := dayBorders.Start().Location()

	startBorder := time.Date(
		0, 0, 0,
		borderStartHour, borderStartMinute, borderStartSecond, borderStartNanosecond,
		borderStartLoc,
	)

	borderEndHour, borderEndMinute, borderEndSecond := dayBorders.End().Clock()
	borderEndNanosecond := dayBorders.End().Nanosecond()
	borderEndLoc := dayBorders.End().Location()

	endBorder := time.Date(
		0, 0, 0,
		borderEndHour, borderEndMinute, borderEndSecond, borderEndNanosecond,
		borderEndLoc,
	)

	return spanStart.Before(startBorder) || spanEnd.After(endBorder)
}

func greaterOrEqualThanDuration(span spaniel.Span, minDuration time.Duration) bool {
	spanDuration := span.End().Sub(span.Start())
	return spanDuration >= minDuration
}

func lessOrEqualThanDuration(span spaniel.Span, maxDuration time.Duration) bool {
	spanDuration := span.End().Sub(span.Start())
	return spanDuration <= maxDuration
}

func FilterTimeSpansWithFunc(spans spaniel.Spans, filter FilterFunc) spaniel.Spans {
	filtered := make(spaniel.Spans, 0, len(spans))

	for _, span := range spans {
		if filter(span) {
			filtered = append(filtered, span)
		}
	}

	return filtered
}

func FilterTimeSpans(spans spaniel.Spans,
	mainSpan spaniel.Span, dayPartSpan spaniel.Span,
	minDuration *time.Duration, maxDuration *time.Duration) spaniel.Spans {

	filters := make([]FilterFunc, 0)

	if mainSpan != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			if notInInterval(span, mainSpan) {
				return false
			}
			return true
		})
	}

	if dayPartSpan != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			if notInDayPart(span, dayPartSpan) {
				return false
			}
			return true
		})
	}

	if minDuration != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			if !greaterOrEqualThanDuration(span, *minDuration) {
				return false
			}
			return true
		})
	}

	if maxDuration != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			if !lessOrEqualThanDuration(span, *maxDuration) {
				return false
			}
			return true
		})
	}

	return FilterTimeSpansWithFunc(spans, mergeSpanFilters(filters...))
}

// ------------filters------------

// ------------complements--------

func FlatComplementOfSpanComplementsRecursive(complements []spaniel.Spans) spaniel.Spans {
	switch len(complements) {
	case 0:
		return spaniel.Spans{}
	case 1:
		return complements[0]
	case 2:
		return complements[0].IntersectionBetween(complements[1])
	default:
		left := FlatComplementOfSpanComplementsRecursive(complements[:len(complements)/2])
		right := FlatComplementOfSpanComplementsRecursive(complements[len(complements)/2:])
		return left.IntersectionBetween(right)
	}
}

func FlatComplementOfSpanComplementsIterative(complements []spaniel.Spans) spaniel.Spans {
	flatComplementOfSpanSet := spaniel.Spans{}
	if len(complements) > 0 {
		spanComplement := complements[0]
		complements = complements[1:]
		flatComplementOfSpanSet = make(spaniel.Spans, len(spanComplement))
		copy(flatComplementOfSpanSet, spanComplement)
	}
	for _, spanComplement := range complements {
		flatComplementOfSpanSet = flatComplementOfSpanSet.IntersectionBetween(spanComplement)
	}
	return flatComplementOfSpanSet
}

func CreateComplementForEachSpan(flatSpanSet spaniel.Spans, mainSpan spaniel.Span) []spaniel.Spans {
	startMainInterval := mainSpan.Start()
	endMainInterval := mainSpan.End()

	complementsOfTimeSpans := make([]spaniel.Spans, 0, len(flatSpanSet))
	for _, span := range flatSpanSet {
		beforePart := spaniel.New(startMainInterval, span.Start())
		afterPart := spaniel.New(span.End(), endMainInterval)

		spanComplement := spaniel.Spans{beforePart, afterPart}

		complementsOfTimeSpans = append(complementsOfTimeSpans, spanComplement)
	}

	return complementsOfTimeSpans
}

//func FlatComplementOfSpanSet(flatSpanSet spaniel.Spans, mainSpan spaniel.Span) spaniel.Spans {
//	complementsOfTimeSpans := CreateComplementForEachSpan(flatSpanSet, mainSpan)
//	return FlatComplementOfSpanComplementsIterative(complementsOfTimeSpans)
//}

// ------------complements--------

func MergeBusyIntervals(freeBusyUser types.FreeBusyUser) spaniel.Spans {
	busyTimeSpans := spaniel.Spans{}
	for _, userSpans := range freeBusyUser.FreeBusy {
		for _, span := range userSpans.FreeBusy {
			busyTimeSpans = append(busyTimeSpans, spaniel.New(span.From, span.To))
		}
	}
	return busyTimeSpans.Union()
}

func TestFreeBusySingle(t *testing.T) {
	var uc EventUseCase
	now := time.Now()

	year, month, day := now.Date()
	loc := now.Location()

	hoursBefore := time.Date(year, month, day, 5, 0, 0, 0, loc)
	hoursAfter := time.Date(year, month, day, 21, 0, 0, 0, loc)

	mainSpan := spaniel.New(hoursBefore, hoursAfter)

	freeBusy := types.FreeBusy{
		Users: []string{"mr.eskov1@mail.ru", "alersh@internet.ru"},
		From:  mainSpan.Start(),
		To:    mainSpan.End(),
	}

	response, err := uc.GetUsersBusyIntervals("2307ff9aa0517ccca3e22af9b8ea952b61138b2337363830", freeBusy)
	assert.NoError(t, err)

	busyFlatTimeSpan := MergeBusyIntervals(response.Data)
	fmt.Printf("%+v\n", busyFlatTimeSpan)

	busyFlatFiltered := FilterTimeSpans(busyFlatTimeSpan, mainSpan, nil, nil, nil)
	fmt.Printf("%+v\n", busyFlatFiltered)

	complements := CreateComplementForEachSpan(busyFlatTimeSpan, mainSpan)

	freeTimeSpansIterative := FlatComplementOfSpanComplementsIterative(complements)
	freeTimeSpansRecursive := FlatComplementOfSpanComplementsRecursive(complements)
	assert.Equal(t, freeTimeSpansIterative, freeTimeSpansRecursive)

	fmt.Printf("%+v\n", freeTimeSpansIterative)

	dayPartSpan := spaniel.New(
		time.Date(year, month, day, 20, 0, 0, 0, loc),
		time.Date(year, month, day, 21, 0, 0, 0, loc),
	)
	maxDuration := 60 * time.Minute
	minDuration := 30 * time.Minute

	filteredFreeTimeSpans := FilterTimeSpans(freeTimeSpansIterative, nil, dayPartSpan, &minDuration, &maxDuration)
	fmt.Printf("%+v\n", filteredFreeTimeSpans)
}
