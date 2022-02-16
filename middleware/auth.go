package middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func AuthMiddleware(c *gin.Context) {
	var cert *x509.Certificate

	authToken := c.GetHeader("Authorization")
	idToken := strings.TrimSpace(strings.Replace(authToken, "Bearer", "", 1))
	if idToken == "" {
		c.JSON(401, gin.H{"error": "You are not authorized"})
		c.Abort()
		return
	}
	rsaPublicKey, err := ioutil.ReadFile("code-red-app.pem")
	if err != nil {
		fmt.Println("error reading file: ", err)
	}
	block, _ := pem.Decode(rsaPublicKey)
	cert, _ = x509.ParseCertificate(block.Bytes)
	pub := cert.PublicKey.(*rsa.PublicKey)

	token, err := jwt.ParseWithClaims(idToken, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return pub, nil
	})
	if token.Valid {
		c.Next()
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		c.JSON(403, gin.H{"error": "You are not authorized"})
		c.AbortWithStatus(401)
		return
	} else if errors.Is(err, jwt.ErrTokenMalformed) {
		c.JSON(http.StatusBadRequest, gin.H{"message": jwt.ErrTokenMalformed.Error()})
		c.Abort()
		return
	}
}
