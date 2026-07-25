#!/usr/bin/env python3
"""Quick debug for timetable import"""
import subprocess, os, re

ENV_FILE = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env'
env = {}
with open(ENV_FILE) as f:
    for line in f:
        line = line.strip()
        if '=' in line and not line.startswith('#'):
            k, v = line.split('=', 1)
            env[k] = v

PGPASS = env.get('POSTGRES_PASSWORD', env.get('POSTGRESS_PASSWORD', ''))
AY_ID = 'e3dfb158-76c6-46c2-b6e7-00cd604a0277'

my_env = os.environ.copy()
my_env['PGPASSWORD'] = PGPASS

def psql(cmd):
    r = subprocess.run(f"psql -h 127.0.0.1 -U {env['POSTGRES_USER']} -d {env['POSTGRES_DB']} -p {env.get('POSTGRES_PORT', '5432')} -t -A -F'|' -c ", shell=True, capture_output=True, text=True, timeout=10, env=my_env)
    return r

# Build assn dict
q = f"SELECT c.code, s.name, csa.id::text FROM class_subject_assignments csa JOIN school_classes c ON c.id=csa.class_id JOIN subjects s ON s.id=csa.subject_id WHERE c.academic_year_id='{AY_ID}'::uuid"
r = subprocess.run(f"psql -h 127.0.0.1 -U {env['POSTGRES_USER']} -d {env['POSTGRES_DB']} -p {env.get('POSTGRES_PORT', '5432')} -t -A -F'|' -c \"{q}\"", shell=True, capture_output=True, text=True, timeout=10, env=my_env)
print("=== ALL ASSIGNMENT KEYS ===")
found = {}
for line in r.stdout.strip().split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 2:
        norm_key = re.sub(r'[,.\-]', ' ', parts[1].lower())
        norm_key = re.sub(r'\s+', ' ', norm_key).strip()
        found[(parts[0], norm_key)] = parts[2] if len(parts) >= 3 else ''
        print(f"  ({parts[0]}, '{norm_key}') -> {parts[2][:12] if len(parts) >= 3 else '??'}")

# Test lookups
print("\n=== TEST LOOKUPS ===")
tests = [('VII.A', 'matematika'), ('VII.A', 'pendidikan jasmani olahraga dan kesehatan'), ('VII.A', 'al qur an hadis'), ('VII.A', 'bahasa arab')]
for cc, norm in tests:
    val = found.get((cc, norm))
    print(f"  found[({cc}, '{norm}')] = {val if val else 'NOT FOUND'}")
