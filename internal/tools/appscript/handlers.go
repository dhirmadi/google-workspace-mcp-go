package appscript

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	drivepb "google.golang.org/api/drive/v3"
	scriptpb "google.golang.org/api/script/v1"

	"github.com/evert/google-workspace-mcp-go/internal/middleware"
	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
	"github.com/evert/google-workspace-mcp-go/internal/services"
)

// --- list_script_projects (core) ---

type ListScriptProjectsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Max results (default 20)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for next page of results"`
}

type ListScriptProjectsOutput struct {
	Projects      []ScriptProjectSummary `json:"projects"`
	NextPageToken string                 `json:"next_page_token,omitempty"`
}

type ScriptProjectSummary struct {
	ScriptID   string `json:"script_id"`
	Title      string `json:"title"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	ParentID   string `json:"parent_id,omitempty"`
}

func createListScriptProjectsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListScriptProjectsInput, ListScriptProjectsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListScriptProjectsInput) (*mcp.CallToolResult, ListScriptProjectsOutput, error) {
		// Use Drive API to search for Apps Script files
		driveSrv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, ListScriptProjectsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.PageSize == 0 {
			input.PageSize = 20
		}

		call := driveSrv.Files.List().
			Q("mimeType='application/vnd.google-apps.script'").
			Fields("files(id, name, createdTime, modifiedTime, parents), nextPageToken").
			PageSize(int64(input.PageSize)).
			OrderBy("modifiedTime desc").
			Context(ctx)
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListScriptProjectsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		projects := make([]ScriptProjectSummary, 0, len(result.Files))
		rb := response.New()
		rb.Header("Script Projects")
		rb.KeyValue("Count", len(result.Files))
		if result.NextPageToken != "" {
			rb.KeyValue("Next page token", result.NextPageToken)
		}
		rb.Blank()

		for _, f := range result.Files {
			sp := ScriptProjectSummary{
				ScriptID:   f.Id,
				Title:      f.Name,
				CreateTime: f.CreatedTime,
				UpdateTime: f.ModifiedTime,
			}
			if len(f.Parents) > 0 {
				sp.ParentID = f.Parents[0]
			}
			projects = append(projects, sp)
			rb.Item("%s", f.Name)
			rb.Line("    ID: %s | Modified: %s", f.Id, f.ModifiedTime)
		}

		return rb.TextResult(), ListScriptProjectsOutput{Projects: projects, NextPageToken: result.NextPageToken}, nil
	}
}

// --- get_script_project (core) ---

type GetScriptProjectInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
}

type GetScriptProjectOutput struct {
	ScriptID   string `json:"script_id"`
	Title      string `json:"title"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
	ParentID   string `json:"parent_id,omitempty"`
	Creator    string `json:"creator,omitempty"`
}

func createGetScriptProjectHandler(factory *services.Factory) mcp.ToolHandlerFor[GetScriptProjectInput, GetScriptProjectOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetScriptProjectInput) (*mcp.CallToolResult, GetScriptProjectOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, GetScriptProjectOutput{}, middleware.HandleGoogleAPIError(err)
		}

		project, err := srv.Projects.Get(input.ScriptID).Context(ctx).Do()
		if err != nil {
			return nil, GetScriptProjectOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := GetScriptProjectOutput{
			ScriptID:   project.ScriptId,
			Title:      project.Title,
			CreateTime: project.CreateTime,
			UpdateTime: project.UpdateTime,
			ParentID:   project.ParentId,
		}
		if project.Creator != nil {
			output.Creator = project.Creator.Email
		}

		rb := response.New()
		rb.Header("Script Project")
		rb.KeyValue("Title", project.Title)
		rb.KeyValue("Script ID", project.ScriptId)
		rb.KeyValue("Created", project.CreateTime)
		rb.KeyValue("Updated", project.UpdateTime)
		if project.ParentId != "" {
			rb.KeyValue("Parent ID", project.ParentId)
		}
		if project.Creator != nil {
			rb.KeyValue("Creator", project.Creator.Email)
		}

		return rb.TextResult(), output, nil
	}
}

// --- get_script_content (core) ---

type GetScriptContentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
}

type GetScriptContentOutput struct {
	ScriptID string       `json:"script_id"`
	Files    []ScriptFile `json:"files"`
}

type ScriptFile struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Source     string `json:"source"`
	CreateTime string `json:"create_time,omitempty"`
	UpdateTime string `json:"update_time,omitempty"`
}

func createGetScriptContentHandler(factory *services.Factory) mcp.ToolHandlerFor[GetScriptContentInput, GetScriptContentOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetScriptContentInput) (*mcp.CallToolResult, GetScriptContentOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, GetScriptContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		content, err := srv.Projects.GetContent(input.ScriptID).Context(ctx).Do()
		if err != nil {
			return nil, GetScriptContentOutput{}, middleware.HandleGoogleAPIError(err)
		}

		files := make([]ScriptFile, 0, len(content.Files))
		rb := response.New()
		rb.Header("Script Content")
		rb.KeyValue("Script ID", content.ScriptId)
		rb.KeyValue("Files", len(content.Files))
		rb.Blank()

		for _, f := range content.Files {
			sf := ScriptFile{
				Name:       f.Name,
				Type:       f.Type,
				Source:     f.Source,
				CreateTime: f.CreateTime,
				UpdateTime: f.UpdateTime,
			}
			files = append(files, sf)

			rb.Item("[%s] %s", f.Type, f.Name)
			rb.Line("    Lines: ~%d", countLines(f.Source))
		}

		return rb.TextResult(), GetScriptContentOutput{ScriptID: content.ScriptId, Files: files}, nil
	}
}

// --- create_script_project (core) ---

type CreateScriptProjectInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	Title     string `json:"title" jsonschema:"required" jsonschema_description:"Title for the new script project"`
	ParentID  string `json:"parent_id,omitempty" jsonschema_description:"Drive file ID to bind the script to (Doc Sheet Slide or Form)"`
}

func createCreateScriptProjectHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateScriptProjectInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateScriptProjectInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		project := &scriptpb.CreateProjectRequest{
			Title:    input.Title,
			ParentId: input.ParentID,
		}

		created, err := srv.Projects.Create(project).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Script Project Created")
		rb.KeyValue("Title", created.Title)
		rb.KeyValue("Script ID", created.ScriptId)
		if created.ParentId != "" {
			rb.KeyValue("Parent ID", created.ParentId)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- update_script_content (core) ---

type UpdateScriptContentInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	Files     string `json:"files" jsonschema:"required" jsonschema_description:"JSON array of file objects with name type and source fields. Type is SERVER_JS HTML or JSON."`
}

func createUpdateScriptContentHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateScriptContentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateScriptContentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		var files []*scriptpb.File
		if err := json.Unmarshal([]byte(input.Files), &files); err != nil {
			return nil, nil, fmt.Errorf("invalid files JSON - provide array of {name, type, source} objects: %w", err)
		}

		content := &scriptpb.Content{
			ScriptId: input.ScriptID,
			Files:    files,
		}

		_, err = srv.Projects.UpdateContent(input.ScriptID, content).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Script Content Updated")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.KeyValue("Files Updated", len(files))
		for _, f := range files {
			rb.Item("[%s] %s", f.Type, f.Name)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- run_script_function (core) ---

type RunScriptFunctionInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID   string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	Function   string `json:"function" jsonschema:"required" jsonschema_description:"The function name to execute"`
	Parameters string `json:"parameters,omitempty" jsonschema_description:"JSON array of parameters to pass to the function"`
	DevMode    bool   `json:"dev_mode,omitempty" jsonschema_description:"Run against the most recently saved version (not deployed)"`
}

func createRunScriptFunctionHandler(factory *services.Factory) mcp.ToolHandlerFor[RunScriptFunctionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input RunScriptFunctionInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		execReq := &scriptpb.ExecutionRequest{
			Function: input.Function,
			DevMode:  input.DevMode,
		}

		if input.Parameters != "" {
			var params []interface{}
			if err := json.Unmarshal([]byte(input.Parameters), &params); err != nil {
				return nil, nil, fmt.Errorf("invalid parameters JSON - provide a JSON array of values: %w", err)
			}
			execReq.Parameters = params
		}

		op, err := srv.Scripts.Run(input.ScriptID, execReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		if op.Error != nil {
			rb.Header("Script Execution Failed")
			rb.KeyValue("Error", op.Error.Message)
			for _, detail := range op.Error.Details {
				detailJSON, _ := json.Marshal(detail)
				rb.Line("  Detail: %s", string(detailJSON))
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
				IsError: true,
			}, nil, nil
		}

		rb.Header("Script Execution Complete")
		rb.KeyValue("Function", input.Function)
		if len(op.Response) > 0 {
			var respMap map[string]interface{}
			if err := json.Unmarshal(op.Response, &respMap); err == nil {
				if result, ok := respMap["result"]; ok {
					resultJSON, _ := json.Marshal(result)
					rb.KeyValue("Result", string(resultJSON))
				} else {
					rb.KeyValue("Result", string(op.Response))
				}
			} else {
				rb.KeyValue("Result", string(op.Response))
			}
		} else {
			rb.KeyValue("Result", "void (no return value)")
		}

		return rb.TextResult(), nil, nil
	}
}

// --- create_deployment (extended) ---

type CreateDeploymentInput struct {
	UserEmail   string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID    string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	VersionNum  int    `json:"version_number,omitempty" jsonschema_description:"Version number to deploy (if omitted uses HEAD)"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of this deployment"`
}

func createCreateDeploymentHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateDeploymentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateDeploymentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		deployConfig := &scriptpb.DeploymentConfig{
			Description: input.Description,
		}
		if input.VersionNum > 0 {
			deployConfig.VersionNumber = int64(input.VersionNum)
		}

		created, err := srv.Projects.Deployments.Create(input.ScriptID, deployConfig).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Deployment Created")
		rb.KeyValue("Deployment ID", created.DeploymentId)
		if created.DeploymentConfig != nil {
			rb.KeyValue("Description", created.DeploymentConfig.Description)
			rb.KeyValue("Version", created.DeploymentConfig.VersionNumber)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- list_deployments (extended) ---

type ListDeploymentsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Max results (default 20)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for next page"`
}

type ListDeploymentsOutput struct {
	Deployments   []DeploymentSummary `json:"deployments"`
	NextPageToken string              `json:"next_page_token,omitempty"`
}

type DeploymentSummary struct {
	DeploymentID string `json:"deployment_id"`
	Description  string `json:"description,omitempty"`
	Version      int64  `json:"version"`
}

func createListDeploymentsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListDeploymentsInput, ListDeploymentsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListDeploymentsInput) (*mcp.CallToolResult, ListDeploymentsOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, ListDeploymentsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.PageSize == 0 {
			input.PageSize = 20
		}

		call := srv.Projects.Deployments.List(input.ScriptID).
			PageSize(int64(input.PageSize)).
			Context(ctx)
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListDeploymentsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		deployments := make([]DeploymentSummary, 0, len(result.Deployments))
		rb := response.New()
		rb.Header("Deployments")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.KeyValue("Count", len(result.Deployments))
		rb.Blank()

		for _, d := range result.Deployments {
			ds := DeploymentSummary{DeploymentID: d.DeploymentId}
			if d.DeploymentConfig != nil {
				ds.Description = d.DeploymentConfig.Description
				ds.Version = d.DeploymentConfig.VersionNumber
			}
			deployments = append(deployments, ds)
			rb.Item("ID: %s", d.DeploymentId)
			if ds.Description != "" {
				rb.Line("    Description: %s", ds.Description)
			}
			rb.Line("    Version: %d", ds.Version)
		}

		return rb.TextResult(), ListDeploymentsOutput{Deployments: deployments, NextPageToken: result.NextPageToken}, nil
	}
}

// --- update_deployment (extended) ---

type UpdateDeploymentInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID     string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	DeploymentID string `json:"deployment_id" jsonschema:"required" jsonschema_description:"The deployment ID to update"`
	VersionNum   int    `json:"version_number,omitempty" jsonschema_description:"New version number for the deployment"`
	Description  string `json:"description,omitempty" jsonschema_description:"New description"`
}

func createUpdateDeploymentHandler(factory *services.Factory) mcp.ToolHandlerFor[UpdateDeploymentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input UpdateDeploymentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		updateReq := &scriptpb.UpdateDeploymentRequest{
			DeploymentConfig: &scriptpb.DeploymentConfig{
				Description: input.Description,
			},
		}
		if input.VersionNum > 0 {
			updateReq.DeploymentConfig.VersionNumber = int64(input.VersionNum)
		}

		updated, err := srv.Projects.Deployments.Update(input.ScriptID, input.DeploymentID, updateReq).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Deployment Updated")
		rb.KeyValue("Deployment ID", updated.DeploymentId)
		if updated.DeploymentConfig != nil {
			rb.KeyValue("Version", updated.DeploymentConfig.VersionNumber)
			rb.KeyValue("Description", updated.DeploymentConfig.Description)
		}

		return rb.TextResult(), nil, nil
	}
}

// --- delete_deployment (extended) ---

type DeleteDeploymentInput struct {
	UserEmail    string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID     string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	DeploymentID string `json:"deployment_id" jsonschema:"required" jsonschema_description:"The deployment ID to delete"`
}

func createDeleteDeploymentHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteDeploymentInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteDeploymentInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		_, err = srv.Projects.Deployments.Delete(input.ScriptID, input.DeploymentID).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Deployment Deleted")
		rb.KeyValue("Deployment ID", input.DeploymentID)
		rb.KeyValue("Script ID", input.ScriptID)

		return rb.TextResult(), nil, nil
	}
}

// --- delete_script_project (extended) ---

type DeleteScriptProjectInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID to delete (moves to Drive trash)"`
}

func createDeleteScriptProjectHandler(factory *services.Factory) mcp.ToolHandlerFor[DeleteScriptProjectInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input DeleteScriptProjectInput) (*mcp.CallToolResult, any, error) {
		// Use Drive API to trash the script file
		driveSrv, err := factory.Drive(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		_, err = driveSrv.Files.Update(input.ScriptID, &drivepb.File{
			Trashed: true,
		}).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Script Project Deleted")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.Line("The project has been moved to Drive trash.")

		return rb.TextResult(), nil, nil
	}
}

// --- list_versions (extended) ---

type ListVersionsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Max results (default 20)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for next page"`
}

type ListVersionsOutput struct {
	Versions      []VersionSummary `json:"versions"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

type VersionSummary struct {
	VersionNumber int64  `json:"version_number"`
	Description   string `json:"description,omitempty"`
	CreateTime    string `json:"create_time"`
}

func createListVersionsHandler(factory *services.Factory) mcp.ToolHandlerFor[ListVersionsInput, ListVersionsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListVersionsInput) (*mcp.CallToolResult, ListVersionsOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, ListVersionsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.PageSize == 0 {
			input.PageSize = 20
		}

		call := srv.Projects.Versions.List(input.ScriptID).
			PageSize(int64(input.PageSize)).
			Context(ctx)
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListVersionsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		versions := make([]VersionSummary, 0, len(result.Versions))
		rb := response.New()
		rb.Header("Script Versions")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.KeyValue("Count", len(result.Versions))
		rb.Blank()

		for _, v := range result.Versions {
			vs := VersionSummary{
				VersionNumber: v.VersionNumber,
				Description:   v.Description,
				CreateTime:    v.CreateTime,
			}
			versions = append(versions, vs)
			rb.Item("v%d: %s", v.VersionNumber, v.Description)
			rb.Line("    Created: %s", v.CreateTime)
		}

		return rb.TextResult(), ListVersionsOutput{Versions: versions, NextPageToken: result.NextPageToken}, nil
	}
}

// --- create_version (extended) ---

type CreateVersionInput struct {
	UserEmail   string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID    string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	Description string `json:"description,omitempty" jsonschema_description:"Description of this version"`
}

func createCreateVersionHandler(factory *services.Factory) mcp.ToolHandlerFor[CreateVersionInput, any] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input CreateVersionInput) (*mcp.CallToolResult, any, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		version := &scriptpb.Version{
			Description: input.Description,
		}

		created, err := srv.Projects.Versions.Create(input.ScriptID, version).Context(ctx).Do()
		if err != nil {
			return nil, nil, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Version Created")
		rb.KeyValue("Version", created.VersionNumber)
		rb.KeyValue("Script ID", input.ScriptID)
		if created.Description != "" {
			rb.KeyValue("Description", created.Description)
		}
		rb.KeyValue("Created", created.CreateTime)

		return rb.TextResult(), nil, nil
	}
}

// --- get_version (extended) ---

type GetVersionInput struct {
	UserEmail  string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID   string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	VersionNum int    `json:"version_number" jsonschema:"required" jsonschema_description:"The version number to retrieve"`
}

type GetVersionOutput struct {
	VersionNumber int64  `json:"version_number"`
	Description   string `json:"description,omitempty"`
	CreateTime    string `json:"create_time"`
	ScriptID      string `json:"script_id"`
}

func createGetVersionHandler(factory *services.Factory) mcp.ToolHandlerFor[GetVersionInput, GetVersionOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetVersionInput) (*mcp.CallToolResult, GetVersionOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, GetVersionOutput{}, middleware.HandleGoogleAPIError(err)
		}

		version, err := srv.Projects.Versions.Get(input.ScriptID, int64(input.VersionNum)).Context(ctx).Do()
		if err != nil {
			return nil, GetVersionOutput{}, middleware.HandleGoogleAPIError(err)
		}

		output := GetVersionOutput{
			VersionNumber: version.VersionNumber,
			Description:   version.Description,
			CreateTime:    version.CreateTime,
			ScriptID:      input.ScriptID,
		}

		rb := response.New()
		rb.Header("Script Version")
		rb.KeyValue("Version", version.VersionNumber)
		rb.KeyValue("Description", version.Description)
		rb.KeyValue("Created", version.CreateTime)

		return rb.TextResult(), output, nil
	}
}

// --- list_script_processes (extended) ---

type ListScriptProcessesInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
	PageSize  int    `json:"page_size,omitempty" jsonschema_description:"Max results (default 20)"`
	PageToken string `json:"page_token,omitempty" jsonschema_description:"Token for next page"`
}

type ListScriptProcessesOutput struct {
	Processes     []ProcessSummary `json:"processes"`
	NextPageToken string           `json:"next_page_token,omitempty"`
}

type ProcessSummary struct {
	State        string `json:"state"`
	FunctionName string `json:"function_name"`
	StartTime    string `json:"start_time"`
	Duration     string `json:"duration,omitempty"`
	Type         string `json:"type"`
}

func createListScriptProcessesHandler(factory *services.Factory) mcp.ToolHandlerFor[ListScriptProcessesInput, ListScriptProcessesOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input ListScriptProcessesInput) (*mcp.CallToolResult, ListScriptProcessesOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, ListScriptProcessesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		if input.PageSize == 0 {
			input.PageSize = 20
		}

		call := srv.Processes.List().
			UserProcessFilterScriptId(input.ScriptID).
			PageSize(int64(input.PageSize)).
			Context(ctx)
		if input.PageToken != "" {
			call = call.PageToken(input.PageToken)
		}

		result, err := call.Do()
		if err != nil {
			return nil, ListScriptProcessesOutput{}, middleware.HandleGoogleAPIError(err)
		}

		processes := make([]ProcessSummary, 0, len(result.Processes))
		rb := response.New()
		rb.Header("Script Processes")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.KeyValue("Count", len(result.Processes))
		rb.Blank()

		for _, p := range result.Processes {
			ps := ProcessSummary{
				State:        p.ProcessStatus,
				FunctionName: p.FunctionName,
				StartTime:    p.StartTime,
				Duration:     p.Duration,
				Type:         p.ProcessType,
			}
			processes = append(processes, ps)
			rb.Item("[%s] %s", p.ProcessStatus, p.FunctionName)
			rb.Line("    Type: %s | Started: %s", p.ProcessType, p.StartTime)
			if p.Duration != "" {
				rb.Line("    Duration: %s", p.Duration)
			}
		}

		return rb.TextResult(), ListScriptProcessesOutput{Processes: processes, NextPageToken: result.NextPageToken}, nil
	}
}

// --- get_script_metrics (extended) ---

type GetScriptMetricsInput struct {
	UserEmail string `json:"user_google_email" jsonschema:"required" jsonschema_description:"The user's Google email address"`
	ScriptID  string `json:"script_id" jsonschema:"required" jsonschema_description:"The Apps Script project ID"`
}

type GetScriptMetricsOutput struct {
	ScriptID   string           `json:"script_id"`
	MetricSets []MetricSetEntry `json:"metric_sets"`
}

type MetricSetEntry struct {
	Name   string `json:"name"`
	Values []int  `json:"values,omitempty"`
}

func createGetScriptMetricsHandler(factory *services.Factory) mcp.ToolHandlerFor[GetScriptMetricsInput, GetScriptMetricsOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GetScriptMetricsInput) (*mcp.CallToolResult, GetScriptMetricsOutput, error) {
		srv, err := factory.Script(ctx, input.UserEmail)
		if err != nil {
			return nil, GetScriptMetricsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		metrics, err := srv.Projects.GetMetrics(input.ScriptID).Context(ctx).Do()
		if err != nil {
			return nil, GetScriptMetricsOutput{}, middleware.HandleGoogleAPIError(err)
		}

		rb := response.New()
		rb.Header("Script Metrics")
		rb.KeyValue("Script ID", input.ScriptID)
		rb.Blank()

		var metricSets []MetricSetEntry
		if len(metrics.ActiveUsers) > 0 {
			entry := MetricSetEntry{Name: "active_users"}
			for _, mv := range metrics.ActiveUsers {
				entry.Values = append(entry.Values, int(mv.Value))
			}
			metricSets = append(metricSets, entry)
			rb.Item("Active Users: %d data points", len(metrics.ActiveUsers))
		}
		if len(metrics.TotalExecutions) > 0 {
			entry := MetricSetEntry{Name: "total_executions"}
			for _, mv := range metrics.TotalExecutions {
				entry.Values = append(entry.Values, int(mv.Value))
			}
			metricSets = append(metricSets, entry)
			rb.Item("Total Executions: %d data points", len(metrics.TotalExecutions))
		}
		if len(metrics.FailedExecutions) > 0 {
			entry := MetricSetEntry{Name: "failed_executions"}
			for _, mv := range metrics.FailedExecutions {
				entry.Values = append(entry.Values, int(mv.Value))
			}
			metricSets = append(metricSets, entry)
			rb.Item("Failed Executions: %d data points", len(metrics.FailedExecutions))
		}

		return rb.TextResult(), GetScriptMetricsOutput{ScriptID: input.ScriptID, MetricSets: metricSets}, nil
	}
}
