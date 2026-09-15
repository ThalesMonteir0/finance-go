package transaction

import (
	"gofinance/dto/ia"
	"time"
)

type Transaction struct {
	Date        time.Time
	Description string
	Amount      float64
	Type        string
}

type TransactionsResponse struct {
	Transactions               []Transaction                `json:"transactions"`
	PercentagePerEstablishment []ia.EstablishmentPercentage `json:"percentage_per_establishment"`
	TotalAmountOut             float64                      `json:"total_amount_out"`
}
