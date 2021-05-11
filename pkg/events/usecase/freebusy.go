package usecase

import (
	"github.com/calendar-bot/pkg/types"
	"github.com/senseyeio/spaniel"
	"time"
)

// ------------span modifiers----

func StretchOutSpan(span spaniel.Span, d time.Duration) spaniel.Span {
	/* nickeskov:
	for example: span is 12:11 - 12:47; d is 15minutes
	start evaluates in 12:00
	end evaluates in 13:00
	*/
	start := span.Start().Truncate(d)
	end := span.End().Truncate(d)
	// nickeskov: if end truncated, add d
	if !end.Equal(span.End()) {
		end = end.Add(d)
	}
	return spaniel.New(start, end)
}

func SplitSpanBy(span spaniel.Span, splitInterval time.Duration) (spanSplit spaniel.Spans, remainder spaniel.Span) {
	spanSplit = spaniel.Spans{}
	remainder = nil
	for span.End().Sub(span.Start()) >= splitInterval {
		nextStart := span.Start().Add(splitInterval)
		spanSplit = append(spanSplit, spaniel.New(span.Start(), nextStart))
		span = spaniel.New(nextStart, span.End())
	}
	if span.Start() != span.End() {
		remainder = span
	}
	return spanSplit, remainder
}

// ------------span modifiers----

// ------------filters------------

type SpanFilterFunc func(span spaniel.Span) bool

type FreeBusyConfig struct {
	DayPart                 *types.DayPart
	StretchBusyIntervalsBy  *time.Duration
	SplitFreeIntervalsBy    *time.Duration
	MinFreeIntervalDuration *time.Duration
	MaxFreeIntervalDuration *time.Duration
}

func MergeSpanFilters(filters ...SpanFilterFunc) SpanFilterFunc {
	return func(span spaniel.Span) bool {
		for _, filter := range filters {
			if !filter(span) {
				return false
			}
		}
		return true
	}
}

func NotInInterval(span, borders spaniel.Span) bool {
	return borders.Start().After(span.Start()) || borders.End().Before(span.End())
}

func NotInDayPart(span spaniel.Span, part types.DayPart) bool {
	hour, minute, second := part.Start.Clock()
	nanosecond := part.Start.Nanosecond()

	year, month, day := span.Start().Date()
	loc := span.Start().Location()

	startBorder := time.Date(year, month, day, hour, minute, second, nanosecond, loc)
	endBorder := startBorder.Add(part.Duration)

	return span.Start().Before(startBorder) || span.End().After(endBorder)
}

func GreaterOrEqualThanDuration(span spaniel.Span, minDuration time.Duration) bool {
	spanDuration := span.End().Sub(span.Start())
	return spanDuration >= minDuration
}

func LessOrEqualThanDuration(span spaniel.Span, maxDuration time.Duration) bool {
	spanDuration := span.End().Sub(span.Start())
	return spanDuration <= maxDuration
}

func FilterSpansWithFunc(spans spaniel.Spans, filter SpanFilterFunc) spaniel.Spans {
	filtered := make(spaniel.Spans, 0, len(spans))
	for _, span := range spans {
		if filter(span) {
			filtered = append(filtered, span)
		}
	}
	return filtered
}

func FilterSpans(spans spaniel.Spans,
	mainBordersSpan spaniel.Span, dayPart *types.DayPart,
	minDuration *time.Duration, maxDuration *time.Duration) spaniel.Spans {

	filters := make([]SpanFilterFunc, 0)

	if mainBordersSpan != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			return !NotInInterval(span, mainBordersSpan)
		})
	}

	if dayPart != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			return !NotInDayPart(span, *dayPart)
		})
	}

	if minDuration != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			return GreaterOrEqualThanDuration(span, *minDuration)
		})
	}

	if maxDuration != nil {
		filters = append(filters, func(span spaniel.Span) bool {
			return LessOrEqualThanDuration(span, *maxDuration)
		})
	}

	return FilterSpansWithFunc(spans, MergeSpanFilters(filters...))
}

// ------------filters------------

// ------------mappers------------

type SpanMapFunc func(span spaniel.Span) spaniel.Span

func TruncateSpanBy(borders spaniel.Span) SpanMapFunc {
	return func(span spaniel.Span) spaniel.Span {
		start := span.Start()
		end := span.End()

		if start.Before(borders.Start()) {
			start = borders.Start()
		}
		if end.After(borders.End()) {
			end = borders.End()
		}

		return spaniel.New(start, end)
	}
}

func MapSpansWithFunc(spans spaniel.Spans, mapper SpanMapFunc) spaniel.Spans {
	mapped := make(spaniel.Spans, 0, len(spans))
	for _, span := range spans {
		mapped = append(mapped, mapper(span))
	}
	return mapped
}

// ------------mappers------------

// ------------complements--------

func FlatComplementOfSpanComplements(complements []spaniel.Spans) spaniel.Spans {
	flatComplementOfSpanSet := spaniel.Spans{}
	if len(complements) > 0 {
		firstSpanComplement := complements[0]
		flatComplementOfSpanSet = make(spaniel.Spans, len(firstSpanComplement))
		copy(flatComplementOfSpanSet, firstSpanComplement)
		complements = complements[1:]
	}
	for _, spanComplement := range complements {
		flatComplementOfSpanSet = flatComplementOfSpanSet.IntersectionBetween(spanComplement)
	}
	return flatComplementOfSpanSet
}

func CreateComplementForEachSpan(flatSpanSet spaniel.Spans, complementBorders spaniel.Span) []spaniel.Spans {
	startMainInterval := complementBorders.Start()
	endMainInterval := complementBorders.End()

	complementsOfTimeSpans := make([]spaniel.Spans, 0, len(flatSpanSet))
	for _, span := range flatSpanSet {
		beforePart := spaniel.New(startMainInterval, span.Start())
		afterPart := spaniel.New(span.End(), endMainInterval)

		spanComplement := spaniel.Spans{beforePart, afterPart}

		complementsOfTimeSpans = append(complementsOfTimeSpans, spanComplement)
	}

	return complementsOfTimeSpans
}

// ------------complements--------

// ------------helpers------------

func MergeBusyIntervals(freeBusyUser types.FreeBusyUser, stretchBy *time.Duration) spaniel.Spans {
	busyTimeSpans := spaniel.Spans{}
	for _, userSpans := range freeBusyUser.FreeBusy {
		for _, span := range userSpans.FreeBusy {
			var span spaniel.Span = span
			if stretchBy != nil {
				span = StretchOutSpan(span, *stretchBy)
			}
			busyTimeSpans = append(busyTimeSpans, span)
		}
	}
	return busyTimeSpans.Union()
}

func CalculateFreeTimeSpans(busy spaniel.Spans, complementBorders spaniel.Span) spaniel.Spans {
	if len(busy) == 0 {
		return spaniel.Spans{complementBorders}
	}
	complements := CreateComplementForEachSpan(busy, complementBorders)
	freeTimeSpans := FlatComplementOfSpanComplements(complements)
	return freeTimeSpans
}

// ------------helpers------------
