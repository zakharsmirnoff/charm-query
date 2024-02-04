## Charm Query (python version)
This repository contains the code for the Go microservice which works as a middle service between your frontend application and the database. 
It's meant to test the hypothesis I describe here: 
*In short: translating plain English to SQL*

The application is just an API service with the following endpoints: 

- ask: ask your question in natural language and get the json with the table-like result from your db
  Request:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews"}'
  ```
  Response:
  ```json
  ```
- execute: provide your SQL query and get the json with the table-like result from your db:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews"}'
  ```
- generate: generate a new query using OpenAI LLM models:
  ```bash
  curl localhost:5000/ask -X POST -H "Content-Type:application/json" -d '{"question": "labels with most reviews"}'
  ```
- add: add your pair of question-query to the vector db
- delete: delete the query from vector db and all its associated questions
### Quickstart:
- If you have Go 1.21.3 + installed, you can just build from source:
```bash
go build -o charm-query .
```
If you don't want to deal with Go, you can just run docker compose which will start Weaviate vector db and CharmQuery app:
```bash
docker compose up -d
```
This app is not meant to be deployed to production (unless you are absolutely confident in what you are doing), rather than serve as a starting point to explore LLM capabilities to translate
natural language to SQL, improved with vector search.

The following application uses the following stack: 
- BYOF (bring your own frontend *lol, I just made this term up, it's actually just naked API*)
- Weaviate for vector search
- OpenAI API (default LLM is gpt-4, default model for embeddings is text-ada-002)
- Virtually any SQL database is supported if you *go get* the necessary driver. Supported drivers are listed here:
