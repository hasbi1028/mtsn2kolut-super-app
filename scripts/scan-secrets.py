#!/usr/bin/env python3
"""Lightweight repository secret scanner for CI/pre-commit use.

Design goals:
- scan only git-tracked files plus staged additions, never ignored runtime .env files;
- avoid printing secret values; report file/line and rule name only;
- keep dependencies to Python stdlib so CI can run it before app installs.
"""
from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MAX_FILE_BYTES = 1_000_000
SKIP_DIR_PARTS = {
    ".git",
    "node_modules",
    ".svelte-kit",
    "build",
    "dist",
    "coverage",
    ".dart_tool",
    ".gradle",
    "logs",
    "tmp",
}
SKIP_SUFFIXES = {
    ".png",
    ".jpg",
    ".jpeg",
    ".gif",
    ".webp",
    ".ico",
    ".pdf",
    ".zip",
    ".gz",
    ".dump",
    ".bin",
    ".dill",
    ".lock",
}
SKIP_NAMES = {
    "package-lock.json",
    "bun.lock",
}
ALLOWLIST_PATTERNS = [
    re.compile(r"example", re.IGNORECASE),
    re.compile(r"placeholder", re.IGNORECASE),
    re.compile(r"change[_-]?me", re.IGNORECASE),
    re.compile(r"dummy", re.IGNORECASE),
    re.compile(r"test[_-]?secret", re.IGNORECASE),
    re.compile(r"fake", re.IGNORECASE),
    re.compile(r"ganti", re.IGNORECASE),
    re.compile(r"isi-", re.IGNORECASE),
    re.compile(r"localhost", re.IGNORECASE),
    re.compile(r"should-not-leak", re.IGNORECASE),
    re.compile(r"super-secret-password", re.IGNORECASE),
    re.compile(r"\[REDACTED[^\]]*\]"),
    re.compile(r"\$\{?[A-Z0-9_]+\}?"),
]


@dataclass(frozen=True)
class Rule:
    name: str
    pattern: re.Pattern[str]


RULES = [
    Rule(
        "assignment-secret",
        re.compile(
            r"^\s*(?:export\s+)?(?:[A-Z0-9_]*(?:API_KEY|SECRET|PASSWORD|PASSWD|TOKEN|JWT_SECRET|WORKER_API_KEY|DATABASE_URL)[A-Z0-9_]*)\s*=\s*['\"]?([^'\"\s]{12,})"
        ),
    ),
    Rule("postgres-url", re.compile(r"postgres(?:ql)?://[^\s'\"]+:[^\s'\"]+@", re.IGNORECASE)),
    Rule("bearer-token", re.compile(r"Bearer\s+[A-Za-z0-9._~+/=-]{20,}", re.IGNORECASE)),
    Rule("private-key", re.compile(r"-----BEGIN (?:RSA |EC |OPENSSH |PGP )?PRIVATE KEY-----")),
]


def git(args: list[str]) -> str:
    return subprocess.check_output(["git", *args], cwd=ROOT, text=True, stderr=subprocess.DEVNULL)


def tracked_files() -> set[Path]:
    files = {ROOT / line for line in git(["ls-files"]).splitlines() if line.strip()}
    # Include staged new files so pre-commit catches secrets before they become tracked.
    for line in git(["diff", "--cached", "--name-only", "--diff-filter=ACMR"]).splitlines():
        if line.strip():
            files.add(ROOT / line)
    return files


def should_skip(path: Path) -> bool:
    rel = path.relative_to(ROOT)
    if path.name in SKIP_NAMES or path.suffix.lower() in SKIP_SUFFIXES:
        return True
    if any(part in SKIP_DIR_PARTS for part in rel.parts):
        return True
    # Runtime env files must stay ignored/untracked. If someone tracks one, flag by path below.
    return False


def allowlisted(line: str) -> bool:
    return any(p.search(line) for p in ALLOWLIST_PATTERNS)


def scan_file(path: Path) -> list[tuple[str, int]]:
    findings: list[tuple[str, int]] = []
    rel = path.relative_to(ROOT).as_posix()
    if ".env" in path.name and not path.name.endswith(".example"):
        findings.append(("tracked-env-file", 1))
        return findings
    if should_skip(path) or not path.exists() or not path.is_file():
        return findings
    try:
        if path.stat().st_size > MAX_FILE_BYTES:
            return findings
        text = path.read_text(encoding="utf-8", errors="ignore")
    except OSError:
        return findings
    for lineno, line in enumerate(text.splitlines(), start=1):
        if allowlisted(line):
            continue
        for rule in RULES:
            if rule.pattern.search(line):
                findings.append((rule.name, lineno))
                break
    return findings


def main() -> int:
    parser = argparse.ArgumentParser(description="Scan git-tracked/staged files for likely secrets.")
    parser.add_argument("--staged", action="store_true", help="scan only staged paths")
    args = parser.parse_args()

    if args.staged:
        paths = {ROOT / line for line in git(["diff", "--cached", "--name-only", "--diff-filter=ACMR"]).splitlines() if line.strip()}
    else:
        paths = tracked_files()

    all_findings: list[tuple[str, str, int]] = []
    for path in sorted(paths):
        for rule, lineno in scan_file(path):
            all_findings.append((path.relative_to(ROOT).as_posix(), rule, lineno))

    if all_findings:
        print("Potential secrets found. Values are intentionally redacted:", file=sys.stderr)
        for rel, rule, lineno in all_findings:
            print(f"- {rel}:{lineno} [{rule}]", file=sys.stderr)
        print("Review each finding. If this is a placeholder, rewrite it to include example/placeholder/[REDACTED].", file=sys.stderr)
        return 1

    print("Secret scan passed: no likely secrets in git-tracked/staged text files.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
