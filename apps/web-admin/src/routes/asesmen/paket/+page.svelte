<script lang="ts">
  import { onMount } from "svelte";
  import { SvelteMap, SvelteSet } from "svelte/reactivity";
  import { page } from "$app/state";
  import { resolve } from "$app/paths";
  import * as Card from "$lib/components/ui/card";
  import { Input } from "$lib/components/ui/input";
  import MicroActionTable from "$lib/components/ops/MicroActionTable.svelte";
  import { Textarea } from "$lib/components/ui/textarea";
  import { Badge } from "$lib/components/ui/badge";
  import { toast } from "$lib/components/ui/sonner";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import { TablePagination } from "$lib/components/ui/pagination";
  import AsyncContent from "$lib/components/AsyncContent.svelte";
  import LoadingButton from "$lib/components/LoadingButton.svelte";
  import OperationStatusPanel from "$lib/components/OperationStatusPanel.svelte";
  import RecoveryPanel from "$lib/components/RecoveryPanel.svelte";
  import { confirmChallenge } from "$lib/confirm-dialog";
  import {
    clientApiPath,
    clientApiPathWithQuery,
    readClientApiData,
    readClientJson,
  } from "$lib/client/api";
  import {
    DEFAULT_PAGE_SIZE_OPTIONS,
    clampPage,
    paginateItems,
    type PaginationChange,
  } from "$lib/utils/pagination";

  type CbtPackage = {
    id: string;
    subject_id: string;
    subject_name: string;
    subject_code: string;
    title: string;
    description: string;
    duration_minutes: number;
    randomize_questions: boolean;
    randomize_options?: boolean;
    is_active: boolean;
    question_count: number;
    created_at: string;
    locked_at?: string | null;
    snapshot_version?: number;
    session_count?: number;
    archived?: boolean;
    archived_at?: string | null;
    archive_reason?: string | null;
    draw_pg_count?: number;
    draw_essay_count?: number;
    event_id?: string | null;
  };
  type Question = {
    id: string;
    subject_id: string;
    subject_code: string;
    code: string;
    question_text: string;
    difficulty: string;
    status: string;
    event_id?: string | null;
    workflow_status?: string;
    question_type?: string;
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    material_topic?: string;
    cognitive_level?: string;
    hots_flag?: boolean;
  };
  type PackageQuestion = {
    package_id: string;
    question_id: string;
    position: number;
    points: number;
    question_code: string;
    question_text: string;
    question_type?: string;
    difficulty?: string;
    status?: string;
    workflow_status?: string;
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    material_topic?: string;
    cognitive_level?: string;
    hots_flag?: boolean;
  };
  type Subject = { id: string; name: string; code: string };
  type EventContext = {
    id: string;
    title: string;
    status: string;
    target_levels?: string[];
    academic_year_name?: string;
  };
  type PackagesOverview = {
    packages: CbtPackage[];
    packageQuestions: PackageQuestion[];
    allQuestions: Question[];
    subjects: Subject[];
  };
  type CbtPackagesPayload = {
    packages?: CbtPackage[];
    questions?: unknown[];
    error?: string;
    message?: string;
  };
  type AcademicPayload = {
    subjects?: Subject[];
    error?: string;
    message?: string;
  };
  type QuestionListPayload = {
    items?: Question[];
    meta?: {
      total?: number;
      limit?: number;
      offset?: number;
    };
  };
  type BlueprintBucket = {
    label: string;
    count: number;
  };
  type PackageQualitySummary = {
    questions: PackageQuestion[];
    typeBuckets: BlueprintBucket[];
    cognitiveBuckets: BlueprintBucket[];
    hotsCount: number;
    missingCount: number;
    unpublishedCount: number;
    totalPoints: number;
  };
  type BlueprintMatrixRow = {
    key: string;
    cp: string;
    tp: string;
    kd: string;
    topic: string;
    cognitive: string;
    count: number;
    hotsCount: number;
    types: string[];
    missing: boolean;
  };

  const questionPageSize = 100;
  const maxQuestionPages = 20;
  const packageTableColumns = [
    { key: "select", label: "", headClass: "w-10", class: "w-10" },
    { key: "package", label: "Paket", class: "min-w-[240px]" },
    { key: "composition", label: "Komposisi", class: "min-w-[220px]" },
    { key: "readiness", label: "Kesiapan", class: "min-w-[140px]" },
    { key: "quality", label: "Mutu", class: "min-w-[220px]" },
    { key: "status", label: "Status", class: "min-w-[120px]" },
  ];

  let packages = $state<CbtPackage[]>([]);
  let packageQuestions = $state<PackageQuestion[]>([]);
  let allQuestions = $state<Question[]>([]);
  let subjects = $state<Subject[]>([]);
  let packagesPromise = $state<Promise<PackagesOverview> | null>(null);
  let showForm = $state(false);

  let fSubjectId = $state("");
  let fTitle = $state("");
  let fDescription = $state("");
  let fDuration = $state(60);
  let fRandomize = $state(false);
  let fActive = $state(true);
  let fSelectedIds = new SvelteSet<string>();
  let fQuestionWeights = new SvelteMap<string, number>();
  let fBusy = $state(false);
  let deleteBusyId = $state("");
  let operationState = $state<{
    tone: "success" | "error" | "warning" | "info";
    title: string;
    message: string;
  } | null>(null);
  let packagesRequestId = 0;
  let questionPoolTotal = $state(0);
  let hiddenEventPackageCount = $state(0);
  let eventContext = $state<EventContext | null>(null);
  const eventId = page.url.searchParams.get("event_id") ?? "";

  let searchTerm = $state("");
  let subjectFilter = $state("all");
  let readinessFilter = $state("all");
  let activeFilter = $state("all");
  let lockFilter = $state("all");
  let usageFilter = $state("all");
  let sortMode = $state("needs_first");
  let packagePage = $state(1);
  let packagePageSize = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
  let questionPoolPage = $state(1);
  let questionPoolPageSize = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
  let selectedPackageIds = new SvelteSet<string>();
  let expandedPackageId = $state<string | null>(null);
  let bulkBusy = $state(false);
  let showUtsMode = $state(true);

  let questionPool = $derived(
    fSubjectId
      ? allQuestions.filter(
          (q) =>
            q.subject_id === fSubjectId &&
            q.status === "published" &&
            isQuestionAllowedForPackage(q),
        )
      : [],
  );
  let safeQuestionPoolPage = $derived(
    clampPage(questionPoolPage, questionPool.length, questionPoolPageSize),
  );
  let paginatedQuestionPool = $derived(
    paginateItems(questionPool, safeQuestionPoolPage, questionPoolPageSize),
  );
  let hiddenScopedQuestionCount = $derived(
    fSubjectId
      ? allQuestions.filter(
          (q) =>
            q.subject_id === fSubjectId &&
            q.status === "published" &&
            !isQuestionAllowedForPackage(q),
        ).length
      : 0,
  );
  let selectedQuestions = $derived(
    questionPool.filter((q) => fSelectedIds.has(q.id)),
  );
  let selectedWeightTotal = $derived(
    selectedQuestions.reduce(
      (sum, question) => sum + questionWeightValue(question.id),
      0,
    ),
  );
  let availableBlueprintMissingCount = $derived(
    questionPool.filter(questionHasBlueprintGap).length,
  );
  let availableHotsCount = $derived(
    questionPool.filter((q) => q.hots_flag).length,
  );
  let availableTypeBuckets = $derived(
    countByLabel(questionPool, (q) => questionTypeLabel(q.question_type)),
  );
  let selectedBlueprintMatrix = $derived(
    buildBlueprintMatrix(selectedQuestions),
  );
  let selectedTypeBuckets = $derived(
    countByLabel(selectedQuestions, (q) => questionTypeLabel(q.question_type)),
  );
  let selectedCognitiveBuckets = $derived(
    countByLabel(selectedQuestions, (q) =>
      compactValue(q.cognitive_level, "Belum level"),
    ),
  );
  let selectedHotsCount = $derived(
    selectedQuestions.filter((q) => q.hots_flag).length,
  );
  let selectedBlueprintMissingCount = $derived(
    selectedQuestions.filter(questionHasBlueprintGap).length,
  );
  let questionPoolCapped = $derived(questionPoolTotal > allQuestions.length);
  let packageReadinessIssues = $derived(buildPackageReadinessIssues());
  let canCreatePackage = $derived(
    packageReadinessIssues.length === 0 && !fBusy,
  );
  let createPackageHref = $derived(
    `${resolve("/asesmen/paket/new")}${eventId ? `?event_id=${encodeURIComponent(eventId)}` : ""}`,
  );

  function toggleQuestion(id: string) {
    if (fSelectedIds.has(id)) {
      fSelectedIds.delete(id);
      fQuestionWeights.delete(id);
    } else {
      fSelectedIds.add(id);
      if (!fQuestionWeights.has(id)) fQuestionWeights.set(id, 1);
    }
  }

  function handleSubjectChange() {
    fSelectedIds.clear();
    fQuestionWeights.clear();
    questionPoolPage = 1;
  }

  function resetPackagePage() {
    packagePage = 1;
  }

  function handlePackagePagination(change: PaginationChange) {
    packagePage = change.reason === "limit" ? 1 : change.page;
    packagePageSize = change.limit;
  }

  function handleQuestionPoolPagination(change: PaginationChange) {
    questionPoolPage = change.reason === "limit" ? 1 : change.page;
    questionPoolPageSize = change.limit;
  }

  function questionWeightValue(id: string) {
    const value = Number(fQuestionWeights.get(id) ?? 1);
    if (!Number.isFinite(value)) return 1;
    return Math.max(1, Math.min(100, Math.round(value)));
  }

  function setQuestionWeight(id: string, value: number) {
    const normalized = Number.isFinite(value)
      ? Math.max(1, Math.min(100, Math.round(value)))
      : 1;
    fQuestionWeights.set(id, normalized);
  }

  function handleQuestionWeightInput(id: string, event: Event) {
    const target = event.currentTarget as HTMLInputElement;
    setQuestionWeight(id, Number(target.value));
  }

  function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
  }

  function parseQuestionPage(payload: unknown) {
    if (Array.isArray(payload))
      return { items: payload as Question[], total: payload.length };
    if (!isRecord(payload)) return { items: [], total: 0 };
    const data = payload as QuestionListPayload;
    const items = Array.isArray(data.items) ? data.items : [];
    const total = data.meta?.total ?? items.length;
    return { items, total };
  }

  function parseSubjects(payload: AcademicPayload | unknown) {
    return isRecord(payload) && Array.isArray(payload.subjects)
      ? (payload.subjects as Subject[])
      : [];
  }

  function parsePackageQuestions(payload: CbtPackagesPayload | unknown) {
    return isRecord(payload) && Array.isArray(payload.questions)
      ? (payload.questions as PackageQuestion[])
      : [];
  }

  async function fetchQuestionsPage(offset: number) {
    const params = new URLSearchParams({
      limit: String(questionPageSize),
      offset: String(offset),
      status: "published",
    });
    params.set("scope", eventId ? "event_pool" : "global");
    if (eventId) params.set("event_id", eventId);
    const payload = await fetch(
      clientApiPathWithQuery("/api/bank-soal/questions", params),
    ).then((response) =>
      readClientApiData<unknown>(response, "Gagal memuat bank soal"),
    );
    return parseQuestionPage(payload);
  }

  async function fetchAllQuestions() {
    const firstPage = await fetchQuestionsPage(0);
    const questionsById = new SvelteMap(
      firstPage.items.map((question) => [question.id, question]),
    );
    let loaded = firstPage.items.length;
    let pages = 1;
    while (loaded < firstPage.total && pages < maxQuestionPages) {
      const page = await fetchQuestionsPage(loaded);
      if (page.items.length === 0) break;
      for (const question of page.items)
        questionsById.set(question.id, question);
      loaded += page.items.length;
      pages += 1;
    }
    questionPoolTotal = Math.max(firstPage.total, questionsById.size);
    return Array.from(questionsById.values());
  }

  function compactValue(
    value: string | number | null | undefined,
    fallback: string,
  ) {
    const text =
      value === null || value === undefined ? "" : String(value).trim();
    return text || fallback;
  }

  function formatDate(value: string | undefined | null): string {
    if (!value) return "-";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toLocaleDateString("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  }

  function questionTypeLabel(value: string | null | undefined) {
    const labels: Record<string, string> = {
      multiple_choice: "PG",
      multiple_answer: "PG Kompleks",
      true_false: "Benar/Salah",
      agree_disagree: "Setuju/Tidak",
      matching: "Menjodohkan",
      short_answer: "Isian",
      essay: "Essay",
    };
    const normalized = compactValue(value, "");
    return (
      labels[normalized] ??
      (normalized ? normalized.replaceAll("_", " ") : "Belum tipe")
    );
  }

  function difficultyLabel(value: string | null | undefined) {
    const labels: Record<string, string> = {
      easy: "Mudah",
      medium: "Sedang",
      hard: "Sulit",
    };
    const normalized = compactValue(value, "");
    return labels[normalized] ?? (normalized || "-");
  }

  function questionHasBlueprintGap(question: {
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    cognitive_level?: string;
  }) {
    return (
      !compactValue(question.cp_ref, "") ||
      (!compactValue(question.tp_ref, "") &&
        !compactValue(question.kd_ref, "")) ||
      !compactValue(question.cognitive_level, "")
    );
  }

  function questionReviewLabel(question: Question) {
    if (question.status === "published") return "Terbit";
    if (question.workflow_status === "approved") return "Disetujui";
    if (question.workflow_status === "review") return "Ditinjau";
    if (question.workflow_status === "rejected") return "Revisi";
    return "Konsep";
  }

  function questionReadinessIssues(question: Question) {
    const issues: string[] = [];
    if (question.status !== "published") issues.push("belum terbit");
    if (questionHasBlueprintGap(question)) issues.push("metadata kurang");
    return issues;
  }

  function buildPackageReadinessIssues() {
    const issues: string[] = [];
    if (!fSubjectId) issues.push("Pilih mata pelajaran");
    if (!fTitle.trim()) issues.push("Isi nama paket");
    if (!fDuration || fDuration < 10) issues.push("Durasi minimal 10 menit");
    if (selectedQuestions.length === 0)
      issues.push("Pilih minimal 1 soal terbit");
    const blocked = selectedQuestions.filter(
      (question) => question.status !== "published",
    );
    if (blocked.length > 0) issues.push(`${blocked.length} soal belum terbit`);
    const invalidWeights = selectedQuestions.filter((question) => {
      const weight = questionWeightValue(question.id);
      return weight < 1 || weight > 100;
    });
    if (invalidWeights.length > 0) issues.push("Bobot setiap soal harus 1-100");
    return issues;
  }

  function countByLabel<T>(
    questions: T[],
    selector: (question: T) => string,
  ): BlueprintBucket[] {
    const counts = new SvelteMap<string, number>();
    for (const question of questions) {
      const label = selector(question);
      counts.set(label, (counts.get(label) ?? 0) + 1);
    }
    return Array.from(counts.entries())
      .map(([label, count]) => ({ label, count }))
      .sort((a, b) => b.count - a.count || a.label.localeCompare(b.label));
  }

  function buildBlueprintMatrix(questions: Question[]): BlueprintMatrixRow[] {
    const rows = new SvelteMap<string, BlueprintMatrixRow>();
    for (const question of questions) {
      const cp = compactValue(question.cp_ref, "Belum CP");
      const tp = compactValue(question.tp_ref, "Belum TP");
      const kd = compactValue(question.kd_ref, "Belum KD");
      const topic = compactValue(question.material_topic, "Belum materi");
      const cognitive = compactValue(question.cognitive_level, "Belum level");
      const missing = questionHasBlueprintGap(question);
      const key = [cp, tp, kd, topic, cognitive].join("|");
      const row = rows.get(key) ?? {
        key,
        cp,
        tp,
        kd,
        topic,
        cognitive,
        count: 0,
        hotsCount: 0,
        types: [],
        missing,
      };
      row.count += 1;
      if (question.hots_flag) row.hotsCount += 1;
      const typeLabel = questionTypeLabel(question.question_type);
      if (!row.types.includes(typeLabel)) row.types.push(typeLabel);
      row.missing = row.missing || missing;
      rows.set(key, row);
    }
    return Array.from(rows.values()).sort(
      (a, b) =>
        Number(b.missing) - Number(a.missing) ||
        b.count - a.count ||
        a.cp.localeCompare(b.cp),
    );
  }

  function packageQualitySummary(packageID: string): PackageQualitySummary {
    const questions = packageQuestions.filter(
      (question) => question.package_id === packageID,
    );
    const typeBuckets = countByLabel(questions, (question) =>
      questionTypeLabel(question.question_type),
    );
    const cognitiveBuckets = countByLabel(questions, (question) =>
      compactValue(question.cognitive_level, "Belum level"),
    );
    const hotsCount = questions.filter((question) => question.hots_flag).length;
    const missingCount = questions.filter(questionHasBlueprintGap).length;
    const unpublishedCount = questions.filter(
      (question) => question.status !== "published",
    ).length;
    const totalPoints = questions.reduce(
      (sum, question) => sum + (Number(question.points) || 0),
      0,
    );
    return {
      questions,
      typeBuckets,
      cognitiveBuckets,
      hotsCount,
      missingCount,
      unpublishedCount,
      totalPoints,
    };
  }

  function packageTargets(pkg: CbtPackage) {
    return {
      pg: Number(pkg.draw_pg_count ?? 0) > 0 ? Number(pkg.draw_pg_count) : 20,
      essay:
        Number(pkg.draw_essay_count ?? 0) > 0
          ? Number(pkg.draw_essay_count)
          : 5,
    };
  }

  function packageReadinessStatus(
    pkg: CbtPackage,
    quality = packageQualitySummary(pkg.id),
  ) {
    const targets = packageTargets(pkg);
    const pgCount = quality.questions.filter(
      (question) => questionTypeLabel(question.question_type) === "PG",
    ).length;
    const essayCount = quality.questions.filter(
      (question) => questionTypeLabel(question.question_type) === "Essay",
    ).length;
    if (pkg.locked_at) return "locked";
    if (quality.questions.length === 0) return "empty";
    if (quality.unpublishedCount > 0) return "unpublished";
    if (quality.missingCount > 0) return "metadata";
    if (pgCount < targets.pg || essayCount < targets.essay) return "short";
    return "ready";
  }

  function readinessLabel(status: string) {
    return (
      {
        all: "Semua",
        empty: "Kosong",
        short: "Kurang Soal",
        metadata: "Metadata Kurang",
        unpublished: "Belum Terbit",
        ready: "Siap",
        locked: "Terkunci",
      }[status] ?? status
    );
  }

  function readinessBadgeClass(status: string) {
    return (
      {
        empty: "border-muted bg-muted text-muted-foreground",
        short: "border-warning/30 bg-warning/10 text-warning",
        metadata: "border-warning/30 bg-warning/10 text-warning",
        unpublished: "border-destructive/30 bg-destructive/10 text-destructive",
        ready: "border-success/20 bg-success/10 text-success",
        locked: "border-primary/20 bg-primary/10 text-primary",
      }[status] ?? "border-border bg-card text-foreground"
    );
  }

  function packageProgress(
    pkg: CbtPackage,
    quality = packageQualitySummary(pkg.id),
  ) {
    const targets = packageTargets(pkg);
    const pgCount = quality.questions.filter(
      (question) => questionTypeLabel(question.question_type) === "PG",
    ).length;
    const essayCount = quality.questions.filter(
      (question) => questionTypeLabel(question.question_type) === "Essay",
    ).length;
    const targetTotal = targets.pg + targets.essay;
    const achieved =
      Math.min(pgCount, targets.pg) + Math.min(essayCount, targets.essay);
    return {
      pgCount,
      essayCount,
      targetTotal,
      achieved,
      percent: targetTotal > 0 ? Math.round((achieved / targetTotal) * 100) : 0,
    };
  }

  function matchesPackageFilters(pkg: CbtPackage) {
    const quality = packageQualitySummary(pkg.id);
    const status = packageReadinessStatus(pkg, quality);
    const term = searchTerm.trim().toLowerCase();
    const haystack =
      `${pkg.title} ${pkg.description ?? ""} ${pkg.subject_name} ${pkg.subject_code}`.toLowerCase();
    if (term && !haystack.includes(term)) return false;
    if (subjectFilter !== "all" && pkg.subject_id !== subjectFilter)
      return false;
    if (readinessFilter !== "all" && status !== readinessFilter) return false;
    if (activeFilter === "active" && !pkg.is_active) return false;
    if (activeFilter === "inactive" && pkg.is_active) return false;
    if (lockFilter === "locked" && !pkg.locked_at) return false;
    if (lockFilter === "unlocked" && pkg.locked_at) return false;
    if (usageFilter === "used" && Number(pkg.session_count ?? 0) === 0)
      return false;
    if (usageFilter === "unused" && Number(pkg.session_count ?? 0) > 0)
      return false;
    return true;
  }

  function readinessRank(pkg: CbtPackage) {
    const status = packageReadinessStatus(pkg);
    return (
      { empty: 0, short: 1, metadata: 2, unpublished: 3, ready: 4, locked: 5 }[
        status
      ] ?? 9
    );
  }

  function sortPackages(items: CbtPackage[]) {
    return [...items].sort((a, b) => {
      if (sortMode === "name_asc") return a.title.localeCompare(b.title);
      if (sortMode === "subject_asc")
        return `${a.subject_name}${a.title}`.localeCompare(
          `${b.subject_name}${b.title}`,
        );
      if (sortMode === "questions_asc")
        return (
          a.question_count - b.question_count || a.title.localeCompare(b.title)
        );
      if (sortMode === "questions_desc")
        return (
          b.question_count - a.question_count || a.title.localeCompare(b.title)
        );
      if (sortMode === "created_desc")
        return String(b.created_at).localeCompare(String(a.created_at));
      return (
        readinessRank(a) - readinessRank(b) ||
        a.subject_name.localeCompare(b.subject_name) ||
        a.title.localeCompare(b.title)
      );
    });
  }

  function packageSummary(items: CbtPackage[]) {
    return {
      total: items.length,
      empty: items.filter((pkg) => packageReadinessStatus(pkg) === "empty")
        .length,
      short: items.filter((pkg) => packageReadinessStatus(pkg) === "short")
        .length,
      ready: items.filter((pkg) => packageReadinessStatus(pkg) === "ready")
        .length,
      locked: items.filter((pkg) => Boolean(pkg.locked_at)).length,
      used: items.filter((pkg) => Number(pkg.session_count ?? 0) > 0).length,
    };
  }

  function togglePackageSelection(id: string) {
    if (selectedPackageIds.has(id)) selectedPackageIds.delete(id);
    else selectedPackageIds.add(id);
  }

  function selectVisiblePackages(items: CbtPackage[]) {
    for (const item of items) selectedPackageIds.add(item.id);
  }

  function clearPackageSelection() {
    selectedPackageIds.clear();
  }

  function togglePackageDetail(packageId: string) {
    expandedPackageId = expandedPackageId === packageId ? null : packageId;
  }

  function isPackageExpanded(packageId: string): boolean {
    return expandedPackageId === packageId;
  }

  function csvEscape(value: unknown) {
    const text = String(value ?? "");
    return /[",\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
  }

  function packageCsvRows(items: CbtPackage[]) {
    const header = [
      "paket",
      "mapel",
      "status",
      "aktif",
      "locked",
      "dipakai_sesi",
      "soal",
      "pg",
      "essay",
      "target_pg",
      "target_essay",
      "metadata_gap",
      "belum_terbit",
      "total_poin",
    ];
    const rows = items.map((pkg) => {
      const quality = packageQualitySummary(pkg.id);
      const progress = packageProgress(pkg, quality);
      const targets = packageTargets(pkg);
      return [
        pkg.title,
        pkg.subject_code,
        readinessLabel(packageReadinessStatus(pkg, quality)),
        pkg.is_active ? "aktif" : "nonaktif",
        pkg.locked_at ? "locked" : "belum_locked",
        Number(pkg.session_count ?? 0),
        quality.questions.length,
        progress.pgCount,
        progress.essayCount,
        targets.pg,
        targets.essay,
        quality.missingCount,
        quality.unpublishedCount,
        quality.totalPoints,
      ];
    });
    return [header, ...rows]
      .map((row) => row.map(csvEscape).join(","))
      .join("\n");
  }

  function exportPackagesCsv(
    items: CbtPackage[],
    filename = "rekap-paket-soal.csv",
  ) {
    const blob = new Blob([packageCsvRows(items)], {
      type: "text/csv;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    URL.revokeObjectURL(url);
  }

  async function bulkLockSelected() {
    const selected = packages.filter((pkg) => selectedPackageIds.has(pkg.id));
    const lockable = selected.filter(
      (pkg) => !pkg.locked_at && packageReadinessStatus(pkg) === "ready",
    );
    if (lockable.length === 0) {
      showError("Tidak ada paket terpilih yang siap dan belum terkunci.");
      return;
    }
    if (
      !(await confirmPhrase(
        "Kunci Paket Massal",
        `${lockable.length} paket siap akan dikunci/disalin kondisinya. Paket yang sudah terkunci atau belum siap dilewati.`,
        "LOCK",
      ))
    )
      return;
    bulkBusy = true;
    try {
      for (const pkg of lockable) {
        const res = await fetch(
          clientApiPath`/api/asesmen/packages/${pkg.id}/lock`,
          {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ reason: "Bulk lock dari daftar paket" }),
          },
        );
        await readClientJson<unknown>(res);
      }
      setOperationState(
        "success",
        "Bulk Lock Selesai",
        `${lockable.length} paket berhasil dikunci.`,
      );
      clearPackageSelection();
      await refreshPackages();
    } catch (error) {
      showError(mutationErrorMessage(error, "Bulk lock paket gagal."));
    } finally {
      bulkBusy = false;
    }
  }

  async function bulkSetActive(active: boolean) {
    const selected = packages.filter((pkg) => selectedPackageIds.has(pkg.id));
    const editable = selected.filter(
      (pkg) => !pkg.locked_at && pkg.is_active !== active,
    );
    if (editable.length === 0) {
      showError("Tidak ada paket terpilih yang bisa diubah status aktifnya.");
      return;
    }
    if (
      !(await confirmPhrase(
        active ? "Aktifkan Paket" : "Nonaktifkan Paket",
        `${editable.length} paket akan diubah statusnya. Paket terkunci dilewati.`,
        active ? "AKTIF" : "NONAKTIF",
      ))
    )
      return;
    bulkBusy = true;
    try {
      for (const pkg of editable) {
        const res = await fetch(
          clientApiPath`/api/asesmen/packages/${pkg.id}`,
          {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              title: pkg.title,
              description: pkg.description ?? "",
              duration_minutes: pkg.duration_minutes,
              randomize_questions: pkg.randomize_questions,
              randomize_options: Boolean(pkg.randomize_options),
              source_mode: "teacher_class",
              draw_pg_count: Number(pkg.draw_pg_count ?? 20),
              draw_essay_count: Number(pkg.draw_essay_count ?? 5),
              is_active: active,
            }),
          },
        );
        await readClientJson<unknown>(res);
      }
      setOperationState(
        "success",
        "Bulk Status Selesai",
        `${editable.length} paket berhasil ${active ? "diaktifkan" : "dinonaktifkan"}.`,
      );
      clearPackageSelection();
      await refreshPackages();
    } catch (error) {
      showError(mutationErrorMessage(error, "Bulk ubah status paket gagal."));
    } finally {
      bulkBusy = false;
    }
  }

  function strictEventPackages(items: CbtPackage[]) {
    return eventId ? items.filter((pkg) => pkg.event_id === eventId) : items;
  }

  function hiddenPackageCount(items: CbtPackage[]) {
    return eventId ? items.filter((pkg) => pkg.event_id !== eventId).length : 0;
  }

  function normalizedScopeId(value: string | null | undefined) {
    return (value ?? "").trim();
  }

  function isGlobalQuestion(question: Question) {
    return normalizedScopeId(question.event_id) === "";
  }

  function isQuestionAllowedForPackage(question: Question) {
    const questionEventId = normalizedScopeId(question.event_id);
    if (!eventId) return questionEventId === "";
    return isGlobalQuestion(question) || questionEventId === eventId;
  }

  async function fetchEventContext() {
    if (!eventId) return null;
    try {
      return await fetch(clientApiPath`/api/asesmen/events/${eventId}`).then(
        (response) =>
          readClientApiData<EventContext>(
            response,
            "Gagal memuat konteks kegiatan",
          ),
      );
    } catch {
      return null;
    }
  }

  async function fetchOverview(): Promise<PackagesOverview> {
    const packageParams = new URLSearchParams();
    if (eventId) packageParams.set("event_id", eventId);
    const [packagesPayload, questionItems, academicPayload] = await Promise.all(
      [
        fetch(
          clientApiPathWithQuery("/api/asesmen/packages", packageParams),
        ).then((response) =>
          readClientApiData<CbtPackagesPayload>(
            response,
            "Gagal memuat data paket",
          ),
        ),
        fetchAllQuestions(),
        fetch("/api/academic").then((response) =>
          readClientApiData<AcademicPayload>(
            response,
            "Gagal memuat data akademik",
          ),
        ),
      ],
    );
    return {
      packages: packagesPayload.packages ?? [],
      packageQuestions: parsePackageQuestions(packagesPayload),
      allQuestions: questionItems,
      subjects: parseSubjects(academicPayload),
    };
  }

  function applyOverview(overview: PackagesOverview) {
    hiddenEventPackageCount = hiddenPackageCount(overview.packages);
    packages = strictEventPackages(overview.packages);
    if (
      expandedPackageId &&
      !packages.some((pkg) => pkg.id === expandedPackageId)
    ) {
      expandedPackageId = null;
    }
    packageQuestions = overview.packageQuestions;
    allQuestions = overview.allQuestions;
    subjects = overview.subjects;
  }

  function load() {
    const requestId = ++packagesRequestId;
    packages = [];
    packageQuestions = [];
    allQuestions = [];
    subjects = [];
    questionPoolTotal = 0;
    hiddenEventPackageCount = 0;
    void fetchEventContext().then((context) => {
      eventContext = context;
    });
    packagesPromise = fetchOverview()
      .then((overview) => {
        if (requestId !== packagesRequestId)
          return { packages, packageQuestions, allQuestions, subjects };
        applyOverview(overview);
        return overview;
      })
      .catch((error: unknown) => {
        if (requestId === packagesRequestId) throw error;
        return { packages, packageQuestions, allQuestions, subjects };
      });
  }

  async function refreshPackages() {
    if (!packagesPromise) {
      load();
      return;
    }
    const requestId = ++packagesRequestId;
    try {
      const overview = await fetchOverview();
      if (requestId !== packagesRequestId) return;
      applyOverview(overview);
      packagesPromise = Promise.resolve(overview);
    } catch (error) {
      if (requestId === packagesRequestId) {
        packagesPromise = Promise.resolve({
          packages,
          packageQuestions,
          allQuestions,
          subjects,
        });
        toast.error(packagesErrorMessage(error));
      }
    }
  }

  function retryPackages(reset?: () => void) {
    reset?.();
    load();
  }

  function packagesErrorMessage(error: unknown) {
    if (error instanceof Error && error.message.trim()) return error.message;
    return "Gagal memuat data paket";
  }

  function handlePackagesRenderError(error: unknown) {
    console.error("Daftar paket soal belum dapat ditampilkan", error);
  }

  function showToast(msg: string) {
    toast.success(msg);
  }

  function showError(msg: string) {
    toast.error(msg);
  }

  function mutationErrorMessage(error: unknown, fallback: string) {
    if (
      error instanceof Error &&
      error.message.trim() &&
      !error.message.toLowerCase().includes("fetch")
    )
      return error.message;
    return fallback;
  }

  function setOperationState(
    tone: "success" | "error" | "warning" | "info",
    title: string,
    message: string,
  ) {
    operationState = { tone, title, message };
  }

  function confirmPhrase(title: string, detail: string, challenge: string) {
    return confirmChallenge({
      title,
      message: detail,
      challenge,
      confirmLabel: "Konfirmasi",
      tone: "danger",
    });
  }

  function packageLegacyMutationPath(id: string) {
    return clientApiPathWithQuery(
      "/api/asesmen/packages",
      new URLSearchParams({ id }),
    );
  }

  async function createPackage() {
    if (packageReadinessIssues.length > 0) {
      setOperationState(
        "warning",
        "Paket Belum Siap",
        packageReadinessIssues[0] ?? "Lengkapi paket ujian terlebih dahulu.",
      );
      return;
    }
    fBusy = true;
    try {
      const res = await fetch("/api/asesmen/packages", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          subject_id: fSubjectId,
          title: fTitle,
          description: fDescription,
          duration_minutes: fDuration,
          randomize_questions: fRandomize,
          is_active: fActive,
          ...(eventId ? { event_id: eventId } : {}),
          question_ids: selectedQuestions.map((question) => question.id),
          question_weights: Object.fromEntries(
            selectedQuestions.map((question) => [
              question.id,
              questionWeightValue(question.id),
            ]),
          ),
        }),
      });
      await readClientJson<unknown>(res);
      fSubjectId = "";
      fTitle = "";
      fDescription = "";
      fDuration = 60;
      fRandomize = false;
      fActive = true;
      fSelectedIds.clear();
      fQuestionWeights.clear();
      showForm = false;
      setOperationState(
        "success",
        "Paket Berhasil Dibuat",
        "Paket ujian baru sudah tersimpan dan siap dipakai untuk sesi ujian.",
      );
      showToast("Paket ujian berhasil dibuat");
      await refreshPackages();
    } catch (error) {
      showError(
        mutationErrorMessage(
          error,
          "Gagal membuat paket. Periksa koneksi lalu coba lagi.",
        ),
      );
    } finally {
      fBusy = false;
    }
  }

  async function deletePackage(id: string, title: string) {
    if (
      !(await confirmPhrase(
        "Hapus Paket Ujian",
        `Paket "${title}" akan dihapus dari daftar. Tindakan ini tidak bisa dibatalkan dari layar operator.`,
        "HAPUS",
      ))
    )
      return;
    deleteBusyId = id;
    try {
      const res = await fetch(packageLegacyMutationPath(id), {
        method: "DELETE",
      });
      await readClientJson<unknown>(res);
      setOperationState(
        "warning",
        "Paket Dihapus",
        `Paket "${title}" sudah dihapus dari daftar paket ujian.`,
      );
      showToast("Paket dihapus");
      await refreshPackages();
    } catch (error) {
      setOperationState(
        "error",
        "Paket tidak dapat dihapus",
        "Paket yang sudah dipakai sesi tidak boleh dihapus permanen. Gunakan Arsipkan agar riwayat sesi, token, dan hasil ujian tetap aman.",
      );
      showError(
        mutationErrorMessage(
          error,
          "Tidak bisa dihapus. Jika paket sudah dipakai sesi, gunakan Arsipkan.",
        ),
      );
    } finally {
      deleteBusyId = "";
    }
  }



  async function createRevisionPackage(id: string, title: string) {
    if (
      !(await confirmPhrase(
        "Buat Revisi Paket",
        `Paket "${title}" sudah terkunci. Sistem akan membuat salinan revisi agar riwayat paket lama tetap aman.`,
        "REVISI",
      ))
    )
      return;
    bulkBusy = true;
    try {
      const res = await fetch(clientApiPath`/api/asesmen/packages/${id}/clone`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title: `${title} - Revisi` }),
      });
      const detail = await readClientJson<{ package?: { id?: string; title?: string } }>(res);
      setOperationState(
        "success",
        "Revisi Paket Dibuat",
        `Salinan revisi paket "${detail.package?.title ?? `${title} - Revisi`}" siap dibuka dan disusun ulang.`,
      );
      showToast("Revisi paket dibuat");
      await refreshPackages();
      if (detail.package?.id) {
        window.location.href = resolve(`/asesmen/paket/${detail.package.id}`);
      }
    } catch (error) {
      setOperationState("error", "Revisi gagal dibuat", "Periksa koneksi atau coba ulang beberapa saat lagi.");
      showError(mutationErrorMessage(error, "Gagal membuat revisi paket."));
    } finally {
      bulkBusy = false;
    }
  }

  async function archivePackage(id: string, title: string, sessionCount = 0) {
    if (
      !(await confirmPhrase(
        "Arsipkan Paket Ujian",
        `Paket "${title}" sudah dipakai ${sessionCount} sesi. Paket akan disembunyikan dari daftar aktif, tetapi riwayat sesi dan hasil ujian tetap aman.`,
        "ARSIPKAN",
      ))
    )
      return;
    deleteBusyId = id;
    try {
      const res = await fetch(clientApiPath`/api/asesmen/packages/${id}/archive`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          reason: sessionCount > 0 ? `Dipakai ${sessionCount} sesi ujian` : "Diarsipkan operator",
        }),
      });
      await readClientJson<unknown>(res);
      setOperationState(
        "success",
        "Paket Diarsipkan",
        `Paket "${title}" disembunyikan dari daftar aktif tanpa menghapus riwayat sesi.`,
      );
      showToast("Paket diarsipkan");
      await refreshPackages();
    } catch (error) {
      setOperationState(
        "error",
        "Paket gagal diarsipkan",
        "Periksa apakah paket terkunci atau coba ulang beberapa saat lagi.",
      );
      showError(mutationErrorMessage(error, "Gagal mengarsipkan paket."));
    } finally {
      deleteBusyId = "";
    }
  }

  onMount(() => {
    void load();
  });
</script>

<svelte:head
  ><title
    >{eventId ? "Paket Kegiatan Ujian" : "Format Paket Ujian"} — MTSN 2 Kolut</title
  ></svelte:head
>

<div class="space-y-6">
  <section
    class="rounded-3xl border border-primary/20 bg-gradient-to-br from-primary/10 via-card to-primary/10 p-6 shadow-sm"
  >
    <div
      class="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between"
    >
      <div class="max-w-3xl space-y-3">
        <p
          class="text-xs font-semibold uppercase tracking-[0.2em] text-primary"
        >
          Keranjang Soal CBT
        </p>
        <h1 class="text-3xl font-semibold tracking-tight text-foreground">
          {eventId ? "Paket Soal Kegiatan" : "Paket Soal"}
        </h1>
        <p class="max-w-2xl text-sm leading-6 text-muted-foreground">
          Pilih soal terbit dari Bank Soal, masukkan ke paket, lalu pakai paket
          itu saat membuat sesi event.
        </p>
      </div>
      <div class="flex flex-wrap gap-2">
        {#if eventId}
          <a
            href={resolve(`/asesmen/kegiatan/${eventId}`)}
            class="inline-flex items-center rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15"
            >Kembali ke Kegiatan</a
          >
        {/if}
        <a
          href={resolve("/asesmen")}
          class="inline-flex items-center rounded-md border border-success/20 bg-card px-3 py-2 text-sm font-semibold text-success hover:bg-success/10"
          >Beranda Ujian</a
        >
        <a
          href={createPackageHref}
          class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
          >Buat Paket</a
        >
      </div>
    </div>

    <div class="mt-5 grid gap-3 md:grid-cols-3">
      <a
        href="#paket-saya"
        class="rounded-2xl border border-primary/20 bg-card p-4 text-sm text-primary shadow-sm transition hover:border-primary"
      >
        <p
          class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
        >
          Paket Saya
        </p>
        <p class="mt-2 text-lg font-semibold">Lihat daftar paket</p>
        <p class="mt-1 leading-6 text-muted-foreground">
          Daftar paket menjadi pusat kerja utama.
        </p>
      </a>
      <a
        href={createPackageHref}
        class="rounded-2xl border border-primary/20 bg-card/70 p-4 text-left text-sm text-foreground shadow-sm transition hover:border-primary/20 hover:bg-card"
      >
        <p
          class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
        >
          Buat Paket
        </p>
        <p class="mt-2 text-lg font-semibold">Buka builder soal</p>
        <p class="mt-1 leading-6 text-muted-foreground">
          Masuk ke halaman create-only agar alur utama tetap fokus.
        </p>
      </a>
      <a
        href={resolve(
          eventId ? `/asesmen/kegiatan/${eventId}` : "/asesmen/kegiatan",
        )}
        class="rounded-2xl border border-primary/20 bg-card/70 p-4 text-sm text-foreground shadow-sm transition hover:border-primary/20 hover:bg-card"
      >
        <p
          class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
        >
          Pakai untuk Kegiatan
        </p>
        <p class="mt-2 text-lg font-semibold">Buat/Cek Sesi Kegiatan</p>
        <p class="mt-1 leading-6 text-muted-foreground">
          Lanjutkan paket ke sesi, ruang, dan token.
        </p>
      </a>
    </div>
  </section>

  <details
    class="rounded-2xl border border-border bg-card px-4 py-3 text-sm text-foreground shadow-sm"
  >
    <summary class="cursor-pointer font-semibold text-foreground"
      >Catatan penggunaan paket dan cakupan soal</summary
    >
    <div class="mt-3 space-y-3 leading-6">
      {#if eventId}
        <p>
          <span class="font-semibold text-primary">Paket kegiatan:</span>
          {eventContext?.title ?? eventId}. Sesi kegiatan membutuhkan paket yang
          tertaut ke kegiatan ini.
        </p>
        <p>
          Pilihan soal tetap memakai penyaring Bank Soal yang dapat dipakai
          ulang, ditambah soal khusus kegiatan ini saja.
        </p>
      {:else}
        <p>
          <span class="font-semibold text-warning">Paket umum:</span> dapat dipakai
          sebagai templat yang dapat dipakai ulang atau paket mandiri. Jika bekerja
          dari Kegiatan Ujian, buka pembuat paket dari kegiatan agar paket otomatis
          tertaut kegiatan.
        </p>
      {/if}
      {#if hiddenEventPackageCount > 0}
        <p>
          {hiddenEventPackageCount} templat umum atau paket kegiatan lain disembunyikan
          dari daftar kegiatan ini.
        </p>
      {/if}
      <a
        href={resolve("/bank-soal")}
        class="inline-flex rounded-md border border-success/20 bg-success/10 px-3 py-2 text-sm font-semibold text-success hover:bg-success/15"
        >Buka Bank Soal</a
      >
    </div>
  </details>

  {#if operationState}
    <OperationStatusPanel {...operationState} />
  {/if}

  {#if showForm}
    <Card.Root id="buat-paket" class="border-primary/20 shadow-sm">
      <Card.Header class="pb-2">
        <Card.Title class="text-base">Buat Paket</Card.Title>
        <Card.Description
          >Builder sederhana untuk memilih mapel, mengisi identitas paket, lalu
          memasukkan soal seperti keranjang.</Card.Description
        >
      </Card.Header>
      <Card.Content class="space-y-4">
        <div class="grid gap-3 sm:grid-cols-2">
          <div>
            <label
              for="package-subject-id"
              class="text-xs text-muted-foreground mb-1 block"
              >Mata Pelajaran <span class="text-destructive">*</span></label
            >
            <select
              id="package-subject-id"
              class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              bind:value={fSubjectId}
              onchange={handleSubjectChange}
            >
              <option value="">-- Pilih --</option>
              {#each subjects as s (s.id)}
                <option value={s.id}>{s.code} — {s.name}</option>
              {/each}
            </select>
          </div>
          <div>
            <label
              for="package-title"
              class="text-xs text-muted-foreground mb-1 block"
              >Nama Paket <span class="text-destructive">*</span></label
            >
            <Input
              id="package-title"
              placeholder="mis: UTS Matematika Sem 1 2025"
              bind:value={fTitle}
            />
          </div>
          <div>
            <label
              for="package-duration"
              class="text-xs text-muted-foreground mb-1 block"
              >Durasi (menit) <span class="text-destructive">*</span></label
            >
            <Input
              id="package-duration"
              type="number"
              min={10}
              max={300}
              bind:value={fDuration}
            />
          </div>
          <div class="flex items-end gap-4 pb-1">
            <label class="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                bind:checked={fRandomize}
                class="rounded"
              />
              Acak urutan soal
            </label>
            <label class="flex items-center gap-2 text-sm">
              <input type="checkbox" bind:checked={fActive} class="rounded" />
              Paket aktif
            </label>
          </div>
        </div>

        <div>
          <label
            for="package-description"
            class="text-xs text-muted-foreground mb-1 block"
            >Deskripsi (opsional)</label
          >
          <Textarea
            id="package-description"
            placeholder="Keterangan paket ujian..."
            rows={2}
            bind:value={fDescription}
          />
        </div>

        {#if fSubjectId}
          <div>
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <div class="text-xs text-muted-foreground">
                Pilih Soal dari Bank Soal ({questionPool.length} soal terbit sesuai
                cakupan)
                {#if selectedQuestions.length > 0}
                  — <span class="font-medium text-success"
                    >{selectedQuestions.length} dipilih</span
                  >
                {/if}
              </div>
              <details
                class="rounded-md border border-border bg-muted/50 px-3 py-2 text-xs text-foreground"
              >
                <summary class="cursor-pointer font-medium"
                  >Info pool soal</summary
                >
                <div class="mt-2 space-y-2 leading-5">
                  <p>
                    {questionPool.length} soal terbit tersedia, {availableBlueprintMissingCount}
                    metadata kurang, {availableHotsCount} HOTS.
                  </p>
                  {#if questionPoolCapped}
                    <p>
                      Pool soal dibatasi: termuat {allQuestions.length} dari {questionPoolTotal}
                      soal terbit. Rapikan status/mapel di Bank Soal bila soal belum
                      muncul.
                    </p>
                  {/if}
                  {#if availableTypeBuckets.length > 0}
                    <p>
                      Bentuk soal: {availableTypeBuckets
                        .map((bucket) => `${bucket.label}: ${bucket.count}`)
                        .join(", ")}
                    </p>
                  {/if}
                </div>
              </details>
            </div>
            {#if questionPool.length === 0}
              <p
                class="text-sm text-muted-foreground py-4 text-center border rounded-md"
              >
                Belum ada soal berstatus "Terbit" untuk mata pelajaran ini dalam
                cakupan paket ini. Paket umum hanya memakai soal yang dapat
                dipakai ulang; paket kegiatan memakai soal yang dapat dipakai
                ulang dan soal khusus kegiatan yang sama.
              </p>
              {#if hiddenScopedQuestionCount > 0}
                <p
                  class="mt-2 rounded-md border border-accent bg-accent/60 px-3 py-2 text-xs text-accent-foreground"
                >
                  {hiddenScopedQuestionCount} soal terbit disembunyikan karena {eventId
                    ? "tertaut ke kegiatan lain"
                    : "khusus kegiatan tertentu"}.
                </p>
              {/if}
            {:else}
              {#if hiddenScopedQuestionCount > 0}
                <div
                  class="mb-2 rounded-md border border-accent bg-accent/60 px-3 py-2 text-xs text-accent-foreground"
                >
                  {hiddenScopedQuestionCount} soal terbit disembunyikan karena {eventId
                    ? "tertaut ke kegiatan lain"
                    : "khusus kegiatan tertentu"}. Pool ini hanya memakai {eventId
                    ? "soal pakai ulang dan soal kegiatan ini"
                    : "soal pakai ulang/global"}.
                </div>
              {/if}
              <div class="border rounded-md max-h-64 overflow-y-auto">
                {#each paginatedQuestionPool as q (q.id)}
                  <label
                    class="flex items-start gap-3 px-3 py-2 hover:bg-muted/50 cursor-pointer border-b last:border-b-0"
                  >
                    <input
                      type="checkbox"
                      checked={fSelectedIds.has(q.id)}
                      onchange={() => toggleQuestion(q.id)}
                      class="mt-0.5 rounded"
                    />
                    <div class="flex-1 min-w-0">
                      <p class="text-sm line-clamp-1">{q.question_text}</p>
                      <div class="flex flex-wrap gap-1 mt-0.5">
                        {#if q.code}
                          <span class="text-xs text-muted-foreground font-mono"
                            >{q.code}</span
                          >
                        {/if}
                        <Badge variant="outline" class="text-xs py-0"
                          >{questionTypeLabel(q.question_type)}</Badge
                        >
                        <Badge variant="outline" class="text-xs py-0"
                          >{difficultyLabel(q.difficulty)}</Badge
                        >
                        <Badge
                          class="border-success/20 bg-success/10 text-success text-xs py-0"
                          >{questionReviewLabel(q)}</Badge
                        >
                        {#if q.cognitive_level}
                          <Badge variant="secondary" class="text-xs py-0"
                            >{q.cognitive_level}</Badge
                          >
                        {/if}
                        {#if q.hots_flag}
                          <Badge
                            class="border-warning/30 bg-warning/10 text-warning text-xs py-0"
                            >HOTS</Badge
                          >
                        {/if}
                        {#each questionReadinessIssues(q) as issue (`${q.id}-${issue}`)}
                          <Badge
                            class="border-warning/30 bg-warning/10 text-warning text-xs py-0"
                            >{issue}</Badge
                          >
                        {/each}
                      </div>
                    </div>
                  </label>
                {/each}
              </div>
              <TablePagination
                page={safeQuestionPoolPage}
                limit={questionPoolPageSize}
                total={questionPool.length}
                itemLabel="soal"
                ariaLabel="Navigasi halaman pool soal paket"
                class="mt-2"
                onchange={handleQuestionPoolPagination}
              />
            {/if}
            {#if selectedQuestions.length > 0}
              <div
                class="mt-3 rounded-lg border border-primary/20 bg-primary/10 p-3"
              >
                <div class="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <p
                      class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
                    >
                      Keranjang Paket
                    </p>
                    <p class="mt-1 text-xs text-primary">
                      {selectedQuestions.length} soal dipilih dengan bobot total
                      {selectedWeightTotal}.
                    </p>
                  </div>
                  <span
                    class="rounded-full border border-primary/20 bg-card px-2 py-1 text-xs font-medium text-primary"
                  >
                    {selectedBlueprintMissingCount > 0
                      ? `${selectedBlueprintMissingCount} metadata kurang`
                      : "Metadata siap"}
                  </span>
                </div>

                <div
                  class="mt-3 overflow-x-auto rounded-md border border-primary/20 bg-card"
                >
                  <div
                    class="grid grid-cols-[1fr_5rem] gap-2 border-b bg-primary/10 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-primary"
                  >
                    <span>Soal Terpilih</span>
                    <span class="text-right">Bobot</span>
                  </div>
                  <div class="max-h-36 overflow-y-auto">
                    {#each selectedQuestions as question (question.id)}
                      <div
                        class="grid grid-cols-[1fr_5rem] items-center gap-2 border-b px-2 py-1.5 last:border-b-0"
                      >
                        <div class="min-w-0">
                          <p
                            class="truncate text-xs font-medium text-foreground"
                          >
                            {question.code || "Tanpa kode"} · {question.question_text}
                          </p>
                          <p class="text-[11px] text-muted-foreground">
                            {questionTypeLabel(question.question_type)} · {difficultyLabel(
                              question.difficulty,
                            )}
                          </p>
                        </div>
                        <Input
                          aria-label={`Bobot ${question.code || question.question_text}`}
                          class="h-8 text-right text-xs"
                          type="number"
                          min={1}
                          max={100}
                          value={questionWeightValue(question.id)}
                          oninput={(event) =>
                            handleQuestionWeightInput(question.id, event)}
                        />
                      </div>
                    {/each}
                  </div>
                </div>

                <details
                  class="mt-3 rounded-md border border-primary/20 bg-card px-3 py-2 text-xs text-foreground"
                >
                  <summary class="cursor-pointer font-semibold text-foreground"
                    >Lihat ringkasan blueprint dan mutu</summary
                  >
                  <div class="mt-3 grid gap-3 xl:grid-cols-[0.75fr_1.25fr]">
                    <div class="space-y-2 text-xs">
                      <div>
                        <p class="mb-1 font-semibold text-muted-foreground">
                          Bentuk soal
                        </p>
                        <div class="flex flex-wrap gap-1">
                          {#each selectedTypeBuckets as bucket (bucket.label)}
                            <Badge variant="outline" class="bg-card"
                              >{bucket.label}: {bucket.count}</Badge
                            >
                          {/each}
                        </div>
                      </div>
                      <div>
                        <p class="mb-1 font-semibold text-muted-foreground">
                          Level kognitif
                        </p>
                        <div class="flex flex-wrap gap-1">
                          {#each selectedCognitiveBuckets as bucket (bucket.label)}
                            <Badge
                              variant={bucket.label === "Belum level"
                                ? "secondary"
                                : "outline"}
                              class="bg-card"
                              >{bucket.label}: {bucket.count}</Badge
                            >
                          {/each}
                        </div>
                      </div>
                      <p
                        class="rounded-md border border-primary/20 bg-card px-2 py-1.5 text-muted-foreground"
                      >
                        Gunakan panel ini untuk mencegah paket terlalu menumpuk
                        di satu CP/KD sebelum sesi ujian dibuat.
                      </p>
                    </div>

                    <div
                      class="overflow-x-auto rounded-md border border-primary/20 bg-card"
                    >
                      <div
                        class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b bg-primary/10 px-2 py-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-primary min-w-[500px]"
                      >
                        <span>CP</span>
                        <span>TP</span>
                        <span>KD / Materi</span>
                        <span>Level</span>
                        <span class="text-right">Soal</span>
                      </div>
                      <div class="max-h-40 overflow-y-auto">
                        {#each selectedBlueprintMatrix as row (row.key)}
                          <div
                            class="grid grid-cols-[1fr_1fr_1fr_0.8fr_0.5fr] gap-2 border-b px-2 py-1.5 text-xs last:border-b-0 min-w-[500px] {row.missing
                              ? 'bg-warning/10'
                              : ''}"
                          >
                            <span
                              class="min-w-0 truncate font-medium text-foreground"
                              title={row.cp}>{row.cp}</span
                            >
                            <span
                              class="min-w-0 truncate text-muted-foreground"
                              title={row.tp}>{row.tp}</span
                            >
                            <span
                              class="min-w-0 truncate text-muted-foreground"
                              title={`${row.kd} / ${row.topic}`}
                              >{row.kd} / {row.topic}</span
                            >
                            <span
                              class="min-w-0 truncate text-muted-foreground"
                              title={row.types.join(", ")}>{row.cognitive}</span
                            >
                            <span
                              class="text-right font-semibold text-foreground"
                            >
                              {row.count}
                              {#if row.hotsCount > 0}
                                <span class="text-warning"
                                  >/{row.hotsCount}</span
                                >
                              {/if}
                            </span>
                          </div>
                        {/each}
                      </div>
                    </div>
                  </div>
                </details>
              </div>
            {/if}
            {#if packageReadinessIssues.length > 0}
              <div
                class="mt-3 rounded-md border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-warning"
              >
                <span class="font-semibold">Belum siap dibuat:</span>
                <span>{packageReadinessIssues.join(", ")}</span>
              </div>
            {:else}
              <div
                class="mt-3 rounded-md border border-success/20 bg-success/10 px-3 py-2 text-xs font-medium text-success"
              >
                Paket siap dibuat dengan soal terbit yang sudah terpilih.
              </div>
            {/if}
          </div>
        {/if}

        <div class="flex gap-2">
          <LoadingButton
            disabled={!canCreatePackage}
            onclick={() => void createPackage()}
            loading={fBusy}
            loadingLabel="Menyimpan..."
          >
            {`Buat Paket${selectedQuestions.length > 0 ? ` (${selectedQuestions.length} soal)` : ""}`}
          </LoadingButton>
          <LoadingButton variant="outline" onclick={() => (showForm = false)}
            >Batal</LoadingButton
          >
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

  <AsyncContent promise={packagesPromise} onerror={handlePackagesRenderError}>
    {#snippet pending()}
      <Card.Root class="overflow-hidden border-border shadow-sm">
        <Card.Content class="space-y-3 p-6">
          {#each Array.from( { length: 5 }, ) as _, index (`package-row-skeleton-${index}`)}
            <div
              class="grid gap-3 lg:grid-cols-[1.2fr_0.6fr_0.5fr_0.5fr_1fr_0.5fr_0.6fr_auto] lg:items-center"
            >
              <Skeleton class="h-5 w-40" />
              <Skeleton class="h-6 w-16" />
              <Skeleton class="h-5 w-14" />
              <Skeleton class="h-5 w-12" />
              <Skeleton class="h-7 w-48" />
              <Skeleton class="h-6 w-12" />
              <Skeleton class="h-6 w-16" />
              <Skeleton class="h-9 w-20 justify-self-end" />
            </div>
          {/each}
        </Card.Content>
      </Card.Root>
    {/snippet}

    {#snippet failed(error, reset)}
      <RecoveryPanel
        title="Paket Ujian Belum Tersaji"
        message={packagesErrorMessage(error)}
        onRetry={() => retryPackages(reset)}
      />
    {/snippet}

    {#snippet children(value)}
      {@const overview = value as PackagesOverview}
      {@const eventPackages = strictEventPackages(overview.packages)}
      {@const currentPackages = eventPackages}
      {@const filteredPackages = sortPackages(
        currentPackages.filter(matchesPackageFilters),
      )}
      {@const safePackagePage = clampPage(
        packagePage,
        filteredPackages.length,
        packagePageSize,
      )}
      {@const paginatedPackages = paginateItems(
        filteredPackages,
        safePackagePage,
        packagePageSize,
      )}
      {@const summary = packageSummary(currentPackages)}
      {@const hiddenPackages = hiddenPackageCount(overview.packages)}
        {#snippet packageDetailPanel(pkg: CbtPackage, quality: PackageQualitySummary, progress: ReturnType<typeof packageProgress>, readiness: string)}
          <div
            class="space-y-4 rounded-2xl border border-border bg-muted/20 p-4 text-sm"
          >
            <div class="flex flex-wrap items-center gap-2">
              <Badge variant="outline" class="font-semibold">
                {pkg.title}
              </Badge>
              <Badge class={`text-xs ${readinessBadgeClass(readiness)}`}>
                {readinessLabel(readiness)}
              </Badge>
              {#if pkg.locked_at}
                <Badge variant="outline" class="text-xs">Terkunci</Badge>
              {/if}
              {#if Number(pkg.session_count ?? 0) > 0}
                <Badge variant="outline" class="text-xs">
                  {pkg.session_count} sesi
                </Badge>
              {/if}
            </div>

            {#if pkg.description}
              <p class="rounded-xl border bg-background p-3 text-muted-foreground">
                {pkg.description}
              </p>
            {/if}

            <section
              class="grid gap-2 rounded-xl border bg-background p-3 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-4"
            >
              <p>
                <span class="font-medium text-foreground">Mapel:</span>
                {pkg.subject_name} ({pkg.subject_code})
              </p>
              <p>
                <span class="font-medium text-foreground">Durasi:</span>
                {pkg.duration_minutes} menit
              </p>
              <p>
                <span class="font-medium text-foreground">Jumlah soal:</span>
                {quality.questions.length || pkg.question_count} soal
              </p>
              <p>
                <span class="font-medium text-foreground">Total poin:</span>
                {quality.totalPoints || "Belum dihitung"}
              </p>
              <p>
                <span class="font-medium text-foreground">Acak soal:</span>
                {pkg.randomize_questions ? "Ya" : "Tidak"}
              </p>
              <p>
                <span class="font-medium text-foreground">Acak opsi:</span>
                {pkg.randomize_options ? "Ya" : "Tidak"}
              </p>
              <p>
                <span class="font-medium text-foreground">Status:</span>
                {pkg.is_active ? "Aktif" : "Nonaktif"}
              </p>
              <p>
                <span class="font-medium text-foreground">Dibuat:</span>
                {formatDate(pkg.created_at)}
              </p>
            </section>

            <section class="space-y-2">
              <div class="flex items-center justify-between gap-2">
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Komposisi Paket
                </p>
                <span class="text-xs text-muted-foreground">
                  {progress.pgCount}/{packageTargets(pkg).pg} PG · {progress.essayCount}/{packageTargets(pkg).essay} Essay
                </span>
              </div>
              <div class="h-2 rounded-full bg-muted">
                <div
                  class="h-2 rounded-full bg-primary"
                  style={`width: ${Math.min(100, progress.percent)}%`}
                ></div>
              </div>
              {#if quality.questions.length > 0}
                <div class="flex flex-wrap gap-1.5">
                  {#each quality.typeBuckets as bucket (bucket.label)}
                    <Badge variant="outline" class="bg-background text-xs">
                      {bucket.label}: {bucket.count}
                    </Badge>
                  {/each}
                  {#each quality.cognitiveBuckets as bucket (bucket.label)}
                    <Badge variant="secondary" class="text-xs">
                      {bucket.label}: {bucket.count}
                    </Badge>
                  {/each}
                  {#if quality.hotsCount > 0}
                    <Badge class="border-warning/30 bg-warning/10 text-warning text-xs">
                      {quality.hotsCount} HOTS
                    </Badge>
                  {/if}
                  {#if quality.missingCount > 0}
                    <Badge class="border-warning/30 bg-warning/10 text-warning text-xs">
                      {quality.missingCount} metadata kurang
                    </Badge>
                  {/if}
                  {#if quality.unpublishedCount > 0}
                    <Badge class="border-destructive/30 bg-destructive/10 text-destructive text-xs">
                      {quality.unpublishedCount} belum terbit
                    </Badge>
                  {/if}
                </div>
              {:else}
                <p class="rounded-xl border bg-background p-3 text-muted-foreground">
                  Rincian soal paket belum tersedia pada daftar ini. Buka halaman paket untuk menyusun isi paket.
                </p>
              {/if}
            </section>

            {#if quality.questions.length > 0}
              <section class="space-y-2">
                <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Cuplikan Isi Paket
                </p>
                <div class="max-h-72 overflow-auto rounded-xl border bg-background">
                  <ol class="divide-y divide-border text-xs">
                    {#each quality.questions.slice(0, 12) as question (question.question_id)}
                      <li class="grid gap-1 p-3 sm:grid-cols-[3rem_1fr_auto] sm:items-start">
                        <span class="font-semibold text-muted-foreground">#{question.position}</span>
                        <div class="min-w-0">
                          <p class="font-mono font-semibold text-primary">
                            {question.question_code}
                          </p>
                          <p class="mt-1 line-clamp-2 break-words text-muted-foreground [overflow-wrap:anywhere]">
                            {question.question_text}
                          </p>
                        </div>
                        <div class="flex flex-wrap gap-1 sm:justify-end">
                          <Badge variant="outline" class="text-[11px]">
                            {questionTypeLabel(question.question_type)}
                          </Badge>
                          <Badge variant="secondary" class="text-[11px]">
                            {difficultyLabel(question.difficulty)}
                          </Badge>
                          <Badge variant="outline" class="text-[11px]">
                            {question.points} poin
                          </Badge>
                        </div>
                      </li>
                    {/each}
                  </ol>
                  {#if quality.questions.length > 12}
                    <p class="border-t border-border p-3 text-xs text-muted-foreground">
                      +{quality.questions.length - 12} soal lainnya. Buka halaman paket untuk melihat semua isi paket.
                    </p>
                  {/if}
                </div>
              </section>
            {/if}
          </div>
        {/snippet}

          {#snippet packageActions(row: unknown)}
            {@const p = row as CbtPackage}
            {@const quality = packageQualitySummary(p.id)}
            {@const readiness = packageReadinessStatus(p, quality)}
            <button
              type="button"
              class="inline-flex items-center rounded-md border border-border px-2 py-1 text-xs font-semibold hover:bg-muted"
              aria-expanded={isPackageExpanded(p.id)}
              aria-controls={`package-detail-${p.id}`}
              aria-label={`${isPackageExpanded(p.id) ? "Tutup pratinjau" : "Pratinjau"} paket ${p.title}`}
              onclick={() => togglePackageDetail(p.id)}
            >
              {isPackageExpanded(p.id) ? "Tutup" : "Pratinjau"}
            </button>
            <a
              href={`${resolve("/asesmen/paket")}/${p.id}`}
              class="inline-flex items-center rounded-md border border-border px-2 py-1 text-xs font-semibold hover:bg-muted"
              title="Buka halaman paket: identitas, isi soal, blueprint, dan kunci paket"
            >
              Buka
            </a>
            {#if p.locked_at}
              <LoadingButton
                size="xs"
                variant="outline"
                onclick={() => createRevisionPackage(p.id, p.title)}
                loading={bulkBusy}
                loadingLabel="Membuat..."
                title="Paket terkunci; buat salinan revisi agar riwayat lama tetap aman"
              >
                Buat Revisi
              </LoadingButton>
            {:else if readiness === "ready"}
              <LoadingButton
                size="xs"
                variant="outline"
                onclick={() => {
                  selectedPackageIds.clear();
                  selectedPackageIds.add(p.id);
                  void bulkLockSelected();
                }}
                loading={bulkBusy}>Lock</LoadingButton
              >
            {/if}
            {#if Number(p.session_count ?? 0) > 0}
              <LoadingButton
                variant="outline"
                size="xs"
                onclick={() => archivePackage(p.id, p.title, Number(p.session_count ?? 0))}
                loading={deleteBusyId === p.id}
                disabled={deleteBusyId !== "" && deleteBusyId !== p.id}
                loadingLabel="Mengarsipkan..."
                title={`Paket sudah dipakai ${p.session_count} sesi, tidak dapat dihapus permanen. Gunakan Arsipkan.`}
              >
                Arsipkan
              </LoadingButton>
            {:else}
              <LoadingButton
                variant="destructive"
                size="xs"
                onclick={() => deletePackage(p.id, p.title)}
                loading={deleteBusyId === p.id}
                disabled={deleteBusyId !== "" && deleteBusyId !== p.id}
                loadingLabel="Menghapus..."
                title="Hapus permanen paket yang belum dipakai sesi"
              >
                Hapus
              </LoadingButton>
            {/if}
          {/snippet}

          {#snippet packageCell(row: unknown, column: { key: string })}
            {@const p = row as CbtPackage}
            {@const quality = packageQualitySummary(p.id)}
            {@const progress = packageProgress(p, quality)}
            {@const readiness = packageReadinessStatus(p, quality)}
            {#if column.key === "select"}
              <input
                type="checkbox"
                class="rounded"
                checked={selectedPackageIds.has(p.id)}
                onchange={() => togglePackageSelection(p.id)}
                aria-label={`Pilih paket ${p.title}`}
              />
            {:else if column.key === "package"}
              <div class="space-y-1">
                <p class="font-semibold text-foreground">{p.title}</p>
                <p class="text-[11px] text-muted-foreground">
                  {p.subject_name} ({p.subject_code}) · {p.duration_minutes} menit
                </p>
                {#if p.description}
                  <p class="line-clamp-1 text-[11px] text-muted-foreground">{p.description}</p>
                {/if}
              </div>
            {:else if column.key === "composition"}
              <div class="space-y-1.5">
                <div class="flex flex-wrap items-center gap-1.5">
                  <Badge variant="secondary" class="text-xs">{p.question_count} soal</Badge>
                  {#if quality.totalPoints > 0}
                    <Badge variant="outline" class="text-xs">{quality.totalPoints} poin</Badge>
                  {/if}
                </div>
                {#if showUtsMode}
                  <div class="min-w-28">
                    <div class="h-1.5 rounded-full bg-muted">
                      <div class="h-1.5 rounded-full bg-primary" style={`width: ${Math.min(100, progress.percent)}%`}></div>
                    </div>
                    <p class="mt-1 text-[11px] text-muted-foreground">
                      {progress.pgCount}/{packageTargets(p).pg} PG · {progress.essayCount}/{packageTargets(p).essay} Essay
                    </p>
                  </div>
                {/if}
              </div>
            {:else if column.key === "readiness"}
              <div class="space-y-1">
                <Badge class={`text-xs ${readinessBadgeClass(readiness)}`}>{readinessLabel(readiness)}</Badge>
                {#if Number(p.session_count ?? 0) > 0}
                  <p class="text-[11px] text-muted-foreground">{p.session_count} sesi</p>
                {/if}
              </div>
            {:else if column.key === "quality"}
              {#if quality.questions.length === 0}
                <span class="text-xs text-muted-foreground">Belum ada rincian</span>
              {:else}
                <div class="flex max-w-sm flex-wrap gap-1">
                  {#each quality.typeBuckets.slice(0, 2) as bucket (bucket.label)}
                    <Badge variant="outline" class="bg-card text-[11px]">{bucket.label}: {bucket.count}</Badge>
                  {/each}
                  {#if quality.hotsCount > 0}
                    <Badge class="border-warning/30 bg-warning/10 text-warning text-[11px]">{quality.hotsCount} HOTS</Badge>
                  {/if}
                  {#if quality.missingCount > 0}
                    <Badge class="border-warning/30 bg-warning/10 text-warning text-[11px]">{quality.missingCount} metadata</Badge>
                  {/if}
                  {#if quality.unpublishedCount > 0}
                    <Badge class="border-destructive/30 bg-destructive/10 text-destructive text-[11px]">{quality.unpublishedCount} belum terbit</Badge>
                  {/if}
                </div>
              {/if}
            {:else if column.key === "status"}
              <div class="flex flex-wrap gap-1">
                {#if p.is_active}
                  <Badge class="bg-primary/15 text-primary border-primary/20 text-xs">Aktif</Badge>
                {:else}
                  <Badge variant="secondary" class="text-xs">Nonaktif</Badge>
                {/if}
                {#if p.randomize_questions}
                  <Badge class="bg-primary/10 text-primary border-primary/20 text-xs">Acak</Badge>
                {/if}
                {#if p.locked_at}
                  <Badge variant="outline" class="text-xs">Terkunci</Badge>
                {/if}
              </div>
            {/if}
          {/snippet}

          {#snippet packageMobile(row: unknown)}
            {@const p = row as CbtPackage}
            {@const quality = packageQualitySummary(p.id)}
            {@const progress = packageProgress(p, quality)}
            {@const readiness = packageReadinessStatus(p, quality)}
            <div class="space-y-3 text-xs">
              <div class="flex items-start gap-3">
                {@render packageCell(p, packageTableColumns[0])}
                <div class="min-w-0 flex-1">{@render packageCell(p, packageTableColumns[1])}</div>
              </div>
              <div class="grid gap-3 sm:grid-cols-2">
                {@render packageCell(p, packageTableColumns[2])}
                {@render packageCell(p, packageTableColumns[3])}
              </div>
              {@render packageCell(p, packageTableColumns[4])}
              {@render packageCell(p, packageTableColumns[5])}
              {#if isPackageExpanded(p.id)}
                <div id={`package-mobile-detail-${p.id}`} class="pt-1" role="region">
                  {@render packageDetailPanel(p, quality, progress, readiness)}
                </div>
              {/if}
              <div class="flex flex-wrap gap-1.5 pt-1">{@render packageActions(p)}</div>
            </div>
          {/snippet}



      <Card.Root
        id="paket-saya"
        class="overflow-hidden border-border shadow-sm"
      >
        <Card.Header class="space-y-4 pb-4">
          <div
            class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between"
          >
            <div>
              <p
                class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
              >
                Paket Saya
              </p>
              <Card.Title class="mt-1 text-base"
                >Daftar Paket ({filteredPackages.length}/{currentPackages.length})</Card.Title
              >
            </div>
            <div class="flex flex-wrap gap-2">
              <LoadingButton
                variant="outline"
                size="sm"
                onclick={() =>
                  exportPackagesCsv(
                    filteredPackages,
                    "rekap-paket-soal-filtered.csv",
                  )}>Export Filter</LoadingButton
              >
              <a
                href={createPackageHref}
                class="inline-flex items-center rounded-md bg-primary px-3 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
                >Buat Paket</a
              >
            </div>
          </div>
          {#if hiddenPackages > 0}
            <Card.Description
              >{hiddenPackages} templat umum atau paket kegiatan lain disembunyikan
              dari daftar kegiatan ini.</Card.Description
            >
          {/if}

          <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
            {#each [["Total", summary.total, "text-foreground"], ["Kosong", summary.empty, "text-muted-foreground"], ["Kurang", summary.short, "text-warning"], ["Siap", summary.ready, "text-success"], ["Terkunci", summary.locked, "text-primary"], ["Dipakai Sesi", summary.used, "text-foreground"]] as item (`summary-${item[0]}`)}
              <div class="rounded-xl border border-border bg-muted/30 p-3">
                <p
                  class="text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground"
                >
                  {item[0]}
                </p>
                <p class={`mt-1 text-2xl font-semibold ${item[2]}`}>
                  {item[1]}
                </p>
              </div>
            {/each}
          </div>

          <div class="rounded-2xl border border-border bg-muted/30 p-3">
            <div
              class="grid gap-3 overflow-x-auto lg:grid-cols-3 xl:grid-cols-[1.3fr_0.9fr_0.9fr_0.8fr_0.8fr_0.8fr_0.9fr]"
            >
              <Input
                placeholder="Cari paket, mapel, kode, deskripsi..."
                bind:value={searchTerm}
                oninput={resetPackagePage}
                class="min-w-0"
              />
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={subjectFilter}
                onchange={resetPackagePage}
              >
                <option value="all">Semua mapel</option>
                {#each subjects as s (s.id)}<option value={s.id}
                    >{s.code} — {s.name}</option
                  >{/each}
              </select>
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={readinessFilter}
                onchange={resetPackagePage}
              >
                {#each ["all", "empty", "short", "metadata", "unpublished", "ready", "locked"] as status (status)}
                  <option value={status}>{readinessLabel(status)}</option>
                {/each}
              </select>
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={activeFilter}
                onchange={resetPackagePage}
              >
                <option value="all">Semua status</option><option value="active"
                  >Aktif</option
                ><option value="inactive">Nonaktif</option>
              </select>
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={lockFilter}
                onchange={resetPackagePage}
              >
                <option value="all">Semua status kunci</option><option
                  value="locked">Terkunci</option
                ><option value="unlocked">Belum terkunci</option>
              </select>
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={usageFilter}
                onchange={resetPackagePage}
              >
                <option value="all">Semua sesi</option><option value="used"
                  >Dipakai sesi</option
                ><option value="unused">Belum dipakai</option>
              </select>
              <select
                class="min-w-0 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                bind:value={sortMode}
                onchange={resetPackagePage}
              >
                <option value="needs_first">Prioritas gap</option><option
                  value="name_asc">Nama A-Z</option
                ><option value="subject_asc">Mapel A-Z</option><option
                  value="questions_asc">Soal sedikit</option
                ><option value="questions_desc">Soal banyak</option><option
                  value="created_desc">Terbaru</option
                >
              </select>
            </div>
            <div
              class="mt-3 flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground"
            >
              <label class="flex items-center gap-2"
                ><input
                  type="checkbox"
                  bind:checked={showUtsMode}
                  class="rounded"
                /> Mode kesiapan UTS: target 20 PG + 5 Essay</label
              >
              <div class="flex flex-wrap gap-2">
                <LoadingButton
                  size="xs"
                  variant="outline"
                  onclick={() => selectVisiblePackages(filteredPackages)}
                  >Pilih hasil filter</LoadingButton
                >
                <LoadingButton
                  size="xs"
                  variant="outline"
                  onclick={clearPackageSelection}
                  >Bersihkan pilihan</LoadingButton
                >
              </div>
            </div>
          </div>

          {#if selectedPackageIds.size > 0}
            <div
              class="flex flex-wrap items-center justify-between gap-3 rounded-2xl border border-primary/20 bg-primary/10 px-4 py-3 text-sm text-primary"
            >
              <p>
                <span class="font-semibold"
                  >{selectedPackageIds.size} paket dipilih.</span
                > Aksi massal hanya memproses paket yang aman; paket terkunci/belum
                siap otomatis dilewati.
              </p>
              <div class="flex flex-wrap gap-2">
                <LoadingButton
                  size="sm"
                  variant="outline"
                  onclick={() =>
                    exportPackagesCsv(
                      currentPackages.filter((pkg) =>
                        selectedPackageIds.has(pkg.id),
                      ),
                      "rekap-paket-soal-selected.csv",
                    )}>Export Terpilih</LoadingButton
                >
                <LoadingButton
                  size="sm"
                  onclick={() => void bulkLockSelected()}
                  loading={bulkBusy}
                  loadingLabel="Lock...">Bulk Lock Ready</LoadingButton
                >
                <LoadingButton
                  size="sm"
                  variant="outline"
                  onclick={() => void bulkSetActive(true)}
                  loading={bulkBusy}>Aktifkan</LoadingButton
                >
                <LoadingButton
                  size="sm"
                  variant="outline"
                  onclick={() => void bulkSetActive(false)}
                  loading={bulkBusy}>Nonaktifkan</LoadingButton
                >
              </div>
            </div>
          {/if}
        </Card.Header>
        <Card.Content class="p-0">
          <MicroActionTable
            rows={paginatedPackages}
            columns={packageTableColumns}
            rowKey={(row) => (row as CbtPackage).id}
            cell={packageCell}
            actions={packageActions}
            mobile={packageMobile}
            tableClass="min-w-[1050px]"
            emptyTitle="Tidak ada paket sesuai filter"
            emptyDescription="Ubah filter atau buat paket baru."
            class="rounded-none border-x-0 shadow-none"
          />

          {#if paginatedPackages.some((pkg) => isPackageExpanded(pkg.id))}
            <div class="space-y-3 border-t border-border bg-muted/10 p-3">
              {#each paginatedPackages.filter((pkg) => isPackageExpanded(pkg.id)) as p (p.id)}
                {@const quality = packageQualitySummary(p.id)}
                {@const progress = packageProgress(p, quality)}
                {@const readiness = packageReadinessStatus(p, quality)}
                <section id={`package-detail-${p.id}`} class="hidden md:block">
                  {@render packageDetailPanel(p, quality, progress, readiness)}
                </section>
              {/each}
            </div>
          {/if}

          <div class="border-t border-border p-3">
            <TablePagination
              page={safePackagePage}
              limit={packagePageSize}
              total={filteredPackages.length}
              itemLabel="paket"
              ariaLabel="Navigasi halaman daftar paket asesmen"
              onchange={handlePackagePagination}
            />
          </div>
        </Card.Content>
      </Card.Root>
    {/snippet}
  </AsyncContent>
</div>
