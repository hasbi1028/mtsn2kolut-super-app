#!/usr/bin/env python3
"""Fill employees.tanggal_lahir from ASN NIP birthdate prefix.

Safe defaults:
- --dry-run only previews candidates.
- --apply updates only rows where tanggal_lahir IS NULL.
- NIP must contain exactly 18 digits.
- First 8 digits must parse as YYYYMMDD.
- Existing tanggal_lahir is never overwritten.
"""
from __future__ import annotations

import argparse
import csv
import datetime as dt
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any
from urllib.parse import urlparse

REPO = Path(__file__).resolve().parents[2]
ENV_FILE = REPO / "services" / "core-api" / ".env"
BACKUP_DIR = REPO / "tmp" / "employee-birthdate-backups"


def load_env_file(path: Path) -> None:
    if not path.exists():
        return
    for raw in path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        if key and key not in os.environ:
            os.environ[key] = value


def db_env() -> dict[str, str]:
    load_env_file(ENV_FILE)
    env = os.environ.copy()
    if env.get("DATABASE_URL"):
        return env
    required = ["DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"]
    if all(env.get(k) for k in required):
        env["PGHOST"] = env["DB_HOST"]
        env["PGPORT"] = env["DB_PORT"]
        env["PGUSER"] = env["DB_USER"]
        env["PGPASSWORD"] = env["DB_PASSWORD"]
        env["PGDATABASE"] = env["DB_NAME"]
        return env
    raise SystemExit("DATABASE_URL or DB_* environment variables are missing")


def psql_base(env: dict[str, str]) -> list[str]:
    if env.get("DATABASE_URL"):
        return ["psql", env["DATABASE_URL"]]
    return ["psql"]


def psql_csv(sql: str, env: dict[str, str]) -> list[dict[str, str]]:
    cmd = psql_base(env) + ["-v", "ON_ERROR_STOP=1", "-P", "footer=off", "--csv", "-c", sql]
    out = subprocess.check_output(cmd, env=env, text=True)
    return list(csv.DictReader(out.splitlines()))


def psql_exec(sql: str, env: dict[str, str]) -> str:
    cmd = psql_base(env) + ["-v", "ON_ERROR_STOP=1", "-At", "-c", sql]
    return subprocess.check_output(cmd, env=env, text=True).strip()


def parse_birthdate(nip: str) -> tuple[str | None, str]:
    digits = re.sub(r"\D", "", nip or "")
    if not digits:
        return None, "nip_empty"
    if len(digits) != 18:
        return None, f"nip_not_18_digits:{len(digits)}"
    raw = digits[:8]
    try:
        parsed = dt.datetime.strptime(raw, "%Y%m%d").date()
    except ValueError:
        return None, "invalid_birthdate_prefix"
    if parsed < dt.date(1940, 1, 1) or parsed > dt.date.today():
        return None, "birthdate_out_of_range"
    return parsed.isoformat(), "valid"


def quote_sql(value: str) -> str:
    return "'" + value.replace("'", "''") + "'"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--apply", action="store_true", help="apply updates; default is dry-run")
    parser.add_argument("--backup-dir", default=str(BACKUP_DIR))
    args = parser.parse_args()

    env = db_env()
    backup_dir = Path(args.backup_dir)
    backup_dir.mkdir(parents=True, exist_ok=True)
    stamp = dt.datetime.now().strftime("%Y%m%d-%H%M%S")
    backup_path = backup_dir / f"employees-birthdates-before-{stamp}.csv"

    rows = psql_csv(
        "SELECT id::text, nama, nip, tanggal_lahir::text AS tanggal_lahir "
        "FROM employees ORDER BY lower(nama), id",
        env,
    )
    with backup_path.open("w", newline="") as fh:
        writer = csv.DictWriter(fh, fieldnames=["id", "nama", "nip", "tanggal_lahir"])
        writer.writeheader()
        writer.writerows(rows)

    candidates: list[dict[str, Any]] = []
    skipped: list[dict[str, Any]] = []
    for row in rows:
        birthdate, reason = parse_birthdate(row.get("nip", ""))
        entry = {
            "id": row["id"],
            "nama": row["nama"],
            "nip": row["nip"],
            "existing_tanggal_lahir": row.get("tanggal_lahir") or "",
            "parsed_tanggal_lahir": birthdate or "",
            "reason": reason,
        }
        if row.get("tanggal_lahir"):
            entry["reason"] = "already_has_birthdate"
            skipped.append(entry)
        elif birthdate:
            candidates.append(entry)
        else:
            skipped.append(entry)

    print(f"mode={'apply' if args.apply else 'dry-run'}")
    print(f"backup={backup_path}")
    print(f"employees_total={len(rows)}")
    print(f"update_candidates={len(candidates)}")
    print(f"skipped={len(skipped)}")
    print("candidates_sample:")
    for item in candidates[:10]:
        print(f"- {item['nama']} | {item['nip']} -> {item['parsed_tanggal_lahir']}")
    reasons: dict[str, int] = {}
    for item in skipped:
        reasons[item["reason"]] = reasons.get(item["reason"], 0) + 1
    print("skip_reasons:")
    for reason, count in sorted(reasons.items()):
        print(f"- {reason}: {count}")

    if not args.apply:
        print("dry_run_only=true")
        return 0

    updated = 0
    for item in candidates:
        sql = (
            "UPDATE employees SET tanggal_lahir = "
            + quote_sql(item["parsed_tanggal_lahir"])
            + "::date WHERE id = "
            + quote_sql(item["id"])
            + "::uuid AND tanggal_lahir IS NULL"
        )
        result = psql_exec(sql, env)
        if result == "UPDATE 1":
            updated += 1
    print(f"updated={updated}")
    totals = psql_csv(
        "SELECT count(*)::int AS total, "
        "count(*) FILTER (WHERE tanggal_lahir IS NOT NULL)::int AS filled, "
        "count(*) FILTER (WHERE tanggal_lahir IS NULL)::int AS missing "
        "FROM employees",
        env,
    )[0]
    print(f"after_total={totals['total']} after_filled={totals['filled']} after_missing={totals['missing']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
