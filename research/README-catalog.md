# Machine-readable iikoServer API catalog

`endpoints.yaml` · `dtos.yaml` · `verbs-from-css.tsv`

Generated 2026-09-22/23 from the 150 `ru.iiko.help/article/api-documentations/*` pages already
mirrored in `iiko-docs/`, re-fetched as raw HTML so the machine-readable parts of the markup
survived. Companion to the prose report in `iikoserver-api.md`, which stays the place for
narrative gotchas.

## Coverage

| | |
|---|---|
| endpoint records | 207 (201 iikoServer `/resto/api`, 6 sibling products, see `api:`) |
| HTTP verb confirmed from markup | 188 of 201 |
| verb inferred | 13 (7 of them undocumented third-party paths) |
| DTOs | 186 |
| fields | 1354 |
| enums | 71 |
| OLAP columns recovered from docs (SALES 114 / TRANSACTIONS 98 / DELIVERIES 35 / STOCK 17) | 264 |
| fields marked `read_only` | 278 |
| fields marked `declared_but_unimplemented` | 19 |
| fields carrying an `aliases` entry (docs spell it 2 ways) | 18 |
| endpoints that mutate production data (`writes: true`) | 64 |

## How verbs were established

The docs never render the HTTP verb as text — it is an image plus a CSS class
`CHHttpRequest_get|post|put|delete`. Extracting all 200 request blocks across the 150 pages
(0 unparsed) gives the verb table in `verbs-from-css.tsv`. That file is evidence: the path is
printed exactly as the page prints it.

`method_confirmed: true` means the verb came from that CSS class. `false` means it was inferred
from an example URL, prose, or convention — those are listed at the bottom of this file.

Three verbs are hand-corrected because the CSS class alone was wrong or unusable. Each carries its
evidence in the record's own `notes`, and the merge step refuses to accept any other unbacked
`method_confirmed: true`:

- `PUT /resto/api/edi/{ediSystem}/orders/ack` — the PUT block printed the `bySeller` URL from the
  block above it; its heading, params table and worked example all say `ack`.
- `PUT /resto/api/employees/byId/{UUID}` — tagged GET, but the heading is "Добавить или заменить
  сотрудника", the page names PUT for full replacement twice in prose, a separate correctly-tagged
  GET block already serves the path, and the parallel `byCode` section is correctly tagged PUT.
- `POST /resto/api/v2/entities/quickLabels/delete` — the POST block's bold path segment is empty in
  the HTML, so it extracts as a bare `/resto/api/v2/`; completed from the page's own example.

## Reading the provenance fields

- `confidence: documented` — stated in the docs. `inferred` — derived from an example or
  convention. `third_party` — seen only in a client in the wild, never in any iiko document.
- `documented: false` — verified absent from all 150 pages, not merely unindexed.
- `api:` — `iikoserver` is this project's target. `iikobiz` / `iikocloud` records are sibling
  products documented on the same pages; they are kept so nobody re-discovers them, and must be
  filtered out before generating an iikoServer client.

## Contradictions are recorded, never resolved

Where the docs disagree with themselves the catalog keeps both readings rather than picking one:

- A field spelled two ways keeps one entry with the other spelling in `aliases`.
- An enum that diverges from the base-types reference page (`kody-bazovykh-tipov`) keeps its own
  values plus a `note` stating the exact delta. These are not all errors: `ProductType` on save
  legitimately excludes `PETROL`, and the store-report XSD legitimately documents
  `FUEL_ACCEPTANCE` / `FUEL_GAGING_DOCUMENT`, which the base-types page omits.
- Where a DTO name is used for structurally unrelated types in different domains, both survive
  under suffixed names with a `NAME COLLISION` note (`ContainerDto__edi` vs
  `ContainerDto__nomenclature`).
- Where the same type was extracted at different depth in two domains, the richer shape wins and
  the note records which endpoints the docs only ever show the reduced shape on.
- A field typed differently by two pages keeps the more specific type and records the disagreement.

### Path contradictions

On two independent pages the HTTP-request header block omits a `/v2/` segment that the page's own
worked example includes. Both spellings are recorded as separate endpoints:

| header block | example URL on the same page |
|---|---|
| `/resto/api/corporation/settings` | `/resto/api/v2/corporation/settings` |
| `/resto/api/entities/accounts/list` | `/resto/api/v2/entities/accounts/list` |

Also `/resto/api/corporation/terminals/search` (header block) vs
`/resto/api/corporation/terminal/search` (prose *and* example, singular).

## Endpoints whose verb is still inferred

Documented, but no CSS-tagged block exists — all appear only as example URLs or in prose, and all
are reads:

`GET /resto/api/corporation/stores/search` ·
`GET /resto/api/employees/waiterTeams/byCode/{teamCode}` ·
`GET /resto/api/licence/info` · `GET /resto/api/v2/corporation/settings` ·
`GET /resto/api/v2/entities/accounts/list` · `GET /resto/api/v2/entities/periodSchedules/byId`

Plus the 7 `documented: false` third-party paths, whose verbs come from the calling client's
source and are unverifiable from any iiko document.

## Integrity

The merge step refuses to emit the catalog if any endpoint or field references a DTO that does not
exist, if a `method_confirmed: true` is not backed by the CSS map or the hand-verified list, or if
an unknown key appears in any record. Current state: 0 problems.

Two records are synthesized by the merge rather than by a domain extraction, and say so in their
own notes: `OlapFilter` (the docs describe a `filterType`-discriminated union in prose but never
name it as one structure) and the `__`-suffixed halves of a collided DTO name.

## Not verified against a live server

Nothing here has been run against a real iiko stand. `/v2/reports/olap/columns` is the
authoritative OLAP field list and needs a live server; the OLAP field DTOs here are only what the
documentation and its examples state.


### `map_of` (added 2026-09-23)

A field with `type: object`, `of: null`, `map_of: <DTO>` is a map keyed by string to that DTO — used where iiko keys entries by column name (`OlapPreset.filters`, `OlapV2RequestBody.filters`). `of: <DTO>` alone still means a single nested object.

### `xml` and `wrapper` (added 2026-09-23)

`research/dtos.yaml` was extracted from field tables, which cannot say whether an XML field is an element, an attribute, or character data. That is fine for every v1 XML DTO except EDI: the EDI examples (`api-edi-5-0`, `api-edi-5-1-6-4`) are attribute-heavy, and elements-only tags decode a real `<order number="…">` into an empty struct with no error.

- `xml: attr` — the field is an XML attribute (`xml:"name,attr"`). `xml: chardata` — the field is the element's text. Absent means element.
- `wrapper: <child>` — a list field is wrapped: `lineItems` with `wrapper: lineItem` means `<lineItems><lineItem>…</lineItem></lineItems>` (`xml:"lineItems>lineItem"`).

Source of the EDI mapping: `pkg/iikoserver/edi/wire.go`, whose round-trip tests use fixtures copied verbatim from the doc examples.
