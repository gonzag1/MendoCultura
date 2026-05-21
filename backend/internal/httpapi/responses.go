package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func fail(c *gin.Context, status int, code string, message string) {
	c.AbortWithStatusJSON(status, errorResponse{Error: code, Message: message})
}

func badRequest(c *gin.Context, message string) {
	fail(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

func conflict(c *gin.Context, message string) {
	fail(c, http.StatusConflict, "CONFLICT", message)
}

func forbidden(c *gin.Context, message string) {
	fail(c, http.StatusForbidden, "FORBIDDEN", message)
}

func notFound(c *gin.Context, message string) {
	fail(c, http.StatusNotFound, "NOT_FOUND", message)
}

func unauthorized(c *gin.Context, message string) {
	fail(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func serverError(c *gin.Context, message string) {
	fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
