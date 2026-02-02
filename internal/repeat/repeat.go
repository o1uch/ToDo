package repeat

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	RepeatIsEmpty     = errors.New("the \"repeat\" field is not set")
	DStartIsIncorrect = errors.New("the \"dstart\" field is set incorrectly")
	UnsupportedRule   = errors.New("the rule used in \"repeat\" is not supported")
	// rules errors
	RuleDExceeded = errors.New("the value of the allowed number of days in the \"d\" rule has been exceeded")
	NoInterval    = errors.New("the rule repetition interval is not specified")
)

func afterNow(date, now time.Time) bool {
	return date.After(now)
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	layout := "20060102"
	var nextDate time.Time

	if repeat == "" {
		return "", RepeatIsEmpty
	}

	validRepeat, err := repeatParseAndValidate(repeat)

	if err != nil {
		return "", err
	}

	dsTime, err := time.Parse(layout, dstart)

	if err != nil {
		return "", DStartIsIncorrect
	}

	nextDate = dsTime

	// на данном шаге имеем: now - time.Time, dsTime = dstart - time.Time, проверенный и корректный repeart - []string

	switch validRepeat[0] {
	case "y":
		for {
			if afterNow(nextDate, now) {
				break
			}
			nextDate = addYear(nextDate)
		}

	}

	strNexDate := nextDate.Format(layout)

	return strNexDate, nil
}

func addYear(date time.Time) time.Time {
	return date.AddDate(1, 0, 0)
}

func repeatParseAndValidate(repeat string) ([]string, error) {

	rules := []string{"d", "y"}

	splitRepeat := strings.Fields(repeat)

	if len(splitRepeat) == 0 {
		return nil, RepeatIsEmpty
	}

	splitRepeat[0] = strings.ToLower(splitRepeat[0])

	if !slices.Contains(rules, splitRepeat[0]) {
		return nil, UnsupportedRule
	}

	switch splitRepeat[0] {
	case "y":
		return splitRepeat, nil
	case "d":
		err := ruleDCheck(splitRepeat)
		if err != nil {
			return nil, err
		}
	}

	return splitRepeat, nil
}

func ruleDCheck(repeat []string) error {

	if len(repeat) == 1 {
		return NoInterval
	}

	repeatNumber, err := strconv.Atoi(repeat[1])

	if err != nil {
		return fmt.Errorf("error checking the number for the specified rule: %w", err)
	}

	if repeatNumber > 400 || repeatNumber < 1 {
		return errors.Join(RuleDExceeded, fmt.Errorf("specified \"%d\", valid 1..400", repeatNumber))
	}

	return nil
}
