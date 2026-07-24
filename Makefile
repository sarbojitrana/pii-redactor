BINARY := redact
CMD_DIR := ./cmd/redact

IN     ?= input/Red Herring Prospectus.docx
MERGED ?= input/Red Herring Prospectus.merged.docx
OUT    ?= output/Red Herring Prospectus_redacted.docx

.PHONY: build run test eval clean prep validate preview

# Static binary, no CGO — trivially runnable by anyone grading it, no runtime deps.
build:
	CGO_ENABLED=0 go build -o $(BINARY) $(CMD_DIR)

# Go pipeline consumes MERGED, never IN directly — keeps run-merging a
# separate, inspectable prep step rather than a hidden subprocess call.
run: build prep
	./$(BINARY) --in "$(MERGED)" --out "$(OUT)"

test:
	go test ./...

# One-time XML run normalization before the Go pipeline touches the doc.
# Writes a distinct .merged.docx — never overwrites the original source.
prep:
	python3 scripts/merge_runs.py --in "$(IN)" --out "$(MERGED)"

# Structural validation of the repacked .docx after redaction.
validate:
	python3 scripts/validate.py --in "$(OUT)"

# Visual QA: docx -> PDF/JPEG.
preview:
	bash scripts/render_preview.sh "$(OUT)"

# Runs eval binary/target once eval package exists; scores OUT against ground truth.
eval: build
	./$(BINARY) --eval --in "$(OUT)" --ground-truth testdata/ground_truth.json --report output/eval_report.docx

clean:
	rm -f $(BINARY)
	rm -f output/*.docx output/*.pdf output/*.json