package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*

gin.Context is a struct provided by the Gin web framework in Go.
It carries request-specific information such as request and response objects,
path parameters, query parameters, form data, and more.
It is used to read data from the HTTP request, write responses,
and control the flow of the middleware chain.

*/

// The parameter 'c' is a pointer to gin.Context because it allows the handler to modify the context (e.g., set response data, abort the request, etc.)
// and ensures efficient memory usage by not copying the entire context struct.
func HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}
