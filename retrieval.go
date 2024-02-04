package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

type vectorDbClient struct {
	dbClient  *weaviate.Client
	className string
}

func getVectorClient(cfg weaviate.Config) (*vectorDbClient, error) {
	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &vectorDbClient{dbClient: client}, nil
}

func findQuery(question string, vDb *vectorDbClient) (q string, source string, err error) {
	q, err = vDb.getQuery(question)
	if err != nil {
		return "", "", err
	}
	if q != "" {
		log.Println("Found the query in the db")
		return q, "db", nil
	} else {
		jsonData, err := json.Marshal(schema)
		if err != nil {
			return "", "", err
		}
		schemaStr := string(jsonData)
		log.Println("Couldn't find similar queries, falling back to generation")
		q, err = generate("You should translate all questions to valid "+os.Getenv("DB_DRIVER")+" queries. You should provide ONLY SQL code, without formatting or explanations. For assistance, here is the schema: "+schemaStr, question+" Provide only SQL code")
		if err != nil {
			return "", "", err
		}
	}

	return q, "generated", nil
}

func (vDb *vectorDbClient) checkClass() (err error) {
	className := os.Getenv("DB_COLLECTION_NAME")

	if className == "" {
		vDb.className = "Default"
	} else {
		vDb.className = className
	}

	exists, err := vDb.dbClient.Schema().ClassExistenceChecker().WithClassName(vDb.className).Do(context.Background())
	if err != nil {
		return err
	}

	if exists {
		_, err := vDb.dbClient.Schema().ClassGetter().WithClassName(vDb.className).Do(context.Background())
		if err != nil {
			return err
		}
		return nil
	} else {
		class := &models.Class{
			Class:      vDb.className,
			Vectorizer: "text2vec-openai",
			ModuleConfig: map[string]interface{}{
				"text2vec-openai": map[string]interface{}{},
			},
		}
		err := vDb.dbClient.Schema().ClassCreator().WithClass(class).Do(context.Background())
		if err != nil {
			return err
		}
		err = vDb.addObject("Are you there?", "SELECT 1;")
		if err != nil {
			return err
		}
		return nil
	}
}

func (vDb *vectorDbClient) getQuery(question string) (query string, err error) {

	err = vDb.checkClass()
	if err != nil {
		return "", err
	}

	fields := []graphql.Field{
		{Name: "question"},
		{Name: "query"},
	}

	nearText := vDb.dbClient.GraphQL().NearTextArgBuilder().WithCertainty(0.92).WithConcepts([]string{question})

	result, err := vDb.dbClient.GraphQL().Get().
		WithClassName(vDb.className).
		WithFields(fields...).
		WithNearText(nearText).
		Do(context.Background())
	if err != nil {
		return "", err
	}

	if result.Errors != nil {
		return "", errors.New(result.Errors[0].Message)
	}

	if result.Data != nil {
		get := result.Data["Get"].(map[string]interface{})
		questions := get[vDb.className].([]interface{})
		if len(questions) != 0 {
			firstQuestion := questions[0].(map[string]interface{})
			query = firstQuestion["query"].(string)
			return query, nil
		}
	}

	return "", nil
}

func (vDb *vectorDbClient) addObject(question string, query string) (err error) {
	err = vDb.checkClass()
	if err != nil {
		return err
	}

	_, err = vDb.dbClient.Data().Creator().WithClassName(vDb.className).WithProperties(map[string]interface{}{
		"question": question,
		"query":    query,
	}).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func (vDb *vectorDbClient) deleteObject(query string) (err error) {
	err = vDb.checkClass()
	if err != nil {
		return err
	}

	res, err := vDb.dbClient.Batch().ObjectsBatchDeleter().WithClassName(vDb.className).WithOutput("minimal").WithWhere(filters.Where().WithPath([]string{"query"}).WithOperator(filters.ContainsAll).WithValueText(query)).Do(context.Background())
	if err != nil {
		return err
	}

	if res.Results.Failed > 0 {
		var errorMessage string
		for _, v := range res.Results.Objects {
			errors := v.Errors
			for _, e := range errors.Error {
				em := "ObjectID: " + v.ID.String() + ", " + "Error: " + e.Message + "; "
				errorMessage += em
			}
		}
		return errors.New(errorMessage)
	}

	return nil
}
