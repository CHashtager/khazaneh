package bot

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type ParsedTransaction struct {
	Kind     string
	Amount   string
	Category string
	Note     string
}

func ParseTransaction(text string) (ParsedTransaction, error) {
	text = strings.TrimSpace(normalizeDigits(text))
	if text == "" {
		return ParsedTransaction{}, errors.New("empty message")
	}

	fields := strings.Fields(text)
	kind := "expense"
	if len(fields) > 0 {
		switch strings.ToLower(fields[0]) {
		case "income", "+", "daramad":
			kind = "income"
			fields = fields[1:]
		case "expense", "-", "kharj":
			kind = "expense"
			fields = fields[1:]
		}
	}
	if len(fields) == 0 {
		return ParsedTransaction{}, errors.New("missing amount")
	}

	amountToken := strings.ReplaceAll(fields[0], ",", "")
	amountToken = strings.ReplaceAll(amountToken, "_", "")
	amount, err := strconv.ParseFloat(amountToken, 64)
	if err != nil || amount <= 0 {
		return ParsedTransaction{}, errors.New("amount must be the first value")
	}

	category := "uncategorized"
	noteFields := []string{}
	if len(fields) > 1 {
		category = fields[1]
		noteFields = fields[2:]
	}

	return ParsedTransaction{
		Kind:     kind,
		Amount:   fmt.Sprintf("%.2f", amount),
		Category: cleanWord(category),
		Note:     strings.Join(noteFields, " "),
	}, nil
}

func normalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '۰', '٠':
			b.WriteRune('0')
		case '۱', '١':
			b.WriteRune('1')
		case '۲', '٢':
			b.WriteRune('2')
		case '۳', '٣':
			b.WriteRune('3')
		case '۴', '٤':
			b.WriteRune('4')
		case '۵', '٥':
			b.WriteRune('5')
		case '۶', '٦':
			b.WriteRune('6')
		case '۷', '٧':
			b.WriteRune('7')
		case '۸', '٨':
			b.WriteRune('8')
		case '۹', '٩':
			b.WriteRune('9')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func cleanWord(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}
