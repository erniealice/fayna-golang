#!/usr/bin/env python3
"""Generate an example group-matrix (subscription-group outcome matrix) DOCX.

Data source: the section matrix (stored profile name
subscription_group_outcome_matrix_single_period_11_v1). The job-template columns are one
table column loop, so a single template serves any column count; the tokens it may use
are the vocabulary in the profile manifest.

With no options this writes the package-owned generic template (no images, title from
{{sheet_title}}). Options produce a workspace-branded authoring artifact for upload
through the Section Templates page; branding is template data, never code:

  --output PATH             write here instead of the package template path
  --title TEXT              fixed title text instead of {{sheet_title}}
  --title-prefix TEXT       fixed text printed before {{sheet_title}}
  --caps                    render the title and name heading in capitals
  --band-caps               print band headings (row_band_label) in capitals and
                            client names (row_client_label) as-is
  --schedule-prefix TEXT    fixed text printed before {{price_schedule_name_display}}
  --no-colour               leave rating cells uncoloured (no fill/text colour tokens)
  --name-heading TEXT       fixed first-column heading instead of {{client_name_label}}
  --portrait                portrait page (default landscape)
  --name-width TWIPS        first-column width (default 3900 landscape, 7800 portrait)
  --logo PATH               image at the top left
  --corner PATH             image at the top right

Only the standard library is used; ZIP metadata is fixed so output is deterministic.
"""

from __future__ import annotations

import argparse
import json
import re
import struct
from pathlib import Path
from xml.sax.saxutils import escape
from zipfile import ZIP_DEFLATED, ZipFile, ZipInfo

HERE = Path(__file__).resolve().parent
MANIFEST = HERE / "subscription-group-outcome-matrix-single-period-11-v1.manifest.json"
DEFAULT_OUTPUT = HERE / "matrix-template-example.docx"
TOKEN_RE = re.compile(r"\{\{([^{}]+)\}\}")
EMU_PER_PT = 12700
MARGIN = 454
FONT = '<w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:cs="Arial"/>'

W_NS = 'xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"'
R_NS = 'xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"'
WP_NS = 'xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"'
A_NS = 'xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"'
PIC_NS = 'xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"'


def token(name: str) -> str:
    return "{{" + name + "}}"


def run(text: str, *, bold: bool = False, caps: bool = False, size: int = 16, extra: str = "") -> str:
    props = FONT + ("<w:b/>" if bold else "") + ("<w:caps/>" if caps else "") + extra + f'<w:sz w:val="{size}"/>'
    return f'<w:r><w:rPr>{props}</w:rPr><w:t xml:space="preserve">{escape(text)}</w:t></w:r>'


def para(runs: list[str], *, align: str = "center", after: int = 0) -> str:
    return (
        f'<w:p><w:pPr><w:spacing w:before="0" w:after="{after}" w:line="240" w:lineRule="auto"/>'
        f'<w:jc w:val="{align}"/></w:pPr>{"".join(runs)}</w:p>'
    )


def cell(width: int, paragraphs: str, *, fill_token: str | None = None, valign: str = "center", borders: bool = True) -> str:
    shade = f'<w:shd w:val="clear" w:color="auto" w:fill="{token(fill_token)}"/>' if fill_token else ""
    nob = "" if borders else (
        '<w:tcBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/></w:tcBorders>'
    )
    return (
        f'<w:tc><w:tcPr><w:tcW w:w="{width}" w:type="dxa"/>{nob}{shade}<w:vAlign w:val="{valign}"/></w:tcPr>'
        f"{paragraphs}</w:tc>"
    )


def image_size(path: Path) -> tuple[int, int]:
    head = path.read_bytes()[:24]
    if head[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"{path} is not a PNG")
    return struct.unpack(">II", head[16:24])


def image_run(rid: str, doc_id: int, width_pt: float, height_pt: float) -> str:
    cx, cy = int(width_pt * EMU_PER_PT), int(height_pt * EMU_PER_PT)
    return (
        '<w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0">'
        f'<wp:extent cx="{cx}" cy="{cy}"/><wp:docPr id="{doc_id}" name="image{doc_id}"/>'
        '<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic>'
        f'<pic:nvPicPr><pic:cNvPr id="{doc_id}" name="image{doc_id}"/><pic:cNvPicPr/></pic:nvPicPr>'
        f'<pic:blipFill><a:blip r:embed="{rid}"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill>'
        f'<pic:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="{cx}" cy="{cy}"/></a:xfrm>'
        '<a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr>'
        "</pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r>"
    )


def heading_paragraphs(args: argparse.Namespace) -> str:
    if args.title:
        title_runs = [run(args.title, bold=True, caps=args.caps, size=18)]
    else:
        title_runs = []
        if args.title_prefix:
            title_runs.append(run(args.title_prefix, bold=True, caps=args.caps, size=18))
        title_runs.append(run(token("sheet_title"), bold=True, caps=args.caps, size=18))
    return (
        para(title_runs, after=60)
        + para([run(token("subscription_group_name_display"), bold=True, size=17)], after=60)
        + para(([run(args.schedule_prefix, bold=True, size=17)] if args.schedule_prefix else []) + [
            run(token("price_schedule_name_display"), bold=True, size=17),
            run(" (", bold=True, size=17),
            run(token("period_name_display"), bold=True, size=17),
            run(")", bold=True, size=17),
        ])
    )


def document_xml(args: argparse.Namespace, page_w: int, page_h: int, images: dict[str, tuple[str, float, float]]) -> str:
    body_w = page_w - 2 * MARGIN
    name_w = args.name_width or (7800 if args.portrait else 3900)
    loop_w = body_w - name_w

    headings = heading_paragraphs(args)
    if images:
        side = 2800
        left = para([image_run("rIdLogo", 1, *images["logo"][1:])], align="left") if "logo" in images else para([], align="left")
        right = para([image_run("rIdCorner", 2, *images["corner"][1:])], align="right") if "corner" in images else para([], align="right")
        header = (
            '<w:tbl><w:tblPr><w:tblW w:w="%d" w:type="dxa"/><w:tblLayout w:type="fixed"/>'
            '<w:tblBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/>'
            '<w:insideH w:val="nil"/><w:insideV w:val="nil"/></w:tblBorders></w:tblPr>'
            '<w:tblGrid><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/></w:tblGrid>'
            "<w:tr>%s%s%s</w:tr></w:tbl>"
        ) % (
            body_w, side, body_w - 2 * side, side,
            cell(side, left, valign="top", borders=False),
            cell(body_w - 2 * side, headings, valign="center", borders=False),
            cell(side, right, valign="top", borders=False),
        )
    else:
        header = headings

    name_heading = run(args.name_heading, bold=True, caps=args.caps) if args.name_heading else run(token("client_name_label"), bold=True, caps=args.caps)
    header_row = (
        '<w:tr><w:trPr><w:tblHeader/><w:trHeight w:val="600" w:hRule="atLeast"/></w:trPr>'
        + cell(name_w, para([name_heading]), valign="top")
        + cell(loop_w, para([run(token("#job_templates")), run(token("job_template_name_display"), bold=True), run(token("/job_templates"))]), valign="top")
        + "</w:tr>"
    )
    spacer_row = (
        '<w:tr><w:trPr><w:trHeight w:val="250" w:hRule="exact"/></w:trPr>'
        + cell(name_w, para([]))
        + cell(loop_w, para([run(token("#job_templates")), run(token("/job_templates"))]))
        + "</w:tr>"
    )
    bold = f'<w:b w:val="{token("row_bold")}"/>'
    if args.band_caps:
        label_runs = [run(token("row_band_label"), caps=True, extra=bold), run(token("row_client_label"), extra=bold)]
    else:
        label_runs = [run(token("row_label_display"), extra=bold)]
    marker = lambda text: f'<w:tr><w:tc><w:tcPr><w:tcW w:w="{name_w}" w:type="dxa"/></w:tcPr>{para([run(text)])}</w:tc></w:tr>'
    template_row = (
        '<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="250" w:hRule="exact"/></w:trPr>'
        + cell(name_w, para(label_runs, align="left"))
        + cell(loop_w, para([
            run(token("#cells")),
            run(token("cell_scaled_label"), extra="" if args.no_colour else f'<w:color w:val="{token("cell_text_hex")}"/>'),
            run(token("/cells")),
        ]), fill_token=None if args.no_colour else "cell_fill_hex")
        + "</w:tr>"
    )
    table = (
        f'<w:tbl><w:tblPr><w:tblW w:w="{body_w}" w:type="dxa"/><w:tblLayout w:type="fixed"/>'
        '<w:tblBorders><w:top w:val="single" w:sz="4" w:color="000000"/><w:left w:val="single" w:sz="4" w:color="000000"/>'
        '<w:bottom w:val="single" w:sz="4" w:color="000000"/><w:right w:val="single" w:sz="4" w:color="000000"/>'
        '<w:insideH w:val="single" w:sz="4" w:color="000000"/><w:insideV w:val="single" w:sz="4" w:color="000000"/></w:tblBorders>'
        '<w:tblCellMar><w:top w:w="0" w:type="dxa"/><w:left w:w="60" w:type="dxa"/><w:bottom w:w="0" w:type="dxa"/><w:right w:w="60" w:type="dxa"/></w:tblCellMar>'
        f'</w:tblPr><w:tblGrid><w:gridCol w:w="{name_w}"/><w:gridCol w:w="{loop_w}"/></w:tblGrid>'
        + header_row + spacer_row
        + marker(token("#rows")) + template_row + marker(token("/rows"))
        + "</w:tbl>"
    )
    orient = "" if args.portrait else ' w:orient="landscape"'
    namespaces = " ".join([W_NS, R_NS, WP_NS, A_NS, PIC_NS]) if images else W_NS
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        f"<w:document {namespaces}><w:body>"
        + header
        + para([], after=200)
        + table
        + f'<w:sectPr><w:pgSz w:w="{page_w}" w:h="{page_h}"{orient}/>'
        f'<w:pgMar w:top="{MARGIN}" w:right="{MARGIN}" w:bottom="{MARGIN}" w:left="{MARGIN}" w:header="0" w:footer="0" w:gutter="0"/>'
        "</w:sectPr></w:body></w:document>"
    )


STYLES = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial" w:cs="Arial"/><w:sz w:val="16"/></w:rPr></w:rPrDefault>
  <w:pPrDefault><w:pPr><w:spacing w:before="0" w:after="0"/></w:pPr></w:pPrDefault></w:docDefaults>
  <w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>"""

ROOT_RELS = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>"""


def content_types(has_png: bool) -> str:
    png = '  <Default Extension="png" ContentType="image/png"/>\n' if has_png else ""
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>\n'
        '<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">\n'
        '  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>\n'
        '  <Default Extension="xml" ContentType="application/xml"/>\n'
        + png
        + '  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>\n'
        '  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>\n'
        "</Types>"
    )


def document_rels(images: dict[str, tuple[str, float, float]]) -> str:
    rels = "".join(
        f'<Relationship Id="{rid}" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/{name}.png"/>'
        for name, rid in (("logo", "rIdLogo"), ("corner", "rIdCorner"))
        if name in images
    )
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        f'<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">{rels}</Relationships>'
    )


def assert_manifest(document: str, args: argparse.Namespace) -> None:
    """Every token must be in the group-matrix vocabulary (scope is checked by Go)."""
    manifest = json.loads(MANIFEST.read_text(encoding="utf-8"))
    allowed: set[str] = set()
    for name, scope in manifest["scopes"].items():
        allowed |= set(scope.get("text", [])) | set(scope.get("style", {}))
        if name != "root":
            allowed |= {f"#{name}", f"/{name}"}
    numbered = re.compile(r"^job_template([1-9][0-9]?)_(.+)$")
    extra = set()
    for found in TOKEN_RE.findall(document):
        match = numbered.match(found)
        if found in allowed or (match and int(match.group(1)) <= manifest["profile"]["max_columns"]):
            continue
        extra.add(found)
    if extra:
        raise SystemExit(f"manifest mismatch: unexpected {sorted(extra)}")


def write_part(archive: ZipFile, name: str, body: bytes) -> None:
    info = ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
    info.compress_type = ZIP_DEFLATED
    info.external_attr = 0o600 << 16
    archive.writestr(info, body)


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--output", type=Path, default=DEFAULT_OUTPUT)
    parser.add_argument("--title")
    parser.add_argument("--title-prefix")
    parser.add_argument("--caps", action="store_true")
    parser.add_argument("--band-caps", action="store_true")
    parser.add_argument("--schedule-prefix")
    parser.add_argument("--no-colour", action="store_true")
    parser.add_argument("--name-heading")
    parser.add_argument("--portrait", action="store_true")
    parser.add_argument("--name-width", type=int)
    parser.add_argument("--logo", type=Path)
    parser.add_argument("--corner", type=Path)
    args = parser.parse_args()

    page_w, page_h = (12240, 18720) if args.portrait else (18720, 12240)
    images: dict[str, tuple[str, float, float]] = {}
    media: dict[str, bytes] = {}
    for name, path, height_pt in (("logo", args.logo, 40.0), ("corner", args.corner, 44.0)):
        if path:
            w, h = image_size(path)
            images[name] = (name, height_pt * w / h, height_pt)
            media[f"word/media/{name}.png"] = path.read_bytes()

    document = document_xml(args, page_w, page_h, images)
    assert_manifest(document, args)
    parts: dict[str, bytes] = {
        "[Content_Types].xml": content_types(bool(images)).encode(),
        "_rels/.rels": ROOT_RELS.encode(),
        "word/_rels/document.xml.rels": document_rels(images).encode(),
        "word/document.xml": document.encode(),
        "word/styles.xml": STYLES.encode(),
        **media,
    }
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with ZipFile(args.output, "w") as archive:
        for name in sorted(parts):
            write_part(archive, name, parts[name])
    print(f"generated {args.output} ({args.output.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
