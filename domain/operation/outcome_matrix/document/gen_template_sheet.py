#!/usr/bin/env python3
"""Generator for the grade-sheet (outcome-matrix) PDF template artifacts.

The grade-sheet PDF is the MMIS-parity COMPOSITE gradesheet (decisions.html Q1,
LOCKED): a roster grid — one row per student, columns = per-period finals + the
year Final — ONE artifact per job_category. It is the printed record teachers
sign, NOT the raw criteria matrix (the CSV already exports that).

Two axes are cleanly separated, exactly like gen_template_block.py:

  1. LAYOUT (this code) — geometry, the ONE table-row loop over {{#students}},
     placeholder KEYS. Generic proto/roster nouns only.
  2. PROFILE (per job_category) — the column SHAPE. v1 ships ACADEMIC_PROFILE:
     two period columns + a year-final column. Column HEADERS are placeholders
     ({{period1_label}} / {{period2_label}} / {{final_label}}) so the semester
     wording is RENDER-TIME DATA (lyngua + the roster's phase labels), never
     baked text — a school that names its terms differently needs no re-author.

Contract highlights baked into this layout (the fycha engine, fycha.md §1):
  - Rows = students → ONE table-row loop {{#students}}…{{/students}} (the pre-scan
    clones the single template row per item — xmlprocessor.go:260-334). No body
    loop, no nested loops: the 2-level cap is spent on the one row loop.
  - No conditionals; every cell is a pre-formatted %v string. The Go builder
    (data.go buildSheetData) seeds EVERY manifest path blank before overlay, so a
    student missing a period simply renders an empty cell (the residual scrub is
    the backstop, the manifest is the guarantee).
  - Title / subtitle / footer are ROOT scalars, processed once (headers/footers
    and the body outside the loop see the root data map).

Alongside the DOCX this emits grade-sheet-template-academic.manifest.json — every
root scalar + the one loop path + its item scalars, derived from the SAME profile
+ emission code. A self-check regexes the generated XML for {{...}} tokens and
asserts set-equality with the manifest, failing the build on any drift.

NOTE: the emitted .docx is an AUTHORING ASSET operators upload via the Grade
Sheet Templates settings page — it is deliberately NOT go:embed'd and NOT wired
as a render fallback (Q1 / entities.html §5: a resolver miss fails loud). ONLY the
manifest is embedded (a small, static blank-seed contract — see manifest_embed.go).

Run:  python3 gen_template_sheet.py
"""

import json
import os
import re
import zipfile

HERE = os.path.dirname(os.path.abspath(__file__))

# ===========================================================================
# PROFILES — one per job_category. A profile declares the COLUMN SHAPE only:
# how many period columns the sheet prints. The header wording is a placeholder
# (render-time data), so a profile never contains prose. v1 ships academic.
# ===========================================================================

ACADEMIC_PROFILE = {
    # Number of per-period final columns (academic = 2 semesters — the education1
    # ground truth: every active academic template has exactly 2 active phases).
    "period_columns": 2,
    # The output artifact basename (…docx / …manifest.json share it).
    "artifact_basename": "grade-sheet-template-academic",
}

# The active profile (single indirection so a future selector can swap it).
PROFILE = ACADEMIC_PROFILE

PERIOD_COLUMNS = PROFILE["period_columns"]
assert PERIOD_COLUMNS >= 1, "a grade sheet needs at least one period column"

# ===========================================================================
# Placeholder-key builders. Every key is a generic noun; the SAME builders feed
# both the emitted layout AND the manifest, so the two cannot drift.
# ===========================================================================

# Root scalars (title/subtitle/header-cells/footer). period<N>_label are the
# per-period column HEADERS; they are render-time data, not baked wording.
ROOT_SCALARS = [
    "sheet_title",
    "section_name",
    "academic_year",
    "name_label",
] + ["period%d_label" % (i + 1) for i in range(PERIOD_COLUMNS)] + [
    "final_label",
    "printed_by",
    "printed_at",
]

# The one table-row loop: one item per student, cells = name + per-period final +
# year final.
STUDENTS_LOOP = "students"
STUDENT_ITEM_SCALARS = ["name"] + ["period%d" % (i + 1) for i in range(PERIOD_COLUMNS)] + ["final"]

MANIFEST = {
    "scalars": ROOT_SCALARS,
    "loops": {
        STUDENTS_LOOP: {"scalars": STUDENT_ITEM_SCALARS, "loops": {}},
    },
}


def manifest_tokens(node):
    """Flatten a manifest node into the set of rendered {{...}} tokens it implies."""
    toks = {"{{%s}}" % s for s in node.get("scalars", [])}
    for loop_key, sub in node.get("loops", {}).items():
        toks.add("{{#%s}}" % loop_key)
        toks.add("{{/%s}}" % loop_key)
        toks |= manifest_tokens(sub)
    return toks


_TOKEN_RE = re.compile(r"\{\{\s*([^{}]*?)\s*\}\}")


def xml_tokens(*xml_parts):
    """Extract the set of {{...}} tokens present across the generated XML parts."""
    found = set()
    for part in xml_parts:
        for m in _TOKEN_RE.finditer(part):
            found.add("{{%s}}" % m.group(1).strip())
    return found


# ===========================================================================
# Generic layout — no profile prose. Placeholder KEYS are the generic contract.
# (A trimmed sibling of gen_template_block.py's OOXML helpers.)
# ===========================================================================

W_NS = 'xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"'

RED = "C00000"
GRAY = "808080"
BLACK = "000000"
WHITE = "FFFFFF"

# Page content width between the 1134-twip margins on US Letter (12240 twips).
CONTENT_W = 12240 - 1134 - 1134  # 9972


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def tok(path):
    return "{{%s}}" % path


def run(text, *, bold=False, color=None, size=None):
    """A text run. size in HALF-POINTS."""
    rpr = []
    if bold:
        rpr.append("<w:b/>")
    if color:
        rpr.append('<w:color w:val="%s"/>' % color)
    if size:
        rpr.append('<w:sz w:val="%d"/><w:szCs w:val="%d"/>' % (size, size))
    rpr_xml = "<w:rPr>%s</w:rPr>" % "".join(rpr) if rpr else ""
    return '<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>' % (rpr_xml, esc(text))


def para(runs, *, align=None, before=None, after=None, keep=False):
    ppr = []
    if keep:
        ppr.append("<w:keepNext/>")
    sp = []
    if before is not None:
        sp.append('w:before="%d"' % before)
    if after is not None:
        sp.append('w:after="%d"' % after)
    if sp:
        ppr.append("<w:spacing %s/>" % " ".join(sp))
    if align:
        ppr.append('<w:jc w:val="%s"/>' % align)
    ppr_xml = "<w:pPr>%s</w:pPr>" % "".join(ppr) if ppr else ""
    if isinstance(runs, str):
        runs = [runs]
    return "<w:p>%s%s</w:p>" % (ppr_xml, "".join(runs))


def cell(paras, *, width, shade=None, valign="center"):
    tcpr = ['<w:tcW w:w="%d" w:type="dxa"/>' % width]
    if shade:
        tcpr.append('<w:shd w:val="clear" w:color="auto" w:fill="%s"/>' % shade)
    tcpr.append('<w:vAlign w:val="%s"/>' % valign)
    if isinstance(paras, str):
        paras = [paras]
    return "<w:tc><w:tcPr>%s</w:tcPr>%s</w:tc>" % ("".join(tcpr), "".join(paras))


def row(cells, *, height=None, cant_split=True):
    trpr = []
    if cant_split:
        trpr.append("<w:cantSplit/>")
    if height:
        trpr.append('<w:trHeight w:val="%d" w:hRule="atLeast"/>' % height)
    trpr_xml = "<w:trPr>%s</w:trPr>" % "".join(trpr) if trpr else ""
    return "<w:tr>%s%s</w:tr>" % (trpr_xml, "".join(cells))


def table(rows, *, grid, borders=(4, BLACK)):
    sz, color = borders
    b = ("<w:tblBorders>"
         + "".join('<w:%s w:val="single" w:sz="%d" w:space="0" w:color="%s"/>' % (s, sz, color)
                   for s in ("top", "left", "bottom", "right", "insideH", "insideV"))
         + "</w:tblBorders>")
    tblpr = ('<w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/>%s'
             '<w:tblCellMar><w:left w:w="72" w:type="dxa"/><w:right w:w="72" w:type="dxa"/></w:tblCellMar>'
             "</w:tblPr>") % (sum(grid), b)
    grid_xml = "<w:tblGrid>%s</w:tblGrid>" % "".join('<w:gridCol w:w="%d"/>' % g for g in grid)
    return "<w:tbl>%s%s%s</w:tbl>" % (tblpr, grid_xml, "".join(rows))


def sectpr(footer_rid):
    return ("<w:sectPr>"
            '<w:footerReference w:type="default" r:id="%s"/>'
            '<w:footerReference w:type="first" r:id="%s"/>'
            '<w:pgSz w:w="12240" w:h="15840"/>'
            '<w:pgMar w:top="1134" w:right="1134" w:bottom="1134" w:left="1134" '
            'w:header="720" w:footer="480" w:gutter="0"/>'
            "</w:sectPr>") % (footer_rid, footer_rid)


# ---------------------------------------------------------------------------
# Column grid: name column takes ~40%, the remaining period+final columns split
# the rest evenly.
# ---------------------------------------------------------------------------

DATA_COLS = PERIOD_COLUMNS + 1  # period columns + year-final
NAME_W = int(CONTENT_W * 0.40)
REST_W = CONTENT_W - NAME_W
COL_W = REST_W // DATA_COLS
# Absorb the integer-division remainder into the name column so the grid sums
# to exactly CONTENT_W (fixed layout).
NAME_W = CONTENT_W - COL_W * DATA_COLS
GRID = [NAME_W] + [COL_W] * DATA_COLS

# Header cells: the roster label + each period label placeholder + the final label.
HEADER_KEYS = ["name_label"] + ["period%d_label" % (i + 1) for i in range(PERIOD_COLUMNS)] + ["final_label"]
# Body (loop template) cells: the student name + each period value + the year final.
BODY_KEYS = ["name"] + ["period%d" % (i + 1) for i in range(PERIOD_COLUMNS)] + ["final"]


def header_cells():
    out = []
    for i, key in enumerate(HEADER_KEYS):
        align = "left" if i == 0 else "center"
        out.append(cell(para([run(tok(key), bold=True, size=18, color=WHITE)], align=align, after=0, keep=True),
                        width=GRID[i], shade=RED))
    return out


def body_cells():
    out = []
    for i, key in enumerate(BODY_KEYS):
        align = "left" if i == 0 else "center"
        out.append(cell(para([run(tok(key), size=18)], align=align, after=0),
                        width=GRID[i]))
    return out


# ---------------------------------------------------------------------------
# Assemble the document body: title, subtitle, roster table.
# ---------------------------------------------------------------------------

body = []
body.append(para([run(tok("sheet_title"), bold=True, size=32)], align="center", after=60))
body.append(para([
    run(tok("section_name"), bold=True, size=22),
    run("   ", size=22),
    run(tok("academic_year"), size=22),
], align="center", after=200))

roster_table = table([
    row(header_cells(), height=340),
    # The ONE table-row loop: marker rows (blanked by the engine) wrap a single
    # template row cloned per student (fycha table-loop pre-scan).
    row([cell(para([run(tok("#" + STUDENTS_LOOP))], after=0), width=sum(GRID))]),
    row(body_cells(), height=300),
    row([cell(para([run(tok("/" + STUDENTS_LOOP))], after=0), width=sum(GRID))]),
], grid=GRID)
body.append(roster_table)

document_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    "<w:document %s><w:body>%s%s</w:body></w:document>"
) % (W_NS, "".join(body), sectpr("rId201"))

R_NS = 'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"'


def footer_xml():
    bar = para([run("", size=2)], after=60)
    text = para([
        run("Printed by: ", size=16, color=GRAY),
        run(tok("printed_by"), size=16, color=GRAY),
        run("   ", size=16, color=GRAY),
        run(tok("printed_at"), size=16, color=GRAY),
    ], align="right", after=0)
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
            "<w:ftr %s %s>%s%s</w:ftr>" % (W_NS, R_NS, bar, text))


content_types = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">'
    '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>'
    '<Default Extension="xml" ContentType="application/xml"/>'
    '<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>'
    '<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>'
    '<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>'
    "</Types>"
)

root_rels = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
    '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>'
    "</Relationships>"
)

doc_rels = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
    '<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>'
    '<Relationship Id="rId201" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>'
    "</Relationships>"
)

styles_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    "<w:styles %s>"
    "<w:docDefaults><w:rPrDefault><w:rPr>"
    '<w:rFonts w:ascii="Helvetica" w:hAnsi="Helvetica" w:cs="Arial"/>'
    '<w:sz w:val="18"/><w:szCs w:val="18"/>'
    "</w:rPr></w:rPrDefault>"
    '<w:pPrDefault><w:pPr><w:spacing w:after="0"/></w:pPr></w:pPrDefault></w:docDefaults>'
    '<w:style w:type="paragraph" w:default="1" w:styleId="Normal">'
    '<w:name w:val="Normal"/></w:style>'
    "</w:styles>"
) % W_NS

footer1_xml = footer_xml()

# Blank-guard self-check: the union of {{...}} tokens across document.xml + the
# footer MUST equal the manifest's flattened token set. A leaked placeholder
# (missing seed) or an undeclared token fails the build here.
_expected = manifest_tokens(MANIFEST)
_actual = xml_tokens(document_xml, footer1_xml)
if _expected != _actual:
    _missing = sorted(_expected - _actual)
    _extra = sorted(_actual - _expected)
    raise SystemExit(
        "manifest / artifact token mismatch — refusing to write\n"
        "  declared-but-not-emitted: %s\n"
        "  emitted-but-not-declared: %s" % (_missing, _extra))

BASENAME = PROFILE["artifact_basename"]
out = os.path.join(HERE, BASENAME + ".docx")
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", root_rels)
    z.writestr("word/document.xml", document_xml)
    z.writestr("word/_rels/document.xml.rels", doc_rels)
    z.writestr("word/styles.xml", styles_xml)
    z.writestr("word/footer1.xml", footer1_xml)

manifest_out = os.path.join(HERE, BASENAME + ".manifest.json")
with open(manifest_out, "w", encoding="utf-8") as f:
    json.dump(MANIFEST, f, indent=2, ensure_ascii=False)
    f.write("\n")

print("wrote", out, os.path.getsize(out), "bytes")
print("wrote", manifest_out,
      "(%d root scalars, 1 loop, %d item scalars)" % (len(MANIFEST["scalars"]), len(STUDENT_ITEM_SCALARS)))
print("self-check OK:", len(_actual), "distinct tokens match the manifest")
