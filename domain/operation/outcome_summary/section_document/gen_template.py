#!/usr/bin/env python3
"""Generate the canonical eleven-slot subscription-group outcome DOCX.

The output is an operator authoring asset, not an embedded runtime fallback.
Only Python's standard library is required; ZIP metadata is fixed so repeated
runs are byte-deterministic.
"""

from __future__ import annotations

import json
import re
from pathlib import Path
from zipfile import ZIP_DEFLATED, ZipFile, ZipInfo


HERE = Path(__file__).resolve().parent
PROFILE = "subscription-group-outcome-matrix-single-period-11-v1"
MANIFEST = HERE / f"{PROFILE}.manifest.json"
OUTPUT = HERE / f"{PROFILE}.docx"
TOKEN_RE = re.compile(r"\{\{([^{}]+)\}\}")


def paragraph(token: str, *, bold: bool = False, size: int = 18) -> str:
    bold_xml = "<w:b/>" if bold else ""
    return (
        '<w:p><w:pPr><w:jc w:val="center"/></w:pPr><w:r><w:rPr>'
        f'<w:rFonts w:ascii="Arial" w:hAnsi="Arial"/>{bold_xml}<w:sz w:val="{size}"/>'
        f'</w:rPr><w:t>{{{{{token}}}}}</w:t></w:r></w:p>'
    )


def cell(text_token: str, width: int, *, header: bool = False, slot: int | None = None) -> str:
    shade = ""
    color = '<w:color w:val="000000"/>'
    bold = "<w:b/>" if header else ""
    if slot is not None:
        shade = f'<w:shd w:val="clear" w:color="auto" w:fill="{{{{job_template{slot}_fill_hex}}}}"/>'
        color = f'<w:color w:val="{{{{job_template{slot}_text_hex}}}}"/>'
    return (
        f'<w:tc><w:tcPr><w:tcW w:w="{width}" w:type="dxa"/>{shade}'
        '<w:vAlign w:val="center"/></w:tcPr><w:p><w:pPr><w:jc w:val="center"/></w:pPr>'
        '<w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/>'
        f'{bold}{color}<w:sz w:val="14"/></w:rPr><w:t>{{{{{text_token}}}}}</w:t></w:r></w:p></w:tc>'
    )


def document_xml() -> str:
    header_cells = [cell("client_name_label", 3270, header=True)]
    for index in range(1, 12):
        header_cells.append(cell(f"job_template{index}_name_display", 1322, header=True))

    row_cells = [
        '<w:tc><w:tcPr><w:tcW w:w="3270" w:type="dxa"/><w:vAlign w:val="center"/></w:tcPr>'
        '<w:p><w:r><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/>'
        '<w:b w:val="{{row_bold}}"/><w:sz w:val="16"/></w:rPr>'
        '<w:t>{{row_label_display}}</w:t></w:r></w:p></w:tc>'
    ]
    for index in range(1, 12):
        row_cells.append(cell(f"job_template{index}_scaled_label", 1322, slot=index))

    # Word applies table borders from tblBorders; the cell helper stays compact.
    table_properties = (
        '<w:tblPr><w:tblW w:w="17812" w:type="dxa"/><w:tblLayout w:type="fixed"/>'
        '<w:tblBorders><w:top w:val="single" w:sz="4" w:color="000000"/>'
        '<w:left w:val="single" w:sz="4" w:color="000000"/>'
        '<w:bottom w:val="single" w:sz="4" w:color="000000"/>'
        '<w:right w:val="single" w:sz="4" w:color="000000"/>'
        '<w:insideH w:val="single" w:sz="4" w:color="000000"/>'
        '<w:insideV w:val="single" w:sz="4" w:color="000000"/></w:tblBorders></w:tblPr>'
    )
    return (
        '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
        '<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">'
        '<w:body>'
        + paragraph("sheet_title", bold=True, size=24)
        + paragraph("subscription_group_name_display", bold=True, size=20)
        + paragraph("price_schedule_name_display", size=18)
        + paragraph("job_template_phase_name_display", size=18)
        + '<w:p><w:pPr><w:spacing w:after="80"/></w:pPr></w:p>'
        + '<w:tbl>'
        + table_properties
        + '<w:tblGrid><w:gridCol w:w="3270"/>'
        + ''.join('<w:gridCol w:w="1322"/>' for _ in range(11))
        + '</w:tblGrid>'
        + '<w:tr><w:trPr><w:tblHeader/><w:trHeight w:val="680" w:hRule="atLeast"/></w:trPr>'
        + ''.join(header_cells)
        + '</w:tr>'
        + '<w:tr><w:tc><w:p><w:r><w:t>{{#rows}}</w:t></w:r></w:p></w:tc></w:tr>'
        + '<w:tr><w:trPr><w:cantSplit/><w:trHeight w:val="227" w:hRule="atLeast"/></w:trPr>'
        + ''.join(row_cells)
        + '</w:tr>'
        + '<w:tr><w:tc><w:p><w:r><w:t>{{/rows}}</w:t></w:r></w:p></w:tc></w:tr>'
        + '</w:tbl>'
        + '<w:sectPr><w:pgSz w:w="18720" w:h="12240" w:orient="landscape"/>'
        '<w:pgMar w:top="454" w:right="454" w:bottom="454" w:left="454" w:header="0" w:footer="0" w:gutter="0"/>'
        '<w:cols w:space="720"/><w:docGrid w:linePitch="360"/></w:sectPr>'
        '</w:body></w:document>'
    )


CONTENT_TYPES = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>"""

ROOT_RELS = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>"""

DOCUMENT_RELS = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>"""

STYLES = """<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:docDefaults><w:rPrDefault><w:rPr><w:rFonts w:ascii="Arial" w:hAnsi="Arial"/><w:sz w:val="16"/></w:rPr></w:rPrDefault></w:docDefaults>
  <w:style w:type="paragraph" w:default="1" w:styleId="Normal"><w:name w:val="Normal"/></w:style>
</w:styles>"""


def assert_manifest(document: str) -> None:
    manifest = json.loads(MANIFEST.read_text(encoding="utf-8"))
    expected = set(manifest["scalars"])
    loop_name = manifest["profile"]["row_loop"]
    expected.update(manifest["loops"][loop_name]["scalars"])
    expected.update({f"#{loop_name}", f"/{loop_name}"})
    found = TOKEN_RE.findall(document)
    if set(found) != expected or len(found) != len(expected):
        missing = sorted(expected - set(found))
        extra = sorted(set(found) - expected)
        raise SystemExit(f"manifest mismatch: missing={missing} extra={extra} duplicates={len(found) - len(set(found))}")


def write_part(archive: ZipFile, name: str, body: str) -> None:
    info = ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
    info.compress_type = ZIP_DEFLATED
    info.external_attr = 0o600 << 16
    archive.writestr(info, body.encode("utf-8"))


def main() -> None:
    document = document_xml()
    assert_manifest(document)
    parts = {
        "[Content_Types].xml": CONTENT_TYPES,
        "_rels/.rels": ROOT_RELS,
        "word/_rels/document.xml.rels": DOCUMENT_RELS,
        "word/document.xml": document,
        "word/styles.xml": STYLES,
    }
    with ZipFile(OUTPUT, "w") as archive:
        for name in sorted(parts):
            write_part(archive, name, parts[name])
    print(f"generated {OUTPUT.name} ({OUTPUT.stat().st_size} bytes)")


if __name__ == "__main__":
    main()
