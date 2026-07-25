"""Import timetable from Roster Excel."""
import openpyxl, subprocess, os, re, sys, tempfile
from uuid import uuid4

DOCS_DIR = '/home/servermtsn2kolut/.hermes/cache/documents'
ENV = {}
for line in open('/home/servermtsn2kolut/mtsn2kolut-super-app/services/core-api/.env'):
    line = line.strip()
    if '=' in line and not line.startswith('#'):
        k, v = line.split('=', 1)
        ENV[k] = v

PGPASS = ENV.get('POSTGRES_PASSWORD', ENV.get('POSTGRESS_PASSWORD', ''))
AY_ID = 'e3dfb158-76c6-46c2-b6e7-00cd604a0277'
SMT_ID = 'ddbe7929-6861-4989-b408-40e27655e8b8'

def psql(q):
    my_env = os.environ.copy(); my_env['PGPASSWORD'] = PGPASS
    fd, path = tempfile.mkstemp(suffix='.sql')
    with os.fdopen(fd, 'w') as f: f.write(q + '\n')
    r = subprocess.run(f"psql -h 127.0.0.1 -U {ENV['POSTGRES_USER']} -d {ENV['POSTGRES_DB']} -p {ENV.get('POSTGRES_PORT','5432')} -t -A -F'|' -f {path}", shell=True, capture_output=True, text=True, timeout=30, env=my_env)
    os.unlink(path)
    return (r.stdout + r.stderr).strip()

def pt(t):
    m = re.match(r'(\d{1,2})[.](\d{2})', str(t).strip().replace(':', '.'))
    return f"{int(m.group(1)):02d}:{m.group(2)}:00" if m else None

ROMAN = {'I':1,'II':2,'III':3,'IV':4,'V':5,'VI':6,'VII':7,'VIII':8,'IX':9,'X':10}
DAYS = {
    1: [(1,'07.30','08.05'),(2,'08.05','08.40'),(3,'08.40','09.15'),(4,'09.15','09.50'),(5,'10.05','10.40'),(6,'10.40','11.15'),(7,'11.40','12.15'),(8,'12.25','13.00'),(9,'13.00','13.35')],
    2: [(1,'07.30','08.05'),(2,'08.05','08.35'),(3,'08.35','09.10'),(4,'09.10','09.45'),(5,'10.00','10.35'),(6,'10.35','11.10'),(7,'11.10','11.45'),(8,'11.45','12.20'),(9,'12.25','13.00')],
    3: [(1,'07.50','08.25'),(2,'08.25','09.00'),(3,'09.15','09.50'),(4,'09.50','10.25'),(5,'10.25','11.00'),(6,'11.00','11.35'),(7,'12.05','12.40'),(8,'12.40','13.15')],
    4: [(1,'07.15','07.50'),(2,'07.50','08.25'),(3,'08.25','09.00'),(4,'09.00','09.35'),(5,'09.45','10.20'),(6,'10.20','10.55'),(7,'10.55','11.30'),(8,'11.30','12.05')],
    5: [(1,'09.00','09.35'),(2,'09.35','10.10'),(3,'10.10','10.45')],
    6: [(1,'07.15','07.50'),(2,'07.50','08.25'),(3,'08.25','09.00'),(4,'09.00','09.35'),(5,'09.45','10.20'),(6,'10.20','10.55'),(7,'10.55','11.30'),(8,'11.30','12.05')]
}
COLS = {4:'IX.A',5:'IX.B',6:'IX.C',7:'VIII.A',8:'VIII.B',9:'VIII.C',10:'VIII.D',11:'VII.A',12:'VII.B',13:'VII.C',14:'VII.D',15:'VII.E'}
DAY_KEYS = {'SENIN':1,'SELASA':2,'RABU':3,'KAMIS':4,"JUM'AT":5,'JUMAT':5,'SABTU':6}
NON = ['shalat','istirahat','upacara','zikir','apel',"jum'at bersih",'latihan','membersihkan','olahraga sehat','shalat dhuha','tadarrus','pembersihan']

ALIAS = {
    'mat':'matematika','m a t':'matematika','ipa':'ilmu pengetahuan alam','ips':'ilmu pengetahuan sosial',
    'bhs indonesia':'bahasa indonesia','bhs.ina':'bahasa indonesia','b indonesia':'bahasa indonesia',
    'bhs inggris':'bahasa inggris','bhs.ingg':'bahasa inggris','b.inggris':'bahasa inggris','b.ingg':'bahasa inggris','bhs/inggris':'bahasa inggris',
    'bhs arab':'bahasa arab','bhs.arab':'bahasa arab',
    'penjas':'pendidikan jasmani olahraga dan kesehatan','penjaskes':'pendidikan jasmani olahraga dan kesehatan',
    'pkn':'pendidikan pancasila','pk n':'pendidikan pancasila',
    'p.seni':'seni budaya','seni budaya':'seni budaya','ski':'sejarah kebudayaan islam',
    'fiqhi':'fikih','fiqih':'fikih',
    'akidah akhlak':'akidah akhlak','akidah ahlak':'akidah akhlak',
    'iinformatika':'informatika','informatika':'informatika',
    "mulok":'muatan lokal keagamaan',"mulok/takfis":'muatan lokal keagamaan',"mulok tahfizd":'muatan lokal keagamaan',"mulok tahfidz":'muatan lokal keagamaan',
}

# Step 1
psql(f"DELETE FROM lesson_period_templates WHERE academic_year_id='{AY_ID}'::uuid")
psql(f"DELETE FROM timetable_slots WHERE semester_id='{SMT_ID}'::uuid")
cnt = sum(1 for d in DAYS for _ in DAYS[d])
for d in DAYS:
    for pn, st, et in DAYS[d]:
        psql(f"INSERT INTO lesson_period_templates VALUES (gen_random_uuid(), '{AY_ID}'::uuid, {d}, {pn}, '{pt(st)}'::time, '{pt(et)}'::time, 'pelajaran', 'JP {pn}', true, NOW(), NOW())")
print(f"Step1: {cnt} period templates")

# Step 2 - ref data
subj_map={}; assn={}; classes={}
for l in psql(f"SELECT code, id::text FROM school_classes WHERE academic_year_id='{AY_ID}'::uuid AND is_active=true").split('\n'):
    p=[x.strip() for x in l.split('|')]
    if len(p)>=2: classes[p[0]]=p[1]

for l in psql("SELECT id::text, name, LOWER(REGEXP_REPLACE(name, '[,.-]', ' ', 'g')) FROM subjects WHERE is_active=true").split('\n'):
    p=[x.strip() for x in l.split('|')]
    if len(p)>=3:
        c=re.sub(r'\s+',' ',p[2]).strip()
        subj_map[c]=p[0]
        subj_map[p[1].lower()]=p[0]

for l in psql(f"SELECT c.code, s.name, csa.id::text FROM class_subject_assignments csa JOIN school_classes c ON c.id=csa.class_id JOIN subjects s ON s.id=csa.subject_id WHERE c.academic_year_id='{AY_ID}'::uuid").split('\n'):
    p=[x.strip() for x in l.split('|')]
    if len(p)>=3:
        nk=re.sub(r'[,.\-]',' ',p[1].lower())
        nk=re.sub(r'\s+',' ',nk).strip()
        # Also store without apostrophe
        nk2=nk.replace("'",' ')
        nk2=re.sub(r'\s+',' ',nk2).strip()
        assn[(p[0],nk)]=p[2]
        if nk2!=nk: assn[(p[0],nk2)]=p[2]  # also store apostrophe-stripped version

ptids={}
for l in psql(f"SELECT day_of_week, period_number, id::text FROM lesson_period_templates WHERE academic_year_id='{AY_ID}'::uuid").split('\n'):
    p=[x.strip() for x in l.split('|')]
    if len(p)>=3: ptids[(int(p[0]),int(p[1]))]=p[2]

print(f"Step2: {len(classes)} classes, {len(subj_map)} subjects, {len(assn)} assn, {len(ptids)} periods")

# Step 3 - Parse roster
wb=openpyxl.load_workbook(f'{DOCS_DIR}/doc_cf93a5b2b842_perubahan ROSTER PELAJARAN SEMESTER GANJIL 2026-2027.xlsx',data_only=True)
ws=wb['ROSTER SM GNJL 2026']

cd=None; slots=0; unk=set(); miss=set()

for ri in range(9, 98):
    row=[ws.cell(row=ri,column=c).value for c in range(1,22)]
    if row[1] and str(row[1]).strip().upper() in DAY_KEYS:
        cd=DAY_KEYS[str(row[1]).strip().upper()]; continue
    if not cd: continue
    pn_raw=str(row[2] or '').strip()
    pn=ROMAN.get(pn_raw)
    if not pn or pn<1 or pn>9: continue
    ptid=ptids.get((cd,pn))
    if not ptid: continue
    start=None; end=None
    for _pn,_st,_et in DAYS[cd]:
        if _pn==pn: start=_st; end=_et; break
    if not start: continue

    for col, cc in COLS.items():
        raw=str(row[col] or '').strip() if col<len(row) and row[col] else ''
        if not raw or raw in ['',' ','-']: continue
        if any(x in raw.upper() for x in NON): continue
        
        # Normalize
        n = re.sub(r'[.\-]', ' ', raw.lower())
        n = re.sub(r'\s+', ' ', n).strip()
        n = ALIAS.get(n, n)
        
        # Find subject
        sid = subj_map.get(n) or subj_map.get(raw.lower())
        # try both with and without apostrophe
        n_apos = n.replace("'", '')
        n_apos = re.sub(r'\s+', ' ', n_apos).strip()
        if not sid: sid=subj_map.get(n_apos) or subj_map.get(raw.lower().replace("'",''))
        if not sid:
            unk.add(f"{raw}→{n}")
            continue
        
        # Find assignment - try multiple key variations
        fn = re.sub(r'[,\-]', ' ', n)
        fn = re.sub(r'\s+', ' ', fn).strip()
        # try with apostrophe, without, and stripped
        candidates = [fn, fn.replace("'",' '), fn.replace("'",'')]
        candidates = [re.sub(r'\s+',' ',c).strip() for c in candidates]
        aid=None
        for cfn in set(candidates):
            aid=assn.get((cc,cfn))
            if aid: break
        if not aid:
            miss.add(f"{cc}/{raw}({fn})")
            continue
        
        # Create slot
        st_t=pt(start); et_t=pt(end)
        if not st_t or not et_t: continue
        
        # Skip if exists
        ex=psql(f"SELECT id FROM timetable_slots WHERE assignment_id='{aid}'::uuid AND day_of_week={cd} AND start_time='{st_t}'::time LIMIT 1")
        
        if ex.strip(): continue
        
        r=psql(f"INSERT INTO timetable_slots (id, assignment_id, day_of_week, start_time, end_time, lesson_period_id, slot_type, lesson_hours, semester_id) VALUES (gen_random_uuid(), '{aid}'::uuid, {cd}, '{st_t}'::time, '{et_t}'::time, '{ptid}'::uuid, 'pelajaran', 1, '{SMT_ID}'::uuid)")
        
        slots += 1
        
        # Debug: print first slot created
        if slots == 1:
            print(f"  FIRST SLOT: {cc} {raw} day={cd} time={st_t}-{et_t}")

print(f"Step3: {slots} slots, {len(unk)} unknown, {len(miss)} missing")
if unk: print(f"  Unk: {list(unk)[:10]}")
if miss: print(f"  Miss: {list(miss)[:10]}")

c1=psql(f"SELECT COUNT(*)::text FROM lesson_period_templates WHERE academic_year_id='{AY_ID}'::uuid")
c2=psql(f"SELECT COUNT(*)::text FROM timetable_slots WHERE semester_id='{SMT_ID}'::uuid")
print(f"Final: {c1} periods, {c2} slots")
