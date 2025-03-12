package middleware

import (
    "fmt"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"error": "no authorization header"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{"error": "invalid authorization format"})
            c.Abort()
            return
        }

        tokenString := parts[1]

        // Parse token
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(secret), nil
        })

        if err != nil {
            fmt.Printf("Token Error: %v\n", err)
            c.JSON(401, gin.H{"error": fmt.Sprintf("invalid token: %v", err)})
            c.Abort()
            return
        }

        if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
            c.Set("claims", claims)
            if sub, ok := claims["sub"].(string); ok {
                c.Set("playerID", sub)
            }
            c.Next()
        } else {
            c.JSON(401, gin.H{"error": "invalid token claims"})
            c.Abort()
            return
        }
    }
}