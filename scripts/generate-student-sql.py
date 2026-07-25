#!/usr/bin/env python3
"""
Generate student import SQL from PDF data.
Distributes students evenly across class sections.

Usage: uv run python3 scripts/generate-student-sql.py | tee scripts/import-students.sql
Then: cd services/core-api && source .env && PGPASSWORD=$POSTGRES_PASSWORD psql -h 127.0.0.1 -U $POSTGRES_USER -d $POSTGRES_DB -p $POSTGRES_PORT -f ../../scripts/import-students.sql
"""
import subprocess, re, sys
from datetime import datetime
from uuid import uuid4

DOCS_DIR = '/home/servermtsn2kolut/.hermes/cache/documents'
PDF_PATH = f'{DOCS_DIR}/doc_ded7b5897f21_LAPORAN BULANAN GANJIL T.P 2026-2027.xlsx - Siswa.pdf'

# Class IDs for 2026/2027
CLASS_IDS = {
    'VII.A': 'ac6218cd-194c-4c6a-ae71-d1de9ca06155',
    'VII.B': '12b750fd-f33b-4556-b60d-fb2e07e7a113',
    'VII.C': 'f3b86457-d0ed-4898-aafd-a859d77813a3',
    'VII.D': '50a7c72c-e4f1-4674-82ce-1e719b722455',
    'VII.E': '498ed32b-68b5-4a93-b9c9-7efd5447d857',
    'VIII.A': '65d9fc64-dc7f-4252-be70-f9c4cdfdbcff',
    'VIII.B': '10bdd6ba-8e7c-4ca7-aeca-7de7c8dd4b74',
    'VIII.C': '44ad9360-5f01-4be6-a53b-1d474fec6b39',
    'VIII.D': 'ae915c90-f305-4fb2-95b4-b9d743b98889',
    'IX.A': '5dbf2382-c32c-42f6-8228-938f53b10129',
    'IX.B': '99c6c9fb-06a7-4d53-a880-46979de99b7c',
    'IX.C': 'f551c72d-7678-41a3-b8be-d621eabe260c',
}

SECTIONS_BY_LEVEL = {
    'VII': ['VII.A', 'VII.B', 'VII.C', 'VII.D', 'VII.E'],
    'VIII': ['VIII.A', 'VIII.B', 'VIII.C', 'VIII.D'],
    'IX': ['IX.A', 'IX.B', 'IX.C'],
}

def parse_pdf():
    r = subprocess.run(['pdftotext', '-layout', PDF_PATH, '-'], capture_output=True, text=True, timeout=10)
    return r.stdout.split('\n')

def main():
    lines = parse_pdf()
    students = []
    
    for i, line in enumerate(lines):
        ls = line.strip()
        if not ls:
            continue
        
        # Extract URT, NIS, NISN, NAME, GENDER, CLASS, PLACE, DOB, PARENT
        m = re.match(r'\s*(\d{1,3})\s+(\d{10,25})\s+(\S{1,20})\s+(.+?)\s+([LP])\s+(VIII[A-D]?|VII[A-E]?|IX[A-C]?)\s+(.+?)\s+(\d{1,2}/\d{1,2}/\d{2,4})\s+(.+?)(?:\s+(.+?))?\s*$', ls)
        if m:
            level = m.group(6).strip()
            students.append({
                'nis': m.group(2).strip(),
                'nisn': m.group(3).strip().replace("'", '').replace("'", ''),
                'nama': m.group(4).strip().title(),
                'gender': m.group(5).upper(),
                'level': level,
                'tempat_lahir': m.group(7).strip().title(),
                'tanggal_lahir': m.group(8).strip(),
                'ayah': m.group(9).strip() if m.group(9) else '',
            })
            continue
        
        # Simpler format (no class in columns)
        m2 = re.match(r'\s*(\d{1,3})\s+(\d{10,25})\s+(\S{1,20})\s+(.+?)\s+([LP])\s+(.+?)\s+(\d{1,2}/\d{1,2}/\d{2,4})\s+(.+?)(?:\s+(.+?))?\s*$', ls)
        if m2:
            level = None
            # Determine level from nearby context (last seen class header)
            students.append({
                'nis': m2.group(2).strip(),
                'nisn': m2.group(3).strip().replace("'", '').replace("'", ''),
                'nama': m2.group(4).strip().title(),
                'gender': m2.group(5).upper(),
                'level': None,  # will be assigned later
                'tempat_lahir': m2.group(6).strip().title(),
                'tanggal_lahir': m2.group(7).strip(),
                'ayah': m2.group(8).strip() if m2.group(8) else '',
            })
    
    # Separate IX - it starts at line ~112
    # From our analysis: first 74 students are VIII, then ~53 are IX, then ~83 are VII
    # Let me re-sort by their actual level
    
    # Actually, the parser already extracts level from the KLS column
    # Let me filter out students without level
    valid = [s for s in students if s['level']]
    
    # Group by level
    by_level = {}
    for s in valid:
        lv = s['level']
        # Normalize: the KLS column might be "VII", "VIII", or "IX"
        # Remove any trailing letters
        base = re.match(r'(VIII|VII|IX)', lv)
        if base:
            base = base.group(1)
            by_level.setdefault(base, []).append(s)
    
    # Distribute students evenly across sections
    result = []
    section_counters = {}
    
    for level, sections in SECTIONS_BY_LEVEL.items():
        students_in_level = by_level.get(level, [])
        section_counters[level] = 0
        for i, s in enumerate(students_in_level):
            section_idx = i % len(sections)
            section_code = sections[section_idx]
            s['class_code'] = section_code
            result.append(s)
            section_counters[level] += 1
    
    # Students that couldn't be matched
    unmatched = len(students) - len(valid)
    
    return result, unmatched

def parse_date(tgl):
    date_val = 'NULL'
    for fmt in ['%d/%m/%Y', '%d-%m-%Y', '%d/%m/%y', '%d-%m-%y']:
        try:
            dt = datetime.strptime(tgl, fmt)
            return f"'{dt.strftime('%Y-%m-%d')}'::date"
        except:
            pass
    return 'NULL'

def generate_sql(students):
    out = ['-- Generated student import script', 'BEGIN;', '']
    inserted = 0
    
    for s in students:
        nis = s['nis']
        nisn = s['nisn'] if s['nisn'] else '0'
        nama = s['nama'].replace("'", "''")
        gender = s['gender']
        class_code = s.get('class_code', 'VII.A')
        tempat = s['tempat_lahir'].replace("'", "''")
        tgl = s['tanggal_lahir']
        ayah = s.get('ayah', '').replace("'", "''")
        
        class_id = CLASS_IDS.get(class_code)
        if not class_id:
            continue
        
        date_val = parse_date(tgl)
        sid = str(uuid4())
        
        out.append(
            f"INSERT INTO students (id, nis, nisn, nama, gender, class_id, tempat_lahir, tanggal_lahir, is_active) "
            f"SELECT '{sid}'::uuid, '{nis}', '{nisn}', '{nama}', '{gender}', '{class_id}'::uuid, '{tempat}', {date_val}, true "
            f"WHERE NOT EXISTS (SELECT 1 FROM students WHERE nis = '{nis}');"
        )
        inserted += 1
    
    out.append('')
    out.append('COMMIT;')
    out.append(f'')
    out.append(f'-- Students to insert: {inserted}')
    
    return '\n'.join(out)

if __name__ == '__main__':
    students, unmatched = main()
    
    # Stats
    by_class = {}
    for s in students:
        cc = s.get('class_code', '?')
        by_class.setdefault(cc, []).append(s['nama'])
    
    print(f'-- Total students parsed: {len(students)}', file=sys.stderr)
    print(f'-- Unmatched (no level): {unmatched}', file=sys.stderr)
    print(f'-- Distribution:', file=sys.stderr)
    for cc in sorted(by_class.keys()):
        names = by_class[cc]
        print(f'--   {cc}: {len(names)} students', file=sys.stderr)
    
    sql = generate_sql(students)
    print(sql)
