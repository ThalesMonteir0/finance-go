package main

import (
	"github.com/gin-gonic/gin"
	"gofinance/controller/statement"
	"gofinance/services/xpdf"
	usecase2 "gofinance/usecase"
)

func main() {
	router := gin.Default()

	pdfService := xpdf.New()
	usecase := usecase2.NewReadStatementUseCase(pdfService)
	crtl := statement.NewController(usecase)

	router.POST("/upload", crtl.ReceiveStatementPDF)

	if err := router.Run(":5005"); err != nil {
		panic(err.Error())
	}
}
