package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/pkg/graphql"
)

var httpClient *http.Client
var expectedNewTodo = `{"createTodo":{"user":{"id":"1"},"text":"todo","done":false}}`
var expectedTodos = `{"todos":[{"text":"todo","done":false,"user":{"name":"user 1"}}]}`
var URL = "localhost:" + defaultPort

func TestMainFunc(t *testing.T) {
	router := NewRouter()
	ts := httptest.NewUnstartedServer(router)
	l, _ := net.Listen("tcp", URL)
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	t.Run("should create new todo", func(t *testing.T) {
		gqlReqBody := graphql.Request{
			Query: "mutation { createTodo(input: { text: \"todo\", userId: \"1\" }) {user {id } text done } }",
		}

		newTodoResponse := executeGraphQLRequest(t, gqlReqBody)

		assert.JSONEq(t, expectedNewTodo, string(newTodoResponse.Data))

		gqlReqBody = graphql.Request{
			Query: "{todos { text done user { name } } }",
		}

		todosResponse := executeGraphQLRequest(t, gqlReqBody)
		// fmt.Println(string(todosResponse.Data))
		assert.JSONEq(t, expectedTodos, string(todosResponse.Data))
	})

}

func executeGraphQLRequest(t *testing.T, reqBody graphql.Request) GraphQLResponse {
	t.Helper()

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "http://localhost:8085/query", bytes.NewBuffer(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if httpClient == nil {
		httpClient = &http.Client{}
	}

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equalf(t, http.StatusOK, resp.StatusCode, "response status code is not 200")

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	gqlResponse := GraphQLResponse{}
	err = json.Unmarshal(respBody, &gqlResponse)
	require.NoError(t, err)

	return gqlResponse
}

type GraphQLResponse struct {
	Data json.RawMessage `json:"data"`
}



