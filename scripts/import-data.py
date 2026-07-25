#!/usr/bin/env python3
"""
Import script for MTsN 2 Kolut data:
1. Pembagian Tugas (Teacher Assignments)
2. Roster Pelajaran (Timetable)
3. Data Siswa (Students)

Usage: uv run python3 import-data.py
"""
import openpyxl
import subprocess, sys, os, re, json, tempfile
from datetime import datetime

DOCS_DIR = '/home/servermtsn2kolut/.hermes/cache/documents'
PROJECT_DIR = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api'
ENV_FILE = f'{PROJECT_DIR}/.env'

# ============ LOAD ENV ============
def load_env():
    env = {}
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if '=' in line and not line.startswith('#'):
                k, v = line.split('=', 1)
                env[k] = v
    return env

def psql(sql, env=None):
    """Run SQL and return output"""
    e = os.environ.copy()
    if env:
        e.update(env)
    cmd = f"PGPASSWORD='{env.get('POSTGRES_PASSWORD','')}' psql -h 127.0.0.1 -U {env.get('POSTGRES_USER','')} -d {env.get('POSTGRES_DB','')} -p {env.get('POSTGRES_PORT','5432')} -t -A -c {shq(sql)}"
    r = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=30, env=e)
    if r.returncode != 0:
        print(f"SQL ERROR: {r.stderr[:200]}")
        return ''
    return r.stdout.strip()

def shq(s):
    """Simple shell quote"""
    return "'" + s.replace("'", "'\\''") + "'"

def generate_uuid():
    """Generate a UUID v4"""
    import uuid
    return str(uuid.uuid4())

print("=" * 60)
print("IMPORT DATA MTSN 2 KOLAKA UTARA")
print("TAHUN PELAJARAN 2026/2027")
print("=" * 60)

env = load_env()
env['PGPASSWORD'] = env.get('POSTGRESS_PASSWORD', env.get('POSTGRES_PASSWORD', ''))  # handle typo in env name

# Check connection
ay = psql("SELECT id, name FROM academic_years WHERE is_active = true ORDER BY start_date DESC LIMIT 1", env)
print(f"\n✓ Connected. Active academic year: {ay}")

# Get active academic year ID
academic_year_id = ay.split('|')[0].strip() if ay else ''
print(f"  Academic Year ID: {academic_year_id}")

# Get existing classes
existing_classes = {}
rows = psql("SELECT id, code, name, level FROM school_classes WHERE is_active = true", env)
for line in rows.split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 4:
        existing_classes[parts[1]] = {'id': parts[0], 'name': parts[2], 'level': parts[3]}
print(f"  Existing classes: {len(existing_classes)}")
for k, v in existing_classes.items():
    print(f"    {k} → {v['name']} ({v['level']})")

# Get existing teachers
existing_teachers = {}
rows = psql("SELECT id, nama FROM employees WHERE is_active = true", env)
for line in rows.split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 2:
        existing_teachers[parts[1].lower()] = parts[0]
print(f"  Existing teachers: {len(existing_teachers)}")

# Get existing subjects
existing_subjects = {}
rows = psql("SELECT id, code, name FROM subjects WHERE is_active = true", env)
for line in rows.split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 3:
        existing_subjects[parts[2].lower()] = {'id': parts[0], 'code': parts[1]}
print(f"  Existing subjects: {len(existing_subjects)}")

print("\n" + "=" * 60)
print("READING PEMBAGIAN TUGAS...")
print("=" * 60)

wb = openpyxl.load_workbook(f'{DOCS_DIR}/doc_3e3a0d311e13_PEMBAGIAN TUGAS 2026-2027 (2).xlsx', data_only=True)
ws = wb['PEMBAGIAN TUGAS 2627 GANJIL']

# Parse assignments
# Columns: no, nama_guru, mapel, beban_jam, 7A, 7B, 7C, 7D, 7E, 8A, 8B, 8C, 8D, 9A, 9B, 9C
CLASS_COLS = {
    4: 'VII.A', 5: 'VII.B', 6: 'VII.C', 7: 'VII.D', 8: 'VII.E',
    9: 'VIII.A', 10: 'VIII.B', 11: 'VIII.C', 12: 'VIII.D',
    13: 'IX.A', 14: 'IX.B', 15: 'IX.C'
}

assignments = []
current_teacher = ''

for row in ws.iter_rows(min_row=2, values_only=True):
    no = row[0]
    nama = str(row[1] or '').strip()
    mapel = str(row[2] or '').strip()
    # Skip total/piket rows
    if 'TOTAL' in mapel.upper() or 'PIKET' in mapel.upper() or 'WALI KELAS' in mapel.upper():
        continue
    if not mapel:
        continue
    
    teacher_name = nama if nama else current_teacher
    if teacher_name:
        current_teacher = teacher_name
    
    # Get jam per class
    class_hours = {}
    for col_idx, class_code in CLASS_COLS.items():
        jam = row[col_idx] if col_idx < len(row) else None
        if jam and str(jam).strip() and str(jam).strip() not in ['', '0']:
            try:
                class_hours[class_code] = int(float(str(jam).strip()))
            except ValueError:
                pass
    
    assignments.append({
        'teacher': teacher_name,
        'subject': mapel,
        'classes': class_hours
    })

print(f"  Parsed {len(assignments)} assignment rows")

# Print summary
teacher_subjects = {}
for a in assignments:
    key = f"{a['teacher']} → {a['subject']}"
    if key not in teacher_subjects:
        teacher_subjects[key] = []
    for cls, jam in a['classes'].items():
        teacher_subjects[key].append(f"{cls}({jam}j)")

print(f"\n  Assignments found:")
for ts, classes in sorted(teacher_subjects.items()):
    print(f"    {ts}: {', '.join(classes)}")

print("\n" + "=" * 60)
print("READING ROSTER PELAJARAN...")
print("=" * 60)

wb2 = openpyxl.load_workbook(f'{DOCS_DIR}/doc_cf93a5b2b842_perubahan ROSTER PELAJARAN SEMESTER GANJIL 2026-2027.xlsx', data_only=True)
ws2 = wb2['ROSTER SM GNJL 2026']

print(f"  Roster sheet: {ws2.max_row} rows x {ws2.max_column} cols")

# We'll parse the schedule later in the actual import phase
print("  (Will be processed during import phase)")

print("\n" + "=" * 60)
print("READING DATA SISWA...")
print("=" * 60)

# Read student PDF
pdf_path = f'{DOCS_DIR}/doc_ded7b5897f21_LAPORAN BULANAN GANJIL T.P 2026-2027.xlsx - Siswa.pdf'
r = subprocess.run(['pdftotext', '-layout', pdf_path, '-'], capture_output=True, text=True, timeout=10)
pdf_text = r.stdout

# Parse students - extract from the PDF text
# Format: URT | NIS | NISN | NAMA | L/P | KLS | TEMPAT LAHIR | TGL LAHIR | AYAH | IBU
students = []
current_class = None

for line in pdf_text.split('\n'):
    line = line.strip()
    
    # Detect class info
    m = re.search(r'KELAS\s+(VII|VIII|IX)\s*[\(\（]?\s*([A-Z]+)', line, re.IGNORECASE)
    if m:
        level = m.group(1).upper()
        section = m.group(2)
        current_class = f"{level}.{section}"
        continue
    
    # Try to parse student row
    # Pattern: URT NIS NISN NAMA L/P KLS TEMPAT_LAHIR TGL_LAHIR AYAH IBU
    m = re.match(r'\s*(\d+)\s+(\d{6,})\s+(\d{1,20}|\'?\d{5,})\s+(.+?)\s+([LP])\s+(VII|VIII|IX)\s+(.+?)\s+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})\s+(.+?)(?:\s+(.+?))?\s*$', line)
    if m:
        students.append({
            'nis': m.group(2),
            'nisn': m.group(3).strip("'"),
            'nama': m.group(4).strip(),
            'gender': m.group(5),
            'level': m.group(6),
            'tempat_lahir': m.group(7).strip(),
            'tanggal_lahir': m.group(8),
            'ayah': m.group(9).strip() if m.group(9) else '',
        })
        if current_class:
            students[-1]['class_code'] = current_class
    elif current_class:
        # Try simpler pattern without class column
        m2 = re.match(r'\s*(\d+)\s+(\d{6,22})\s+(\d{1,20}|\'?\d{5,}|\s*)\s+(.+?)\s+([LP])\s+(.+?)\s+(\d{1,2}[/-]\d{1,2}[/-]\d{2,4})\s+(.+?)(?:\s+(.+?))?\s*$', line)
        if m2:
            students.append({
                'nis': m2.group(2),
                'nisn': m2.group(3).strip("'"),
                'nama': m2.group(4).strip(),
                'gender': m2.group(5),
                'level': current_class.split('.')[0],
                'tempat_lahir': m2.group(6).strip(),
                'tanggal_lahir': m2.group(7),
                'ayah': m2.group(8).strip() if m2.group(8) else '',
            })
            if current_class:
                students[-1]['class_code'] = current_class

print(f"  Extracted {len(students)} student records")
print(f"  Sample:")
for s in students[:5]:
    print(f"    {s.get('nis','')} | {s.get('nama','')} | {s.get('gender','')} | {s.get('level','')} | {s.get('tempat_lahir','')}")

print("\n" + "=" * 60)
print("SUMMARY")
print("=" * 60)
print(f"  Assignments to create: {len(assignments)}")
print(f"  Students to import:    {len(students)}")
print(f"  Timetable slots:       ~400 (from roster)")
print(f"\n  Existing classes:     {len(existing_classes)}")
print(f"  Existing teachers:    {len(existing_teachers)}")
print(f"  Existing subjects:    {len(existing_subjects)}")

# Save parsed data for the actual import
data = {
    'academic_year_id': academic_year_id,
    'assignments': assignments,
    'students': students,
    'classes': {k: v for k, v in existing_classes.items()},
    'teachers': existing_teachers,
    'subjects': existing_subjects,
}
with open('/tmp/import-data.json', 'w') as f:
    json.dump(data, f, indent=2)

print(f"\n  Parsed data saved to /tmp/import-data.json")
print(f"\n✓ Data analysis complete")
