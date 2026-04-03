# Project Stuff

A physical asset management system with spreadsheet editing and label printing.

## Overview

Stuff manages fixed and non-fixed physical assets for the organization. It
provides a spreadsheet-style UI for viewing and editing asset data, and can
generate printable labels for each asset row.

## How It Works

1. User opens the spreadsheet UI to view all tracked assets
2. Assets can be created, edited, and deleted inline
3. Each asset row can generate a printable label
4. Labels contain asset identifiers and metadata for physical tagging

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌──────────────────────┐
│   Client    │────▶│  Stuff Server   │────▶│     SQL Database     │
│  (Browser)  │◀────│                 │◀────│  (Postgres / SQLite) │
└─────────────┘     └─────────────────┘     └──────────────────────┘
        │
        ▼
┌─────────────────┐
│  Label Printer  │
└─────────────────┘
```
