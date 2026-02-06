# Tool Inventory

**Total: 136 tools** across 12 Google Workspace services.

Comment tools (read/create/reply/resolve) for Docs, Sheets, and Slides are implemented via a shared `comments` package using the Drive API. They are counted under each parent service (4 tools x 3 services = 12 comment tools included in the total).

> The original spec listed 133 tools. The correct count is **136** — the discrepancy was from undercounting the 12 shared comment tools across Docs (4), Sheets (4), and Slides (4). This is now reconciled.

## Summary

| Service | Core | Extended | Complete | Total |
|---------|------|----------|----------|-------|
| Gmail | 4 | 9 | 2 | 15 |
| Drive | 7 | 7 | 2 | 16 |
| Calendar | 5 | 1 | 0 | 6 |
| Docs | 3 | 6 | 10 | 19 |
| Sheets | 3 | 6 | 5 | 14 |
| Chat | 4 | 0 | 0 | 4 |
| Forms | 2 | 1 | 3 | 6 |
| Slides | 2 | 3 | 4 | 9 |
| Tasks | 5 | 1 | 6 | 12 |
| Contacts | 4 | 4 | 7 | 15 |
| Search | 1 | 1 | 1 | 3 |
| Apps Script | 7 | 10 | 0 | 17 |
| **TOTAL** | **47** | **49** | **40** | **136** |

---

## Gmail (15 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `search_gmail_messages` | core | yes | Search emails with Gmail query syntax |
| `get_gmail_message_content` | core | yes | Get full content of a single message |
| `get_gmail_messages_content_batch` | core | yes | Get content of multiple messages (max 25) |
| `send_gmail_message` | core | no | Send email with optional reply threading |
| `get_gmail_attachment_content` | extended | yes | Get attachment data |
| `get_gmail_thread_content` | extended | yes | Get all messages in a thread |
| `modify_gmail_message_labels` | extended | no | Add/remove labels from message |
| `list_gmail_labels` | extended | yes | List all labels |
| `manage_gmail_label` | extended | no | Create/update/delete labels |
| `draft_gmail_message` | extended | no | Create/update/send drafts |
| `list_gmail_filters` | extended | yes | List email filters |
| `create_gmail_filter` | extended | no | Create email filter |
| `delete_gmail_filter` | extended | no | Delete email filter |
| `get_gmail_threads_content_batch` | complete | yes | Batch get thread contents |
| `batch_modify_gmail_message_labels` | complete | no | Batch label modifications |

## Drive (16 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `search_drive_files` | core | yes | Search files with Drive query |
| `get_drive_file_content` | core | yes | Get file content (text extraction) |
| `get_drive_file_download_url` | core | yes | Get download URL for file |
| `create_drive_file` | core | no | Create new file |
| `import_to_google_doc` | core | no | Import file to Google Doc format |
| `share_drive_file` | core | no | Share file with users/groups |
| `get_drive_shareable_link` | core | yes | Get shareable link |
| `list_drive_items` | extended | yes | List files in folder |
| `copy_drive_file` | extended | no | Copy a file |
| `update_drive_file` | extended | no | Update file content/metadata |
| `update_drive_permission` | extended | no | Modify existing permission |
| `remove_drive_permission` | extended | no | Remove sharing permission |
| `transfer_drive_ownership` | extended | no | Transfer file ownership |
| `batch_share_drive_file` | extended | no | Share multiple files at once |
| `get_drive_file_permissions` | complete | yes | List all permissions on file |
| `check_drive_file_public_access` | complete | yes | Check if file is public |

## Calendar (6 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `list_calendars` | core | yes | List user's calendars |
| `get_events` | core | yes | Get events in time range |
| `create_event` | core | no | Create calendar event |
| `modify_event` | core | no | Update existing event |
| `delete_event` | **core** | no | Delete calendar event |
| `query_freebusy` | extended | yes | Query free/busy times |

> `delete_event` promoted from extended to **core** — create+modify without delete is an awkward UX gap.

## Docs (19 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `get_doc_content` | core | yes | Get document content |
| `create_doc` | core | no | Create new document |
| `modify_doc_text` | core | no | Insert/replace text with formatting |
| `export_doc_to_pdf` | extended | yes | Export document as PDF |
| `search_docs` | extended | yes | Search documents |
| `find_and_replace_doc` | extended | no | Find and replace text |
| `list_docs_in_folder` | extended | yes | List docs in Drive folder |
| `insert_doc_elements` | extended | no | Insert paragraphs, lists, etc. |
| `update_paragraph_style` | extended | no | Update text styling |
| `insert_doc_image` | complete | no | Insert image into document |
| `update_doc_headers_footers` | complete | no | Modify headers/footers |
| `batch_update_doc` | complete | no | Batch document updates |
| `inspect_doc_structure` | complete | yes | Debug document structure |
| `create_table_with_data` | complete | no | Create table with data |
| `debug_table_structure` | complete | yes | Debug table structure |
| `read_document_comments` | complete | yes | Read comments (via Drive API, shared) |
| `create_document_comment` | complete | no | Add comment (via Drive API, shared) |
| `reply_to_document_comment` | complete | no | Reply to comment (via Drive API, shared) |
| `resolve_document_comment` | complete | no | Resolve comment (via Drive API, shared) |

## Sheets (14 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `create_spreadsheet` | core | no | Create new spreadsheet |
| `read_sheet_values` | core | yes | Read cell values |
| `modify_sheet_values` | core | no | Write/update cell values |
| `list_spreadsheets` | extended | yes | List spreadsheets |
| `get_spreadsheet_info` | extended | yes | Get spreadsheet metadata |
| `format_sheet_range` | extended | no | Format cells (bold, colors, borders, alignment, number format) |
| `add_conditional_formatting` | extended | no | Add conditional formatting rules |
| `update_conditional_formatting` | extended | no | Update conditional formatting rules |
| `delete_conditional_formatting` | extended | no | Delete conditional formatting rules |
| `create_sheet` | complete | no | Create new sheet tab |
| `read_spreadsheet_comments` | complete | yes | Read comments (via Drive API, shared) |
| `create_spreadsheet_comment` | complete | no | Add comment (via Drive API, shared) |
| `reply_to_spreadsheet_comment` | complete | no | Reply to comment (via Drive API, shared) |
| `resolve_spreadsheet_comment` | complete | no | Resolve comment (via Drive API, shared) |

## Chat (4 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `send_chat_message` | core | no | Send chat message |
| `get_chat_messages` | core | yes | Get messages from space |
| `search_chat_messages` | core | yes | Search chat messages |
| `list_chat_spaces` | **core** | yes | List chat spaces |

> Chat tools renamed with `chat_` prefix to avoid collision with Gmail tool names.
> `list_chat_spaces` promoted from extended to **core** — can't send messages without knowing the space ID.

## Forms (6 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `create_form` | core | no | Create new form |
| `get_form` | core | yes | Get form details |
| `list_form_responses` | extended | yes | List form responses |
| `set_publish_settings` | complete | no | Set form publish settings |
| `get_form_response` | complete | yes | Get single response |
| `batch_update_form` | complete | no | Batch form updates |

## Slides (9 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `create_presentation` | core | no | Create new presentation |
| `get_presentation` | core | yes | Get presentation details |
| `batch_update_presentation` | extended | no | Batch presentation updates |
| `get_page` | extended | yes | Get single slide/page |
| `get_page_thumbnail` | extended | yes | Get slide thumbnail |
| `read_presentation_comments` | complete | yes | Read comments (via Drive API, shared) |
| `create_presentation_comment` | complete | no | Add comment (via Drive API, shared) |
| `reply_to_presentation_comment` | complete | no | Reply to comment (via Drive API, shared) |
| `resolve_presentation_comment` | complete | no | Resolve comment (via Drive API, shared) |

## Tasks (12 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `get_task` | core | yes | Get task details |
| `list_tasks` | core | yes | List tasks in list |
| `create_task` | core | no | Create new task |
| `update_task` | core | no | Update task |
| `list_task_lists` | **core** | yes | List task lists |
| `delete_task` | extended | no | Delete task |
| `get_task_list` | complete | yes | Get task list details |
| `create_task_list` | complete | no | Create task list |
| `update_task_list` | complete | no | Update task list |
| `delete_task_list` | complete | no | Delete task list |
| `move_task` | complete | no | Move task position |
| `clear_completed_tasks` | complete | no | Clear completed tasks |

> `list_task_lists` promoted from complete to **core** — without it, you can't use ANY task tools (they all require `task_list_id`).

## Contacts (15 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `search_contacts` | core | yes | Search contacts (People API) |
| `get_contact` | core | yes | Get contact details |
| `list_contacts` | core | yes | List all contacts |
| `create_contact` | core | no | Create new contact |
| `update_contact` | extended | no | Update contact |
| `delete_contact` | extended | no | Delete contact |
| `list_contact_groups` | extended | yes | List contact groups |
| `get_contact_group` | extended | yes | Get group details |
| `batch_create_contacts` | complete | no | Batch create contacts |
| `batch_update_contacts` | complete | no | Batch update contacts |
| `batch_delete_contacts` | complete | no | Batch delete contacts |
| `create_contact_group` | complete | no | Create contact group |
| `update_contact_group` | complete | no | Update contact group |
| `delete_contact_group` | complete | no | Delete contact group |
| `modify_contact_group_members` | complete | no | Add/remove group members |

## Search (3 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `search_custom` | core | yes | Custom search using Google CSE (requires `GOOGLE_CSE_ID`) |
| `search_custom_siterestrict` | extended | yes | Site-restricted search |
| `get_search_engine_info` | complete | yes | Get search engine config |

## Apps Script (17 tools)

| Tool | Tier | Read-Only | Description |
|------|------|-----------|-------------|
| `list_script_projects` | core | yes | List script projects |
| `get_script_project` | core | yes | Get project details |
| `get_script_content` | core | yes | Get script source code |
| `create_script_project` | core | no | Create new project |
| `update_script_content` | core | no | Update script source |
| `run_script_function` | core | no | Execute script function (API executable only, requires edit access) |
| `generate_trigger_code` | core | yes | Generate trigger code |
| `create_deployment` | extended | no | Create deployment |
| `list_deployments` | extended | yes | List deployments |
| `update_deployment` | extended | no | Update deployment |
| `delete_deployment` | extended | no | Delete deployment |
| `delete_script_project` | extended | no | Delete project |
| `list_versions` | extended | yes | List script versions |
| `create_version` | extended | no | Create new version |
| `get_version` | extended | yes | Get version details |
| `list_script_processes` | extended | yes | List running processes |
| `get_script_metrics` | extended | yes | Get execution metrics |
