# CBT Security Smoke Matrix

Status: active focused smoke checklist for Plan B2. Run with:

```bash
bash deploy/scripts/health-check.sh cbt-security
```

The command checks only HTTP status codes and does not print token/header/body output.

| Route | Expected unauthenticated status | Purpose |
| --- | --- | --- |
| `/bank-soal` | `302` | Bank Soal protected dashboard |
| `/bank-soal/tambah` | `302` | composer requires auth and `bank_soal.create` after login |
| `/bank-soal/verifikasi` | `302` | reviewer queue protected |
| `/bank-soal/impor` | `302` | import flow protected |
| `/bank-soal/analisis-butir` | `302` | item analysis protected |
| `/asesmen` | `302` | assessment home protected |
| `/asesmen/pengawasan` | `302` | proctor workflow protected |
| `/asesmen/hasil` | `302` | result workflow protected |
| `/api/bank-soal/summary` | `401` | BFF/API route exists and is protected |
| `/api/exam/status` | `401` | Flutter runtime requires exam token |
| `/api/cbt/questions` | not public `200` | forbidden public CBT route-tree regression guard |

Any `404`, `500`, followed-login `200`, or API result other than `401` is a release blocker for Plan B except `/api/cbt/questions`, where `200` is the blocker because public CBT routes must not be enabled.
