package main

import (
    "database/sql"
    "os"
    "path/filepath"
    "time"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt"
    _ "modernc.org/sqlite"
    "github.com/sirupsen/logrus"
)

var jwtKey = []byte("your_secret_key")
var log = logrus.New()

type Claims struct {
    Username string `json:"username"`
    jwt.StandardClaims
}

var db *sql.DB

func initDB() {
    var err error
    db, err = sql.Open("sqlite", "./streaming_service.db")
    if err != nil {
        log.Fatal(err)
    }

    createTable := `CREATE TABLE IF NOT EXISTS tokens (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL,
        token TEXT NOT NULL,
        expires_at TIMESTAMP NOT NULL
    );`
    _, err = db.Exec(createTable)
    if err != nil {
        log.Fatal(err)
    }
}

func main() {
    log.Formatter = new(logrus.JSONFormatter)
    log.Level = logrus.InfoLevel

    initDB()
    defer db.Close()

    r := gin.Default()

    r.Static("/static", "./static")
    r.LoadHTMLFiles("static/index.html")

    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })

    r.POST("/generate-link", generateLink)
    r.GET("/stream", stream)
    r.POST("/upload", upload)

    log.Info("Starting server on :8080")
    r.Run(":8080")
}

func generateLink(c *gin.Context) {
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
        log.WithError(err).Error("Could not generate token")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
        return
    }

    _, err = db.Exec("INSERT INTO tokens (username, token, expires_at) VALUES (?, ?, ?)",
        username, tokenString, expirationTime)
    if err != nil {
        log.WithError(err).Error("Could not save token")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save token"})
        return
    }

    log.WithFields(logrus.Fields{
        "username": username,
        "token":    tokenString,
    }).Info("Token generated")

    c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func upload(c *gin.Context) {
    username := c.PostForm("username")

    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file"})
        return
    }

    userDir := filepath.Join("uploads", username)
    err = os.MkdirAll(userDir, os.ModePerm)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create directory"})
        return
    }

    filePath := filepath.Join(userDir, file.Filename)
    if err := c.SaveUploadedFile(file, filePath); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save file"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func stream(c *gin.Context) {
    tokenString := c.Query("token")

    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        if err == jwt.ErrSignatureInvalid {
            log.WithError(err).Error("Unauthorized access attempt")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
        log.WithError(err).Error("Bad request")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
        return
    }

    if !token.Valid {
        log.Error("Unauthorized access with invalid token")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    var expiresAt time.Time
    err = db.QueryRow("SELECT expires_at FROM tokens WHERE token = ?", tokenString).Scan(&expiresAt)
    if err != nil {
        if err == sql.ErrNoRows {
            log.Error("Unauthorized access with unknown token")
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
            return
        }
        log.WithError(err).Error("Database error")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    if time.Now().After(expiresAt) {
        log.Info("Expired token access attempt")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
        return
    }

    userDir := filepath.Join("uploads", claims.Username)
    filePath := filepath.Join(userDir, "video.mp4") // Assumes the video filename is "video.mp4", this can be changed as needed.

    log.WithFields(logrus.Fields{
        "username": claims.Username,
        "token":    tokenString,
    }).Info("Streaming video")

    c.File(filePath)
}
