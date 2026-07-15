#!/usr/bin/env python3
"""Generator for report-card-template.docx (fayna outcome_summary).

Produces a VALID .docx (a zip of OOXML) authored for the fycha `doctemplate`
engine (packages/fycha-golang/services/doctemplate). It uses ONLY the placeholder
+ section-loop syntax that engine supports:

  {{key}}            simple placeholder (root scalar; leaks verbatim if unresolved)
  {{#subjects}} ...  table-row loop start (its own <w:tr>)
  {{/subjects}}      table-row loop end   (its own <w:tr>)

No conditionals, no dynamic images, %v-stringify only. The Go builder
(buildReportCardData) emits EVERY key as a pre-formatted string so nothing leaks.

Block order (plan 20260713-report-card-documents §Render Contract / §5):
  COVER -> GRADES-HEADER -> SUBJECT TABLE {{#subjects}}..{{/subjects}}
  -> GRADE BOUNDARIES (fixed IB 7-row) -> FOOTER.

Styling is functional-first (MMIS red #C00000 approximated); pixel parity is a
later pass (W6) against a real MMIS card.

Run:  python3 gen_template.py   (writes report-card-template.docx beside this file)
"""

import os
import zipfile

W = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
RED = "C00000"
BLACK = "000000"


def esc(s):
    return (s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;"))


def run(text, *, bold=False, color=None, size=None):
    rpr = []
    if bold:
        rpr.append("<w:b/>")
    if color:
        rpr.append('<w:color w:val="%s"/>' % color)
    if size:
        rpr.append('<w:sz w:val="%d"/><w:szCs w:val="%d"/>' % (size, size))
    rpr_xml = "<w:rPr>%s</w:rPr>" % "".join(rpr) if rpr else ""
    return '<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>' % (rpr_xml, esc(text))


def para(runs, *, align=None, spacing_after=None):
    ppr = []
    if align:
        ppr.append('<w:jc w:val="%s"/>' % align)
    if spacing_after is not None:
        ppr.append('<w:spacing w:after="%d"/>' % spacing_after)
    ppr_xml = "<w:pPr>%s</w:pPr>" % "".join(ppr) if ppr else ""
    if isinstance(runs, str):
        runs = [runs]
    return "<w:p>%s%s</w:p>" % (ppr_xml, "".join(runs))


def cell(runs, *, width=None, shade=None, align=None):
    tcpr = []
    if width:
        tcpr.append('<w:tcW w:w="%d" w:type="dxa"/>' % width)
    if shade:
        tcpr.append('<w:shd w:val="clear" w:color="auto" w:fill="%s"/>' % shade)
    tcpr.append('<w:vAlign w:val="center"/>')
    tcpr_xml = "<w:tcPr>%s</w:tcPr>" % "".join(tcpr)
    return "<w:tc>%s%s</w:tc>" % (tcpr_xml, para(runs, align=align or "center"))


def row(cells):
    return "<w:tr>%s</w:tr>" % "".join(cells)


def table(rows, *, grid=None):
    borders = (
        "<w:tblBorders>"
        '<w:top w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        '<w:left w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        '<w:bottom w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        '<w:right w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        '<w:insideH w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        '<w:insideV w:val="single" w:sz="4" w:space="0" w:color="808080"/>'
        "</w:tblBorders>"
    )
    tblpr = (
        "<w:tblPr>"
        '<w:tblW w:w="5000" w:type="pct"/>'
        + borders
        + "</w:tblPr>"
    )
    grid_xml = ""
    if grid:
        grid_xml = "<w:tblGrid>%s</w:tblGrid>" % "".join(
            '<w:gridCol w:w="%d"/>' % w for w in grid
        )
    return "<w:tbl>%s%s%s</w:tbl>" % (tblpr, grid_xml, "".join(rows))


# ---- Subject table --------------------------------------------------------
# Columns: Subject | Crit A | B | C | D | Total/32 | Sem 1 | Sem 2 | MYP Overall
# Crit A-D + Total have NO per-criterion source on education1 (job_outcome_line
# is per-subject, not per-criterion) -> the builder emits them as a dash; the
# columns are retained structurally for the W6 MMIS-parity pass.
HEAD = ["Subject", "A", "B", "C", "D", "Total /32",
        "Semester 1", "Semester 2", "MYP Overall"]
GRID = [3200, 500, 500, 500, 500, 900, 1100, 1100, 1300]

header_cells = [
    cell([run(h, bold=True, color=RED, size=18)], width=GRID[i], shade="F2F2F2")
    for i, h in enumerate(HEAD)
]

# {{#subjects}} marker row (its own <w:tr>, single cell whose full text is the marker)
loop_start = row([cell([run("{{#subjects}}")], width=sum(GRID))])
# template row cloned per subject
tmpl_keys = ["subject_name", "crit_a", "crit_b", "crit_c", "crit_d",
             "criteria_total", "sem1_band", "sem2_band", "myp_overall"]
tmpl_cells = []
for i, k in enumerate(tmpl_keys):
    bold = (k == "myp_overall")
    color = RED if k == "myp_overall" else None
    align = "left" if k == "subject_name" else "center"
    tmpl_cells.append(cell([run("{{%s}}" % k, bold=bold, color=color)],
                           width=GRID[i], align=align))
tmpl_row = row(tmpl_cells)
loop_end = row([cell([run("{{/subjects}}")], width=sum(GRID))])

subjects_table = table([row(header_cells), loop_start, tmpl_row, loop_end], grid=GRID)

# ---- Grade boundaries (fixed IB MYP 7-row) --------------------------------
BOUNDARIES = [
    ("7", "28-32", "Produces high-quality, frequently innovative work; comprehensive understanding."),
    ("6", "24-27", "Produces high-quality work; wide-ranging understanding, consistently applied."),
    ("5", "19-23", "Produces generally high-quality work; secure understanding, usually applied well."),
    ("4", "15-18", "Produces good-quality work; solid understanding, sometimes applied."),
    ("3", "10-14", "Produces work of an acceptable quality; basic understanding, occasionally applied."),
    ("2", "6-9", "Produces work of limited quality; minimal understanding, rarely applied."),
    ("1", "1-5", "Produces work of very limited quality; very limited understanding."),
]
bh = [cell([run(h, bold=True, color=RED, size=16)], shade="F2F2F2")
      for h in ["Grade", "Boundary /32", "General descriptor"]]
brows = [row(bh)]
for g, b, d in BOUNDARIES:
    brows.append(row([
        cell([run(g, bold=True)], width=900),
        cell([run(b)], width=1400),
        cell([run(d)], width=8000, align="left"),
    ]))
boundaries_table = table(brows, grid=[900, 1400, 8000])

# ---- Document body --------------------------------------------------------
body_parts = []

# COVER
body_parts.append(para([run("{{school_name}}", bold=True, color=RED, size=32)],
                       align="center", spacing_after=40))
body_parts.append(para([run("MYP Report Card", bold=True, size=26)],
                       align="center", spacing_after=40))
body_parts.append(para([run("Academic Year {{academic_year}}", size=20)],
                       align="center", spacing_after=200))

# STUDENT HEADER
body_parts.append(para([run("Student: ", bold=True), run("{{student_name}}")]))
body_parts.append(para([run("Grade / Section: ", bold=True),
                        run("{{grade_level}} {{section_name}}")]))
body_parts.append(para([run("LRN: ", bold=True), run("{{lrn}}")], spacing_after=200))

# GRADES HEADER
body_parts.append(para([run("Academic Achievement", bold=True, color=RED, size=22)],
                       spacing_after=80))
body_parts.append(subjects_table)
body_parts.append(para([run(
    "MYP Overall is the stored year-final achievement level (1-7). Semester "
    "columns are the recomputed semestral-progress bands. Criterion columns "
    "(A-D, Total) are shown when per-criterion data is available.", size=14)],
    spacing_after=200))

# GRADE BOUNDARIES
body_parts.append(para([run("MYP Grade Boundaries", bold=True, color=RED, size=22)],
                       spacing_after=80))
body_parts.append(boundaries_table)

# FOOTER
body_parts.append(para([run("", size=14)], spacing_after=200))
body_parts.append(para([run(
    "Printed by {{printed_by}} on {{printed_at}}. This is a system-generated "
    "report card.", size=14, color="808080")], align="center"))

sectpr = (
    "<w:sectPr>"
    '<w:pgSz w:w="12240" w:h="15840"/>'
    '<w:pgMar w:top="720" w:right="720" w:bottom="720" w:left="720" '
    'w:header="480" w:footer="480" w:gutter="0"/>'
    "</w:sectPr>"
)

document_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<w:document xmlns:w="%s">'
    "<w:body>%s%s</w:body>"
    "</w:document>"
) % (W, "".join(body_parts), sectpr)

# ---- Minimal supporting parts ---------------------------------------------
content_types = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">'
    '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>'
    '<Default Extension="xml" ContentType="application/xml"/>'
    '<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>'
    '<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>'
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
    "</Relationships>"
)

styles_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<w:styles xmlns:w="%s">'
    '<w:docDefaults><w:rPrDefault><w:rPr>'
    '<w:rFonts w:ascii="Calibri" w:hAnsi="Calibri"/>'
    '<w:sz w:val="20"/><w:szCs w:val="20"/>'
    "</w:rPr></w:rPrDefault></w:docDefaults>"
    '<w:style w:type="paragraph" w:default="1" w:styleId="Normal">'
    '<w:name w:val="Normal"/></w:style>'
    "</w:styles>"
) % W

out = os.path.join(os.path.dirname(os.path.abspath(__file__)),
                   "report-card-template.docx")
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", root_rels)
    z.writestr("word/document.xml", document_xml)
    z.writestr("word/_rels/document.xml.rels", doc_rels)
    z.writestr("word/styles.xml", styles_xml)

print("wrote", out, os.path.getsize(out), "bytes")
