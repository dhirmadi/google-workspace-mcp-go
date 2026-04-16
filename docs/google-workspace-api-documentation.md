# Official Google Workspace API documentation (online)

This page lists **primary URLs** for the Google APIs that this MCP server uses heavily, plus broader hubs useful for **onboarding, support, or indexing** external documentation.

> **Naming note:** If “Page” meant **Google Slides** (presentations and slide pages), use the Slides links below. This repository implements **Slides**, not the **Google Sites** API. If you need Sites, see [Google Sites API](https://developers.google.com/sites).

---

## Top-level hubs

| Resource | URL |
|----------|-----|
| Google Workspace for Developers (umbrella) | [https://developers.google.com/workspace](https://developers.google.com/workspace) |
| Workspace developer resources | [https://developers.google.com/workspace/resources](https://developers.google.com/workspace/resources) |
| Google Cloud — Workspace | [https://cloud.google.com/workspace](https://cloud.google.com/workspace) |
| Google Cloud Console (API enablement, OAuth clients) | [https://console.cloud.google.com/](https://console.cloud.google.com/) |

---

## APIs used by this server (Calendar, Mail, Drive, Sheets, Slides, Docs)

Use these roots as **starting points** for guides, REST reference, quotas, and release notes.

| Product | Developer documentation |
|---------|-------------------------|
| **Google Calendar** | [https://developers.google.com/calendar/api](https://developers.google.com/calendar/api) |
| **Gmail** (Mail) | [https://developers.google.com/gmail/api](https://developers.google.com/gmail/api) |
| **Google Drive** | [https://developers.google.com/drive/api](https://developers.google.com/drive/api) |
| **Google Sheets** | [https://developers.google.com/sheets/api](https://developers.google.com/sheets/api) |
| **Google Slides** | [https://developers.google.com/slides](https://developers.google.com/slides) |
| **Google Docs** | [https://developers.google.com/docs/api](https://developers.google.com/docs/api) |

### Related: Drive comments on Docs / Sheets / Slides

Several tools in this repo read or write **file comments** via the **Drive API** (not a separate “Comments API”). Keep the Drive documentation above in scope when working on comment behavior.

---

## Machine-readable API metadata (for indexing or code generation)

| Resource | URL |
|----------|-----|
| **API Discovery** (overview) | [https://developers.google.com/discovery](https://developers.google.com/discovery) |

Discovery documents describe REST resources, methods, and JSON schemas. They complement narrative docs; they do **not** replace OAuth, quota, or product-policy documentation.

---

## Indexing or crawling (practical guidance)

- Google does **not** ship a single downloadable bundle of all Workspace prose docs. Typical approaches are **scoped crawls** from the hub and each API root you care about, or indexing **Discovery** artifacts where a structured surface is enough.
- Respect **`robots.txt`** on each host (e.g. `developers.google.com`, `cloud.google.com`), crawl rate limits, and terms of use. Start from each site’s robots file to discover **sitemap** hints when available.
- **Scope:** Limit crawls to paths under `/workspace`, `/gmail`, `/drive`, `/calendar`, `/sheets`, `/slides`, `/docs`, etc., rather than the entire `developers.google.com` tree, unless you intend to index unrelated products.

---

## Repository pointers

- **Scopes and OAuth** used by this server: [`docs/auth-and-scopes.md`](auth-and-scopes.md)
- **Tools per service:** [`docs/tools-inventory.md`](tools-inventory.md)
