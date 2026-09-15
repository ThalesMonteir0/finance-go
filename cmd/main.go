package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gofinance/client/verboo"
	"gofinance/config"
	"gofinance/controller/statement"
	"gofinance/services/xpdf"
	"gofinance/usecase"
)

// CORS libera o acesso a partir do front e responde ao preflight OPTIONS.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	env := config.LoadEnv()

	router := gin.Default()

	router.Use(CORS())

	iaClient := verboo.New(env.VerbooAPIKey)
	pdfService := xpdf.New()
	usecase := usecase.NewReadStatementUseCase(pdfService, iaClient)
	crtl := statement.NewController(usecase)

	//routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})
	router.POST("/upload", crtl.ReceiveStatementPDF)

	if err := router.Run(":5005"); err != nil {
		panic(err.Error())
	}
}
