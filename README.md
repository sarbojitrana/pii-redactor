# PII Redaction Tool — README

## Tech Stack

- **Language:** Go 1.22
- **Standard library only** for core logic: `regexp` (RE2 engine — no
  lookahead/lookbehind/backreferences, a known limitation for a few edge
  cases), `encoding/xml`, `archive/zip`
- **No third-party NLP/NER library** (no spaCy, no Presidio) — see
  "Why hybrid, not pure regex or pure NER" below for why a gazetteer was
  chosen over NER for this document
- **`.docx` handling:** direct manipulation of the OOXML format — a `.docx`
  is unzipped, `word/document.xml` is parsed, redacted, re-serialized, and
  the archive is repacked (`internal/docx`)
- **Project layout:**
  ```
  cmd/redact/     — CLI entrypoint that runs the full redaction pipeline
  cmd/eval/       — CLI entrypoint that scores detector output against
                     ground truth and produces a metrics report
  internal/detect/    — Detector interface + regex, gazetteer, and
                         address detectors
  internal/docx/      — docx unpack / XML parse / repack
  internal/mapper/    — deterministic real-value → fake-value mapping
  internal/redactor/  — orchestrates detect → map → replace over the doc
  internal/eval/      — precision/recall/F1 scoring + report.docx generator
  testdata/           — ground_truth.json, synthetic_pii_snippet.docx,
                         gazetteer source lists
  ```

## How to Run

**Build/run everything with Go directly** (no separate build step needed —
`go run` compiles and runs in one command):

### 1. Redact a document

```bash
go run ./cmd/redact \
    --in input/Red_Herring_Prospectus.docx \
    --out output/redacted.docx
```

- `--in` — path to the source `.docx` to redact (required)
- `--out` — path to write the redacted `.docx` (required)

### 2. Evaluate detector accuracy

```bash
go run ./cmd/eval \
    --in testdata/synthetic_pii_snippet.docx \
    --ground-truth testdata/ground_truth.json \
    --out output/eval_report_synthetic.docx
```

- `--in` — path to the `.docx` to run detectors against (required)
- `--ground-truth` — path to a JSON file of expected `{category, value}`
  entries (default: `testdata/ground_truth.json`)
- `--out` — path to write the generated metrics report `.docx` (default:
  `output/eval_report.docx`)

Prints a precision/recall/F1 table per category to stdout and writes the
same as a formatted `.docx` report.

**Note:** eval must be run against whichever document the ground truth was
actually labeled against — running it against the real prospectus while
using ground truth built from the synthetic snippet (or vice versa) will
produce meaningless all-zero TP counts, since the two texts don't share any
PII values to match against.

### 3. (Optional) Compile a binary instead of `go run`

```bash
go build -o bin/redact ./cmd/redact
go build -o bin/eval ./cmd/eval
./bin/redact --in input/Red_Herring_Prospectus.docx --out output/redacted.docx
```

## Approach

Hybrid detection, implemented in Go:

- **Regex-based detectors** for structurally well-defined PII: `EMAIL`, `PHONE`,
  `SSN`, `CREDIT_CARD` (Luhn-validated), `IP_ADDRESS`, `DOB` (label-anchored,
  e.g. "Date of Birth: ..."), and `ADDRESS` (anchored on the Indian
  "City - PIN State" pattern). Also added `CIN`, `PAN`, and `ISIN` as
  India-specific structured identifiers found in the source document.
- **Gazetteer-based detectors** for `NAME` and `COMPANY`. These have no
  reliable format to regex against, so instead of NER we curated a list of
  real names/companies found in the source prospectus (promoters, officers,
  group companies, banks, law firms) and match against that list, longest
  match first, with overlap resolution so a multi-word name isn't
  double-redacted as sub-strings.
- A **mapper** maintains a `real value → fake value` lookup per category so
  every occurrence of the same real value is replaced with the same fake
  value throughout the document (e.g. `Kushal Subbayya Hegde` always becomes
  the same fake name).
- The `.docx` is unpacked, `word/document.xml` is parsed per-paragraph,
  detectors run on each paragraph's text, matches are replaced, and the
  document is repacked.

## Why hybrid, not pure regex or pure NER

Pure regex cannot distinguish a person's name or a company name from any
other capitalized phrase in a legal document (e.g. "Book Built Offer",
"Companies Act") — there's no format signal to anchor on. Pure NER (spaCy/
Presidio) was considered but judged unnecessary overhead for this document:
the cast of real names/companies is small and closed (a single company's
IPO prospectus), so a curated gazetteer gets equivalent recall on this
document with far less complexity, at the direct cost of **not
generalizing** to a name/company never seen before — the explicit tradeoff
we're making, and the seam where NER would replace the gazetteer if this
became a general-purpose tool.

## Results

Evaluation run against `testdata/synthetic_pii_snippet.docx` with
`testdata/ground_truth.json` (see `Evaluation_Strategy_and_Metrics.docx`
for full methodology and interpretation):

| Category | TP | FP | FN | Precision | Recall | F1 |
|---|---|---|---|---|---|---|
| EMAIL | 2 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| PHONE | 0 | 1 | 2 | 0.00 | 0.00 | 0.00 |
| IP_ADDRESS | 2 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| DOB | 1 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| CREDIT_CARD | 1 | 0 | 0 | 1.00 | 1.00 | 1.00 |
| ADDRESS | 0 | 2 | 2 | 0.00 | 0.00 | 0.00 |

- **EMAIL, IP_ADDRESS, DOB, CREDIT_CARD:** perfect on this test set — all
  format-anchored with no ambiguity.
- **PHONE:** broken. The regex's digit-count window is too permissive —
  it misses real phone numbers and also misfires on unrelated numeric
  strings (in the real prospectus, this causes SEBI registration numbers
  like `000013004` to be flagged as phone numbers).
- **ADDRESS:** broken. Ground-truth addresses span a line break (matching
  how addresses actually appear in the real document); the current regex
  does not match across that break, so it misses the real spans and
  produces false positives instead.
- **NAME, COMPANY:** not scored numerically — see "Known limitations"
  below and the qualitative notes in `Evaluation_Strategy_and_Metrics.docx`.
- **SSN:** implemented, but zero instances in both the real document and
  this evaluation run, so no score was produced.

Run against the real `Red_Herring_Prospectus.docx` (qualitative, from
manual inspection of `output/redacted.docx`):

- Most individual names, most company names, registered/corporate office
  addresses, emails, and most phone numbers were successfully redacted and
  replaced with consistent fake values (e.g. `KSH INTERNATIONAL LIMITED` →
  `Sample Industries 7466 Private Limited`, `Sarthak Malvadkar` →
  `Vivaan Gupta`).
- Known leaks in this run: the original CIN (`U28129PN1979PLC141032`),
  website URLs (`www.kshinternational.com`), one surviving occurrence of
  `Kushal Subbayya Hegde`, and `EVEREST FAMILY TRUST` (while other family
  trust names were correctly redacted).

## Deliverables / Files in This Submission

| File | What it is |
|---|---|
| `README.md` | This file — approach, tech stack, usage, results, limitations |
| `Evaluation_Strategy_and_Metrics.docx` | Evaluation methodology write-up + full metrics table + interpretation (deliverable #4) |
| `output/redacted.docx` | The redacted output of `Red_Herring_Prospectus.docx` |
| `output/eval_report_synthetic.docx` | Auto-generated metrics report from `cmd/eval`, same numbers as the Results table above |
| `testdata/synthetic_pii_snippet.docx` | Hand-built synthetic test document with known PII values, used as the ground-truth-scored eval target |
| `testdata/ground_truth.json` | Expected `{category, value}` PII entries matching the synthetic snippet |
| `testdata/names.txt`, `testdata/companies.txt` | Curated gazetteer source lists of real names/companies found in the source prospectus |
| Source code (`cmd/`, `internal/`) | The redaction tool and eval tool — see "How to Run" above |

## Known limitations / false negatives observed

- **CIN not redacted.** A `CategoryCIN` detector exists but a bug currently
  leaves the original CIN (`U28129PN1979PLC141032`) unchanged in output —
  detector implemented but not wired into the redaction pass correctly.
- **Website URLs are not a covered category** (`www.kshinternational.com`
  and similar survive). Not in the original 9 required categories, and we
  ran out of time to add a `WEBSITE` detector.
- **Some individual name occurrences survive** (e.g. one instance of
  `Kushal Subbayya Hegde` was found unredacted while most occurrences of the
  same name were correctly caught elsewhere). Root cause is almost certainly
  that occurrence being split across multiple Word XML `<w:r>` runs so it
  isn't a contiguous string when the paragraph is read, or it falling inside
  a table cell / header that wasn't included in the paragraph sweep.
- **Family trust names are inconsistently redacted** — some
  (`ANNAPURNA FAMILY TRUST` style) were caught, others (`EVEREST FAMILY
  TRUST`) were not, despite both being in the gazetteer — likely a
  case-sensitivity or all-caps vs. title-case mismatch between the
  gazetteer entry and how the name appears at that specific occurrence.
- **Phone detector over-matches.** SEBI registration numbers
  (`000013004`, `000011179`) are numeric strings in the digit-count range
  the phone regex accepts, and get flagged as phone numbers. The regex
  needs a stricter anchor (e.g. requiring a `+91`/STD-code prefix or
  adjacency to phone-labeling context) to cut this false-positive class.
- **Not covered at all:** SEBI registration numbers, ROC registration
  numbers, GST numbers, UPI IDs, QR code images, and logos (Nuvama, ICICI,
  etc.) — outside the 9 required PII categories, flagged here as scope
  gaps rather than bugs.

## What we'd extend next

Adding a new PII category means implementing the `Detector` interface
(`Category() Category` + `Find(text string) []Match`) and registering it —
nothing else in the pipeline changes. A `WEBSITE` detector and a tightened
`PHONE` regex (to stop matching SEBI registration numbers) are the two
highest-value next additions.
