package usecase

import (
	"fmt"
	"github.com/calendar-bot/pkg/types"
	"github.com/senseyeio/spaniel"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFreeBusySingle(t *testing.T) {
	var uc EventUseCase
	now := time.Now().AddDate(0, 0, 0)

	year, month, day := now.Date()
	loc := now.Location()

	hoursBefore := time.Date(year, month, day, 1, 0, 0, 0, loc)
	hoursAfter := time.Date(year, month, day, 23, 0, 0, 0, loc)

	mainSpan := spaniel.New(hoursBefore, hoursAfter)

	freeBusy := types.FreeBusy{
		Users: []string{"mr.eskov1@mail.ru", "alersh@internet.ru"},
		From:  mainSpan.Start(),
		To:    mainSpan.End(),
	}

	response, err := uc.GetUsersBusyIntervals("5b86108c1a0c76e99ea1b1cf7d41acab61138b2337363830", freeBusy)
	assert.NoError(t, err)

	stretchBy := 15 * time.Minute
	fmt.Printf("busyFlatTimeSpanUnstretched: %+v\n", MergeBusyIntervals(response.Data, nil))
	busyFlatTimeSpan := MergeBusyIntervals(response.Data, &stretchBy)
	fmt.Printf("busyFlatTimeSpan: %+v\n", busyFlatTimeSpan)

	busyFlatTruncated := MapSpansWithFunc(busyFlatTimeSpan, TruncateSpanBy(mainSpan))
	fmt.Printf("busyFlatTruncated: %+v\n", busyFlatTruncated)

	freeTimeSpansIterative := CalculateFreeTimeSpans(busyFlatTruncated, mainSpan)

	fmt.Printf("freeTimeSpansIterative: %+v\n", freeTimeSpansIterative)

	splitBy := 30 * time.Minute
	freeTimeSplit := make(spaniel.Spans, 0, len(freeTimeSpansIterative))
	for _, span := range freeTimeSpansIterative {
		spanSplit, _ := SplitSpanBy(span, splitBy)
		freeTimeSplit = append(freeTimeSplit, spanSplit...)
	}
	fmt.Printf("freeTimeSplit: %+v\n", freeTimeSplit)

	dayPart := DayPart{
		Start:    time.Date(year, month, day, 21, 0, 0, 0, loc),
		Duration: 2 * time.Hour,
	}
	//maxDuration := 60 * time.Minute
	//minDuration := 30 * time.Minute

	filteredFreeTimeSpans := FilterSpans(freeTimeSplit, nil, &dayPart, nil, nil)
	fmt.Printf("filteredFreeTimeSpans: %+v\n", filteredFreeTimeSpans)
}

func TestGetUsersFreeIntervals(t *testing.T) {
	var uc EventUseCase
	now := time.Now().AddDate(0, 0, 0)

	year, month, day := now.Date()
	loc := now.Location()

	hoursBefore := time.Date(year, month, day, 1, 0, 0, 0, loc)
	hoursAfter := time.Date(year, month, day+4, 23, 0, 0, 0, loc)

	mainSpan := spaniel.New(hoursBefore, hoursAfter)

	freeBusy := types.FreeBusy{
		Users: []string{"mr.eskov1@mail.ru", "alersh@internet.ru"},
		From:  mainSpan.Start(),
		To:    mainSpan.End(),
	}

	stretchBusyIntervalsBy := 15 * time.Minute
	splitFreeIntervalsBy := 30 * time.Minute
	//maxDuration := 60 * time.Minute
	//minDuration := 30 * time.Minute

	freeIntervals, err := uc.GetUsersFreeIntervals(
		"5b86108c1a0c76e99ea1b1cf7d41acab61138b2337363830",
		freeBusy,
		FreeBusyConfig{
			DayPart: &DayPart{
				Start:    time.Date(year, month, day, 21, 0, 0, 0, loc),
				Duration: 2 * time.Hour,
			},
			StretchBusyIntervalsBy: &stretchBusyIntervalsBy,
			SplitFreeIntervalsBy:   &splitFreeIntervalsBy,
			//MinFreeIntervalDuration: &minDuration,
			//MaxFreeIntervalDuration: &maxDuration,
		},
	)
	assert.NoError(t, err)

	fmt.Printf("%+v\n", freeIntervals)
}

func TestStretchOutSpan(t *testing.T) {
	stretchDur := 15 * time.Minute

	expectedSpan := spaniel.New(
		time.Date(0, 0, 0, 12, 00, 0, 0, time.UTC),
		time.Date(0, 0, 0, 13, 00, 0, 0, time.UTC),
	)

	testSpan := spaniel.New(
		time.Date(0, 0, 0, 12, 11, 0, 0, time.UTC),
		time.Date(0, 0, 0, 12, 47, 0, 0, time.UTC),
	)

	stretched := StretchOutSpan(testSpan, stretchDur)

	assert.Equal(t, expectedSpan, stretched)
}

func TestSplitSpanBy(t *testing.T) {
	splitInterval := 15 * time.Minute

	testSpan := spaniel.New(
		time.Date(0, 0, 0, 12, 4, 0, 0, time.UTC),
		time.Date(0, 0, 0, 13, 7, 0, 0, time.UTC),
	)

	spanSplit, remainder := SplitSpanBy(testSpan, splitInterval)

	assert.NotNil(t, remainder)
	assert.Equal(t, 3*time.Minute, remainder.End().Sub(remainder.Start()))

	for _, span := range spanSplit {
		assert.Equal(t, splitInterval, span.End().Sub(span.Start()))
	}
}

func TestTruncSpanBy(t *testing.T) {
	testSpan := spaniel.New(
		time.Date(0, 0, 0, 11, 55, 0, 0, time.UTC),
		time.Date(0, 0, 0, 13, 07, 0, 0, time.UTC),
	)

	borders := spaniel.New(
		time.Date(0, 0, 0, 12, 00, 0, 0, time.UTC),
		time.Date(0, 0, 0, 13, 00, 0, 0, time.UTC),
	)

	expected := spaniel.New(
		time.Date(0, 0, 0, 12, 00, 0, 0, time.UTC),
		time.Date(0, 0, 0, 13, 00, 0, 0, time.UTC),
	)

	truncated := TruncateSpanBy(borders)(testSpan)

	assert.Equal(t, expected, truncated)
}
