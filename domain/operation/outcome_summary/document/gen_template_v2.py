#!/usr/bin/env python3
"""Generator for report-card-template-v2.docx (fayna outcome_summary).

The v2 FAITHFUL block-layout template: a page-accurate rebuild of the operator's
printed report card (spec: docs/plan/20260713-report-card-documents/
codex-pdf-spec.md — geometry measured from the live PDF content stream).

Layout (4 Word sections):
  S1 cover      — logo + right title block, identity lines, letter, signatory.
  S2 grades     — first-page header (logo + corner + title + identity block),
                  continuation header (identity block); body = {{#subjects}}
                  BODY loop, each block a heading line + a 4-column criteria
                  table with a nested {{#criteria}} TABLE loop (the engine's
                  documented 2-level nesting).
  S3 boundary   — the full 7-row ascending grade-boundary table (long IB
                  descriptors), no header.
  S4 formation  — identity header (red underlines); conduct tables
                  ({{#conduct_rows}} root-scope table loop + group conduct),
                  descriptor legend, attendance scaffold (no data source — the
                  cells are intentionally blank), certificate of transfer.
Shared footer on every page: the red bar + "Printed by: … | … | Page N"
(native PAGE field).

Engine constraints honored (fycha doctemplate): ONE body-level loop per
document ({{#subjects}}); one row-loop per table; headers/footers processed
with root data; no conditionals — every key is emitted by the Go builder.

Run:  python3 gen_template_v2.py   (writes report-card-template-v2.docx here)
"""

import os
import zipfile

HERE = os.path.dirname(os.path.abspath(__file__))

W_NS = 'xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"'
R_NS = 'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"'
WP_NS = 'xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"'
A_NS = 'xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"'
PIC_NS = 'xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"'
ALL_NS = " ".join([W_NS, R_NS, WP_NS, A_NS, PIC_NS])

RED = "C00000"
GRAY = "808080"
BLACK = "000000"
WHITE = "FFFFFF"

CONTENT_W = 9972  # twips between the 1134 margins (498.6 pt)

EMU_PER_PT = 12700


def esc(s):
    return s.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")


def run(text, *, bold=False, color=None, size=None, underline=None):
    """A text run. size in HALF-POINTS. underline: None | 'black' | 'red'."""
    rpr = []
    if bold:
        rpr.append("<w:b/>")
    if underline == "black":
        rpr.append('<w:u w:val="single"/>')
    elif underline == "red":
        rpr.append('<w:u w:val="single" w:color="%s"/>' % RED)
    if color:
        rpr.append('<w:color w:val="%s"/>' % color)
    if size:
        rpr.append('<w:sz w:val="%d"/><w:szCs w:val="%d"/>' % (size, size))
    rpr_xml = "<w:rPr>%s</w:rPr>" % "".join(rpr) if rpr else ""
    return '<w:r>%s<w:t xml:space="preserve">%s</w:t></w:r>' % (rpr_xml, esc(text))


def tab():
    return "<w:r><w:tab/></w:r>"


def page_field(size=16, color=GRAY):
    rpr = '<w:rPr><w:color w:val="%s"/><w:sz w:val="%d"/><w:szCs w:val="%d"/></w:rPr>' % (color, size, size)
    return ('<w:fldSimple w:instr=" PAGE "><w:r>%s<w:t>1</w:t></w:r></w:fldSimple>' % rpr)


def para(runs, *, align=None, before=None, after=None, line=None, keep=False,
         tabs=None, border_bottom=None, indent_left=None):
    """tabs: list of (val, pos, leader|None). border_bottom: (sz_eighths, color)."""
    ppr = []
    if keep:
        ppr.append("<w:keepNext/>")
    if tabs:
        t = "".join(
            '<w:tab w:val="%s" w:pos="%d"%s/>' % (v, p, ' w:leader="%s"' % l if l else "")
            for v, p, l in tabs)
        ppr.append("<w:tabs>%s</w:tabs>" % t)
    sp = []
    if before is not None:
        sp.append('w:before="%d"' % before)
    if after is not None:
        sp.append('w:after="%d"' % after)
    if line is not None:
        sp.append('w:line="%d" w:lineRule="exact"' % line)
    if sp:
        ppr.append("<w:spacing %s/>" % " ".join(sp))
    if indent_left is not None:
        ppr.append('<w:ind w:left="%d"/>' % indent_left)
    if border_bottom:
        ppr.append('<w:pBdr><w:bottom w:val="single" w:sz="%d" w:space="1" w:color="%s"/></w:pBdr>'
                   % border_bottom)
    if align:
        ppr.append('<w:jc w:val="%s"/>' % align)
    ppr_xml = "<w:pPr>%s</w:pPr>" % "".join(ppr) if ppr else ""
    if isinstance(runs, str):
        runs = [runs]
    return "<w:p>%s%s</w:p>" % (ppr_xml, "".join(runs))


def image(rid, w_pt, h_pt, doc_id):
    cx, cy = int(w_pt * EMU_PER_PT), int(h_pt * EMU_PER_PT)
    return (
        '<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        '<wp:extent cx="%(cx)d" cy="%(cy)d"/>'
        '<wp:docPr id="%(id)d" name="img%(id)d"/>'
        '<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">'
        '<pic:pic>'
        '<pic:nvPicPr><pic:cNvPr id="%(id)d" name="img%(id)d"/><pic:cNvPicPr/></pic:nvPicPr>'
        '<pic:blipFill><a:blip r:embed="%(rid)s"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>'
        '<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%(cx)d" cy="%(cy)d"/></a:xfrm>'
        '<a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>'
        '</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>'
    ) % {"cx": cx, "cy": cy, "id": doc_id, "rid": rid}


def cell(paras, *, width=None, shade=None, valign="center", borders=None, no_margins=False):
    """borders: dict side→(sz, color) or 'none' for an explicitly borderless side."""
    tcpr = []
    if width:
        tcpr.append('<w:tcW w:w="%d" w:type="dxa"/>' % width)
    if borders is not None:
        sides = []
        for side in ("top", "left", "bottom", "right"):
            spec = borders.get(side)
            if spec == "none":
                sides.append('<w:%s w:val="nil"/>' % side)
            elif spec:
                sides.append('<w:%s w:val="single" w:sz="%d" w:space="0" w:color="%s"/>' % (side, spec[0], spec[1]))
        if sides:
            tcpr.append("<w:tcBorders>%s</w:tcBorders>" % "".join(sides))
    if shade:
        tcpr.append('<w:shd w:val="clear" w:color="auto" w:fill="%s"/>' % shade)
    if no_margins:
        tcpr.append('<w:tcMar><w:left w:w="0" w:type="dxa"/><w:right w:w="0" w:type="dxa"/></w:tcMar>')
    tcpr.append('<w:vAlign w:val="%s"/>' % valign)
    if isinstance(paras, str):
        paras = [paras]
    return "<w:tc><w:tcPr>%s</w:tcPr>%s</w:tc>" % ("".join(tcpr), "".join(paras))


def row(cells, *, height=None, hrule="exact", cant_split=True):
    trpr = []
    if cant_split:
        trpr.append("<w:cantSplit/>")
    if height:
        trpr.append('<w:trHeight w:val="%d" w:hRule="%s"/>' % (height, hrule))
    trpr_xml = "<w:trPr>%s</w:trPr>" % "".join(trpr) if trpr else ""
    return "<w:tr>%s%s</w:tr>" % (trpr_xml, "".join(cells))


def table(rows, *, grid, width=None, borders=None, layout_fixed=True, cell_margin=57):
    """borders: (sz, color) for a full grid, or None for a borderless table."""
    b = ""
    if borders:
        sz, color = borders
        b = ("<w:tblBorders>"
             + "".join('<w:%s w:val="single" w:sz="%d" w:space="0" w:color="%s"/>' % (s, sz, color)
                       for s in ("top", "left", "bottom", "right", "insideH", "insideV"))
             + "</w:tblBorders>")
    w = width if width else sum(grid)
    layout = '<w:tblLayout w:type="fixed"/>' if layout_fixed else ""
    tblpr = ('<w:tblPr><w:tblW w:w="%d" w:type="dxa"/>%s%s'
             '<w:tblCellMar><w:left w:w="%d" w:type="dxa"/><w:right w:w="%d" w:type="dxa"/></w:tblCellMar>'
             "</w:tblPr>") % (w, b, layout, cell_margin, cell_margin)
    grid_xml = "<w:tblGrid>%s</w:tblGrid>" % "".join('<w:gridCol w:w="%d"/>' % g for g in grid)
    return "<w:tbl>%s%s%s</w:tbl>" % (tblpr, grid_xml, "".join(rows))


# ---------------------------------------------------------------------------
# Shared pieces
# ---------------------------------------------------------------------------

def identity_block(underline):
    """The two-line student identity block (grades/formation headers)."""
    gap = "   "
    line1 = para([
        run("Name: ", size=16),
        run("{{student_name}}", bold=True, size=16, underline=underline),
        run(gap + "Grade Level / Section: ", size=16),
        run("{{grade_section}}", bold=True, size=16, underline=underline),
    ], after=40)
    line2 = para([
        run("Academic Year: ", size=16),
        run("{{academic_year}}", bold=True, size=16, underline=underline),
        run(gap + "LRN: ", size=16),
        run("{{client_reference}}", bold=True, size=16, underline=underline),
        run(gap + "Adviser: ", size=16),
        run("{{adviser}}", bold=True, size=16, underline=underline),
    ], after=80)
    return line1 + line2


def footer_xml():
    bar = para([run("", size=2)], border_bottom=(45, RED), after=60)
    text = para([
        run("Printed by: {{printed_by_name}} | {{printed_at_long}} | Page ",
            size=16, color=GRAY),
        page_field(),
    ], align="right", after=0)
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
            "<w:ftr %s>%s%s</w:ftr>" % (ALL_NS, bar, text))


def header_empty():
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
            "<w:hdr %s>%s</w:hdr>" % (ALL_NS, para([run("", size=2)], after=0)))


def header_grades_first():
    """S2 first page: logo | REPORT CARD | corner sweep, then identity block."""
    banner = table([
        row([
            cell(para([image("rId101", 113.69, 42.52, 11)], after=0), width=3324,
                 valign="top", borders={s: "none" for s in ("top", "left", "bottom", "right")}, no_margins=True),
            cell(para([run("REPORT CARD", bold=True, size=16)], align="center", after=0),
                 width=3324, valign="bottom",
                 borders={s: "none" for s in ("top", "left", "bottom", "right")}),
            cell(para([image("rId102", 67.92, 42.52, 12)], align="right", after=0), width=3324,
                 valign="top", borders={s: "none" for s in ("top", "left", "bottom", "right")}, no_margins=True),
        ], height=880, hrule="atLeast", cant_split=True),
    ], grid=[3324, 3324, 3324], borders=None)
    spacer = para([run("", size=8)], after=60)
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
            "<w:hdr %s>%s%s%s</w:hdr>" % (ALL_NS, banner, spacer, identity_block("black")))


def header_identity(underline):
    return ('<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
            "<w:hdr %s>%s</w:hdr>" % (ALL_NS, identity_block(underline)))


def sectpr(*, header_first=None, header_default=None, footer="rId201",
           title_pg=False, top=1134):
    refs = []
    if header_first:
        refs.append('<w:headerReference w:type="first" r:id="%s"/>' % header_first)
    if header_default:
        refs.append('<w:headerReference w:type="default" r:id="%s"/>' % header_default)
    if footer:
        refs.append('<w:footerReference w:type="default" r:id="%s"/>' % footer)
        refs.append('<w:footerReference w:type="first" r:id="%s"/>' % footer)
    tp = "<w:titlePg/>" if title_pg else ""
    return ("<w:sectPr>%s"
            '<w:pgSz w:w="12240" w:h="15840"/>'
            '<w:pgMar w:top="%d" w:right="1134" w:bottom="1134" w:left="1134" '
            'w:header="1134" w:footer="480" w:gutter="0"/>%s</w:sectPr>') % ("".join(refs), top, tp)


def section_break(sp):
    return "<w:p><w:pPr>%s</w:pPr></w:p>" % sp


# ---------------------------------------------------------------------------
# S1 — cover
# ---------------------------------------------------------------------------

body = []

nb = {s: "none" for s in ("top", "left", "bottom", "right")}
cover_banner = table([
    row([
        cell(para([image("rId103", 189.48, 70.87, 13)], after=0), width=4986, valign="top",
             borders=nb, no_margins=True),
        cell([
            para([run("MYP Report Card", bold=True, size=32)], align="right", after=60),
            para([run("Academic Year {{academic_year_spaced}}", size=20)], align="right", after=0),
        ], width=4986, valign="top", borders=nb, no_margins=True),
    ], height=1420, hrule="atLeast", cant_split=True),
], grid=[4986, 4986], borders=None)
body.append(cover_banner)

body.append(para([run("Student's Name: {{student_name}}", size=22)], before=700, after=80))
body.append(para([run("Grade level & Section: {{grade_section}}", size=22)], after=80))
body.append(para([run("Homeroom Adviser: {{adviser}}", size=22)], after=80))

LETTER = [
    "Dear Parents and Guardians,",
    "Warm greetings from MMIS.",
    "We are pleased to share your child's MYP Final Report Card, which reflects "
    "their overall learning progress for the academic year.",
    "As an IB MYP Candidate School, our reporting is criteria-based, focusing on "
    "your child's level of achievement in relation to subject objectives. The "
    "final rating represents your child's performance based on demonstrated "
    "understanding and consistent work over time. Please refer to the MYP Overall "
    "Achievement Grade and Boundary to better understand your child's performance.",
    "We encourage you to review the report with your child and continue guiding "
    "them in their learning journey. Should you have any questions, please feel "
    "free to reach out.",
    "Thank you for your continued partnership.",
]
body.append(para([run(LETTER[0], size=20)], before=560, after=200))
for p in LETTER[1:]:
    body.append(para([run(p, size=20)], after=200, line=283))
body.append(para([run("Warm regards,", size=20)], before=200, after=400))
body.append(para([run("Ms. Mia Villamor Young", bold=True, size=22)], after=20))
body.append(para([run("School Director / Acting School Principal", size=20)], after=20))
body.append(para([run("Maria Montessori International School", size=20)], after=0))

body.append(section_break(sectpr(header_default="rId301")))  # S1 end (empty header)

# ---------------------------------------------------------------------------
# S2 — grades (the {{#subjects}} body loop)
# ---------------------------------------------------------------------------

SUBJ_GRID = [6232, 1246, 1246, 1248]
FULLB = (5, BLACK)


def subj_header_cell(text, width, align):
    return cell(para([run(text, bold=True, size=16, color=WHITE)], align=align, after=0, keep=True),
                width=width, shade=RED)


def subj_cells(vals, *, bold=False, color=None, shade=None, keep=True, black_cell=None):
    out = []
    for i, v in enumerate(vals):
        align = "left" if i == 0 else "center"
        sh = shade
        if black_cell is not None and i == black_cell:
            out.append(cell(para([run("", size=16)], align="center", after=0, keep=keep),
                            width=SUBJ_GRID[i], shade=BLACK))
            continue
        out.append(cell(para([run(v, bold=bold, size=16, color=color)], align=align, after=0, keep=keep),
                        width=SUBJ_GRID[i], shade=sh))
    return out


body.append(para([run("{{#subjects}}")]))

body.append(para([
    run("{{subject_title}}", size=16),
    tab(),
    run("{{teacher_line}}", size=16),
], tabs=[("right", CONTENT_W, None)], before=160, after=50, keep=True))

subject_table = table([
    row([subj_header_cell("Assessment Criterion", SUBJ_GRID[0], "left"),
         subj_header_cell("Highest Level", SUBJ_GRID[1], "center"),
         subj_header_cell("Semester 1", SUBJ_GRID[2], "center"),
         subj_header_cell("Semester 2", SUBJ_GRID[3], "center")],
        height=255),
    row([cell(para([run("{{#criteria}}")], after=0), width=CONTENT_W)]),
    row(subj_cells(["{{crit_label}}", "8", "{{crit_sem1}}", "{{crit_sem2}}"]), height=255),
    row([cell(para([run("{{/criteria}}")], after=0), width=CONTENT_W)]),
    row(subj_cells(["Criteria Total", "32", "{{sem1_total}}", "{{sem2_total}}"],
                   bold=True, color=WHITE, shade=RED), height=255),
    row(subj_cells(["Semestral Progress", "7", "{{sem1_band}}", "{{sem2_band}}"],
                   bold=True), height=255),
    row(subj_cells(["MYP Overall Achievement Grade", "7", "", "{{myp_overall}}"],
                   bold=True, keep=False, black_cell=2), height=255),
], grid=SUBJ_GRID, borders=FULLB)
body.append(subject_table)
body.append(para([run("", size=8)], after=60))

body.append(para([run("{{/subjects}}")]))

body.append(section_break(sectpr(header_first="rId302", header_default="rId303", title_pg=True)))

# ---------------------------------------------------------------------------
# S3 — grade boundary
# ---------------------------------------------------------------------------

BOUNDARIES = [
    ("1", "1-5",
     "Produces work of very limited quality. Conveys many significant misunderstandings or lacks "
     "understanding of most concepts and contexts. Very rarely demonstrate critical or creative "
     "thinking. Very inflexible, rarely using knowledge or skills."),
    ("2", "6-9",
     "Produces work of limited quality. Expresses misunderstandings or significant gaps in "
     "understanding for many concepts and contexts. Infrequently demonstrates critical or creative "
     "thinking. Generally inflexible in the use of knowledge and skills, infrequently applying "
     "knowledge and skills."),
    ("3", "10-14",
     "Produces work of an acceptable quality. Communicates basic understanding of many concepts "
     "and contexts, with occasionally significant misunderstandings or gaps. Begins to demonstrate "
     "some basic critical and creative thinking. Is often inflexible in the use of knowledge and "
     "skills, requiring support even in familiar classroom situations."),
    ("4", "15-18",
     "Produces good-quality work. Communicates basic understanding of most concepts and contexts "
     "with few misunderstandings and minor gaps. Often demonstrates basic critical and creative "
     "thinking. Uses knowledge and skills with some flexibility in familiar classroom situations "
     "but requires support in unfamiliar situations."),
    ("5", "19-23",
     "Produces generally high-quality work. Communicates secure understanding of concepts and "
     "contexts. Demonstrates critical and creative thinking, sometimes with sophistication. Uses "
     "knowledge and skills in familiar classroom and, with support, some unfamiliar real-world "
     "situations."),
    ("6", "24-27",
     "Produces high-quality, occasionally innovative work. Communicates extensive understanding of "
     "concepts and contexts. Demonstrates critical and creative thinking, frequently with "
     "sophistication. Uses knowledge and skills in familiar and unfamiliar classroom and real-world "
     "situations, often with independence."),
    ("7", "28-32",
     "Produces high-quality, frequently innovative work. Communicates comprehensive, nuanced "
     "understanding of concepts and contexts. Consistently demonstrates sophisticated critical and "
     "creative thinking. Frequently transfers knowledge and skills with independence and expertise "
     "in a variety of complex classroom and real-world situations."),
]

BND_GRID = [767, 2301, 6904]
body.append(para([run("MYP Overall Achievement Grade and Boundary", bold=True, size=24)],
                 before=60, after=120))
bnd_rows = [row([
    cell(para([run("Grade", size=16, color=WHITE)], align="center", after=0), width=BND_GRID[0], shade=RED),
    cell(para([run("Boundary Guidelines", size=16, color=WHITE)], align="center", after=0), width=BND_GRID[1], shade=RED),
    cell(para([run("Descriptors", size=16, color=WHITE)], align="center", after=0), width=BND_GRID[2], shade=RED),
], height=567, hrule="atLeast")]
for g, b, d in BOUNDARIES:
    bnd_rows.append(row([
        cell(para([run(g, size=16)], align="center", after=0), width=BND_GRID[0]),
        cell(para([run(b, size=16)], align="center", after=0), width=BND_GRID[1]),
        cell(para([run(d, size=16)], align="left", after=0), width=BND_GRID[2]),
    ], height=850, hrule="atLeast"))
body.append(table(bnd_rows, grid=BND_GRID, borders=FULLB))

body.append(section_break(sectpr(header_default="rId301")))

# ---------------------------------------------------------------------------
# S4 — formation (conduct + legend + attendance + certificate)
# ---------------------------------------------------------------------------

DEP_GRID = [5669, 2268, 2268]
red_bottom = {"top": "none", "left": "none", "right": "none", "bottom": (5, RED)}


def dep_header_row(title):
    return row([
        cell(para([run(title, bold=True, size=18, color=RED)], align="left", after=0),
             width=DEP_GRID[0], borders=red_bottom),
        cell(para([run("1st Semester", bold=True, size=18, color=RED)], align="center", after=0),
             width=DEP_GRID[1], borders=red_bottom),
        cell(para([run("2nd Semester", bold=True, size=18, color=RED)], align="center", after=0),
             width=DEP_GRID[2], borders=red_bottom),
    ], height=360, hrule="atLeast")


def dep_data_row(vals):
    return row([
        cell(para([run(vals[0], size=16)], align="left", after=0), width=DEP_GRID[0], borders=red_bottom),
        cell(para([run(vals[1], size=16)], align="center", after=0), width=DEP_GRID[1], borders=red_bottom),
        cell(para([run(vals[2], size=16)], align="center", after=0), width=DEP_GRID[2], borders=red_bottom),
    ], height=340, hrule="atLeast")


subject_dep_table = table([
    dep_header_row("Subject Deportment"),
    row([cell(para([run("{{#conduct_rows}}")], after=0), width=sum(DEP_GRID))]),
    dep_data_row(["{{conduct_title}}", "{{conduct_sem1}}", "{{conduct_sem2}}"]),
    row([cell(para([run("{{/conduct_rows}}")], after=0), width=sum(DEP_GRID))]),
], grid=DEP_GRID, borders=None)
body.append(subject_dep_table)
body.append(para([run("", size=8)], after=120))

homeroom_table = table([
    dep_header_row("Homeroom Deportment"),
    dep_data_row(["Grade", "{{group_conduct_sem1}}", "{{group_conduct_sem2}}"]),
], grid=DEP_GRID, borders=None)
body.append(homeroom_table)
body.append(para([run("", size=8)], after=120))

body.append(para([run("Deportment Grade Descriptors", bold=True, size=18)], after=60))
LEGEND = [
    ("90% - 100%", "Outstanding (O)"),
    ("85% - 89%", "Very Satisfactory (VS)"),
    ("80% - 84%", "Satisfactory (S)"),
    ("75% - 79%", "Fairly Satisfactory (FS)"),
    ("74% and below", "Did not meet expectations (NM)"),
]
LEG_GRID = [2268, 4535]
leg_rows = [row([
    cell(para([run("Grade Boundary", bold=True, size=16)], align="left", after=0), width=LEG_GRID[0], borders=red_bottom),
    cell(para([run("Descriptors", bold=True, size=16)], align="left", after=0), width=LEG_GRID[1], borders=red_bottom),
], height=284, hrule="atLeast")]
for bnd, desc in LEGEND:
    leg_rows.append(row([
        cell(para([run(bnd, size=16)], align="left", after=0), width=LEG_GRID[0], borders=red_bottom),
        cell(para([run(desc, size=16)], align="left", after=0), width=LEG_GRID[1], borders=red_bottom),
    ], height=284, hrule="atLeast"))
body.append(table(leg_rows, grid=LEG_GRID, borders=None))
body.append(para([run("", size=8)], after=120))

body.append(para([run("Attendance", bold=True, size=18)], after=60))
MONTHS = ["July", "August", "September", "October", "November", "December",
          "January", "February", "March", "April", "May", "Total"]
# Column widths tuned so no month header wraps at 7 pt ("September" is the
# widest and gets the extra twips shaved off the label column).
ATT_GRID = [1611, 689, 689, 779, 689, 689, 689, 689, 689, 689, 689, 689, 692]
att_rows = [row(
    [cell(para([run("", size=14)], after=0), width=ATT_GRID[0], borders=red_bottom)] +
    [cell(para([run(m, size=14)], align="center", after=0), width=ATT_GRID[i + 1], borders=red_bottom)
     for i, m in enumerate(MONTHS)],
    height=284, hrule="atLeast")]
for label in ("Days of School", "Days Present", "Times Tardy"):
    att_rows.append(row(
        [cell(para([run(label, size=14)], align="left", after=0), width=ATT_GRID[0], borders=red_bottom)] +
        [cell(para([run("", size=14)], align="center", after=0), width=ATT_GRID[i + 1], borders=red_bottom)
         for i in range(12)],
        height=284, hrule="atLeast"))
body.append(table(att_rows, grid=ATT_GRID, borders=None, cell_margin=10))
body.append(para([run("", size=8)], after=160))

body.append(para([run("CERTIFICATE OF TRANSFER", bold=True, size=18)], align="center", after=120))

CERT_TAB = 4700


def cert_half(heading, admit_label):
    return [
        para([run(heading, size=16)], after=120, keep=True),
        para([run(admit_label, size=16), tab()],
             tabs=[("right", CERT_TAB, "underscore")], after=120),
        para([run("Date", size=16), tab()],
             tabs=[("right", CERT_TAB, "underscore")], after=280),
        para([run("Maria Corazon Villamor-Young", size=16)], align="right", after=20),
        para([run("School Principal", size=16)], align="right", after=0),
    ]


cert_table = table([row([
    cell(cert_half("Eligible for transfer and admission", "To"),
         width=4986, valign="top",
         borders={"top": "none", "left": "none", "bottom": "none", "right": (5, RED)}),
    cell(cert_half("Cancellation of Transfer Eligibility", "Has been admitted to"),
         width=4986, valign="top", borders=nb),
], cant_split=True)], grid=[4986, 4986], borders=None)
body.append(cert_table)

# ---------------------------------------------------------------------------
# Assemble document
# ---------------------------------------------------------------------------

document_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    "<w:document %s><w:body>%s%s</w:body></w:document>"
) % (ALL_NS, "".join(body), sectpr(header_default="rId304"))

content_types = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">'
    '<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>'
    '<Default Extension="xml" ContentType="application/xml"/>'
    '<Default Extension="png" ContentType="image/png"/>'
    '<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>'
    '<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>'
    + "".join('<Override PartName="/word/header%d.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>' % i
              for i in (1, 2, 3, 4))
    + '<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>'
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
    '<Relationship Id="rId103" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/logo.png"/>'
    '<Relationship Id="rId201" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>'
    '<Relationship Id="rId301" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>'
    '<Relationship Id="rId302" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header2.xml"/>'
    '<Relationship Id="rId303" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header3.xml"/>'
    '<Relationship Id="rId304" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header4.xml"/>'
    "</Relationships>"
)

header2_rels = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    '<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">'
    '<Relationship Id="rId101" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/logo.png"/>'
    '<Relationship Id="rId102" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/corner.png"/>'
    "</Relationships>"
)

styles_xml = (
    '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
    "<w:styles %s>"
    "<w:docDefaults><w:rPrDefault><w:rPr>"
    '<w:rFonts w:ascii="Helvetica" w:hAnsi="Helvetica" w:cs="Arial"/>'
    '<w:sz w:val="16"/><w:szCs w:val="16"/>'
    "</w:rPr></w:rPrDefault>"
    "<w:pPrDefault><w:pPr><w:spacing w:after="
    '"0"/></w:pPr></w:pPrDefault></w:docDefaults>'
    '<w:style w:type="paragraph" w:default="1" w:styleId="Normal">'
    '<w:name w:val="Normal"/></w:style>'
    "</w:styles>"
) % W_NS

out = os.path.join(HERE, "report-card-template-v2.docx")
with zipfile.ZipFile(out, "w", zipfile.ZIP_DEFLATED) as z:
    z.writestr("[Content_Types].xml", content_types)
    z.writestr("_rels/.rels", root_rels)
    z.writestr("word/document.xml", document_xml)
    z.writestr("word/_rels/document.xml.rels", doc_rels)
    z.writestr("word/_rels/header2.xml.rels", header2_rels)
    z.writestr("word/styles.xml", styles_xml)
    z.writestr("word/header1.xml", header_empty())
    z.writestr("word/header2.xml", header_grades_first())
    z.writestr("word/header3.xml", header_identity("black"))
    z.writestr("word/header4.xml", header_identity("red"))
    z.writestr("word/footer1.xml", footer_xml())
    with open(os.path.join(HERE, "asset-logo.png"), "rb") as f:
        z.writestr("word/media/logo.png", f.read())
    with open(os.path.join(HERE, "asset-corner.png"), "rb") as f:
        z.writestr("word/media/corner.png", f.read())

print("wrote", out, os.path.getsize(out), "bytes")
