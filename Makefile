# Nom du fichier binaire
BINARY=streaming-service

# Variables pour le chemin du fichier de base de données et les dossiers statiques
DB_FILE=./streaming_service.db
STATIC_DIR=./static

# Commande pour compiler le projet
build:
	@echo "Building the project..."
	go build -o $(BINARY) main.go

# Commande pour exécuter le projet
run: build
	@echo "Running the project..."
	./$(BINARY)

# Commande pour nettoyer les fichiers générés (binaire et base de données)
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY) $(DB_FILE)

# Commande pour initialiser la base de données (au cas où tu voudrais une commande dédiée)
initdb:
	@echo "Initializing the database..."
	sqlite3 $(DB_FILE) "CREATE TABLE IF NOT EXISTS tokens (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT NOT NULL, token TEXT NOT NULL, expires_at TIMESTAMP NOT NULL);"

# Commande pour générer des liens de streaming (pour tester l'API)
gen-link:
	@echo "Generating a streaming link..."
	curl -X POST -d "username=testuser" http://localhost:8080/generate-link

# Commande pour afficher l'aide
help:
	@echo "Usage:"
	@echo "  make build     Build the project"
	@echo "  make run       Run the project"
	@echo "  make clean     Clean the project (remove binary and database)"
	@echo "  make initdb    Initialize the database"
	@echo "  make gen-link  Generate a streaming link"
	@echo "  make help      Show this help message"

# La commande par défaut est d'afficher l'aide
.DEFAULT_GOAL := help
