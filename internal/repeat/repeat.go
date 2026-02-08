package repeat

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const layout = "20060102"

var (
	RepeatIsEmpty     = errors.New("the \"repeat\" field is not set")
	DStartIsIncorrect = errors.New("the \"dstart\" field is set incorrectly")
	UnsupportedRule   = errors.New("the rule used in \"repeat\" is not supported")

	ErrIncorrectInterval = errors.New("the specified interval in the repetition rule is incorrect")
	NoInterval           = errors.New("the rule repetition interval is not specified")
)

type Rule interface {
	Next(time.Time) time.Time
}

// YRule

type YRule struct{}

func (y *YRule) Next(from time.Time) time.Time {
	return from.AddDate(1, 0, 0)
}

// DRule

type DRule struct {
	interval int
}

func ruleDCheck(repeat []string) (DRule, error) {

	if len(repeat) == 1 {
		return DRule{}, NoInterval
	}

	repeatNumber, err := strconv.Atoi(repeat[1])

	if err != nil {
		return DRule{}, fmt.Errorf("error checking the number for the specified rule: %w", err)
	}

	if repeatNumber > 400 || repeatNumber < 1 {
		return DRule{}, errors.Join(ErrIncorrectInterval, fmt.Errorf("specified \"%d\", valid 1..400", repeatNumber))
	}

	return DRule{interval: repeatNumber}, nil
}

func (d *DRule) Next(from time.Time) time.Time {
	return from.AddDate(0, 0, d.interval)
}

// WRule

type WRule struct {
	days map[time.Weekday]bool
}

func ruleWCheck(repeat []string) (WRule, error) {
	week := WRule{days: map[time.Weekday]bool{}}

	if len(repeat) != 2 {
		return WRule{}, NoInterval
	}

	interval := strings.Split(repeat[1], ",")

	for _, v := range interval {
		day, err := strconv.Atoi(v)

		if err != nil {
			return WRule{}, fmt.Errorf("error checking the number for the specified rule: %w", err)
		}

		if day > 7 || day <= 0 {
			return WRule{}, errors.Join(ErrIncorrectInterval, fmt.Errorf("specified \"%d\", valid 1..7", day))
		}

		if day == 7 {
			day = 0
		}

		week.days[time.Weekday(day)] = true

	}

	return week, nil
}

func (w *WRule) Next(from time.Time) time.Time {

	date := from.AddDate(0, 0, 1)

	for {
		if w.days[date.Weekday()] {
			return date
		}
		date = date.AddDate(0, 0, 1)
	}
}

// MRule

type MRule struct {
	months map[int]bool
	days   map[int]bool
}

func ruleMCheck(repeat []string) (MRule, error) {
	month := MRule{
		months: make(map[int]bool),
		days:   make(map[int]bool),
	}
	var DaysInterval, MonthsInterval []string

	if len(repeat) < 2 || len(repeat) > 3 {
		return MRule{}, NoInterval
	}

	if len(repeat) == 3 {

		MonthsInterval = strings.Split(repeat[2], ",")

		for _, m := range MonthsInterval {
			mon, err := strconv.Atoi(m)

			if err != nil {
				return MRule{}, fmt.Errorf("error checking the number for the specified rule: %w", err)
			}

			if mon < 1 || mon > 12 {
				return MRule{}, errors.Join(ErrIncorrectInterval, fmt.Errorf("specified \"%d\", valid 1..12", mon))
			}

			month.months[mon] = true
		}

	}

	DaysInterval = strings.Split(repeat[1], ",")

	for _, d := range DaysInterval {
		day, err := strconv.Atoi(d)

		if err != nil {
			return MRule{}, fmt.Errorf("error checking the number for the specified rule: %w", err)
		}

		if day == 0 || day < -2 || day > 31 {
			return MRule{}, errors.Join(ErrIncorrectInterval, fmt.Errorf("specified \"%d\", valid -2, -1 and 1..31", day))
		}

		month.days[day] = true
	}

	return month, nil

}

func DaysInMonth(date time.Time) int {
	nextMonth := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
	return nextMonth.Day()
}

func (m *MRule) Next(from time.Time) time.Time {

	date := from.AddDate(0, 0, 1)

	for {
		month := int(date.Month())

		// проверка месяца
		if len(m.months) != 0 && !m.months[month] {
			date = date.AddDate(0, 0, 1)
			continue
		}

		day := date.Day()
		dim := DaysInMonth(date)

		match := false

		// обычные дни
		if m.days[day] {
			match = true
		}

		// последний день
		if m.days[-1] && day == dim {
			match = true
		}

		// предпоследний
		if m.days[-2] && day == dim-1 {
			match = true
		}

		if match {
			return date
		}

		date = date.AddDate(0, 0, 1)
	}
}

// main logic

func ParseAndValidateRules(repeat string) (Rule, error) {

	splitRepeat := strings.Fields(repeat)

	if len(splitRepeat) == 0 {
		return nil, RepeatIsEmpty
	}

	splitRepeat[0] = strings.ToLower(splitRepeat[0])

	switch splitRepeat[0] {
	case "y":
		return &YRule{}, nil
	case "d":
		DRule, err := ruleDCheck(splitRepeat)

		if err != nil {
			return nil, err
		}
		return &DRule, nil
	case "w":
		WRule, err := ruleWCheck(splitRepeat)
		if err != nil {
			return nil, err
		}
		return &WRule, nil
	case "m":
		MRule, err := ruleMCheck(splitRepeat)
		if err != nil {
			return nil, err
		}
		return &MRule, nil
	}

	return nil, UnsupportedRule
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {

	var nextDate time.Time

	validRepeat, err := ParseAndValidateRules(repeat)

	if err != nil {
		return "", err
	}

	dsTime, err := time.Parse(layout, dstart)

	if err != nil {
		return "", DStartIsIncorrect
	}

	nextDate = dsTime

	for {
		if nextDate.After(now) {
			break
		}
		nextDate = validRepeat.Next(nextDate)
	}

	strNexDate := nextDate.Format(layout)

	return strNexDate, nil
}
