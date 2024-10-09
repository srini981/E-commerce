package middleware

import (
	"fmt"
	"net/http"

	token "monk/tokens"

	"github.com/gin-gonic/gin"
)

func Authentication(c *gin.Context) {
	ClientToken := c.Request.Header.Get("token")
	if ClientToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No Authorization Header Provided"})
		c.Abort()
		return
	}
	claims, err := token.ValidateToken(ClientToken)
	if err != "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		c.Abort()
		return
	}
	fmt.Println(claims)
	c.Set("email", claims.Email)
	c.Set("uid", claims.Id)
	c.Next()
}
