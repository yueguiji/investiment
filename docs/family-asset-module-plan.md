# Family Asset Module Plan

## Goal

Build a household finance module inside the existing Wails desktop app with:

- household asset registry
- debt amortization and auto-updating liability balances
- historical snapshots for charts
- AI-powered household finance analysis

## Phases

### Phase 1: Domain foundation

- add household domain models
- add CRUD service methods
- add dashboard summary aggregation
- add snapshot persistence

Acceptance:

- data tables migrate successfully
- core CRUD APIs compile
- summary and snapshot APIs return stable data

### Phase 2: Debt engine

- support mortgage and other liabilities
- generate amortization schedules
- auto-update outstanding principal by month
- include debt ratio and payment pressure in summary

Acceptance:

- fixed-rate debt schedule is reproducible
- monthly balances match expected amortization

### Phase 3: Frontend information architecture

- split asset area into overview, registry, debt plans, AI analysis
- add charts backed by snapshots
- support create, edit, delete workflows for each household record type

Acceptance:

- pages render in Wails and production build
- CRUD flows are manually verified

### Phase 4: AI analysis

- reuse shared AI settings
- add prompt-template-driven household analysis
- auto-trigger analysis after data changes
- keep manual re-run button

Acceptance:

- AI config is reused from settings
- household analysis persists and refreshes correctly

### Phase 5: Hardening

- build verification
- regression pass for existing asset pages
- targeted smoke tests for new views

## Notes

- Regional benchmarking should be based on versioned benchmark data, not LLM-only inference.
- Snapshots should be append-friendly and chart-oriented to avoid expensive on-demand recomputation.
