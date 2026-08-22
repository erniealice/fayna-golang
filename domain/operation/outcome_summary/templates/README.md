# Outcome Summary HTML Templates

This directory contains the Go HTML templates embedded by the parent
`outcome_summary` package. They render full-page shells, HTMX content fragments,
tables, cards, and drawer forms from Fayna view models.

## Key templates

- `list.html` switches between the flat outcome-summary table and the tabbed
  section landing, including its boosted-navigation content fragment.
- `template-settings.html` renders the document-template binding table, while
  `template-drawer-form.html` renders its upload or replacement drawer.
- `section-download-drawer.html` renders the category, period, and CSV/PDF choices
  for a native browser download.
- `section.html`, `client-card.html`, and the summary templates render progressively
  more detailed outcome views.

Templates use Pyeza components and runtime-provided route and label data. They must
not hardcode education URLs or vocabulary: school-admin supplies the report-card
overlays, while other consumers retain generic outcome-summary terminology.
