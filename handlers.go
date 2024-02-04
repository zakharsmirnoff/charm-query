package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

type Payload struct {
	Question string `json:"question,omitempty"`
	Query    string `json:"query,omitempty"`
}

type Result struct {
	Data   []interface{} `json:"data,omitempty"`
	Query  string        `json:"query,omitempty"`
	Source string        `json:"source,omitempty"`
}

func parseReq(r *http.Request) (p *Payload, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func askHandler(w http.ResponseWriter, r *http.Request, vDb *vectorDbClient) {
	p, err := parseReq(r)

	if err != nil {
		http.Error(w, "Couldn't parse the request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("The user question: " + p.Question)

	dbQuery, source, err := findQuery(p.Question, vDb)
	log.Println("The query: " + dbQuery)
	if err != nil {
		http.Error(w, "Error when searching for query "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := executeQuery(dbQuery)
	if err != nil {
		vDb.deleteObject(dbQuery)
		http.Error(w, "Error when executing the query "+err.Error(), http.StatusBadRequest)
		return
	}
	if source == "generated" {
		err = vDb.addObject(p.Question, dbQuery)
		if err != nil {
			log.Println(err)
		}
	}

	if result != nil {
		r := &Result{}
		r.Data = result
		r.Query = dbQuery
		r.Source = source
		jsonData, err := json.Marshal(r)
		if err != nil {
			http.Error(w, err.Error()+"The generated query: "+dbQuery, http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "Probably the table was empty", http.StatusBadRequest)
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	p, err := parseReq(r)

	if err != nil {
		http.Error(w, "Couldn't parse the request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("The user's query: " + p.Query)

	result, err := executeQuery(p.Query)
	if err != nil {
		http.Error(w, "Error when executing the query "+err.Error(), http.StatusBadRequest)
		return
	}
	if result != nil {
		r := &Result{}
		r.Data = result
		r.Query = p.Query
		r.Source = "manual"
		jsonData, err := json.Marshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "Probably the table was empty", http.StatusBadRequest)
	}
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	p, err := parseReq(r)

	if err != nil {
		http.Error(w, "Couldn't parse the request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("The user's query to generate: " + p.Question)

	jsonData, err := json.Marshal(schema)
	if err != nil {
		http.Error(w, "Couldn't convert schema to json: "+err.Error(), http.StatusBadRequest)
		return
	}
	schemaStr := string(jsonData)
	q, err := generate("You should translate all questions to valid "+os.Getenv("DB_DRIVER")+" queries. You should provide ONLY SQL code, without formatting or explanations. For assistance, here is the schema: "+schemaStr, p.Question+" Provide only SQL code")
	if err != nil {
		http.Error(w, "Couldn't generate: "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := executeQuery(q)
	if err != nil {
		http.Error(w, "Error when executing the query "+err.Error(), http.StatusBadRequest)
		return
	}

	if result != nil {
		r := &Result{}
		r.Data = result
		r.Query = q
		r.Source = "generated"
		jsonData, err := json.Marshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	} else {
		http.Error(w, "Probably the table was empty", http.StatusBadRequest)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request, vDb *vectorDbClient) {
	p, err := parseReq(r)

	if err != nil {
		http.Error(w, "Couldn't parse the request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("New query-question pair to add: " + p.Query + "-" + p.Question)

	err = vDb.addObject(p.Question, p.Query)
	if err != nil {
		http.Error(w, "Couldn't add the object: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func deleteHandler(w http.ResponseWriter, r *http.Request, vDb *vectorDbClient) {
	p, err := parseReq(r)
	if err != nil {
		http.Error(w, "Couldn't parse the request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Trying to delete all pairs with this query: " + p.Query)

	if err = vDb.deleteObject(p.Query); err != nil {
		http.Error(w, "Couldn't delete the object: "+err.Error(), http.StatusBadRequest)
		return
	}
}
