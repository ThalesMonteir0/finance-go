package repositories

import (
	"context"
	"database/sql"
	"gofinance/dto/transaction"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

func (tr *TransactionRepository) Save(ctx context.Context, transactions []transaction.Transaction) error {
	tx, err := tr.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("insert into transactions (description, amount, type, date) values ($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, transct := range transactions {
		_, err = stmt.Exec(transct.Description, transct.Amount, transct.Type, transct.Date)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
