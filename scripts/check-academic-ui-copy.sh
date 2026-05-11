#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

python3 - <<'PY'
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path.cwd()
TARGETS = [
    ROOT / "apps/web-admin/src/routes/akademik",
    ROOT / "apps/web-admin/src/routes/students/+page.svelte",
]

# Istilah yang tidak boleh tampil sebagai teks langsung untuk operator madrasah.
PATTERNS = [
    r"\bsprint\b",
    r"\bendpoint\b",
    r"\bapi\b",
    r"\bpayload\b",
    r"\bjson\b",
    r"\bcommit\b",
    r"\bdeploy\b",
    r"\bmigration\b",
    r"\bseed\b",
    r"\bdebug\b",
    r"\bdry[- ]run\b",
    r"\bapply rollover\b",
    r"\bpreview rollover\b",
    r"\brollover\b",
    r"\bmatrix\b",
    r"\bbackend\b",
    r"\btoken\b",
    r"\bchallenge\b",
    r"\bdrawer\b",
    r"\blifecycle\b",
    r"\blegacy\b",
    r"\bslot\b",
    r"\bassignment\b",
]
TECHNICAL_RE = re.compile("|".join(f"(?:{p})" for p in PATTERNS), re.IGNORECASE)

STRING_RE = re.compile(r"(['\"`])((?:\\.|(?!\1).)*?)\1", re.DOTALL)
ATTRIBUTE_OR_CODE_HINTS = (
    "class=",
    "id=",
    "href=",
    "aria-",
    "data-",
    "bind:",
    "on:",
    "{#",
    "{/",
    "{@",
    "import ",
    "from ",
    "let ",
    "const ",
    "type ",
    "interface ",
    "function ",
    "=>",
    "json(",
    "fetch(",
    "goto(",
    "localStorage",
    "console.",
    "throw ",
    "new ",
    "$:",
)


def iter_files() -> list[Path]:
    files: list[Path] = []
    for target in TARGETS:
        if target.is_dir():
            files.extend(target.rglob("*.svelte"))
            files.extend(target.rglob("*.ts"))
        elif target.exists():
            files.append(target)
    return sorted(set(files))


def looks_code_only(line: str, literal: str) -> bool:
    stripped = line.strip()
    lower = stripped.lower()
    # Route paths, object keys, endpoint internals, and internal identifiers are allowed.
    if any(hint.lower() in lower for hint in ATTRIBUTE_OR_CODE_HINTS):
        return True
    if re.search(r"[A-Za-z0-9_]+\s*:\s*['\"`]", stripped):
        return True
    if re.search(r"[A-Za-z0-9_]+\s*=\s*['\"`]", stripped):
        return True
    if "${" in literal:
        return True
    if literal.startswith("/") or literal.startswith("?"):
        return True
    if re.fullmatch(r"[a-z0-9_./:-]+", literal):
        return True
    return False


violations: list[str] = []
code_only = 0
for path in iter_files():
    rel = path.relative_to(ROOT)
    text = path.read_text(encoding="utf-8")
    for line_no, line in enumerate(text.splitlines(), 1):
        if not TECHNICAL_RE.search(line):
            continue
        for match in STRING_RE.finditer(line):
            literal = match.group(2)
            if not TECHNICAL_RE.search(literal):
                continue
            if looks_code_only(line, literal):
                code_only += 1
                continue
            violations.append(f"{rel}:{line_no}: {literal}")

if violations:
    print("Istilah teknis masih tampil pada copy Akademik:", file=sys.stderr)
    for item in violations:
        print(f"- {item}", file=sys.stderr)
    print("\nGunakan bahasa operator madrasah seperti layanan sistem, data yang dikirim, pratinjau, cek data sebelum impor, tabel penugasan, jam pelajaran, atau kalimat konfirmasi.", file=sys.stderr)
    sys.exit(1)

print(f"PASS: tidak ada istilah teknis user-facing pada UI Akademik. Code-only diabaikan: {code_only}.")
PY
