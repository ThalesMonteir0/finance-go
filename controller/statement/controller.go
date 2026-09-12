package statement

import (
	"context"
	"github.com/gin-gonic/gin"
	"gofinance/dto/filters"
	"mime/multipart"
	"net/http"
)

type readStatementUseCase interface {
	Execute(ctx context.Context, file multipart.File, filters filters.Filter) error
}
type Controller struct {
	readStatementUC readStatementUseCase
}

func NewController(readStatementUC readStatementUseCase) *Controller {
	return &Controller{
		readStatementUC: readStatementUC,
	}
}

func (co *Controller) ReceiveStatementPDF(c *gin.Context) {
	ctx := c.Request.Context()
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var filter filters.Filter
	startDate, ok := c.GetQuery("startDate")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate is required"})
		return
	}
	endDate, ok := c.GetQuery("endDate")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate is required"})
		return
	}

	filter.StartDate = startDate
	filter.EndDate = endDate

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err = co.readStatementUC.Execute(ctx, file, filter); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, "ok")
}
