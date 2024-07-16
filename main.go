package main;

import (
    "github.com/gin-gonic/gin"
    "time"
    "net/http"
    "github.com/golang-jwt/jwt"

)

var jwtKey = []byte("your_secret_key")

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

func main() {
    r := gin.Default()

    r.POST("/generate-link", generateLink)
    r.GET("/stream", stream)

    r.Run(":8080")
}

func generateLink(c *gin.Context) {
    // Get the username from the request
    username := c.PostForm("username")
    
    expirationTime := time.Now().Add(15 * time.Minute)
    claims := &Claims{
        Username: username,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: expirationTime.Unix(),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func stream(c *gin.Context) {
    tokenString := c.Query("token")

    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        if err == jwt.ErrSignatureInvalid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
        c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
        return
    }

    if !token.Valid {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    // Stream logic here, e.g., serve a video file
    c.File("/path/to/your/video.mp4")
}
