package xpdf

import (
	"github.com/ledongthuc/pdf"
	"gofinance/dto/filters"
	"gofinance/dto/transaction"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Service struct{}

func New() *Service {
	return &Service{}
}

var dateRegex = regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`)
var valueRegex = regexp.MustCompile(`^-?[\d\.]+,\d{2}$`)

func (s *Service) Read(path string) (io.Reader, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader, err := r.GetPlainText()
	return reader, err
}

func (s *Service) ParseItauStatement(reader io.Reader, filters filters.Filter) ([]transaction.Transaction, float64, error) {
	linesString, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}

	lines := strings.Split(string(linesString), "\n")

	startDt, err := time.Parse("02/01/2006", filters.StartDate)
	if err != nil {
		return nil, 0, err
	}
	endDt, err := time.Parse("02/01/2006", filters.EndDate)
	if err != nil {
		return nil, 0, err
	}

	var transactions []transaction.Transaction
	var TotalAmountOut float64

	for i := 0; i < len(lines)-2; i++ {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		if !dateRegex.MatchString(line) {
			continue
		}

		date, err := time.Parse("02/01/2006", line)
		if err != nil {
			return nil, 0, err
		}

		if date.Before(startDt) || date.After(endDt) {
			continue
		}

		description := strings.TrimSpace(lines[i+1])
		value := strings.TrimSpace(lines[i+2])

		if description == "SALDO DO DIA" {
			continue
		}

		if !valueRegex.MatchString(value) {
			continue
		}

		value = strings.ReplaceAll(value, ".", "")
		value = strings.ReplaceAll(value, ",", ".")

		amount, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}

		matchedDev, err := regexp.MatchString(`(?i)DEV PIX`, description)
		if err != nil {
			continue
		}

		if amount < 0 || matchedDev {
			TotalAmountOut += amount
		}

		txType := "IN"
		if amount < 0 {
			txType = "OUT"
		}

		transactions = append(transactions, transaction.Transaction{
			Date:        date,
			Description: description,
			Amount:      amount,
			Type:        txType,
		})
	}

	return transactions, TotalAmountOut, nil
}
