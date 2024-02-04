package main

import (
	"database/sql"
	"log"
	"os"
)

var schema []interface{}

func getSchema(varExists bool) {
	dbDriver := os.Getenv("DB_DRIVER")
	var schemaQ string
	var err error
	if !varExists {
		schemaQ, err = generate("You should provide ONLY SQL code, without explanations or markdown. ", "I need an SQL query to get the schema of the "+dbDriver+" database. The name of the database is "+os.Getenv("DB_COLLECTION_NAME")+" but if the name is not necessary for the query to succeed, don't specify it. If there is no such query, please provide the closest one which can fetch the information in a most accurate way.")
		if err != nil {
			log.Println("Couldn't generate the schema query: " + err.Error())
		}
	} else {
		schemaQ = os.Getenv("SCHEMA_QUERY")
	}

	schema, err = executeQuery(schemaQ)
	if err != nil {
		log.Println("Couldn't retrieve the schema: " + err.Error())
	}
}

func executeQuery(query string) ([]interface{}, error) {
	db, err := sql.Open(os.Getenv("DB_DRIVER"), os.Getenv("DB_PATH"))
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	count := len(columns)

	var v []interface{}

	for rows.Next() {
		values := make([]interface{}, count)
		valuePtrs := make([]interface{}, count)
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i := range columns {
			val := valuePtrs[i].(*interface{})
			b, ok := (*val).([]byte)
			if ok {
				m[columns[i]] = string(b)
			} else {
				m[columns[i]] = *val
			}
		}
		v = append(v, m)
	}
	return v, nil
}
