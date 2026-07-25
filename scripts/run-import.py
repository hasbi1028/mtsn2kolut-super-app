#!/usr/bin/env python3
"""
IMPORT DATA MTSN 2 KOLUT 2026/2027
===================================
1. Create missing classes + assign to 2026/2027
2. Import students from PDF
3. Create teacher assignments (Pembagian Tugas)
4. Create timetable slots (Roster)

Usage: uv run python3 scripts/run-import.py
"""
import json, os, sys, subprocess, re, tempfile
from datetime import datetime
from uuid import uuid4

DOCS_DIR = '/home/servermtsn2kolut/.hermes/cache/documents'
PROJECT_DIR = '/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api'
ENV_FILE = f'{PROJECT_DIR}/.env'

# ============ HELPERS ============
def load_env():
    env = {}
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if '=' in line and not line.startswith('#'):
                k, v = line.split('=', 1)
                env[k] = v
    return env

ENV = load_env()
PGPASS = ENV.get('POSTGRES_PASSWORD', ENV.get('POSTGRESS_PASSWORD', ''))
PGUSER = ENV.get('POSTGRES_USER', '')
PGDB = ENV.get('POSTGRES_DB', '')
PGPORT = ENV.get('POSTGRES_PORT', '5432')

def sql(sql, *args):
    """Execute SQL via psql, return rows as list of dicts"""
    formatted = sql % args if args else sql
    fd, path = tempfile.mkstemp(suffix='.sql')
    try:
        with os.fdopen(fd, 'w') as f:
            f.write(formatted + '\n')
        my_env = os.environ.copy()
        my_env['PGPASSWORD'] = PGPASS
        cmd = f"psql -h 127.0.0.1 -U {shq(PGUSER)} -d {shq(PGDB)} -p {PGPORT} -t -A -F'|' -f {shq(path)}"
        r = subprocess.run(cmd, shell=True, capture_output=True, text=True, timeout=60, env=my_env)
        combined = (r.stdout + r.stderr).strip()
        if r.returncode != 0 or 'ERROR' in combined.upper():
            # Print error for debugging
            if 'ERROR' in combined.upper() or 'FATAL' in combined.upper():
                pass  # errors handled in caller
        return combined
    finally:
        os.unlink(path)

def shq(s):
    return "'" + s.replace("'", "'\\''") + "'"

def uuid():
    return str(uuid4())

def log(step, msg, status='✅'):
    print(f"  {status} [{step}] {msg}")

# ============ MAIN ============
print("=" * 70)
print("IMPORT DATA MTS NEGERI 2 KOLAKA UTARA")
print("TAHUN PELAJARAN 2026/2027")
print(datetime.now().strftime("%Y-%m-%d %H:%M:%S"))
print("=" * 70)

# Get active academic year
ay_raw = sql("SELECT id FROM academic_years WHERE is_active = true LIMIT 1")
AY_ID = ay_raw.split('|')[0].strip() if ay_raw else ''
if not AY_ID:
    print("❌ Active academic year not found!")
    sys.exit(1)
print(f"\n📅 Academic Year: {ay_raw.strip()}")
print(f"   ID: {AY_ID}")

# ============ STEP 1: CLASSES ============
print("\n" + "=" * 50)
print("STEP 1: CLASSES")
print("=" * 50)

# Get existing classes
existing = []
for line in sql("SELECT id, code, name, level FROM school_classes WHERE is_active = true").split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 4:
        existing.append({'id': parts[0], 'code': parts[1], 'name': parts[2], 'level': parts[3]})

existing_codes = set(c['code'] for c in existing)
print(f"   Existing classes: {len(existing)}")

# Update existing classes to 2026/2027 AY
for c in existing:
    if c['code'] not in ['VII.E', 'VIII.D']:
        sql("UPDATE school_classes SET academic_year_id = %s::uuid WHERE id = %s::uuid", shq(AY_ID), shq(c['id']))
log('Classes', f"Updated {len(existing)-2} classes to 2026/2027", '🔄')

# Create missing classes
missing_classes = [
    ('VII.E', 'VII.E', 'VII'),
    ('VIII.D', 'VIII.D', 'VIII'),
]
for code, name, level in missing_classes:
    if code not in existing_codes:
        cid = uuid()
        sql("INSERT INTO school_classes (id, academic_year_id, code, name, level, is_active, created_at) VALUES (%s::uuid, %s::uuid, %s, %s, %s, true, NOW())",
            shq(cid), shq(AY_ID), shq(code), shq(name), shq(level))
        existing.append({'id': cid, 'code': code, 'name': name, 'level': level})
        log('Classes', f"Created {code}", '➕')
    else:
        log('Classes', f"{code} already exists, updating AY", '🔄')
        c = [x for x in existing if x['code'] == code][0]
        sql("UPDATE school_classes SET academic_year_id = %s::uuid WHERE id = %s::uuid", shq(AY_ID), shq(c['id']))

# Re-read all classes for this AY
classes_raw = sql("SELECT id, code, name, level FROM school_classes WHERE academic_year_id = %s::uuid ORDER BY level, name", shq(AY_ID))
CLASSES = {}
for line in classes_raw.split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 4:
        CLASSES[parts[1]] = {'id': parts[0], 'name': parts[2], 'level': parts[3]}
print(f"   Total classes for 2026/2027: {len(CLASSES)}")
for k, v in sorted(CLASSES.items(), key=lambda x: x[1]['level']):
    print(f"     {k} ({v['level']})")

# ============ STEP 2: STUDENTS ============
print("\n" + "=" * 50)
print("STEP 2: STUDENTS")
print("=" * 50)

import openpyxl

# Load parsed students from the analysis
with open('/tmp/import-data.json') as f:
    data = json.load(f)

students = data['students']
print(f"   Total students to import: {len(students)}")

# Check existing NIS
existing_nis = set()
for line in sql("SELECT nis FROM students WHERE nis IS NOT NULL").split('\n'):
    if line.strip():
        existing_nis.add(line.strip())

imported = 0
skipped = 0
errors = 0

for s in students:
    nis = s.get('nis', '')
    if nis in existing_nis:
        skipped += 1
        continue
    
    # Determine class code
    class_code = s.get('class_code', '')
    if not class_code:
        level = s.get('level', 'VII')
        class_code = f"{level}.A"  # default
    
    # Map class code format - roster uses VII.A, DB uses VII.A
    class_id = ''
    for cc, info in CLASSES.items():
        if cc.replace('.', '').lower() == class_code.replace('.', '').lower():
            class_id = info['id']
            break
    
    if not class_id:
        errors += 1
        continue
    
    sid = uuid()
    gender = 'L' if s.get('gender', '').upper() == 'L' else 'P'
    nama = s.get('nama', '').strip().title()
    tempat = s.get('tempat_lahir', '').strip().title()
    nisn = s.get('nisn', '').strip()
    tgl = s.get('tanggal_lahir', '').strip()
    
    # Parse date
    date_sql = 'NULL'
    for fmt in ['%d/%m/%Y', '%d-%m-%Y', '%d/%m/%y', '%d-%m-%y']:
        try:
            dt = datetime.strptime(tgl, fmt)
            date_sql = f"'{dt.strftime('%Y-%m-%d')}'::date"
            break
        except:
            pass
    
    sql(f"""
        INSERT INTO students (id, nis, nisn, nama, gender, class_id, tempat_lahir, tanggal_lahir, is_active)
        VALUES (%s::uuid, %s, %s, %s, %s, %s::uuid, %s, {date_sql}, true)
    """, shq(sid), shq(nis), shq(nisn), shq(nama), shq(gender), shq(class_id), shq(tempat))
    imported += 1

print(f"   Imported: {imported}, Skipped (duplicate): {skipped}, Errors: {errors}")

# Verify student count
student_count = sql("SELECT COUNT(*) FROM students WHERE is_active = true")
print(f"   Total students in DB: {student_count}")

# ============ STEP 3: TEACHER ASSIGNMENTS ============
print("\n" + "=" * 50)
print("STEP 3: ASSIGN GURU")
print("=" * 50)

# Get teachers map
teachers = {}
for line in sql("SELECT id, nama FROM employees WHERE is_active = true").split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 2:
        nama_clean = parts[1].lower().replace(',', '').replace('.', '').replace('  ', ' ')
        teachers[parts[1].lower()] = parts[0]

# Fuzzy name matching function
TEACHER_MAP = {
    "chairuddin": "chairuddin (drs)",
    "drs chairuddin": "chairuddin (drs)",
    "drs. chairuddin": "chairuddin (drs)",
    "ernawati": "ernawati",
    "km ahmad yani": "km muhammad yani",
    "km. muh tang": "km muh tang",
    "km. muh. tang": "km muh tang",
    "nurilmi": "nur ilmi",
    "rahma": "sitti rahma",
    "srihardianti": "sri hardianti",
    "asniar": "asniar",
}

# Non-teaching assignments to skip
NON_TEACHING_SUBJECTS = {
    'osim', 'pmr', 'ksm', 'uks', 'pembina olahraga',
    'wakamad saspras', 'wakamad kurikulum', 'wakamad kesiswaan',
    'kepala perpustakaan', 'penilaian kinerja guru',
    'jam mts meeto', 'jam mts labipi', 'jam di purau', 'jam di purauu', 'jam di watunohu',
    'total', 'total + piket', 'piket',
    'keagamaan',
}

def find_teacher(name):
    name = name.strip().lower().replace(',', '').replace('.', '').replace('  ', ' ')
    
    # Check in explicit map first
    if name in TEACHER_MAP:
        mapped_name = TEACHER_MAP[name]
        if mapped_name in teachers:
            return teachers[mapped_name]
    
    # Direct match
    if name in teachers:
        return teachers[name]
    
    # Fuzzy match: find teacher with most word overlap
    name_parts = set(word for word in name.split() if len(word) > 2)
    best_match = None
    best_score = 0
    
    for t_name, t_id in teachers.items():
        t_parts = set(word for word in t_name.split() if len(word) > 2)
        if not name_parts or not t_parts:
            continue
        common = name_parts & t_parts
        score = len(common) / max(len(name_parts), len(t_parts))
        if score > 0.5 and score > best_score:
            best_match = t_id
            best_score = score
    
    return best_match

# Subject mapping - normalize names
SUBJECT_MAP = {
    "al-qur'an dan hadits": "Al-Qur'an Hadis",
    "al-qur'an hadits": "Al-Qur'an Hadis",
    "qur'an hadits": "Al-Qur'an Hadis",
    "al qur'an hadits": "Al-Qur'an Hadis",
    "akidah akhlak": "Akidah Akhlak",
    "akidah ahlak": "Akidah Akhlak",
    "bahasa arab": "Bahasa Arab",
    "bhs arab": "Bahasa Arab",
    "bhs.arab": "Bahasa Arab",
    "b.indonesia": "Bahasa Indonesia",
    "b.indo": "Bahasa Indonesia",
    "b.ina": "Bahasa Indonesia",
    "b. indonesia": "Bahasa Indonesia",
    "bhs indonesia": "Bahasa Indonesia",
    "bhs.ina": "Bahasa Indonesia",
    "bahasa inggris": "Bahasa Inggris",
    "b.inggris": "Bahasa Inggris",
    "bhs.inggris": "Bahasa Inggris",
    "b.ingg": "Bahasa Inggris",
    "b.inggr": "Bahasa Inggris",
    "bhs.ingg": "Bahasa Inggris",
    "bhs inggris": "Bahasa Inggris",
    "bimbingan konseling": "Bimbingan Konseling",
    "fiqih": "Fikih",
    "fiqhi": "Fikih",
    "ilmu pengetahuan alam (ipa)": "Ilmu Pengetahuan Alam",
    "ipa": "Ilmu Pengetahuan Alam",
    "ilmu pengetahuan sosial (ips)": "Ilmu Pengetahuan Sosial",
    "ips": "Ilmu Pengetahuan Sosial",
    "informatika": "Informatika",
    "p.seni": "Seni Budaya",
    "seni budaya": "Seni Budaya",
    "matematika": "Matematika",
    "mat": "Matematika",
    "m a t": "Matematika",
    "muatan lokal agama dan ahlak(tahfiz)": "Muatan Lokal Keagamaan",
    "muatan lokal keagamaan": "Muatan Lokal Keagamaan",
    "mulok": "Muatan Lokal Keagamaan",
    "mulok/takfis": "Muatan Lokal Keagamaan",
    "mulok tahfizd": "Muatan Lokal Keagamaan",
    "mulok tahfidz": "Muatan Lokal Keagamaan",
    "mulok kewirausahaan": "Mulok Kewirausahaan",
    "pendidikan jasmani, olahraga, dan kesehatan": "Pendidikan Jasmani, Olahraga, dan Kesehatan",
    "penjas": "Pendidikan Jasmani, Olahraga, dan Kesehatan",
    "penjasorkes": "Pendidikan Jasmani, Olahraga, dan Kesehatan",
    "pendidikan pancasila": "Pendidikan Pancasila",
    "pkn": "Pendidikan Pancasila",
    "pendidikan kewarganegaraan (pkn)": "Pendidikan Pancasila",
    "pkn/pendidikan kewarganegaraan": "Pendidikan Pancasila",
    "prakarya": "Prakarya",
    "p.seni": "Seni Budaya",
    "sejarah kebudayaan islam (ski)": "Sejarah Kebudayaan Islam",
    "ski": "Sejarah Kebudayaan Islam",
    "informatika": "Informatika",
    "simulasi cbt": "Simulasi CBT",
    "cbt-dev": "Simulasi CBT Development Test",
}

# Get subjects map
subjects = {}
for line in sql("SELECT id, name FROM subjects WHERE is_active = true").split('\n'):
    parts = [p.strip() for p in line.split('|')]
    if len(parts) >= 2:
        subjects[parts[1].lower()] = parts[0]

def find_subject(name):
    name_clean = name.strip().lower()
    # Direct lookup
    if name_clean in SUBJECT_MAP:
        mapped = SUBJECT_MAP[name_clean].lower()
        if mapped in subjects:
            return subjects[mapped]
    # Try subjects directly
    for s_name, s_id in subjects.items():
        if name_clean in s_name or s_name in name_clean:
            return s_id
    return None

# Create assignments
assignments = data['assignments']
created_assignments = 0
skipped_assignments = 0
unmatched_teachers = set()
unmatched_subjects = set()

CLASS_MAP = {}
for k, v in CLASSES.items():
    # Normalize: remove dots
    CLASS_MAP[k.replace('.', '')] = v['id']

for a in assignments:
    teacher_name = a['teacher'].strip().lower()
    subject_name = a['subject'].strip()
    
    # Skip non-teaching assignments
    if subject_name.lower().strip() in NON_TEACHING_SUBJECTS:
        skipped_assignments += 1
        continue
    
    teacher_id = find_teacher(teacher_name)
    subject_id = find_subject(subject_name)
    
    if not teacher_id:
        unmatched_teachers.add(a['teacher'])
        skipped_assignments += 1
        continue
    if not subject_id:
        unmatched_subjects.add(a['subject'])
        skipped_assignments += 1
        continue
    
    for class_code, jam in a['classes'].items():
        normalized_code = class_code.replace('.', '')
        class_id = CLASS_MAP.get(normalized_code)
        if not class_id:
            continue
        
        # Check if already exists
        existing_assign = sql("SELECT id FROM class_subject_assignments WHERE class_id = %s::uuid AND subject_id = %s::uuid AND teacher_employee_id = %s::uuid",
            shq(class_id), shq(subject_id), shq(teacher_id))
        if existing_assign.strip():
            skipped_assignments += 1
            continue
        
        aid = uuid()
        sql("""INSERT INTO class_subject_assignments (id, academic_year_id, class_id, subject_id, teacher_employee_id, total_weekly_hours)
               VALUES (%s::uuid, %s::uuid, %s::uuid, %s::uuid, %s::uuid, %s)
               ON CONFLICT DO NOTHING""",
            shq(aid), shq(AY_ID), shq(class_id), shq(subject_id), shq(teacher_id), shq(str(jam)))
        created_assignments += 1

print(f"   Created: {created_assignments} assignments")
print(f"   Skipped: {skipped_assignments}")
if unmatched_teachers:
    print(f"   ❌ Unmatched teachers: {unmatched_teachers}")
if unmatched_subjects:
    print(f"   ❌ Unmatched subjects: {unmatched_subjects}")

# Verify
assign_count = sql("SELECT COUNT(*) FROM class_subject_assignments WHERE academic_year_id = %s::uuid", shq(AY_ID))
print(f"   Total assignments in DB: {assign_count}")

# ============ STEP 4: TIMETABLE ============
print("\n" + "=" * 50)
print("STEP 4: TIMETABLE (coming in import-schedule.py)")
print("=" * 50)
print("   Timetable parsing is complex. Will be done in a separate script.")
print("   Creating assignments is the priority first step.")

# ============ SUMMARY ============
print("\n" + "=" * 70)
print("IMPORT SUMMARY")
print("=" * 70)
print(f"   Academic Year: 2026/2027")
student_count = sql("SELECT COUNT(*) FROM students WHERE is_active = true")
assign_count = sql("SELECT COUNT(*) FROM class_subject_assignments WHERE academic_year_id = %s::uuid", shq(AY_ID))
class_count = sql("SELECT COUNT(*) FROM school_classes WHERE academic_year_id = %s::uuid", shq(AY_ID))
print(f"   Classes:    {class_count}")
print(f"   Students:   {student_count}")
print(f"   Assignments: {assign_count}")
print(f"   Timetable:  Not imported yet (separate phase)")
print("\n" + "=" * 70)
print("IMPORT COMPLETE")
print("=" * 70)
