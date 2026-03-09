package search

import (
	"testing"

	"github.com/curtisnewbie/miso/flow"
)

func TestSearch(t *testing.T) {
	rail := flow.EmptyRail()
	resp, err := Search(rail, "my-key", SearchReq{
		Query:         "Who is Leo Messi?",
		SearchDepth:   "basic",
		MaxResults:    5,
		IncludeAnswer: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Query: %v", resp.Query)
	t.Logf("Answer: %v", resp.Answer)
	t.Logf("ResponseTime: %v", resp.ResponseTime)
	for i, r := range resp.Results {
		t.Logf("Result[%d]: %v - %v (score: %v)", i, r.Title, r.URL, r.Score)
	}
}
