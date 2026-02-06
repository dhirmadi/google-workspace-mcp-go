package appscript

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/services"
)

var serviceIcons = []mcp.Icon{{
	Source:   "https://www.gstatic.com/images/branding/product/1x/apps_script_48dp.png",
	MIMEType: "image/png",
	Sizes:    []string{"48x48"},
}}

// Register registers all Apps Script tools (core + extended) with the MCP server.
func Register(server *mcp.Server, factory *services.Factory) {
	// --- Core tools (7) ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_script_projects",
		Icons:       serviceIcons,
		Description: "List Apps Script projects owned by or shared with the user via Drive search.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Script Projects",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListScriptProjectsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_script_project",
		Icons:       serviceIcons,
		Description: "Get metadata for an Apps Script project including title, create/update times.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Script Project",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetScriptProjectHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_script_content",
		Icons:       serviceIcons,
		Description: "Get the source code files of an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Script Content",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetScriptContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_script_project",
		Icons:       serviceIcons,
		Description: "Create a new Apps Script project, optionally bound to a Google Doc, Sheet, Slide, or Form.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Script Project",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateScriptProjectHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_script_content",
		Icons:       serviceIcons,
		Description: "Update the source code files of an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Script Content",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createUpdateScriptContentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "run_script_function",
		Icons:       serviceIcons,
		Description: "Execute a function in an Apps Script project. The script must be deployed as an API executable and the user must have edit access. Rate limit: ~30 calls/min.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Run Script Function",
			OpenWorldHint: ptrBool(true),
		},
	}, createRunScriptFunctionHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate_trigger_code",
		Icons:       serviceIcons,
		Description: "Generate Apps Script trigger code for common automation patterns (time-based, spreadsheet, form, document events).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Generate Trigger Code",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGenerateTriggerCodeHandler())

	// --- Extended tools (10) ---

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_deployment",
		Icons:       serviceIcons,
		Description: "Create a new deployment for an Apps Script project (web app, API executable, or add-on).",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Deployment",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateDeploymentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_deployments",
		Icons:       serviceIcons,
		Description: "List all deployments for an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Deployments",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListDeploymentsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_deployment",
		Icons:       serviceIcons,
		Description: "Update an existing deployment for an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:          "Update Deployment",
			IdempotentHint: true,
			OpenWorldHint:  ptrBool(true),
		},
	}, createUpdateDeploymentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_deployment",
		Icons:       serviceIcons,
		Description: "Delete a deployment from an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Deployment",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createDeleteDeploymentHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_script_project",
		Icons:       serviceIcons,
		Description: "Delete an Apps Script project by moving it to Drive trash.",
		Annotations: &mcp.ToolAnnotations{
			Title:           "Delete Script Project",
			DestructiveHint: ptrBool(true),
			OpenWorldHint:   ptrBool(true),
		},
	}, createDeleteScriptProjectHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_versions",
		Icons:       serviceIcons,
		Description: "List all versions of an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Script Versions",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListVersionsHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_version",
		Icons:       serviceIcons,
		Description: "Create a new immutable version of an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Create Script Version",
			OpenWorldHint: ptrBool(true),
		},
	}, createCreateVersionHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_version",
		Icons:       serviceIcons,
		Description: "Get details of a specific version of an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Script Version",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetVersionHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_script_processes",
		Icons:       serviceIcons,
		Description: "List running or recently completed processes for an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "List Script Processes",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createListScriptProcessesHandler(factory))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_script_metrics",
		Icons:       serviceIcons,
		Description: "Get execution metrics (active users, executions, failures) for an Apps Script project.",
		Annotations: &mcp.ToolAnnotations{
			Title:         "Get Script Metrics",
			ReadOnlyHint:  true,
			OpenWorldHint: ptrBool(true),
		},
	}, createGetScriptMetricsHandler(factory))
}

func ptrBool(b bool) *bool { return &b }
