package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"ikimeia_stream/m/set"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	_ "github.com/xeodou/go-sqlcipher"
)

var jwtSecret = os.Getenv("JWT_SECRET")
var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbName = os.Getenv("DB_NAME")
var jwtKey = []byte(jwtSecret)
var log = logrus.New()

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var db *sql.DB

type Service struct {
	Db *sql.DB
	Security Claims
}

func initDB() {
	var err error
    fdbFile := dbName + "?_auth&_auth_user=" + dbUser + "&_auth_pass=" + dbPassword + "&_auth_crypt=sha1"
	db, err = sql.Open("sqlite3", fdbFile)
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
	err := godotenv.Load()
	// var count int
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Formatter = new(logrus.JSONFormatter)
	log.Level = logrus.InfoLevel
	set.PrintSomething("server running")

	// time.AfterFunc(5000, func(){
	// 	fname := "data/serverOutput" + strconv.FormatInt(int64(count), 6) + ".csv"
	// 	count = count + 1
	// 	err := uploadCSV(fname)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// })
	// go func() {
		
	// }()
	initDB()
	er := uploadCSV("output.csv")
	if er != nil {
		log.Fatal("Error uploading")
	}
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
	r.POST("/album", postAlbums)

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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	_, err = db.Exec("INSERT INTO tokens (username, token, expires_at) VALUES (?, ?, ?)",
		username, tokenString, expirationTime)
	if err != nil {
		log.WithError(err).Error("Could not save token")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "Could not save token"})
		return
	}

	log.WithFields(logrus.Fields{
		"username": username,
		"token":    tokenString,
	}).Info("Token generated")

	c.IndentedJSON(http.StatusOK, gin.H{"token": tokenString})
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
	if tokenString == "" {
		log.Error("Token is missing in the query")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

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

	log.Info("User: ", claims.Username)

	// Remplacez "ghost" et "video.mkv" par les noms d'utilisateur et de fichier appropriés
	userDir := filepath.Join("uploads", claims.Username)
	filePath := filepath.Join(userDir, "img.png") // Assumes the video filename is "video.mkv", adjust as needed.

	log.WithFields(logrus.Fields{
		"username": claims.Username,
		"token":    tokenString,
	}).Info("Streaming video")

	// Vérifiez que le fichier existe avant de le servir
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.WithError(err).Error("File not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		log.WithError(err).Error("Error opening file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	defer file.Close()

	// Détection dynamique du type MIME
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		log.WithError(err).Error("Error reading file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	contentType := http.DetectContentType(buffer)
	c.Header("Content-Disposition", "inline")
	c.Header("Content-Type", contentType)

	// Réouvrir le fichier car nous avons déjà lu quelques bytes
	file.Seek(0, 0)
	c.File(filePath)
}

// album represents data about a record album.
type album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}

var albums = []album{
    {ID: "1", Title: "Album 1", Artist: "Artist 1", Price: 19.99},
    {ID: "2", Title: "Album 2", Artist: "Artist 2", Price: 24.99},
}

func postAlbums(c *gin.Context) {
    var newAlbum album

    // Call BindJSON to bind the received JSON to
    // newAlbum.
    if err := c.BindJSON(&newAlbum); err != nil {
        return
    }

    // Add the new album to the slice.
    albums = append(albums, newAlbum)
    c.IndentedJSON(http.StatusCreated, newAlbum)
}

func uploadCSV(filePath string ) error {

	// Exécuter une requête SQL pour extraire les données
	rows, err := db.Query("SELECT * FROM tokens")
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	// Obtenir les colonnes de la requête
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("Failed to get columns: %v", err)
	}
    
    file, err := os.Create(filePath)
    if err!= nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
	defer writer.Flush()
    // Écrire les en-têtes des colonnes dans le fichier CSV
	if err := writer.Write(columns); err != nil {
		log.Fatalf("Failed to write headers to CSV: %v", err)
	}

	// Créer une slice d'interfaces pour stocker les valeurs de chaque ligne
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Lire les lignes et les écrire dans le fichier CSV
	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		record := make([]string, len(columns))
		for i, value := range values {
			if value != nil {
				record[i] = fmt.Sprintf("%v", value)
			}
		}

		if err := writer.Write(record); err != nil {
			log.Fatalf("Failed to write record to CSV: %v", err)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

    set.PrintSomething("success")
    return nil
}