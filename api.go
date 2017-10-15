package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/graphql-go/graphql"
)

type event struct {
	Events []struct {
		Action string `json:"action"`
		Target struct {
			Repository string `json:"repository"`
			Tag        string `json:"tag"`
		} `json:"target"`
	} `json:"events"`
}

func handleRegistryEvent(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("can't read body")
		return
	}
	ev := event{}
	json.Unmarshal(body, &ev)
	if ev.Events[0].Action == "push" && ev.Events[0].Target.Tag == "latest" {
		log.Printf("%s %s %s, may need restart ...", ev.Events[0].Action, ev.Events[0].Target.Repository, ev.Events[0].Target.Tag)
		r := record{}
		getRecord(&r, ev.Events[0].Target.Repository)
		go restartDockerWithNewImage(&r, false)
	}
}

var glRecord = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Record",
		Fields: graphql.Fields{
			"repo": &graphql.Field{
				Type: graphql.String,
			},
			"version": &graphql.Field{
				Type: graphql.String,
			},
			"time": &graphql.Field{
				Type: graphql.DateTime,
			},
			"ports": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Query",
				Fields: graphql.Fields{
					"containers": &graphql.Field{
						Type: graphql.NewList(glRecord),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							var records []record
							err := queryRecords(&records)
							return records, err
						},
					},
				},
			}),
		Mutation: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "Mutation",
				Fields: graphql.Fields{
					"ports": &graphql.Field{
						Type: graphql.NewList(glRecord),
						Args: graphql.FieldConfigArgument{
							"value": &graphql.ArgumentConfig{
								Type: graphql.NewNonNull(graphql.String),
							},
							"container": &graphql.ArgumentConfig{
								Type: graphql.NewNonNull(graphql.String),
							},
						},
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							repo := p.Args["container"].(string)
							r := updateRecord(repo, record{
								Ports: p.Args["value"].(string),
							})
							getRecord(r, repo)
							restartDockerWithNewImage(r, true)
							var records []record
							err := queryRecords(&records)
							return records, err
						},
					},
				},
			}),
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("wrong result, unexpected errors: %v", result.Errors)
	}
	return result
}

func handleGraphQL(w http.ResponseWriter, r *http.Request) {
	result := executeQuery(r.URL.Query().Get("query"), schema)
	json.NewEncoder(w).Encode(result)
}

var upgrader = websocket.Upgrader{}

func handleLogger(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Origin")
	repo := r.URL.Query()["repo"][0]

	buf, err := attachLog(repo)
	if err != nil {
		log.Printf("attach to container fail:%s", err.Error())
		return
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		line, err := buf.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("read log fail:%s", err.Error())
			return
		}
		if err == io.EOF {
			time.Sleep(time.Second)
		}
		if len(line) > 0 {
			err = c.WriteMessage(1, []byte(line+"\n"))
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
}

func webTask(port int) {
	http.Handle("/", http.FileServer(assetFS()))
	http.HandleFunc("/event", handleRegistryEvent)
	http.HandleFunc("/log", handleLogger)
	http.HandleFunc("/graphql", handleGraphQL)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
