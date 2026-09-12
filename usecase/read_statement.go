package usecase

import (
	"context"
	"fmt"
	"gofinance/dto/filters"
	"gofinance/dto/transaction"
	"io"
	"mime/multipart"
	"os"
)

type pdfService interface {
	Read(path string) (io.Reader, error)
	ParseItauStatement(reader io.Reader, filters filters.Filter) ([]transaction.Transaction, error)
}

type transactionRepository interface {
	Save(ctx context.Context, transaction []transaction.Transaction) error
}

type ReadStatementUseCase struct {
	pdfSvc                pdfService
	transactionRepository transactionRepository
}

func NewReadStatementUseCase(pdfSvc pdfService) *ReadStatementUseCase {
	return &ReadStatementUseCase{
		pdfSvc: pdfSvc,
		//transactionRepository: transactionRepo,
	}
}

func (c *ReadStatementUseCase) Execute(ctx context.Context, file multipart.File, filters filters.Filter) error {
	fileTemp, err := os.CreateTemp("./file", "*.pdf")
	if err != nil {
		return err
	}
	defer fileTemp.Close()
	defer os.Remove(fileTemp.Name())

	_, err = io.Copy(fileTemp, file)
	if err != nil {
		return err
	}

	reader, err := c.pdfSvc.Read(fileTemp.Name())
	if err != nil {
		return err
	}

	transactions, err := c.pdfSvc.ParseItauStatement(reader, filters)
	if err != nil {
		return err
	}

	for _, transaction := range transactions {
		fmt.Println(transaction.Date)
		fmt.Println(transaction.Amount)
		fmt.Println(transaction.Type)
		fmt.Println(transaction.Description)
	}

	//if err = c.transactionRepository.Save(ctx, transactions); err != nil {
	//	return err
	//}

	return nil
}
