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

	response, err := uc.GetUsersBusyIntervals("d9859057d9c47daf841fd7edbbaf3a2d61138b2337363830", freeBusy)
	assert.NoError(t, err)

	busyFlatTimeSpan := MergeBusyIntervals(response.Data)
	fmt.Printf("%+v\n", busyFlatTimeSpan)

	busyFlatFiltered := FilterTimeSpans(busyFlatTimeSpan, mainSpan, nil, nil, nil)
	fmt.Printf("%+v\n", busyFlatFiltered)

	complements := CreateComplementForEachSpan(busyFlatFiltered, mainSpan)

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

func TestGetUsersFreeIntervals(t *testing.T) {
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

	dayPartSpan := spaniel.New(
		time.Date(year, month, day, 20, 0, 0, 0, loc),
		time.Date(year, month, day, 21, 0, 0, 0, loc),
	)
	maxDuration := 60 * time.Minute
	minDuration := 30 * time.Minute

	freeIntervals, err := uc.GetUsersFreeIntervals(
		"d9859057d9c47daf841fd7edbbaf3a2d61138b2337363830",
		freeBusy,
		FreeBusyConfig{
			DayPartSpan:             dayPartSpan,
			MinFreeIntervalDuration: &minDuration,
			MaxFreeIntervalDuration: &maxDuration,
		},
	)
	assert.NoError(t, err)

	fmt.Printf("%+v\n", freeIntervals)
}
