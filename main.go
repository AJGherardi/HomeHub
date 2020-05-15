package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/go-ble/ble"
	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/samsarahq/thunder/batch"
	"github.com/samsarahq/thunder/graphql"
	"github.com/samsarahq/thunder/graphql/introspection"
	"github.com/samsarahq/thunder/reactive"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	groupsCollection  *mongo.Collection
	devicesCollection *mongo.Collection
	webKeysCollection *mongo.Collection
	appKeysCollection *mongo.Collection
	devKeysCollection *mongo.Collection
	netCollection     *mongo.Collection
	write             *ble.Characteristic
	cln               ble.Client
	mdns              *zeroconf.Server
)

func main() {
	// Get ref to collections
	groupsCollection = getCollection("groups")
	devicesCollection = getCollection("devices")
	webKeysCollection = getCollection("webKeys")
	appKeysCollection = getCollection("appKeys")
	devKeysCollection = getCollection("devKeys")
	netCollection = getCollection("net")
	netCollection.DeleteMany(context.TODO(), bson.D{})

	// Check if configured
	if getNetData(netCollection).ID == primitive.NilObjectID {
		// Setup the mdns service
		mdns, _ = zeroconf.Register("hub", "_alexandergherardi._tcp", "local.", 8080, nil, nil)
	} else {
		// Connect and get write characteristic if hub is configured
		cln, write = connectToProxy()
		fmt.Println("con")
	}
	// Build schema
	schema := schema()
	introspection.AddIntrospectionToSchema(schema)
	// Serve graphql
	http.Handle("/graphql", graphql.HTTPHandler(schema))
	http.ListenAndServe(":8080", nil)
	// connectAndServe(schema)
}

type request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type response struct {
	Data   interface{} `json:"data"`
	Errors []string    `json:"errors"`
}

func connectAndServe(s *graphql.Schema, m ...graphql.MiddlewareFunc) {
	// Connect to service
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/hub"}
	log.Printf("connecting to %s", u.String())
	socket, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
	for {
		// Writes json response to web socket
		writeResponse := func(value interface{}, err error) {
			response := response{}
			if err != nil {
				response.Errors = []string{err.Error()}
			} else {
				response.Data = value
			}
			socket.WriteJSON(response)
		}
		// Reed request from web socket
		var req request
		socket.ReadJSON(&req)
		// Parse the query
		query, err := graphql.Parse(req.Query, req.Variables)
		if err != nil {
			writeResponse(nil, err)
			return
		}
		// Get schema from query type
		schema := s.Query
		if query.Kind == "mutation" {
			schema = s.Mutation
		}
		if err := graphql.PrepareQuery(schema, query.SelectionSet); err != nil {
			writeResponse(nil, err)
			return
		}
		// Run middleware and query
		var wg sync.WaitGroup
		e := graphql.Executor{}
		wg.Add(1)
		runner := reactive.NewRerunner(context.TODO(), func(ctx context.Context) (interface{}, error) {
			defer wg.Done()
			ctx = batch.WithBatching(ctx)
			// Add middlewares
			var middlewares []graphql.MiddlewareFunc
			middlewares = append(middlewares, m...)
			// Last function is the query
			middlewares = append(middlewares, func(input *graphql.ComputationInput, next graphql.MiddlewareNextFunc) *graphql.ComputationOutput {
				output := next(input)
				output.Current, output.Error = e.Execute(input.Ctx, schema, nil, input.ParsedQuery)
				return output
			})
			// Run middlewares and get output
			output := graphql.RunMiddlewares(middlewares, &graphql.ComputationInput{
				Ctx:         ctx,
				ParsedQuery: query,
				Query:       req.Query,
				Variables:   req.Variables,
			})
			current, err := output.Current, output.Error
			// Check for error
			if err != nil {
				if graphql.ErrorCause(err) == context.Canceled {
					return nil, err
				}
				writeResponse(nil, err)
				return nil, err
			}
			// Send response if successful
			writeResponse(current, nil)
			return nil, nil
		}, graphql.DefaultMinRerunInterval)
		// Wait until work group is finished then stop the runner
		wg.Wait()
		runner.Stop()
	}
}
