# System Seed KaTeX & Arabic Question Models Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task after explicit approval.

**Goal:** Add a safe, reviewable system/seed question set that stress-tests full KaTeX math rendering and Arabic/RTL rendering across all currently supported Bank Soal models.

**Architecture:** Use direct, audited database authoring for `author_username='system'` only. Keep all new items in `draft` / `draft` workflow, with stable codes, rich HTML/LaTeX fields, options JSON synchronized with legacy option columns, and post-insert verification reports. Current production schema supports `multiple_choice` and `essay`; other models are represented as pedagogical variants inside those supported types unless/ until schema adds new question types.

**Tech Stack:** PostgreSQL, Go core-api Bank Soal schema, SvelteKit RichContent/KaTeX renderer, KaTeX-compatible LaTeX, RTL Arabic HTML.

---

## Scope Recommendation

### A. KaTeX stress-test seed set

Create **20 Matematika system seed questions**:

- 16 multiple choice
- 4 essay
- Target level split:
  - VII: 6 soal
  - VIII: 8 soal
  - IX: 6 soal
- Status: `draft`
- Workflow: `draft`
- Author: `system`

Coverage of KaTeX symbols:

1. Superscript/subscript: `$x^2$`, `$a_n$`
2. Fractions: `$$\\frac{a}{b}$$`
3. Roots: `$\\sqrt{49}$`, `$\\sqrt[3]{8}$`
4. Multiplication/division: `$\\times$`, `$\\div$`
5. Inequality: `$<$`, `$>$`, `$\\le$`, `$\\ge$`
6. Set notation: `$\\in$`, `$\\notin$`, `$\\cup$`, `$\\cap$`
7. Geometry: `$\\angle ABC$`, `$\\triangle ABC$`, `$90^\\circ$`
8. Greek symbols: `$\\pi$`, `$\\theta$`, `$\\alpha$`
9. Ratio/proportion: `$3:4$`, `$\\frac{x}{y}$`
10. Linear equation/system: `$2x+3=11$`, `$\\begin{cases}...\\end{cases}$`
11. Coordinate/table notation: `$(x,y)$`
12. Statistics: `$\\bar{x}$`, `$\\sum x_i$`
13. Display multi-step solution in explanation
14. Mixed inline + display math in same stem
15. Options containing KaTeX
16. Explanation containing KaTeX and Indonesian prose

### B. Arabic/RTL seed set

Create **20 Bahasa Arab system seed questions**:

- 16 multiple choice
- 4 essay
- Target level split:
  - VII: 6 soal
  - VIII: 8 soal
  - IX: 6 soal
- Status: `draft`
- Workflow: `draft`
- Author: `system`

Coverage of Arabic model variants:

1. Mufradat: arti kata Arab ke Indonesia
2. Mufradat: Indonesia ke Arab
3. Melengkapi kalimat rumpang
4. Menentukan dhamir yang tepat
5. Menentukan isim/fi'il/harf sederhana
6. Mudzakkar-muannats
7. Mufrad-mutsanna-jamak sederhana
8. Kata tanya: من، ما، أين، متى، كيف
9. Membaca teks pendek dan menjawab pertanyaan
10. Menyusun kata menjadi kalimat sederhana
11. Menentukan terjemahan kalimat
12. Menentukan harakat sederhana pada kata umum
13. Sinonim/antonim sederhana
14. Dialog pendek madrasah/kelas
15. Essay: menerjemahkan kalimat pendek
16. Essay: membuat 2 kalimat tentang madrasah/keluarga

HTML/RTL standards:

- Use Arabic text inside `<span dir="rtl" lang="ar">...</span>` for inline Arabic.
- Use `<div dir="rtl" lang="ar" class="arabic-text">...</div>` for passages.
- Avoid malformed Arabic, mixed direction ambiguity, and unverified ayat/hadith text unless sourced.
- Use everyday MTs-safe context: madrasah, keluarga, kelas, waktu, benda sekolah.

### C. Supported model boundary

Current schema has only:

- `multiple_choice`
- `essay`

So “seluruh kemungkinan model soal” should be implemented as **pedagogical variants** under these two types first. Do not invent new `question_type` values yet because that risks breaking renderer, package builder, and scoring.

If later Bapak wants true new types, create a separate feature plan for:

- true/false
- matching
- multi-select
- short answer
- ordering
- fill-in-the-blank

---

## Safety Rules

- No production deploy/restart needed for data-only inserts.
- Backup DB before insert.
- Export before/after CSV.
- Only touch rows with `author_username='system'` and new seed code prefixes.
- New rows remain `draft` and `workflow_status='draft'`.
- Do not submit/publish automatically.
- All AI-created content must be reviewable by guru/operator before package use.

---

## Proposed Code Prefixes

KaTeX/Matematika:

```text
SYS-MTK-KATEX-VII-PG-0001..0006
SYS-MTK-KATEX-VIII-PG-0001..0006
SYS-MTK-KATEX-IX-PG-0001..0004
SYS-MTK-KATEX-VII-ES-0001
SYS-MTK-KATEX-VIII-ES-0001..0002
SYS-MTK-KATEX-IX-ES-0001
```

Bahasa Arab:

```text
SYS-BARAB-RTL-VII-PG-0001..0006
SYS-BARAB-RTL-VIII-PG-0001..0006
SYS-BARAB-RTL-IX-PG-0001..0004
SYS-BARAB-RTL-VII-ES-0001
SYS-BARAB-RTL-VIII-ES-0001..0002
SYS-BARAB-RTL-IX-ES-0001
```

---

## Task 1: Preflight audit

**Objective:** Confirm subjects, current schema, and no conflicting codes.

**Files:**
- Read DB only.
- Output: `reports/bank-soal-review/system-seed-katex-arabic-YYYYMMDD/preflight.tsv`

**Steps:**
1. Query subject IDs for `Matematika` and `Bahasa Arab`.
2. Query distinct `question_type` values.
3. Query existing codes with prefixes `SYS-MTK-KATEX-%` and `SYS-BARAB-RTL-%`.
4. Abort if code conflicts exist.

**Verification:**
- Subjects found.
- `question_type` support confirmed as `multiple_choice`, `essay`.
- Code conflicts = 0.

---

## Task 2: Draft KaTeX item bank CSV

**Objective:** Generate 20 Matematika seed questions with full KaTeX coverage.

**Files:**
- Create: `reports/bank-soal-review/system-seed-katex-arabic-YYYYMMDD/katex_draft.csv`

**Rules:**
- Populate `question_text`, `stem_html`, `stem_latex`, `options`, `option_a`-`option_d`, `answer_key`, `explanation`, `explanation_html`.
- Verify LaTeX uses single commands like `\frac`, not double-escaped `\\frac` in DB text.
- Include both inline `$...$` and display `$$...$$` examples.

**Verification:**
- 20 rows.
- 16 PG have answer key A-D.
- 4 essay have empty `answer_key` and `options=[]`.
- Every row includes at least one KaTeX expression.

---

## Task 3: Draft Arabic/RTL item bank CSV

**Objective:** Generate 20 Bahasa Arab seed questions covering common MTs model variants.

**Files:**
- Create: `reports/bank-soal-review/system-seed-katex-arabic-YYYYMMDD/arabic_draft.csv`

**Rules:**
- Use safe MTs context only.
- Wrap Arabic with `dir="rtl" lang="ar"`.
- Avoid unsourced religious text. If using Islamic context, use simple vocabulary like مسجد، مدرسة، كتاب, not quoted ayat/hadith.
- Keep Indonesian instructions clear.

**Verification:**
- 20 rows.
- Arabic text displays with RTL wrappers in `stem_html` / options HTML.
- No malformed mojibake or broken Arabic glyphs.

---

## Task 4: Agent swarm content review

**Objective:** Review drafts before DB insert.

**Agents:**
1. Matematika/KaTeX reviewer: check math correctness, keys, LaTeX render risk.
2. Bahasa Arab reviewer: check Arabic grammar/vocabulary/harakat and MTs appropriateness.
3. Technical renderer reviewer: check JSON, HTML, RTL, KaTeX delimiters.
4. Assessment-quality reviewer: check duplicates, difficulty, cognitive level, distractors, explanations.

**Output:**
- `review-katex.md`
- `review-arabic.md`
- `review-technical.md`
- `review-assessment.md`
- `README.md` with consolidated go/no-go.

**Gate:**
- Do not insert until critical/revisi wajib items are fixed in CSV.

---

## Task 5: Backup and insert as draft

**Objective:** Insert reviewed items safely.

**Files:**
- Create: `apply_system_seed_katex_arabic.sql`
- Create: `before_insert.csv`
- Create: `after_insert.csv`
- Create: DB backup dump path + SHA256.

**SQL rules:**
- Insert only if code does not exist.
- `author_username='system'`.
- `status='draft'`.
- `workflow_status='draft'`.
- `version_group_id = id` if applicable.
- `media_asset_ids = []` unless a row intentionally uses an uploaded asset.

**Verification:**
- Inserted count = 40.
- Duplicate codes = 0.
- Missing explanation = 0.
- PG options mismatch = 0.
- Essay answer_key nonempty = 0.

---

## Task 6: UI/render smoke test

**Objective:** Confirm KaTeX and Arabic display in Bank Soal UI.

**Steps:**
1. Open Bank Soal list/detail for sample KaTeX rows.
2. Confirm superscript/subscript, fractions, radicals, inequalities, and multi-line display math render correctly.
3. Open sample Arabic rows.
4. Confirm RTL text direction, Arabic options, and passages are readable.
5. Capture screenshots for audit.

**Verification:**
- No clipped superscripts/subscripts.
- No literal `\frac` displayed unless intentionally plain text.
- Arabic not reversed/garbled.
- Options align and remain selectable.

---

## Task 7: Final report

**Objective:** Produce Telegram-friendly summary and attach files.

**Files:**
- `README.md`
- `final_verify.tsv`
- `katex_samples.png`
- `arabic_samples.png`
- CSV after insert.

**Report:**
- Counts inserted by subject/type/level.
- KaTeX symbol coverage checklist.
- Arabic model coverage checklist.
- Backup path + SHA256.
- Any known limitations.

---

## Rollback Plan

If needed, remove only rows with the two new prefixes:

```sql
DELETE FROM cbt_questions
WHERE author_username = 'system'
  AND (
    code LIKE 'SYS-MTK-KATEX-%'
    OR code LIKE 'SYS-BARAB-RTL-%'
  );
```

Run only after backup/approval.
