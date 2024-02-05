## Charm Query
This repository contains the code for the Go microservice which works as a middle service between your frontend application and the database. 
It's meant to test the hypothesis I describe here: 

*In short: translating plain English to SQL*

Demo

[streamlit-main-2024-02-05-13-02-80.webm](https://github.com/zakharsmirnoff/charm-query/assets/89240654/5dc615c9-c9f4-4842-99ff-0e54e5c9ab10)

*This demo is recorded using a Python version of the app, you can find it here: https://github.com/zakharsmirnoff/charm-query-py*

*Go version doesn't have frontend and it's using a different stack, though the functionality is almost the same*

The application is just an API service with the following endpoints: 

- ask: ask your question in natural language and get the json with the table-like result from your db
  Request:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews limit 5"}'
  ```
  Response:
  ```json
  {"data":[{"label":"self-released","num_reviews":420},{"label":"drag city","num_reviews":272},{"label":"sub pop","num_reviews":268},{"label":"thrill jockey","num_reviews":244},{"label":"merge","num_reviews":239}],"query":"SELECT label, COUNT(*) as num_reviews\nFROM labels\nGROUP BY label\nORDER BY num_reviews DESC\nLIMIT 5;","source":"db"}
  ```
- execute: provide your SQL query and get the json with the table-like result from your db:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"query": "SELECT label, COUNT(*) as num_reviews FROM labels GROUP BY label ORDER BY num_reviews DESC LIMIT 5"}'
  ```
  Response:
  ```json
  {"data":[{"label":"self-released","num_reviews":420},{"label":"drag city","num_reviews":272},{"label":"sub pop","num_reviews":268},{"label":"thrill jockey","num_reviews":244},{"label":"merge","num_reviews":239}],"query":"SELECT label, COUNT(*) as num_reviews FROM labels GROUP BY label ORDER BY num_reviews DESC LIMIT 5;","source":"manual"}
  ```
- generate: generate a new query using OpenAI LLM models:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews limit 5"}'
  ```
  Response:
  ```json
  {"data":[{"label":"self-released","num_reviews":420},{"label":"drag city","num_reviews":272},{"label":"sub pop","num_reviews":268},{"label":"thrill jockey","num_reviews":244},{"label":"merge","num_reviews":239}],"query":"SELECT label, COUNT(*) as num_reviews\nFROM labels\nGROUP BY label\nORDER BY num_reviews DESC\nLIMIT 5;","source":"generated"}
  ```
- add: add your pair of question-query to the vector db
  ```bash
  curl localhost:5000/add -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews limit 10", "query": "SELECT label, COUNT(*) as num_reviews FROM labels GROUP BY label ORDER BY num_reviews DESC LIMIT 10;"}'
  ```
  Response: 201 Created
- delete: delete the query from vector db and all its associated questions
  ```bash
  curl localhost:5000/delete -X POST -H "Content-Type:application/json" -d '{"query": "SELECT label, COUNT(*) as num_reviews FROM labels GROUP BY label ORDER BY num_reviews DESC LIMIT 10;"}'
  ```
  Response: 200

So, basically you will always get a json with data, source (can be either "generated", "db" or "manual") and query

### Quickstart:
- Clone the repo
- Set environment variables in .env file:
```text
OPENAI_API_KEY=<your openai_key>
DB_PATH=<connection string to your db>
DB_COLLECTION_NAME=<the name of your db which will create a class/collection in Weaviate> # optional, if you don't specify, it will be set to 'Default'. If you plan to test multiple SQL databases, you'd better set this variable
DB_DRIVER=sqlite3
SCHEMA_QUERY=<sql query to fetch the schema of your db> #optional, will be generated if not specified
```
- Then run Docker compose to start Weaviate and CharmQuery: 
```bash
docker compose up -d
```
- If you want to build Go files, you need to modify/delete/move the Docker files so it won't conflict with the Go binary

This app is not meant to be deployed to production (unless you are absolutely confident in what you are doing), rather than serve as a starting point to explore LLM capabilities to translate
natural language to SQL, improved with vector search.

The application uses the following stack: 
- BYOF (bring your own frontend *lol, I just made this term up, it's actually just naked API*)
- Weaviate for vector search
- OpenAI API (default LLM is gpt-4, default model for embeddings is text-ada-002)
- Virtually any SQL database is supported if you *go get* the necessary driver. Supported drivers are listed here: https://go.dev/wiki/SQLDrivers

__This service is a work in progress, so imminent and drastic changes are very much possible__
