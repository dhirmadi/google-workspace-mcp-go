package search

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/ptr"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/googleg_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Custom Search tools (core + extended + complete) with the MCP server.
// The cseID parameter is the Google Custom Search Engine ID from the GOOGLE_CSE_ID env var.
func Register(server *mcp.Server, factory *services.Factory, cseID string) {
	// --- Core tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_custom",
		Icons:       serviceIcons,
		Description: "Perform a web search using Google Custom Search JSON API. Requires GOOGLE_CSE_ID to be configured.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Custom Web Search",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchCustomHandler(factory, cseID))

	// --- Extended tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_custom_siterestrict",
		Icons:       serviceIcons,
		Description: "Perform a site-restricted search using Google Custom Search. Limits results to the sites configured in the search engine.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Site-Restricted Search",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createSearchCustomSiterestrictHandler(factory, cseID))

	// --- Complete tools ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_search_engine_info",
		Icons:       serviceIcons,
		Description: "Get configuration details of the Custom Search Engine including search settings and site restrictions.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Search Engine Info",
			ReadOnlyHint:  true,
			OpenWorldHint: ptr.Bool(true),
		},
	}, createGetSearchEngineInfoHandler(factory, cseID))
}
