package search

import (
	"github.com/curtisnewbie/miso/errs"
	"github.com/curtisnewbie/miso/miso"
)

var (
	SearchURL = "https://api.tavily.com/search"
)

// https://docs.tavily.com/documentation/api-reference/endpoint/search
type SearchReq struct {
	// The search query to execute with Tavily (required).
	Query string `json:"query"`
	// Controls the latency vs. relevance tradeoff: "advanced", "basic" (default), "fast", "ultra-fast".
	SearchDepth string `json:"search_depth,omitempty"`
	// Max number of relevant chunks (max 500 chars each) returned per source (1-3). Only available when SearchDepth is "advanced".
	ChunksPerSource int `json:"chunks_per_source,omitempty"`
	// The maximum number of search results to return (0-20, default 5).
	MaxResults int `json:"max_results,omitempty"`
	// The category of the search: "general" (default), "news", "finance".
	Topic string `json:"topic,omitempty"`
	// Time range to filter results by publish/update date: "day", "week", "month", "year" (or "d", "w", "m", "y").
	TimeRange string `json:"time_range,omitempty"`
	// Return results published after this date (format: YYYY-MM-DD).
	StartDate string `json:"start_date,omitempty"`
	// Return results published before this date (format: YYYY-MM-DD).
	EndDate string `json:"end_date,omitempty"`
	// Include an LLM-generated answer. true/"basic" for a quick answer, "advanced" for a detailed answer.
	IncludeAnswer any `json:"include_answer,omitempty"`
	// Include cleaned and parsed HTML content. true/"markdown" for markdown format, "text" for plain text.
	IncludeRawContent any `json:"include_raw_content,omitempty"`
	// Also perform an image search and include the results in the response.
	IncludeImages bool `json:"include_images,omitempty"`
	// When IncludeImages is true, also add a descriptive text for each image.
	IncludeImageDescriptions bool `json:"include_image_descriptions,omitempty"`
	// Whether to include the favicon URL for each result.
	IncludeFavicon bool `json:"include_favicon,omitempty"`
	// A list of domains to specifically include in the search results (max 300).
	IncludeDomains []string `json:"include_domains,omitempty"`
	// A list of domains to specifically exclude from the search results (max 150).
	ExcludeDomains []string `json:"exclude_domains,omitempty"`
	// Boost search results from a specific country. Only available when Topic is "general".
	Country string `json:"country,omitempty"`
	// Tavily automatically configures search parameters based on query content and intent.
	AutoParameters bool `json:"auto_parameters,omitempty"`
	// Only return results containing the exact quoted phrase(s) in the query.
	ExactMatch bool `json:"exact_match,omitempty"`
	// Whether to include credit usage information in the response.
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type SearchResp struct {
	Query          string              `json:"query"`
	Answer         string              `json:"answer"`
	Images         []SearchImage       `json:"images"`
	Results        []SearchResult      `json:"results"`
	AutoParameters *AutoParametersInfo `json:"auto_parameters,omitempty"`
	ResponseTime   float64             `json:"response_time"`
	Usage          *UsageInfo          `json:"usage,omitempty"`
	RequestID      string              `json:"request_id"`
}

type SearchResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent string  `json:"raw_content"`
	Favicon    string  `json:"favicon"`
}

type SearchImage struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type AutoParametersInfo struct {
	Topic       string `json:"topic"`
	SearchDepth string `json:"search_depth"`
}

type UsageInfo struct {
	Credits int `json:"credits"`
}

// Search executes a search query using Tavily Search API.
//
// https://docs.tavily.com/documentation/api-reference/endpoint/search
func Search(rail miso.Rail, apiKey string, req SearchReq) (SearchResp, error) {
	var resp SearchResp
	if req.Query == "" {
		return resp, errs.NewErrf("Query required")
	}
	err := miso.NewClient(rail, SearchURL).
		AddAuthBearer(apiKey).
		Require2xx().
		PostJson(req).
		Json(&resp)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
