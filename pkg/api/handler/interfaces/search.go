package interfaces

import "github.com/gin-gonic/gin"

type SearchHandler interface {
	GlobalSearch(ctx *gin.Context)
	Autocomplete(ctx *gin.Context)
	SearchTaxonomy(ctx *gin.Context)
}
