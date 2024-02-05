package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		log.Fatal("DB_PATH environment variable is not set or empty")
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY variable is not set or empty")
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		log.Fatal("DB_DRIVER variable is not set or empty")
	}

	schemaQ := os.Getenv("SCHEMA_QUERY")
	if schemaQ == "" {
		log.Println("Schema query is not provided or empty, will generate it")
		getSchema(false)
	} else {
		getSchema(true)
	}

	db, err := sql.Open(dbDriver, dbPath)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to the database: ", err)
	}

	log.Println("Ping was successfull, proceeding ..")
	db.Close()

	cfg := weaviate.Config{
		Host:   "weaviate:8080",
		Scheme: "http",
	}

	vDb, err := getVectorClient(cfg)
	if err != nil {
		log.Fatal("Couldn't create a Weaviate client", err)
	}

	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		askHandler(w, r, vDb)
	})

	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		executeHandler(w, r)
	})

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		generateHandler(w, r)
	})

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		addHandler(w, r, vDb)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		deleteHandler(w, r, vDb)
	})

	log.Println("Starting server on port 5000")
	err = http.ListenAndServe(":5000", nil)
	if err != nil {
		log.Println("Error while starting the server: ", err)
	}
}
