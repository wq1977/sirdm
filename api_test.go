package main

import (
	"reflect"
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/testutil"
)

func Test_QueryRecords(t *testing.T) {
	initdb()
	defer closedb()
	expected := map[string]interface{}{}
	query := "{user{repo version time }}"
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if !reflect.DeepEqual(result.Data, expected) {
		t.Fatalf("wrong result, graphql result diff: %v", testutil.Diff(expected, result))
	}
}
