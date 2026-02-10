package appscript

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/evert/google-workspace-mcp-go/internal/pkg/response"
)

// --- generate_trigger_code (core) ---

type GenerateTriggerCodeInput struct {
	TriggerType  string `json:"trigger_type" jsonschema:"required" jsonschema_description:"Type of trigger: time_based spreadsheet_open spreadsheet_edit form_submit document_open,enum=time_based,enum=spreadsheet_open,enum=spreadsheet_edit,enum=form_submit,enum=document_open"`
	FunctionName string `json:"function_name" jsonschema:"required" jsonschema_description:"Name of the function to trigger"`
	Interval     string `json:"interval,omitempty" jsonschema_description:"For time-based triggers: every_minute every_5_minutes every_10_minutes every_15_minutes every_30_minutes hourly daily weekly,enum=every_minute,enum=every_5_minutes,enum=every_10_minutes,enum=every_15_minutes,enum=every_30_minutes,enum=hourly,enum=daily,enum=weekly"`
}

type GenerateTriggerCodeOutput struct {
	Code        string `json:"code"`
	TriggerType string `json:"trigger_type"`
}

func createGenerateTriggerCodeHandler() mcp.ToolHandlerFor[GenerateTriggerCodeInput, GenerateTriggerCodeOutput] {
	return func(ctx context.Context, req *mcp.CallToolRequest, input GenerateTriggerCodeInput) (*mcp.CallToolResult, GenerateTriggerCodeOutput, error) {
		code, err := generateTriggerCode(input.TriggerType, input.FunctionName, input.Interval)
		if err != nil {
			return nil, GenerateTriggerCodeOutput{}, err
		}

		rb := response.New()
		rb.Header("Generated Trigger Code")
		rb.KeyValue("Trigger Type", input.TriggerType)
		rb.KeyValue("Function", input.FunctionName)
		rb.Blank()
		rb.Raw("```javascript")
		rb.Raw(code)
		rb.Raw("```")
		rb.Blank()
		rb.Line("To install: paste this code into your Apps Script project and run the createTrigger function once.")

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: rb.Build()}},
		}, GenerateTriggerCodeOutput{Code: code, TriggerType: input.TriggerType}, nil
	}
}

func generateTriggerCode(triggerType, functionName, interval string) (string, error) {
	var sb strings.Builder

	switch triggerType {
	case "time_based":
		if interval == "" {
			interval = "hourly"
		}
		sb.WriteString(fmt.Sprintf(`/**
 * Creates a time-based trigger for %s.
 * Run this function once to install the trigger.
 */
function createTrigger() {
`, functionName))
		switch interval {
		case "every_minute":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyMinutes(1)
    .create();
`, functionName))
		case "every_5_minutes":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyMinutes(5)
    .create();
`, functionName))
		case "every_10_minutes":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyMinutes(10)
    .create();
`, functionName))
		case "every_15_minutes":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyMinutes(15)
    .create();
`, functionName))
		case "every_30_minutes":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyMinutes(30)
    .create();
`, functionName))
		case "hourly":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyHours(1)
    .create();
`, functionName))
		case "daily":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .everyDays(1)
    .atHour(9)
    .create();
`, functionName))
		case "weekly":
			sb.WriteString(fmt.Sprintf(`  ScriptApp.newTrigger('%s')
    .timeBased()
    .onWeekDay(ScriptApp.WeekDay.MONDAY)
    .atHour(9)
    .create();
`, functionName))
		default:
			return "", fmt.Errorf("unknown interval %q - use: every_minute, every_5_minutes, every_10_minutes, every_15_minutes, every_30_minutes, hourly, daily, weekly", interval)
		}
		sb.WriteString("}\n")

	case "spreadsheet_open":
		sb.WriteString(fmt.Sprintf(`/**
 * Creates an onOpen trigger for %s.
 * Run this function once to install the trigger.
 */
function createTrigger() {
  var ss = SpreadsheetApp.getActive();
  ScriptApp.newTrigger('%s')
    .forSpreadsheet(ss)
    .onOpen()
    .create();
}
`, functionName, functionName))

	case "spreadsheet_edit":
		sb.WriteString(fmt.Sprintf(`/**
 * Creates an onEdit trigger for %s.
 * Run this function once to install the trigger.
 */
function createTrigger() {
  var ss = SpreadsheetApp.getActive();
  ScriptApp.newTrigger('%s')
    .forSpreadsheet(ss)
    .onEdit()
    .create();
}
`, functionName, functionName))

	case "form_submit":
		sb.WriteString(fmt.Sprintf(`/**
 * Creates a form submit trigger for %s.
 * Run this function once to install the trigger.
 */
function createTrigger() {
  var form = FormApp.getActiveForm();
  ScriptApp.newTrigger('%s')
    .forForm(form)
    .onFormSubmit()
    .create();
}
`, functionName, functionName))

	case "document_open":
		sb.WriteString(fmt.Sprintf(`/**
 * Creates an onOpen trigger for %s.
 * Run this function once to install the trigger.
 */
function createTrigger() {
  var doc = DocumentApp.getActiveDocument();
  ScriptApp.newTrigger('%s')
    .forDocument(doc)
    .onOpen()
    .create();
}
`, functionName, functionName))

	default:
		return "", fmt.Errorf("unknown trigger type %q - use: time_based, spreadsheet_open, spreadsheet_edit, form_submit, document_open", triggerType)
	}

	return sb.String(), nil
}

func countLines(s string) int {
	if s == "" {
		return 0
	}
	n := 1
	for _, c := range s {
		if c == '\n' {
			n++
		}
	}
	return n
}
