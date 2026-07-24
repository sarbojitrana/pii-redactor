
__init__.py
accept_changes.py
comment.py
merge_runs.py
office
templates
---
"""Merge adjacent identically-formatted runs in a DOCX.

Word fragments paragraph text across many <w:r> elements (revision ids,
spell-check markers, editing history), which makes find-and-replace on
word/document.xml unreliable — the string you're looking for is split
across runs. This coalesces adjacent runs whose formatting (<w:rPr>) is
identical, strips rsid attributes and proofErr markers, and consolidates the
text elements — <w:t>, and <w:delText> for text inside a tracked deletion.

Rendering is unchanged. The text you search is what Word draws, which is not
always the bytes in the file: an element without xml:space="preserve" has its
edge whitespace trimmed before it reaches the page, so `<w:t>Hello </w:t>`
followed by `<w:t>world</w:t>` reads "Helloworld" and merges to exactly that.

Runs in two different <w:ins>/<w:del> wrappers are never merged: that would
rewrite tracked-change structure, collapsing separate revisions into one.

Only word/document.xml is processed (not headers, footers, or footnotes).

Usage:
    python merge_runs.py unpacked/                  # after unzip, before editing
    python merge_runs.py document.docx              # rewrite in place
    python merge_runs.py document.docx -o out.docx
"""


import argparse
import sys
import tempfile
import zipfile
from pathlib import Path

import defusedxml.minidom

from office.helpers import XML_SPACE, rendered_text, rezip, safe_extract

WORDML_NS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"


def merge_runs(input_dir: str) -> tuple[int, str]:
    doc_xml = Path(input_dir) / "word" / "document.xml"

    if not doc_xml.exists():
        return 0, f"Error: {doc_xml} not found"

    try:
        dom = defusedxml.minidom.parseString(doc_xml.read_text(encoding="utf-8"))
        root = dom.documentElement
        run_names = _run_tag_names(root)

