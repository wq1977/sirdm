package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ahmetalpbalkan/dlog"
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

var glState = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Record",
		Fields: graphql.Fields{
			"repo": &graphql.Field{
				Type: graphql.String,
			},
			"state": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

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
			"env": &graphql.Field{
				Type: graphql.String,
			},
			"vols": &graphql.Field{
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
					"state": &graphql.Field{
						Type: graphql.NewList(glState),
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							var records []record
							err := queryRecords(&records)
							if err != nil {
								return nil, err
							}
							res := make([]struct {
								Repo  string `json:"repo"`
								State string `json:"state"`
							}, len(records))
							for idx, container := range records {
								state, err := queryContainerState(container.Repository)
								if err != nil {
									return nil, err
								}
								res[idx].Repo = container.Repository
								res[idx].State = state
							}
							return res, nil
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
					"env": &graphql.Field{
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
								Env: p.Args["value"].(string),
							})
							getRecord(r, repo)
							restartDockerWithNewImage(r, true)
							var records []record
							err := queryRecords(&records)
							return records, err
						},
					},
					"vol": &graphql.Field{
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
								Vols: p.Args["value"].(string),
							})
							getRecord(r, repo)
							restartDockerWithNewImage(r, true)
							var records []record
							err := queryRecords(&records)
							return records, err
						},
					}},
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
	reader := dlog.NewReader(buf)
	s := bufio.NewScanner(reader)
	for s.Scan() {
		c.WriteMessage(1, s.Bytes())
	}
	if err := s.Err(); err != nil {
		log.Printf("read error: %v", err)
	}
}

func webTask(port int) {
	http.Handle("/", http.FileServer(assetFS()))
	//http.Handle("/", http.FileServer(http.Dir("web/dist")))
	http.HandleFunc("/event", handleRegistryEvent)
	http.HandleFunc("/log", handleLogger)
	http.HandleFunc("/graphql", handleGraphQL)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
