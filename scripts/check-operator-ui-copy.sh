#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

python3 <<'PY'
from pathlib import Path
import re
import sys
from collections import Counter

root = Path.cwd()
scan_roots = [root / 'apps/web-admin/src/routes']
paths = []
for scan_root in scan_roots:
    if scan_root.exists():
        paths.extend(scan_root.rglob('*.svelte'))

patterns = {
    'endpoint/API/base URL': r'\b(endpoint|Endpoint|api|API|base URL|Base URL)\b',
    'payload/JSON': r'\b(payload|Payload|JSON|json)\b',
    'token': r'\b(token|Token)\b',
    'debug': r'\b(debug|Debug)\b',
    'deploy/release/build/version': r'\b(deploy|Deploy|deployment|Deployment|release|Release|build|Build|version|Version)\b',
    'matrix': r'\b(matrix|Matrix|matriks|Matriks)\b',
    'bulk/selected': r'\b(bulk|Bulk|selected|Selected)\b',
    'role/RBAC/permission': r'\b(RBAC|role|Role|permission|Permission)\b',
    'session/participant': r'\b(session|Session|participant|Participant)\b',
    'proctoring': r'\b(proctoring|Proctoring)\b',
    'randomization/readiness': r'\b(randomization|Randomization|readiness|Readiness)\b',
    'package/event': r'\b(package|Package|event|Event)\b',
    'template': r'\b(template|Template)\b',
    'validation/duplicate/parser': r'\b(validation|Validation|validator|Validator|duplicate|Duplicate|parser|Parser)\b',
    'workflow/cycle': r'\b(workflow|Workflow|cycle|Cycle)\b',
    'queue/retry/failed': r'\b(queue|Queue|retry|Retry|failed|Failed)\b',
    'error': r'\b(error|Error)\b',
    'submit': r'\b(submit|Submit)\b',
    'slug/SEO': r'\b(slug|Slug|SEO)\b',
    'audit/analytics/compliance': r'\b(audit|Audit|analytics|Analytics|compliance|Compliance)\b',
    'import/export/CSV': r'\b(import|Import|export|Export|CSV|csv)\b',
    'developer ops': r'\b(backend|Backend|migration|Migration|seed|Seed|legacy|Legacy|lifecycle|Lifecycle|drawer|Drawer)\b',
}
compiled = {name: re.compile(pattern) for name, pattern in patterns.items()}
string_re = re.compile(r'''(?P<quote>["'`])(?P<body>(?:\\.|(?!\1).)*?)(?P=quote)''')

# Terms that are acceptable in visible text for this madrasah app.
allow_literal_contains = [
    'Impor', 'impor', 'Unduh', 'unduh', 'Ekspor', 'ekspor',
    'Pilih File', 'file', 'File',
]

def is_code_only(line: str, literal: str) -> bool:
    s = line.strip()
    if len(literal.strip()) < 3:
        return True
    if '${' in literal:
        return True
    if literal.startswith('#'):
        return True
    if s.startswith(('import ', 'export ', 'type ', 'interface ', 'const ', 'let ', 'function ')):
        return True
    if re.search(r'\b(from|as)\s+["\']', s):
        return True
    if re.search(r'(class|id|href|src|accept|value|name|aria-|data-|bind:|on:|let:|slot=|kind|variant|size)=', s):
        return True
    if re.search(r'fetch\(|goto\(|apiFetch\(|clientApiPath|resolve\(|new URL\(|console\.|JSON\.stringify|JSON\.parse|Blob\(', s):
        return True
    if literal.startswith(('/', './', '../', '$lib', 'http', 'GET ', 'POST ', 'PUT ', 'PATCH ', 'DELETE ', 'text/', 'application/')):
        return True
    if re.fullmatch(r'[a-zA-Z0-9_./:;?&=%-]+', literal) and ' ' not in literal:
        return True
    return False

findings = []
code_only = 0
for path in sorted(paths):
    text = path.read_text(errors='ignore')
    for line_no, line in enumerate(text.splitlines(), 1):
        for match in string_re.finditer(line):
            literal = match.group('body')
            hits = [name for name, rx in compiled.items() if rx.search(literal)]
            if not hits:
                continue
            if any(allowed in literal for allowed in allow_literal_contains):
                code_only += 1
                continue
            if is_code_only(line, literal):
                code_only += 1
                continue
            rel = str(path.relative_to(root))
            findings.append((rel, line_no, ', '.join(hits), literal.strip().replace('\n', ' ')[:220]))

if findings:
    print('FAIL: ditemukan istilah teknis user-facing pada UI operator:')
    for rel, line_no, hits, literal in findings[:200]:
        print(f'- {rel}:{line_no} [{hits}] {literal}')
    if len(findings) > 200:
        print(f'... dan {len(findings) - 200} temuan lain')
    sys.exit(1)

print(f'PASS: tidak ada istilah teknis user-facing pada UI operator. Code-only diabaikan: {code_only}.')
PY
