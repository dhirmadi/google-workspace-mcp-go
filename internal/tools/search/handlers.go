package search

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	customsearch "google.golang.org/api/customsearch/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- search_custom (core) ---

type SearchCustomInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query        string `json:"q" jsonschema:"required" jsonschema_description:"The search query"`
	Num          int    `json:"num,omitempty" jsonschema_description:"Number of results (1-10 default 10)"`
	Start        int    `json:"start,omitempty" jsonschema_description:"Index of first result (1-based default 1)"`
	Safe         string `json:"safe,omitempty" jsonschema_description:"Safe search level: active moderate or off (default off),enum=active,enum=moderate,enum=off"`
	SearchType   string `json:"search_type,omitempty" jsonschema_description:"Set to image for image search"`
	SiteSearch   string `json:"site_search,omitempty" jsonschema_description:"Restrict to a specific site/domain"`
	DateRestrict string `json:"date_restrict,omitempty" jsonschema_description:"Restrict by date e.g. d5 (5 days) m3 (3 months)"`
	FileType     string `json:"file_type,omitempty" jsonschema_description:"Filter by file type e.g. pdf doc"`
	Language     string `json:"language,omitempty" jsonschema_description:"Language code e.g. lang_en"`
}

type SearchCustomOutput struct {
	Results        []SearchResult `json:"results"`
	TotalResults   string         `json:"total_results"`
	SearchTime     float64        `json:"search_time_seconds"`
	NextStartIndex int            `json:"next_start_index,omitempty"`
}

type SearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

func createSearchCustomHandler(factory *services.Factory, cseID string) mcp.ToolHandlerFor[SearchCustomInput, SearchCustomOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchCustomInput) (*mcp.CallToolResult, SearchCustomOutput, error) {
		if cseID == "" {
			return nil, SearchCustomOutput{}, fmt.Errorf("GOOGLE_CSE_ID environment variable is not set - configure a Custom Search Engine at programmablesearchengine.google.com")
		}

		srv, err := factory.CustomSearch(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchCustomOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.Num == 0 {
			input.Num = 10
		}
		if input.Start == 0 {
			input.Start = 1
		}

		call := srv.Cse.List().Cx(cseID).Q(input.Query).
			Num(int64(input.Num)).
			Start(int64(input.Start)).
			Context(ctx)

		if input.Safe != "" {
			call = call.Safe(input.Safe)
		}
		if input.SearchType == "image" {
			call = call.SearchType("image")
		}
		if input.SiteSearch != "" {
			call = call.SiteSearch(input.SiteSearch)
		}
		if input.DateRestrict != "" {
			call = call.DateRestrict(input.DateRestrict)
		}
		if input.FileType != "" {
			call = call.FileType(input.FileType)
		}
		if input.Language != "" {
			call = call.Lr(input.Language)
		}

		result, err := call.Do()
		if err != nil {
			return nil, SearchCustomOutput{}, middleware.HandleGoogleAPIError(err)
		}

		return buildSearchResponse(result, input.Query)
	}
}

// --- search_custom_siterestrict (extended) ---

type SearchSiterestrictInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Query        string `json:"q" jsonschema:"required" jsonschema_description:"The search query"`
	Num          int    `json:"num,omitempty" jsonschema_description:"Number of results (1-10 default 10)"`
	Start        int    `json:"start,omitempty" jsonschema_description:"Index of first result (1-based default 1)"`
	Safe         string `json:"safe,omitempty" jsonschema_description:"Safe search level: active moderate or off (default off),enum=active,enum=moderate,enum=off"`
	DateRestrict string `json:"date_restrict,omitempty" jsonschema_description:"Restrict by date e.g. d5 (5 days) m3 (3 months)"`
}

func createSearchCustomSiterestrictHandler(factory *services.Factory, cseID string) mcp.ToolHandlerFor[SearchSiterestrictInput, SearchCustomOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SearchSiterestrictInput) (*mcp.CallToolResult, SearchCustomOutput, error) {
		if cseID == "" {
			return nil, SearchCustomOutput{}, fmt.Errorf("GOOGLE_CSE_ID environment variable is not set - configure a Custom Search Engine at programmablesearchengine.google.com")
		}

		srv, err := factory.CustomSearch(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchCustomOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.Num == 0 {
			input.Num = 10
		}
		if input.Start == 0 {
			input.Start = 1
		}

		call := srv.Cse.Siterestrict.List().Cx(cseID).Q(input.Query).
			Num(int64(input.Num)).
			Start(int64(input.Start)).
			Context(ctx)

		if input.Safe != "" {
			call = call.Safe(input.Safe)
		}
		if input.DateRestrict != "" {
			call = call.DateRestrict(input.DateRestrict)
		}

		result, err := call.Do()
		if err != nil {
			return nil, SearchCustomOutput{}, middleware.HandleGoogleAPIError(err)
		}

		return buildSearchResponse(result, input.Query)
	}
}

// --- get_search_engine_info (complete) ---

type GetSearchEngineInfoInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
}

type SearchEngineInfoOutput struct {
	CSEID        string `json:"cse_id"`
	Title        string `json:"title,omitempty"`
	TotalResults string `json:"total_results,omitempty"`
}

func createGetSearchEngineInfoHandler(factory *services.Factory, cseID string) mcp.ToolHandlerFor[GetSearchEngineInfoInput, SearchEngineInfoOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetSearchEngineInfoInput) (*mcp.CallToolResult, SearchEngineInfoOutput, error) {
		if cseID == "" {
			return nil, SearchEngineInfoOutput{}, fmt.Errorf("GOOGLE_CSE_ID environment variable is not set - configure a Custom Search Engine at programmablesearchengine.google.com")
		}

		srv, err := factory.CustomSearch(ctx, input.UserEmail)
		if err != nil {
			return nil, SearchEngineInfoOutput{}, middleware.HandleGoogleAPIError(err)
		}

		// Perform a minimal search to get engine context info
		result, err := srv.Cse.List().Cx(cseID).Q("test").Num(1).Context(ctx).Do()
		if err != nil {
			return nil, SearchEngineInfoOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := SearchEngineInfoOutput{
			CSEID: cseID,
		}

		rb := response.New()
		rb.Header("Search Engine Info")
		rb.KeyValue("CSE ID", cseID)

		if result.SearchInformation != nil {
			output.TotalResults = result.SearchInformation.TotalResults
			rb.KeyValue("Total Results Available", result.SearchInformation.TotalResults)
			rb.KeyValue("Search Time", fmt.Sprintf("%.3fs", result.SearchInformation.SearchTime))
		}

		return rb.TextResult(), output, nil
	}
}

// --- Helper functions ---

func buildSearchResponse(result *customsearch.Search, query string) (*mcp.CallToolResult, SearchCustomOutput, error) {
	results := make([]SearchResult, 0, len(result.Items))

	rb := response.New()
	rb.Header("Search Results")
	rb.KeyValue("Query", query)

	if result.SearchInformation != nil {
		rb.KeyValue("Total Results", result.SearchInformation.TotalResults)
		rb.KeyValue("Search Time", fmt.Sprintf("%.3fs", result.SearchInformation.SearchTime))
	}
	rb.Blank()

	for i, item := range result.Items {
		sr := SearchResult{
			Title:   item.Title,
			Link:    item.Link,
			Snippet: item.Snippet,
		}
		results = append(results, sr)
		rb.Item("%d. %s", i+1, item.Title)
		rb.Line("   %s", item.Link)
		if item.Snippet != "" {
			rb.Line("   %s", item.Snippet)
		}
	}

	var nextStart int
	if result.Queries != nil && len(result.Queries.NextPage) > 0 {
		nextStart = int(result.Queries.NextPage[0].StartIndex)
	}

	output := SearchCustomOutput{
		Results:        results,
		NextStartIndex: nextStart,
	}
	if result.SearchInformation != nil {
		output.TotalResults = result.SearchInformation.TotalResults
		output.SearchTime = result.SearchInformation.SearchTime
	}

	return rb.TextResult(), output, nil
}
