<script lang="ts">
  import { onMount } from "svelte";

  import { resolve } from "$app/paths";
  import BarChart3Icon from "@lucide/svelte/icons/bar-chart-3";
  import BookOpenCheckIcon from "@lucide/svelte/icons/book-open-check";
  import ClipboardCheckIcon from "@lucide/svelte/icons/clipboard-check";
  import EyeIcon from "@lucide/svelte/icons/eye";
  import FileQuestionIcon from "@lucide/svelte/icons/file-question";
  import HistoryIcon from "@lucide/svelte/icons/history";
  import Layers3Icon from "@lucide/svelte/icons/layers-3";
  import ListFilterIcon from "@lucide/svelte/icons/list-filter";
  import PackageIcon from "@lucide/svelte/icons/package";
  import PencilIcon from "@lucide/svelte/icons/pencil";
  import PlusIcon from "@lucide/svelte/icons/plus";
  import PrinterIcon from "@lucide/svelte/icons/printer";
  import RefreshCcwIcon from "@lucide/svelte/icons/refresh-ccw";
  import SearchIcon from "@lucide/svelte/icons/search";
  import SettingsIcon from "@lucide/svelte/icons/settings";
  import SparklesIcon from "@lucide/svelte/icons/sparkles";
  import Trash2Icon from "@lucide/svelte/icons/trash-2";
  import ArchiveIcon from "@lucide/svelte/icons/archive";
  import UploadIcon from "@lucide/svelte/icons/upload";
  import * as Table from "$lib/components/ui/table";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Skeleton } from "$lib/components/ui/skeleton";
  import { TablePagination } from "$lib/components/ui/pagination";
  import AsyncContent from "$lib/components/AsyncContent.svelte";
  import LoadingButton from "$lib/components/LoadingButton.svelte";
  import RecoveryPanel from "$lib/components/RecoveryPanel.svelte";
  import RichContent from "$lib/components/RichContent.svelte";
  import {
    clientApiPathWithQuery,
    readClientApiData,
    readClientJson,
  } from "$lib/client/api";
  import {
    canCreateBankSoal,
    canDeleteBankSoal,
    canImportBankSoal,
    canManageBankSoalSettings,
    canPublishBankSoal,
    canReviewBankSoal,
  } from "$lib/bank-soal/access";
  import { displayName } from "$lib/utils/display-name";
  import {
    DEFAULT_PAGE_SIZE_OPTIONS,
    normalizePage,
    normalizePageSize,
    type PaginationChange,
  } from "$lib/utils/pagination";

  type PageData = {
    user?: {
      role?: string;
      roles?: string[];
      permissions?: string[];
    };
  };

  type PageMode = "dashboard" | "list" | "combined";

  type Subject = {
    id: string;
    name: string;
    code?: string;
  };

  type Author = { username: string; display_name: string };

  type OptionItem = {
    label?: string;
    text?: string;
    html?: string;
    latex?: string;
    match_label?: string;
    match_text?: string;
    match_html?: string;
    is_distractor?: boolean;
  };

  type Question = {
    id: string;
    event_id?: string | null;
    subject_id?: string;
    subject_name?: string;
    subject_code?: string;
    code?: string;
    question_text?: string;
    question_type?: string;
    stem_html?: string;
    stem_latex?: string;
    stimulus_html?: string;
    options?: OptionItem[];
    option_a?: string;
    option_b?: string;
    option_c?: string;
    option_d?: string;
    option_e?: string;
    answer_key?: string;
    explanation?: string;
    explanation_html?: string;
    writer_notes?: string;
    difficulty?: string;
    status?: string;
    workflow_status?: string;
    author_username?: string;
    author_display_name?: string;
    reviewer_username?: string;
    reviewer_display_name?: string;
    review_notes?: string;
    authoring_mode?: string;
    suggested_mode?: string;
    academic_phase?: string;
    target_level?: string | null;
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    indicator_ref?: string;
    material_topic?: string;
    cognitive_level?: string;
    hots_flag?: boolean;
    created_at?: string;
    updated_at?: string;
    package_count?: number;
    answer_count?: number;
    is_locked?: boolean;
    usage?: {
      package_count?: number;
      answer_count?: number;
      is_locked?: boolean;
    };
  };

  type QuestionListResponse = {
    items?: Question[];
    meta?: {
      total?: number;
      limit?: number;
      offset?: number;
    };
  };

  type AcademicPayload = {
    subjects?: Subject[];
    error?: string;
    message?: string;
  };

  type SummaryCountPayload = Partial<
    Record<
      StatusKey | "package_usage" | "total" | "unsubmitted" | "workflow_draft",
      number
    >
  >;

  type BankSoalSummaryResponse = {
    counts?: SummaryCountPayload;
    by_subject?: Array<{
      subject_id?: string;
      subject_name?: string;
      subject_code?: string;
      total?: number;
    }>;
    by_cognitive_level?: Array<{ cognitive_level?: string; total?: number }>;
    recent?: Question[];
  };

  type WorkflowFilter =
    | "draft"
    | "submitted"
    | "review"
    | "revision_needed"
    | "reviewed"
    | "approved"
    | "published"
    | "rejected"
    | "archived";
  type PublicationFilter = "" | "draft" | "published";
  type QuestionTypeFilter =
    | ""
    | "multiple_choice"
    | "multiple_answer"
    | "true_false"
    | "agree_disagree"
    | "matching"
    | "short_answer"
    | "essay";
  type HotsFilter = "" | "yes" | "no";
  type StatusKey =
    | "all"
    | "unpublished"
    | "draft"
    | "review"
    | "rejected"
    | "approved"
    | "published";

  type StatusCounts = Record<StatusKey, number>;

  type BankSoalOverview = {
    questions: Question[];
    subjects: Subject[];
    totalItems: number;
    counts: StatusCounts;
    page: number;
    limit: number;
    offset: number;
    summary?: BankSoalSummaryResponse;
  };

  type SummaryCard = {
    key: StatusKey;
    label: string;
    helper: string;
    value: number;
    tone: "slate" | "amber" | "red" | "green" | "emerald";
    active: boolean;
  };

  type SubjectDistribution = {
    label: string;
    name: string;
    value: number;
    color: string;
  };

  type BloomComposition = {
    key: string;
    label: string;
    value: number;
    percent: number;
  };

  type ActivityItem = {
    id: string;
    actor: string;
    action: string;
    object: string;
    time: string;
    tone: string;
  };

  let { data, mode = "combined" }: { data: PageData; mode?: PageMode } =
    $props();

  const DEFAULT_PAGE_SIZE = DEFAULT_PAGE_SIZE_OPTIONS[0];
  const emptyCounts: StatusCounts = {
    all: 0,
    unpublished: 0,
    draft: 0,
    review: 0,
    rejected: 0,
    approved: 0,
    published: 0,
  };

  const workflowOptions: Array<{ value: WorkflowFilter; label: string }> = [
    { value: "draft", label: "Draft" },
    { value: "submitted", label: "Diajukan" },
    { value: "review", label: "Review" },
    { value: "revision_needed", label: "Perlu Revisi" },
    { value: "reviewed", label: "Sudah Direview" },
    { value: "approved", label: "Disetujui" },
    { value: "published", label: "Published" },
    { value: "rejected", label: "Ditolak" },
    { value: "archived", label: "Arsip" },
  ];

  const publicationOptions: Array<{ value: PublicationFilter; label: string }> =
    [
      { value: "", label: "Semua ketersediaan" },
      { value: "draft", label: "Belum siap dipakai" },
      { value: "published", label: "Siap dipakai" },
    ];

  const questionTypeOptions: Array<{
    value: QuestionTypeFilter;
    label: string;
  }> = [
    { value: "", label: "Semua tipe" },
    { value: "multiple_choice", label: "PG" },
    { value: "multiple_answer", label: "PG kompleks" },
    { value: "true_false", label: "Benar/Salah" },
    { value: "agree_disagree", label: "Setuju/Tidak" },
    { value: "matching", label: "Menjodohkan" },
    { value: "short_answer", label: "Isian singkat" },
    { value: "essay", label: "Uraian" },
  ];

  const hotsOptions: Array<{ value: HotsFilter; label: string }> = [
    { value: "", label: "Semua HOTS" },
    { value: "yes", label: "HOTS" },
    { value: "no", label: "Non-HOTS" },
  ];

  const workflowLabels: Record<string, string> = {
    draft: "Draft",
    submitted: "Diajukan",
    review: "Review",
    revision_needed: "Perlu Revisi",
    reviewed: "Sudah Direview",
    approved: "Disetujui",
    published: "Published",
    rejected: "Ditolak",
    archived: "Arsip",
  };

  const publicationLabels: Record<string, string> = {
    draft: "Belum Terbit",
    published: "Terbit",
    archived: "Arsip",
  };

  const questionTypeLabels: Record<string, string> = {
    multiple_choice: "PG",
    multiple_answer: "PG Kompleks",
    true_false: "Benar/Salah",
    agree_disagree: "Setuju/Tidak",
    matching: "Menjodohkan",
    short_answer: "Isian Singkat",
    essay: "Uraian",
  };

  const difficultyLabels: Record<string, string> = {
    easy: "Mudah",
    medium: "Sedang",
    hard: "Sulit",
  };

  const subjectChartColors = [
    "bg-accent",
    "bg-primary",
    "bg-rose-500",
    "bg-violet-500",
    "bg-warning",
    "bg-cyan-500",
    "bg-lime-600",
    "bg-orange-600",
  ];

  const bloomOrder = ["C1", "C2", "C3", "C4", "C5", "C6"];
  const bloomLabels: Record<string, string> = {
    C1: "C1 · Mengingat",
    C2: "C2 · Memahami",
    C3: "C3 · Menerapkan",
    C4: "C4 · Menganalisis",
    C5: "C5 · Mengevaluasi",
    C6: "C6 · Mencipta",
  };

  let search = $state("");
  let subjectFilter = $state("");
  let workflowFilters = $state<WorkflowFilter[]>([]);
  let workflowDropdownOpen = $state(false);
  let publicationFilter = $state<PublicationFilter>("");
  let questionTypeFilter = $state<QuestionTypeFilter>("");
  let hotsFilter = $state<HotsFilter>("");
  let authorFilter = $state("");
  let authors = $state<Author[]>([]);
  let authorsLoaded = $state(false);
  let currentPage = $state(1);
  let pageSize = $state<number>(DEFAULT_PAGE_SIZE);
  let questions = $state<Question[]>([]);
  let subjects = $state<Subject[]>([]);
  let totalItems = $state(0);
  let counts = $state<StatusCounts>({ ...emptyCounts });
  let questionsPromise = $state<Promise<BankSoalOverview> | null>(null);
  let refreshing = $state(false);
  let deletingQuestionId = $state<string | null>(null);
  let archivingQuestionId = $state<string | null>(null);
  let restoringQuestionId = $state<string | null>(null);
  let revisingQuestionId = $state<string | null>(null);
  let expandedQuestionId = $state<string | null>(null);
  let revisionError = $state("");
  let restoreError = $state("");
  let deleteError = $state("");
  let requestId = 0;
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  let roles = $derived(
    data.user?.roles ?? (data.user?.role ? [data.user.role] : []),
  );
  let canCreate = $derived(canCreateBankSoal(data.user));
  let canImport = $derived(canImportBankSoal(data.user));
  let canDelete = $derived(canDeleteBankSoal(data.user));
  let canReview = $derived(canReviewBankSoal(data.user));
  let canPublish = $derived(canPublishBankSoal(data.user));
  let canSettings = $derived(canManageBankSoalSettings(data.user));
  let canQuality = $derived(canAccessQuality(data.user));
  let canSeeAnswerKey = $derived(canReview || canPublish || canQuality);
  let roleLabel = $derived.by(() => {
    if (roles.includes("admin")) return "Admin";
    if (roles.includes("guru")) return "Guru";
    return roles.length > 0 ? roles.join(", ") : "Pengguna";
  });
  let hasFilters = $derived(
    Boolean(
      search.trim() ||
        subjectFilter ||
        workflowFilters.length > 0 ||
        publicationFilter ||
        questionTypeFilter ||
        hotsFilter ||
        authorFilter,
    ),
  );
  let selectedSubject = $derived(
    subjects.find((subject) => subject.id === subjectFilter) ?? null,
  );
  let selectedAuthor = $derived(
    authors.find((author) => author.username === authorFilter) ?? null,
  );
  let composerHref = $derived(resolve("/bank-soal/tambah"));
  let importHref = $derived(resolve("/bank-soal/impor"));
  let printHref = $derived(resolve("/bank-soal/cetak"));
  let reviewRouteHref = $derived(reviewHref());
  let analysisHref = $derived(resolve("/bank-soal/analisis-butir"));
  let mapelKdHref = $derived(resolve("/bank-soal/mapel-kd"));
  let settingsHref = $derived(resolve("/bank-soal/pengaturan"));
  let packageHref = $derived(resolve("/asesmen/paket"));
  let listHref = $derived(resolve("/bank-soal"));
  let summarySubjectDistribution = $state<SubjectDistribution[]>([]);
  let summaryBloomComposition = $state<BloomComposition[]>([]);
  let summaryRecentActivities = $state<ActivityItem[]>([]);
  let summaryPackageUsage = $state(0);
  let totalPackageUsage = $derived.by(() =>
    questions.reduce(
      (sum, question) =>
        sum + (question.package_count ?? question.usage?.package_count ?? 0),
      0,
    ),
  );
  let subjectDistribution = $derived.by(() =>
    buildSubjectDistribution(questions, subjects),
  );
  let bloomComposition = $derived.by(() => buildBloomComposition(questions));
  let recentActivities = $derived.by(() => buildRecentActivities(questions));
  let showDashboard = $derived(mode === "dashboard" || mode === "combined");
  let showList = $derived(mode === "list" || mode === "combined");
  let summaryCards = $derived<SummaryCard[]>([
    {
      key: "all",
      label: "Semua",
      helper: "stok sesuai filter",
      value: counts.all,
      tone: "slate",
      active: workflowFilters.length === 0 && !publicationFilter,
    },
    {
      key: "unpublished",
      label: "Belum Terbit",
      helper: "ringkasan belum tayang",
      value: counts.unpublished,
      tone: "slate",
      active: workflowFilters.length === 0 && publicationFilter === "draft",
    },
    {
      key: "draft",
      label: "Draft / Belum Diajukan",
      helper: "belum masuk verifikasi",
      value: counts.draft,
      tone: "slate",
      active: workflowFiltersEqual(["draft"]) && publicationFilter === "draft",
    },
    {
      key: "review",
      label: "Verifikasi",
      helper: "diajukan/review",
      value: counts.review,
      tone: "amber",
      active: workflowFiltersEqual(["submitted", "review"]) && !publicationFilter,
    },
    {
      key: "rejected",
      label: "Revisi",
      helper: "perlu perbaikan",
      value: counts.rejected,
      tone: "red",
      active: workflowFiltersEqual(["revision_needed", "rejected"]) && publicationFilter === "draft",
    },
    {
      key: "approved",
      label: "Disetujui",
      helper: "siap diterbitkan",
      value: counts.approved,
      tone: "green",
      active: workflowFilters.length === 1 && workflowFilters[0] === "approved" && publicationFilter === "draft",
    },
    {
      key: "published",
      label: "Terbit",
      helper: "siap dipakai paket",
      value: counts.published,
      tone: "emerald",
      active: workflowFilters.length === 0 && publicationFilter === "published",
    },
  ]);

  function buildBaseParams(limit: number, offset: number): URLSearchParams {
    const params = new URLSearchParams();
    params.set("limit", String(limit));
    params.set("offset", String(offset));
    if (search.trim()) params.set("q", search.trim());
    if (subjectFilter) params.set("subject_id", subjectFilter);
    if (questionTypeFilter) params.set("question_type", questionTypeFilter);
    if (hotsFilter) params.set("hots", hotsFilter);
    if (authorFilter) params.set("author_username", authorFilter);
    return params;
  }

  function appendStatusParams(params: URLSearchParams, key: StatusKey) {
    if (key === "unpublished") {
      params.set("status", "draft");
      return;
    }
    if (key === "draft") {
      params.set("workflow_status", "draft");
      params.set("status", "draft");
      return;
    }
    if (key === "review") {
      params.append("workflow_status", "submitted");
      params.append("workflow_status", "review");
      return;
    }
    if (key === "rejected") {
      params.append("workflow_status", "revision_needed");
      params.append("workflow_status", "rejected");
      params.set("status", "draft");
      return;
    }
    if (key === "approved") {
      params.set("workflow_status", "approved");
      params.set("status", "draft");
      return;
    }
    if (key === "published") {
      params.set("status", "published");
    }
  }

  function buildListParams(page: number, limit = pageSize): URLSearchParams {
    const params = buildBaseParams(
      limit,
      Math.max(0, (page - 1) * limit),
    );
    workflowFilters.forEach((status) => params.append("workflow_status", status));
    if (publicationFilter) params.set("status", publicationFilter);
    return params;
  }

  function buildCountParams(key: StatusKey): URLSearchParams {
    const params = buildBaseParams(1, 0);
    appendStatusParams(params, key);
    return params;
  }

  async function fetchQuestionList(page: number, limit = pageSize): Promise<QuestionListResponse> {
    return fetch(
      clientApiPathWithQuery("/api/bank-soal/questions", buildListParams(page, limit)),
    ).then((response) =>
      readClientApiData<QuestionListResponse>(
        response,
        "Gagal memuat daftar soal",
      ),
    );
  }

  async function fetchQuestionCount(key: StatusKey): Promise<number> {
    const payload = await fetch(
      clientApiPathWithQuery("/api/bank-soal/questions", buildCountParams(key)),
    ).then((response) =>
      readClientApiData<QuestionListResponse>(
        response,
        "Gagal memuat ringkasan status soal",
      ),
    );
    return payload.meta?.total ?? payload.items?.length ?? 0;
  }

  async function fetchSubjects(): Promise<Subject[]> {
    const payload = await fetch("/api/bank-soal/soal-support/subjects").then(
      (response) =>
        readClientApiData<AcademicPayload>(
          response,
          "Gagal memuat data mata pelajaran",
        ),
    );
    return payload.subjects ?? [];
  }

  async function fetchQuestionSummary(): Promise<BankSoalSummaryResponse> {
    return fetch("/api/bank-soal/summary").then((response) =>
      readClientApiData<BankSoalSummaryResponse>(
        response,
        "Gagal memuat ringkasan bank soal",
      ),
    );
  }

  function normalizeSummaryCounts(
    summaryCounts: SummaryCountPayload | undefined,
    fallback: StatusCounts = emptyCounts,
  ): StatusCounts {
    const all = summaryCounts?.all ?? summaryCounts?.total ?? fallback.all;
    const unpublished = summaryCounts?.unpublished ?? summaryCounts?.draft ?? fallback.unpublished;
    const review = summaryCounts?.review ?? fallback.review;
    const rejected = summaryCounts?.rejected ?? fallback.rejected;
    const approved = summaryCounts?.approved ?? fallback.approved;
    const published = summaryCounts?.published ?? fallback.published;
    const draft =
      summaryCounts?.unsubmitted ??
      summaryCounts?.workflow_draft ??
      (summaryCounts?.unpublished !== undefined
        ? summaryCounts?.draft ?? Math.max(0, unpublished - review - rejected - approved)
        : Math.max(0, unpublished - review - rejected - approved));

    return {
      all,
      unpublished,
      draft,
      review,
      rejected,
      approved,
      published,
    };
  }

  function applySummaryPayload(summary: BankSoalSummaryResponse | undefined) {
    if (!summary) return;
    if (summary.counts) {
      counts = normalizeSummaryCounts(summary.counts, counts);
      summaryPackageUsage = summary.counts.package_usage ?? summaryPackageUsage;
    }
    summarySubjectDistribution = (summary.by_subject ?? [])
      .slice(0, 8)
      .map((item, index) => {
        const name = item.subject_name || item.subject_code || "Tanpa Mapel";
        return {
          label: shortSubjectLabel(name),
          name,
          value: item.total ?? 0,
          color: subjectChartColors[index % subjectChartColors.length],
        };
      });
    const bloomTotal = Math.max(
      1,
      (summary.by_cognitive_level ?? []).reduce(
        (sum, item) => sum + (item.total ?? 0),
        0,
      ),
    );
    summaryBloomComposition = (summary.by_cognitive_level ?? []).map((item) => {
      const key = normalizeBloomLevel(item.cognitive_level);
      const value = item.total ?? 0;
      return {
        key,
        label: bloomLabels[key] ?? key,
        value,
        percent: Math.round((value / bloomTotal) * 100),
      };
    });
    summaryRecentActivities = buildRecentActivities(summary.recent ?? []);
  }

  async function fetchOverview(page = currentPage, limit = pageSize): Promise<BankSoalOverview> {
    const [questionPayload, loadedSubjects, summaryPayload] = await Promise.all(
      [
        fetchQuestionList(page, limit),
        subjects.length > 0 ? Promise.resolve(subjects) : fetchSubjects(),
        fetchQuestionSummary(),
      ],
    );
    const summaryCounts = summaryPayload.counts ?? {};
    return {
      questions: questionPayload.items ?? [],
      subjects: loadedSubjects,
      totalItems:
        questionPayload.meta?.total ??
        summaryCounts.all ??
        questionPayload.items?.length ??
        0,
      counts: normalizeSummaryCounts(summaryCounts),
      page,
      limit: questionPayload.meta?.limit ?? limit,
      offset:
        questionPayload.meta?.offset ?? Math.max(0, (page - 1) * limit),
      summary: summaryPayload,
    };
  }

  function currentOverview(): BankSoalOverview {
    return {
      questions,
      subjects,
      totalItems,
      counts,
      page: currentPage,
      limit: pageSize,
      offset: Math.max(0, (currentPage - 1) * pageSize),
    };
  }

  function applyOverview(
    overview: BankSoalOverview,
    summary?: BankSoalSummaryResponse,
  ) {
    questions = overview.questions;
    if (
      expandedQuestionId &&
      !questions.some((question) => question.id === expandedQuestionId)
    ) {
      expandedQuestionId = null;
    }
    subjects = overview.subjects;
    totalItems = overview.totalItems;
    counts = overview.counts;
    currentPage = overview.page;
    pageSize = normalizePageSize(overview.limit, DEFAULT_PAGE_SIZE_OPTIONS, pageSize);
    applySummaryPayload(summary);
  }

  function syncUrl(page: number, limit = pageSize) {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams();
    if (search.trim()) params.set("q", search.trim());
    if (subjectFilter) params.set("subject_id", subjectFilter);
    workflowFilters.forEach((status) => params.append("workflow_status", status));
    if (publicationFilter) params.set("status", publicationFilter);
    if (questionTypeFilter) params.set("question_type", questionTypeFilter);
    if (hotsFilter) params.set("hots", hotsFilter);
    if (authorFilter) params.set("author_username", authorFilter);
    if (page > 1) params.set("page", String(page));
    if (limit !== DEFAULT_PAGE_SIZE) params.set("limit", String(limit));
    const query = params.toString();
    window.history.replaceState(
      {},
      "",
      query ? `${window.location.pathname}?${query}` : window.location.pathname,
    );
  }

  function load(page = currentPage, markRefreshing = false, limit = pageSize) {
    const nextPage = Math.max(1, page);
    currentPage = nextPage;
    pageSize = limit;
    syncUrl(nextPage, limit);
    if (markRefreshing) refreshing = true;
    const activeRequestId = ++requestId;
    const promise = fetchOverview(nextPage, limit)
      .then((overview) => {
        if (activeRequestId !== requestId) return currentOverview();
        applyOverview(overview, overview.summary);
        return overview;
      })
      .catch((error: unknown) => {
        if (activeRequestId !== requestId) return currentOverview();
        throw error;
      })
      .finally(() => {
        if (activeRequestId === requestId) refreshing = false;
      });
    questionsPromise = promise;
  }

  function applyFilters(event?: SubmitEvent) {
    event?.preventDefault();
    load(1, true, pageSize);
  }

  function onSearchInput(event: Event) {
    search = (event.currentTarget as HTMLInputElement).value;
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => load(1, true, pageSize), 350);
  }

  function clearFilters() {
    if (searchTimer) clearTimeout(searchTimer);
    search = "";
    subjectFilter = "";
    workflowFilters = [];
    publicationFilter = "";
    questionTypeFilter = "";
    hotsFilter = "";
    authorFilter = "";
    load(1, true, pageSize);
  }

  function setSummaryFilter(key: StatusKey) {
    if (key === "all") {
      workflowFilters = [];
      publicationFilter = "";
    } else if (key === "unpublished") {
      workflowFilters = [];
      publicationFilter = "draft";
    } else if (key === "draft") {
      workflowFilters = ["draft"];
      publicationFilter = "draft";
    } else if (key === "review") {
      workflowFilters = ["submitted", "review"];
      publicationFilter = "";
    } else if (key === "rejected") {
      workflowFilters = ["revision_needed", "rejected"];
      publicationFilter = "draft";
    } else if (key === "approved") {
      workflowFilters = ["approved"];
      publicationFilter = "draft";
    } else if (key === "published") {
      workflowFilters = [];
      publicationFilter = "published";
    } else {
      workflowFilters = [key as WorkflowFilter];
      publicationFilter = "";
    }
    load(1, true, pageSize);
  }

  function workflowFiltersEqual(expected: WorkflowFilter[]): boolean {
    return (
      workflowFilters.length === expected.length &&
      expected.every((status) => workflowFilters.includes(status))
    );
  }

  function readInitialFilters() {
    if (typeof window === "undefined") return;
    const params = new URLSearchParams(window.location.search);
    search = params.get("q") ?? "";
    subjectFilter = params.get("subject_id") ?? "";
    workflowFilters = normalizeWorkflowFilters(params.getAll("workflow_status"));
    publicationFilter = normalizePublicationFilter(params.get("status"));
    questionTypeFilter = normalizeQuestionTypeFilter(
      params.get("question_type"),
    );
    hotsFilter = normalizeHotsFilter(params.get("hots"));
    authorFilter = params.get("author_username") ?? "";
    currentPage = normalizePage(params.get("page"), 1);
    pageSize = normalizePageSize(params.get("limit"), DEFAULT_PAGE_SIZE_OPTIONS, DEFAULT_PAGE_SIZE);
  }

  function handlePagination(change: PaginationChange) {
    load(change.reason === "limit" ? 1 : change.page, true, change.limit);
  }

  function normalizeWorkflowFilters(values: string[]): WorkflowFilter[] {
    const allowed = new Set(workflowOptions.map((option) => option.value));
    const selected: WorkflowFilter[] = [];
    for (const value of values) {
      for (const part of value.split(",")) {
        const normalized = part.trim() as WorkflowFilter;
        if (!allowed.has(normalized) || selected.includes(normalized)) continue;
        selected.push(normalized);
      }
    }
    return selected;
  }

  function toggleWorkflowFilter(value: WorkflowFilter) {
    workflowFilters = workflowFilters.includes(value)
      ? workflowFilters.filter((item) => item !== value)
      : [...workflowFilters, value];
    load(1, true);
  }

  function workflowFilterSummary(): string {
    if (workflowFilters.length === 0) return "Semua status pengerjaan";
    if (workflowFilters.length === 1) return workflowLabel(workflowFilters[0]);
    return `${workflowFilters.length} status dipilih`;
  }

  function normalizePublicationFilter(value: string | null): PublicationFilter {
    return value === "draft" || value === "published" ? value : "";
  }

  function normalizeQuestionTypeFilter(
    value: string | null,
  ): QuestionTypeFilter {
    return questionTypeOptions.some((option) => option.value === value)
      ? (value as QuestionTypeFilter)
      : "";
  }

  function normalizeHotsFilter(value: string | null): HotsFilter {
    return value === "yes" || value === "no" ? value : "";
  }

  function questionHref(question: Question): string {
    return resolve(`/bank-soal/soal/${encodeURIComponent(question.id)}`);
  }

  function reviewHref(): string {
    const params = new URLSearchParams();
    if (subjectFilter) params.set("subject_id", subjectFilter);
    return params.toString()
      ? `${resolve("/bank-soal/verifikasi")}?${params.toString()}`
      : resolve("/bank-soal/verifikasi");
  }

  function questionPratinjauHtml(question: Question): string {
    const html = (question.stem_html || question.question_text || "").trim();
    return html || "(Isi soal belum tersedia)";
  }

  function questionDetailHtml(question: Question): string {
    const html = (
      question.stem_html ||
      question.stem_latex ||
      question.question_text ||
      ""
    ).trim();
    return html || "(Isi soal belum tersedia)";
  }

  function questionStimulusHtml(question: Question): string {
    return (question.stimulus_html || "").trim();
  }

  function questionExplanationHtml(question: Question): string {
    return (question.explanation_html || question.explanation || "").trim();
  }

  function questionOptionRows(question: Question) {
    const optionRows = Array.isArray(question.options)
      ? question.options
          .map((option, index) => ({
            label: compactText(
              option.label || String.fromCharCode(65 + index),
              String.fromCharCode(65 + index),
            ),
            html: compactText(option.html || option.latex || option.text, ""),
          }))
          .filter((option) => option.html)
      : [];
    if (optionRows.length > 0) return optionRows;
    return [
      ["A", question.option_a],
      ["B", question.option_b],
      ["C", question.option_c],
      ["D", question.option_d],
      ["E", question.option_e],
    ]
      .map(([label, html]) => ({ label: String(label), html: compactText(html, "") }))
      .filter((option) => option.html);
  }

  function toggleQuestionDetail(questionId: string) {
    expandedQuestionId = expandedQuestionId === questionId ? null : questionId;
  }

  function isQuestionExpanded(questionId: string): boolean {
    return expandedQuestionId === questionId;
  }

  function normalizeBloomLevel(value: string | undefined | null): string {
    const upper = (value ?? "").toUpperCase();
    const match = upper.match(/C[1-6]/);
    return match?.[0] ?? "Lainnya";
  }

  function shortSubjectLabel(value: string): string {
    const compact = value.replace(/[^A-Za-z0-9]/g, "").toUpperCase();
    return (compact || "MAPEL").slice(0, 4);
  }

  function buildSubjectDistribution(
    items: Question[],
    loadedSubjects: Subject[],
  ): SubjectDistribution[] {
    const subjectNames = new Map<string, string>();
    loadedSubjects.forEach((subject) => {
      subjectNames.set(subject.id, subject.name);
    });

    const totals = new Map<string, { name: string; value: number }>();
    items.forEach((question) => {
      const key = question.subject_id || question.subject_name || "unknown";
      const name =
        question.subject_name ||
        (question.subject_id ? subjectNames.get(question.subject_id) : "") ||
        "Tanpa Mapel";
      const current = totals.get(key) ?? { name, value: 0 };
      current.value += 1;
      totals.set(key, current);
    });

    return Array.from(totals.values())
      .sort((a, b) => b.value - a.value)
      .slice(0, 8)
      .map((item, index) => ({
        label: shortSubjectLabel(item.name),
        name: item.name,
        value: item.value,
        color: subjectChartColors[index % subjectChartColors.length],
      }));
  }

  function buildBloomComposition(items: Question[]): BloomComposition[] {
    const totals = new Map<string, number>();
    items.forEach((question) => {
      const key = normalizeBloomLevel(question.cognitive_level);
      totals.set(key, (totals.get(key) ?? 0) + 1);
    });
    const total = Math.max(
      1,
      Array.from(totals.values()).reduce((sum, value) => sum + value, 0),
    );
    const ordered = [
      ...bloomOrder,
      ...Array.from(totals.keys()).filter((key) => !bloomOrder.includes(key)),
    ];
    return ordered.map((key) => {
      const value = totals.get(key) ?? 0;
      return {
        key,
        label: bloomLabels[key] ?? key,
        value,
        percent: Math.round((value / total) * 100),
      };
    });
  }

  function buildRecentActivities(items: Question[]): ActivityItem[] {
    return [...items]
      .sort(
        (a, b) =>
          new Date(b.updated_at ?? b.created_at ?? 0).getTime() -
          new Date(a.updated_at ?? a.created_at ?? 0).getTime(),
      )
      .slice(0, 5)
      .map((question) => ({
        id: question.id,
        actor: displayName(
          {
            display_name:
              question.author_display_name || question.reviewer_display_name,
            username: question.author_username || question.reviewer_username,
          },
          "Tim Bank Soal",
        ),
        action: activityAction(question),
        object: `${compactText(question.code, "Tanpa kode")} · ${compactText(question.material_topic || question.subject_name, "Soal")}`,
        time: formatDate(question.updated_at || question.created_at),
        tone: question.workflow_status ?? question.status ?? "draft",
      }));
  }

  function activityAction(question: Question): string {
    if (question.workflow_status === "review") return "mengirim untuk review";
    if (question.workflow_status === "approved") return "menyetujui";
    if (question.workflow_status === "rejected") return "meminta revisi";
    if (question.status === "published") return "menerbitkan";
    return "memperbarui draft";
  }

  function activityToneClass(tone: string): string {
    if (tone === "approved" || tone === "published")
      return "border-primary/20 bg-primary/10 text-primary";
    if (tone === "review")
      return "border-warning/30 bg-warning/10 text-warning";
    if (tone === "rejected")
      return "border-destructive/30 bg-destructive/10 text-destructive";
    return "border-border bg-muted/50 text-foreground";
  }

  function compactText(
    value: string | undefined | null,
    fallback = "-",
  ): string {
    const trimmed = (value ?? "").trim();
    return trimmed || fallback;
  }

  function subjectLabel(question: Question): string {
    const code = compactText(question.subject_code, "");
    const name = compactText(question.subject_name, "");
    if (name && code) return `${code} - ${name}`;
    return name || code || "Mapel belum tersedia";
  }

  function questionTypeLabel(value: string | undefined): string {
    return (
      questionTypeLabels[value ?? ""] ??
      compactText(value, "Tipe belum dipilih")
    );
  }

  function workflowLabel(value: string | undefined): string {
    return workflowLabels[value ?? ""] ?? compactText(value, "Belum ada alur");
  }

  function publicationLabel(value: string | undefined): string {
    return (
      publicationLabels[value ?? ""] ?? compactText(value, "Belum ada status")
    );
  }

  function difficultyLabel(value: string | undefined): string {
    return (
      difficultyLabels[value ?? ""] ??
      compactText(value, "Kesulitan belum diisi")
    );
  }

  function formatDate(value: string | undefined): string {
    if (!value) return "-";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "-";
    return new Intl.DateTimeFormat("id-ID", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    }).format(date);
  }

  function gradeLabel(question: Question): string {
    const level = compactText(question.target_level, "");
    const phase = compactText(question.academic_phase, "");
    const levelLabel =
      level && level.startsWith("Kelas ")
        ? level
        : level
          ? `Tingkat ${level}`
          : "";
    if (levelLabel && phase) return `${levelLabel} / Fase ${phase}`;
    return levelLabel || (phase ? `Fase ${phase}` : "Level belum diisi");
  }

  function questionUsageLocked(question: Question): boolean {
    const packageCount =
      question.package_count ?? question.usage?.package_count ?? 0;
    const answerCount =
      question.answer_count ?? question.usage?.answer_count ?? 0;
    return Boolean(
      question.is_locked ??
        question.usage?.is_locked ??
        (packageCount > 0 || answerCount > 0),
    );
  }

  function questionUsageText(question: Question): string {
    const packageCount =
      question.package_count ?? question.usage?.package_count ?? 0;
    const answerCount =
      question.answer_count ?? question.usage?.answer_count ?? 0;
    if (packageCount > 0 && answerCount > 0)
      return `${packageCount} paket, ${answerCount} jawaban`;
    if (packageCount > 0) return `${packageCount} paket`;
    if (answerCount > 0) return `${answerCount} jawaban`;
    return "Belum dipakai";
  }

  function isQuickEditable(question: Question): boolean {
    return (
      (question.workflow_status === "draft" ||
        question.workflow_status === "rejected" ||
        question.workflow_status === "revision_needed") &&
      (question.status ?? "draft") === "draft" &&
      !questionUsageLocked(question)
    );
  }

  function isSafeDeletable(question: Question): boolean {
    return (
      ["draft", "rejected"].includes(question.workflow_status ?? "draft") &&
      (question.status ?? "draft") === "draft" &&
      !questionUsageLocked(question)
    );
  }

  function quickEditLabel(question: Question): string {
    if (question.workflow_status === "revision_needed") return "Edit Revisi";
    return "Edit";
  }

  function canReturnToRevision(question: Question): boolean {
    return (
      canReview &&
      question.workflow_status === "approved" &&
      (question.status ?? "draft") === "draft" &&
      !questionUsageLocked(question)
    );
  }

  function canCreateRevision(question: Question): boolean {
    return (
      canCreate &&
      !isQuickEditable(question) &&
      (question.workflow_status === "approved" ||
        question.status === "published" ||
        questionUsageLocked(question))
    );
  }

  function canArchive(question: Question): boolean {
    const archiveable = ["approved", "published", "rejected"];
    return (
      canPublish &&
      !questionUsageLocked(question) &&
      question.status !== "archived" &&
      (archiveable.includes(question.workflow_status ?? "") ||
        question.status === "published")
    );
  }

  function canRestoreArchive(question: Question): boolean {
    return (
      canPublish &&
      question.workflow_status === "archived" &&
      question.status === "archived" &&
      !questionUsageLocked(question)
    );
  }

  function hasSecondaryActions(question: Question): boolean {
    return (
      canReturnToRevision(question) ||
      canCreateRevision(question) ||
      canArchive(question) ||
      canRestoreArchive(question) ||
      (canDelete && isSafeDeletable(question))
    );
  }

  function revisionDefaultNote(question: Question): string {
    if (question.status === "published" || questionUsageLocked(question))
      return "Membuat versi revisi baru agar riwayat paket/ujian lama tetap aman.";
    return "Dikembalikan ke revisi untuk perbaikan setelah review.";
  }

  async function returnToRevision(question: Question) {
    if (!canReturnToRevision(question) || revisingQuestionId) return;
    const notes = window.prompt(
      "Catatan revisi untuk soal ini:",
      revisionDefaultNote(question),
    );
    if (notes === null) return;
    revisingQuestionId = question.id;
    revisionError = "";
    try {
      await fetch(
        `/api/bank-soal/questions/${encodeURIComponent(question.id)}/workflow`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ action: "return_revision", notes }),
        },
      ).then((response) => readClientJson<Question>(response));
      load(currentPage, true);
    } catch (error) {
      revisionError =
        error instanceof Error
          ? error.message
          : "Soal belum dapat dikembalikan ke revisi";
    } finally {
      revisingQuestionId = null;
    }
  }

  async function createRevision(question: Question) {
    if (!canCreateRevision(question) || revisingQuestionId) return;
    const notes = window.prompt(
      "Catatan untuk draft revisi baru:",
      revisionDefaultNote(question),
    );
    if (notes === null) return;
    revisingQuestionId = question.id;
    revisionError = "";
    try {
      const row = await fetch(
        `/api/bank-soal/questions/${encodeURIComponent(question.id)}/revision`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ notes }),
        },
      ).then((response) => readClientJson<Question>(response));
      window.location.href = questionHref(row);
    } catch (error) {
      revisionError =
        error instanceof Error
          ? error.message
          : "Draft revisi baru belum dapat dibuat";
    } finally {
      revisingQuestionId = null;
    }
  }

  async function deleteQuestion(question: Question) {
    if (!canDelete || !isSafeDeletable(question) || deletingQuestionId) return;
    const code = compactText(question.code, "tanpa kode");
    const ok = window.confirm(
      `Hapus permanen soal ${code}? Tindakan ini hanya untuk soal draft atau soal ditolak yang belum dipakai di paket/jawaban siswa.`,
    );
    if (!ok) return;

    deletingQuestionId = question.id;
    deleteError = "";
    try {
      await fetch(
        `/api/bank-soal/questions/${encodeURIComponent(question.id)}`,
        { method: "DELETE" },
      ).then((response) => readClientJson<null>(response));
      load(
        Math.min(
          currentPage,
          Math.max(1, Math.ceil(Math.max(0, totalItems - 1) / pageSize)),
        ),
        true,
      );
    } catch (error) {
      deleteError =
        error instanceof Error ? error.message : "Soal belum dapat dihapus";
    } finally {
      deletingQuestionId = null;
    }
  }

  async function archiveQuestion(question: Question) {
    if (!canArchive(question) || archivingQuestionId) return;
    const code = compactText(question.code, "tanpa kode");
    const ok = window.confirm(
      `Arsipkan soal "${code}"? Soal akan disembunyikan dari daftar aktif tetapi tetap tersimpan untuk riwayat.`,
    );
    if (!ok) return;
    archivingQuestionId = question.id;
    try {
      await fetch(
        `/api/bank-soal/questions/${encodeURIComponent(question.id)}/workflow`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ action: "archive" }),
        },
      ).then((response) => readClientJson<unknown>(response));
      load(currentPage, true);
    } catch {
      // error handled via re-fetch
    } finally {
      archivingQuestionId = null;
    }
  }

  async function restoreArchive(question: Question) {
    if (!canRestoreArchive(question) || restoringQuestionId) return;
    const code = compactText(question.code, "tanpa kode");
    const notes = window.prompt(
      `Pulihkan soal arsip "${code}"? Soal akan kembali menjadi draft ditolak dan bisa direview/dihapus jika belum dipakai.`,
      "Dipulihkan dari arsip untuk ditinjau ulang.",
    );
    if (notes === null) return;
    restoringQuestionId = question.id;
    restoreError = "";
    try {
      await fetch(
        `/api/bank-soal/questions/${encodeURIComponent(question.id)}/workflow`,
        {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ action: "restore_archive", notes }),
        },
      ).then((response) => readClientJson<Question>(response));
      load(currentPage, true);
    } catch (error) {
      restoreError =
        error instanceof Error
          ? error.message
          : "Soal arsip belum dapat dipulihkan";
    } finally {
      restoringQuestionId = null;
    }
  }

  function workflowBadgeClass(value: string | undefined): string {
    switch (value) {
      case "review":
        return "border-warning/30 bg-warning/10 text-warning";
      case "approved":
        return "border-primary/20 bg-primary/10 text-primary";
      case "rejected":
        return "border-destructive/30 bg-destructive/10 text-destructive";
      case "draft":
        return "border-border bg-muted/50 text-foreground";
      default:
        return "border-border bg-card text-muted-foreground";
    }
  }

  function publicationBadgeClass(value: string | undefined): string {
    if (value === "published")
      return "border-success/20 bg-success/10 text-success";
    return "border-border bg-card text-muted-foreground";
  }

  function summaryCardClass(card: SummaryCard): string {
    const active = card.active
      ? "border-primary bg-primary/10 shadow-sm"
      : "border-border bg-card hover:border-primary/20 hover:bg-primary/10";
    return `rounded-lg border p-4 text-left transition ${active}`;
  }

  function summaryValueClass(card: SummaryCard): string {
    if (card.tone === "amber") return "text-warning";
    if (card.tone === "red") return "text-destructive";
    if (card.tone === "green") return "text-success";
    if (card.tone === "emerald") return "text-primary";
    return "text-foreground";
  }

  function handleQuestionsRenderError(error: unknown, reset: () => void) {
    console.error("Daftar Bank Soal belum dapat ditampilkan", error);
    reset();
  }

  function canAccessQuality(user: PageData["user"]): boolean {
    const userRoles = user?.roles ?? (user?.role ? [user.role] : []);
    const permissions = user?.permissions ?? [];
    return (
      userRoles.includes("admin") || permissions.includes("bank_soal.analytics")
    );
  }

  async function fetchAuthors() {
    if (authorsLoaded) return;
    try {
      const payload = await fetch("/api/bank-soal/soal-support/authors").then(
        (response) =>
          readClientApiData<Author[]>(
            response,
            "Gagal memuat data pembuat soal",
          ),
      );
      authors = payload ?? [];
      authorsLoaded = true;
    } catch {
      // silently fail, dropdown shows empty
    }
  }

  onMount(() => {
    readInitialFilters();
    load(currentPage);
    fetchAuthors();
    return () => {
      if (searchTimer) clearTimeout(searchTimer);
    };
  });
</script>

<svelte:head>
  <title>Kelola Soal - Bank Soal - MTsN 2 Kolaka Utara</title>
</svelte:head>

<div class="space-y-5">
  {#if showList && !showDashboard}
    <section
      class="rounded-xl border border-border bg-card p-4 shadow-sm md:p-5"
    >
      <div
        class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between"
      >
        <div class="min-w-0">
          <p
            class="text-[11px] font-semibold uppercase tracking-[0.18em] text-primary"
          >
            Kelola Soal
          </p>
          <h1
            class="mt-1 text-2xl font-semibold tracking-tight text-foreground md:text-3xl"
          >
            Daftar Bank Soal
          </h1>
          <p class="mt-2 max-w-3xl text-sm leading-6 text-muted-foreground">
            Cari, filter, edit, dan cek status soal. Tambah dan impor tetap
            tersedia sebagai aksi cepat, bukan area kerja terpisah di daftar.
          </p>
        </div>
        <div class="flex shrink-0 flex-wrap gap-2">
          {#if canCreate}
            <Button href={composerHref}>
              <PlusIcon class="size-4" />
              Tambah Soal
            </Button>
          {/if}
          <Button href={printHref} variant="outline">
            <PrinterIcon class="size-4" />
            Cetak Soal Saya
          </Button>
          {#if canReview}
            <Button href={reviewRouteHref} variant="outline">
              <ClipboardCheckIcon class="size-4" />
              Verifikasi
            </Button>
          {/if}
          {#if canImport}
            <Button href={importHref} variant="outline">
              <UploadIcon class="size-4" />
              Impor
            </Button>
          {/if}
          {#if canQuality}
            <Button href={analysisHref} variant="outline">Mutu Soal</Button>
          {/if}
        </div>
      </div>
    </section>
  {/if}

  {#if showDashboard}
    <section
      class="overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-br from-emerald-950 via-emerald-900 to-amber-900 text-white shadow-sm"
    >
      <div class="relative p-5 md:p-6">
        <div
          class="absolute right-6 top-6 hidden h-28 w-28 rounded-full bg-amber-300/20 blur-2xl lg:block"
        ></div>
        <div
          class="relative flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between"
        >
          <div class="max-w-3xl space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <Badge
                variant="outline"
                class="border-white/25 bg-white/10 text-white"
                >SCS · Bank Soal</Badge
              >
              <Badge
                variant="outline"
                class="border-primary/20 bg-emerald-200/15 text-emerald-50"
                >{roleLabel}</Badge
              >
            </div>
            <div>
              <h1 class="text-2xl font-semibold tracking-tight md:text-3xl">
                Dashboard Bank Soal
              </h1>
              <p class="mt-2 max-w-2xl text-sm leading-6 text-emerald-50/85">
                Ringkasan koleksi soal MTsN 2 Kolaka Utara: tulis, verifikasi,
                impor, dan siapkan butir terbaik untuk paket asesmen.
              </p>
            </div>
            <div class="flex flex-wrap gap-2 text-xs text-emerald-50/80">
              <span class="rounded-full bg-white/10 px-3 py-1"
                >Tahun Pelajaran 2025/2026</span
              >
              <span class="rounded-full bg-white/10 px-3 py-1"
                >Semester berjalan</span
              >
              <span class="rounded-full bg-white/10 px-3 py-1"
                >{subjects.length} mapel terdaftar</span
              >
            </div>
          </div>
          <div class="flex flex-wrap gap-2">
            {#if canReview}
              <Button
                href={reviewRouteHref}
                variant="outline"
                class="border-white/30 bg-white/10 text-white hover:bg-white/20"
              >
                <ClipboardCheckIcon class="size-4" />
                Verifikasi
              </Button>
            {/if}
            <Button
              href={printHref}
              variant="outline"
              class="border-white/30 bg-white/10 text-white hover:bg-white/20"
            >
              <PrinterIcon class="size-4" />
              Cetak Soal Saya
            </Button>
            {#if canImport}
              <Button
                href={importHref}
                variant="outline"
                class="border-white/30 bg-white/10 text-white hover:bg-white/20"
              >
                <UploadIcon class="size-4" />
                Impor
              </Button>
            {/if}
            {#if canCreate}
              <Button
                href={composerHref}
                class="bg-card text-primary hover:bg-primary/10"
              >
                <PlusIcon class="size-4" />
                Soal Baru
              </Button>
            {/if}
          </div>
        </div>
      </div>
    </section>

    <section
      class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4"
      aria-label="Ringkasan utama bank soal"
    >
      <button
        type="button"
        class={summaryCardClass(summaryCards[0])}
        aria-pressed={summaryCards[0].active}
        onclick={() => setSummaryFilter("all")}
      >
        <div class="flex items-center justify-between gap-3">
          <span
            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >Total Soal</span
          >
          <FileQuestionIcon class="size-5 text-primary" />
        </div>
        <span
          class={`mt-2 block text-3xl font-semibold ${summaryValueClass(summaryCards[0])}`}
          >{counts.all}</span
        >
        <span class="mt-1 block text-xs text-muted-foreground"
          >stok sesuai filter aktif</span
        >
      </button>
      <button
        type="button"
        class={summaryCardClass(summaryCards[4])}
        aria-pressed={summaryCards[4].active}
        onclick={() => setSummaryFilter("approved")}
      >
        <div class="flex items-center justify-between gap-3">
          <span
            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >Telah Diverifikasi</span
          >
          <BookOpenCheckIcon class="size-5 text-primary" />
        </div>
        <span
          class={`mt-2 block text-3xl font-semibold ${summaryValueClass(summaryCards[4])}`}
          >{counts.approved + counts.published}</span
        >
        <span class="mt-1 block text-xs text-muted-foreground"
          >disetujui atau sudah terbit</span
        >
      </button>
      <a
        href={packageHref}
        class="rounded-lg border border-border bg-card p-4 text-left transition hover:border-primary/20 hover:bg-primary/10"
      >
        <div class="flex items-center justify-between gap-3">
          <span
            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >Dipakai di Paket</span
          >
          <PackageIcon class="size-5 text-warning" />
        </div>
        <span class="mt-2 block text-3xl font-semibold text-foreground"
          >{totalPackageUsage}</span
        >
        <span class="mt-1 block text-xs text-muted-foreground"
          >estimasi dari daftar termuat</span
        >
      </a>
      <button
        type="button"
        class={summaryCardClass(summaryCards[2])}
        aria-pressed={summaryCards[2].active}
        onclick={() => setSummaryFilter("review")}
      >
        <div class="flex items-center justify-between gap-3">
          <span
            class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >Menunggu Verifikasi</span
          >
          <ClipboardCheckIcon class="size-5 text-warning" />
        </div>
        <span
          class={`mt-2 block text-3xl font-semibold ${summaryValueClass(summaryCards[2])}`}
          >{counts.review}</span
        >
        <span class="mt-1 block text-xs text-muted-foreground"
          >perlu keputusan reviewer</span
        >
      </button>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.4fr_1fr]">
      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-foreground">
              Distribusi soal per mata pelajaran
            </h2>
            <p class="mt-1 text-xs text-muted-foreground">
              Diurutkan dari daftar soal yang sedang termuat
            </p>
          </div>
          <BarChart3Icon class="size-5 text-muted-foreground" />
        </div>
        {#if subjectDistribution.length > 0}
          <div class="mt-5 flex h-44 items-end gap-2">
            {#each subjectDistribution as item (item.name)}
              <div
                class="flex min-w-0 flex-1 flex-col items-center gap-2"
                title={`${item.name}: ${item.value} soal`}
              >
                <div
                  class="flex h-32 w-full items-end justify-center rounded-t bg-muted/50 px-1"
                >
                  <div
                    class={`w-full max-w-8 rounded-t ${item.color}`}
                    style={`height: ${Math.max(12, Math.round((item.value / Math.max(...subjectDistribution.map((entry) => entry.value), 1)) * 100))}%`}
                  ></div>
                </div>
                <span
                  class="max-w-full truncate font-mono text-[10px] uppercase tracking-wide text-muted-foreground"
                  >{item.label}</span
                >
              </div>
            {/each}
          </div>
        {:else}
          <p
            class="mt-6 rounded-lg border border-dashed border-border bg-muted/50 p-4 text-sm text-muted-foreground"
          >
            Distribusi mapel akan muncul setelah daftar soal dimuat.
          </p>
        {/if}
      </div>

      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-foreground">
              Komposisi level kognitif
            </h2>
            <p class="mt-1 text-xs text-muted-foreground">
              Taksonomi Bloom C1-C6
            </p>
          </div>
          <Layers3Icon class="size-5 text-muted-foreground" />
        </div>
        <div class="mt-4 space-y-3">
          {#each bloomComposition as item (item.key)}
            <div>
              <div class="mb-1 flex items-center justify-between gap-3 text-xs">
                <span class="font-medium text-foreground">{item.label}</span>
                <span class="font-mono text-muted-foreground"
                  >{item.value} · {item.percent}%</span
                >
              </div>
              <div class="h-2 rounded-full bg-muted">
                <div
                  class="h-2 rounded-full bg-primary"
                  style={`width: ${Math.max(item.percent, item.value > 0 ? 6 : 0)}%`}
                ></div>
              </div>
            </div>
          {/each}
        </div>
      </div>
    </section>

    <section class="grid gap-4 xl:grid-cols-[1.35fr_1fr]">
      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-foreground">
              Aktivitas Terbaru
            </h2>
            <p class="mt-1 text-xs text-muted-foreground">
              Perubahan terbaru dari daftar soal yang termuat
            </p>
          </div>
          <HistoryIcon class="size-5 text-muted-foreground" />
        </div>
        {#if recentActivities.length > 0}
          <div class="mt-4 divide-y divide-border">
            {#each recentActivities as item (item.id)}
              <div class="flex items-center gap-3 py-3">
                <div
                  class="flex size-9 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-semibold text-primary"
                >
                  {item.actor.slice(0, 2).toUpperCase()}
                </div>
                <div class="min-w-0 flex-1">
                  <p class="truncate text-sm text-foreground">
                    <span class="font-semibold text-foreground"
                      >{item.actor}</span
                    >
                    {item.action}
                    <span class="font-mono text-xs">{item.object}</span>
                  </p>
                  <p class="mt-0.5 text-xs text-muted-foreground">
                    {item.time}
                  </p>
                </div>
                <Badge variant="outline" class={activityToneClass(item.tone)}
                  >{workflowLabel(item.tone)}</Badge
                >
              </div>
            {/each}
          </div>
        {:else}
          <p
            class="mt-4 rounded-lg border border-dashed border-border bg-muted/50 p-4 text-sm text-muted-foreground"
          >
            Aktivitas terbaru akan muncul setelah ada soal yang dimuat atau
            diperbarui.
          </p>
        {/if}
      </div>

      <div class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex items-start justify-between gap-3">
          <div>
            <h2 class="text-base font-semibold text-foreground">Aksi Cepat</h2>
            <p class="mt-1 text-xs text-muted-foreground">
              Pintasan alur utama Bank Soal
            </p>
          </div>
          <SparklesIcon class="size-5 text-warning" />
        </div>
        <div class="mt-4 grid gap-2">
          <a
            href={listHref}
            class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
          >
            <span
              class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-violet-100 text-violet-800"
              ><FileQuestionIcon class="size-5" /></span
            >
            <span class="min-w-0 flex-1"
              ><span class="block text-sm font-semibold text-foreground"
                >Buka daftar soal</span
              ><span class="block text-xs text-muted-foreground"
                >Kelola filter, status, dan pagination soal</span
              ></span
            >
          </a>
          {#if canCreate}
            <a
              href={composerHref}
              class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
            >
              <span
                class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/15 text-primary"
                ><PlusIcon class="size-5" /></span
              >
              <span class="min-w-0 flex-1"
                ><span class="block text-sm font-semibold text-foreground"
                  >Tambah soal baru</span
                ><span class="block text-xs text-muted-foreground"
                  >PG, essay, benar/salah, menjodohkan</span
                ></span
              >
            </a>
          {/if}
          <a
            href={packageHref}
            class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
          >
            <span
              class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-warning/15 text-warning"
              ><PackageIcon class="size-5" /></span
            >
            <span class="min-w-0 flex-1"
              ><span class="block text-sm font-semibold text-foreground"
                >Buat paket asesmen</span
              ><span class="block text-xs text-muted-foreground"
                >Gunakan soal terbit di modul Asesmen</span
              ></span
            >
          </a>
          {#if canImport}
            <a
              href={importHref}
              class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
            >
              <span
                class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-accent text-accent-foreground"
                ><UploadIcon class="size-5" /></span
              >
              <span class="min-w-0 flex-1"
                ><span class="block text-sm font-semibold text-foreground"
                  >Impor dari Word/Excel</span
                ><span class="block text-xs text-muted-foreground"
                  >Pratinjau, pemetaan, lalu konfirmasi</span
                ></span
              >
            </a>
          {/if}
          {#if canReview}
            <a
              href={reviewRouteHref}
              class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
            >
              <span
                class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-destructive/15 text-destructive"
                ><ClipboardCheckIcon class="size-5" /></span
              >
              <span class="min-w-0 flex-1"
                ><span class="block text-sm font-semibold text-foreground"
                  >Verifikasi antrean</span
                ><span class="block text-xs text-muted-foreground"
                  >Setujui atau kembalikan untuk revisi</span
                ></span
              >
            </a>
          {/if}
          <a
            href={analysisHref}
            class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
          >
            <span
              class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-indigo-100 text-indigo-800"
              ><BarChart3Icon class="size-5" /></span
            >
            <span class="min-w-0 flex-1"
              ><span class="block text-sm font-semibold text-foreground"
                >Mutu Soal</span
              ><span class="block text-xs text-muted-foreground"
                >Pantau kualitas, HOTS, dan pemakaian soal</span
              ></span
            >
          </a>
          <a
            href={mapelKdHref}
            class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
          >
            <span
              class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/15 text-primary"
              ><Layers3Icon class="size-5" /></span
            >
            <span class="min-w-0 flex-1"
              ><span class="block text-sm font-semibold text-foreground"
                >Mapel & KD</span
              ><span class="block text-xs text-muted-foreground"
                >Cek coverage metadata dan materi</span
              ></span
            >
          </a>
          {#if canSettings}
            <a
              href={settingsHref}
              class="group flex items-center gap-3 rounded-lg border border-border bg-muted/50 p-3 transition hover:border-primary/20 hover:bg-primary/10"
            >
              <span
                class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-border text-foreground"
                ><SettingsIcon class="size-5" /></span
              >
              <span class="min-w-0 flex-1"
                ><span class="block text-sm font-semibold text-foreground"
                  >Pengaturan & SOP</span
                ><span class="block text-xs text-muted-foreground"
                  >Alur kerja, standar kualitas, dan integrasi</span
                ></span
              >
            </a>
          {/if}
        </div>
      </div>
    </section>
  {/if}

  {#if showList}
    <section class="rounded-lg border border-border bg-card p-4 shadow-sm">
      <form class="space-y-4" onsubmit={applyFilters}>
        <div
          class="grid min-w-0 gap-3 lg:grid-cols-[minmax(0,1fr)_auto] lg:items-end"
        >
          <div class="min-w-0 space-y-1">
            <label
              for="bank-soal-search"
              class="text-xs font-semibold text-muted-foreground"
              >Cari soal</label
            >
            <div class="relative min-w-0">
              <SearchIcon
                class="pointer-events-none absolute left-2.5 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
              />
              <Input
                id="bank-soal-search"
                value={search}
                oninput={onSearchInput}
                placeholder="Kode, isi soal, materi..."
                class="h-9 min-w-0 pl-8"
              />
            </div>
          </div>
          <div class="flex min-w-0 flex-wrap gap-2 lg:justify-end">
            <Button type="submit" variant="outline" class="h-9"
              ><ListFilterIcon class="size-4" />Terapkan</Button
            >
            <LoadingButton
              variant="outline"
              class="h-9"
              loading={refreshing}
              loadingLabel="Memuat"
              title="Muat ulang daftar soal"
              aria-label="Muat ulang daftar soal"
              onclick={() => load(currentPage, true)}
              ><RefreshCcwIcon class="size-4" /></LoadingButton
            >
            {#if hasFilters}<Button
                type="button"
                variant="ghost"
                class="h-9 text-primary"
                onclick={clearFilters}>Bersihkan</Button
              >{/if}
          </div>
        </div>

        <div class="border-t border-border/50 pt-3">
          <p class="text-xs font-semibold text-muted-foreground">
            Status cepat
          </p>
          <div class="mt-2 flex flex-wrap gap-2">
            {#each summaryCards as card (card.key)}
              <button
                type="button"
                class={`rounded-full border px-3 py-1.5 text-xs font-semibold transition ${card.active ? "border-primary bg-primary/10 text-primary" : "border-border bg-background text-muted-foreground hover:border-primary/30 hover:text-foreground"}`}
                aria-pressed={card.active}
                onclick={() => setSummaryFilter(card.key)}
              >
                {card.label}<span class="ml-1 text-[11px] opacity-70"
                  >{card.value}</span
                >
              </button>
            {/each}
          </div>
        </div>

        <div class="space-y-3 border-t border-border/50 pt-3">
          <div>
            <p class="text-xs font-semibold text-muted-foreground">
              Filter lanjutan
            </p>
            <p class="text-[11px] text-muted-foreground">
              Mapel, pembuat, tipe, ketersediaan, dan HOTS.
            </p>
          </div>
          <div
            class="grid min-w-0 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6"
          >
            <div class="min-w-0 space-y-1">
              <label
                for="bank-soal-subject"
                class="text-xs font-semibold text-muted-foreground"
                >Mata pelajaran</label
              ><select
                id="bank-soal-subject"
                bind:value={subjectFilter}
                onchange={() => load(1, true)}
                class="h-9 min-w-0 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                ><option value="">Semua mapel</option
                >{#each subjects as subject (subject.id)}<option
                    value={subject.id}
                    >{subject.code
                      ? `${subject.code} - ${subject.name}`
                      : subject.name}</option
                  >{/each}</select
              >
            </div>
            <div class="min-w-0 space-y-1">
              <label
                for="bank-soal-author"
                class="text-xs font-semibold text-muted-foreground"
                >Pembuat</label
              ><select
                id="bank-soal-author"
                bind:value={authorFilter}
                onchange={() => load(1, true)}
                class="h-9 min-w-0 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                ><option value="">Semua pembuat</option
                >{#each authors as author (author.username)}<option
                    value={author.username}>{author.display_name}</option
                  >{/each}</select
              >
            </div>
            <div class="relative min-w-0 space-y-1">
              <span
                id="bank-soal-workflow-label"
                class="text-xs font-semibold text-muted-foreground"
                >Status pengerjaan</span
              ><button
                type="button"
                aria-labelledby="bank-soal-workflow-label"
                aria-haspopup="listbox"
                aria-expanded={workflowDropdownOpen}
                onclick={() => (workflowDropdownOpen = !workflowDropdownOpen)}
                class="flex h-9 min-w-0 w-full items-center justify-between gap-2 rounded-md border border-border bg-card px-3 text-left text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                ><span class="truncate">{workflowFilterSummary()}</span
                ><span class="text-xs text-muted-foreground">▾</span></button
              >
              {#if workflowDropdownOpen}
                <div
                  class="absolute z-30 mt-1 max-h-72 w-full min-w-64 overflow-y-auto rounded-md border border-border bg-card p-2 shadow-lg"
                  role="listbox"
                  aria-multiselectable="true"
                >
                  <button
                    type="button"
                    class="mb-2 h-8 w-full rounded-md px-2 text-left text-xs font-semibold text-muted-foreground hover:bg-muted"
                    onclick={() => {
                      workflowFilters = [];
                      load(1, true);
                    }}>Semua status pengerjaan</button
                  >
                  {#each workflowOptions as option (option.value)}
                    <label
                      class="flex cursor-pointer items-center gap-2 rounded-md px-2 py-2 text-sm hover:bg-muted"
                    >
                      <input
                        type="checkbox"
                        class="h-4 w-4 rounded border-border"
                        checked={workflowFilters.includes(option.value)}
                        onchange={() => toggleWorkflowFilter(option.value)}
                      />
                      <span>{option.label}</span>
                    </label>
                  {/each}
                  <div class="mt-2 flex justify-end border-t border-border/60 pt-2">
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onclick={() => (workflowDropdownOpen = false)}>Selesai</Button
                    >
                  </div>
                </div>
              {/if}
            </div>
            <div class="min-w-0 space-y-1">
              <label
                for="bank-soal-publication"
                class="text-xs font-semibold text-muted-foreground"
                >Ketersediaan</label
              ><select
                id="bank-soal-publication"
                bind:value={publicationFilter}
                onchange={() => load(1, true)}
                class="h-9 min-w-0 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                >{#each publicationOptions as option (option.value)}<option
                    value={option.value}>{option.label}</option
                  >{/each}</select
              >
            </div>
            <div class="min-w-0 space-y-1">
              <label
                for="bank-soal-type"
                class="text-xs font-semibold text-muted-foreground">Tipe</label
              ><select
                id="bank-soal-type"
                bind:value={questionTypeFilter}
                onchange={() => load(1, true)}
                class="h-9 min-w-0 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                >{#each questionTypeOptions as option (option.value)}<option
                    value={option.value}>{option.label}</option
                  >{/each}</select
              >
            </div>
            <div class="min-w-0 space-y-1">
              <label
                for="bank-soal-hots"
                class="text-xs font-semibold text-muted-foreground">HOTS</label
              ><select
                id="bank-soal-hots"
                bind:value={hotsFilter}
                onchange={() => load(1, true)}
                class="h-9 min-w-0 w-full rounded-md border border-border bg-card px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
                >{#each hotsOptions as option (option.value)}<option
                    value={option.value}>{option.label}</option
                  >{/each}</select
              >
            </div>
          </div>
        </div>
      </form>

      {#if hasFilters}
        <div
          class="mt-3 flex flex-wrap items-center gap-2 border-t border-border/50 pt-3 text-xs text-muted-foreground"
          aria-label="Filter aktif"
        >
          <span class="font-semibold text-foreground">Filter aktif:</span>
          {#if search.trim()}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                search = "";
                load(1, true);
              }}>Cari: {search.trim()} ×</Button
            >{/if}
          {#if selectedSubject}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                subjectFilter = "";
                load(1, true);
              }}>Mapel: {selectedSubject.name} ×</Button
            >{/if}
          {#if selectedAuthor}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                authorFilter = "";
                load(1, true);
              }}>Pembuat: {selectedAuthor.display_name} ×</Button
            >{/if}
          {#each workflowFilters as status (status)}
            <Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                workflowFilters = workflowFilters.filter((item) => item !== status);
                load(1, true);
              }}>Status: {workflowLabel(status)} ×</Button
            >
          {/each}
          {#if publicationFilter}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                publicationFilter = "";
                load(1, true);
              }}>Ketersediaan: {publicationLabel(publicationFilter)} ×</Button
            >{/if}
          {#if questionTypeFilter}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                questionTypeFilter = "";
                load(1, true);
              }}>Tipe: {questionTypeLabel(questionTypeFilter)} ×</Button
            >{/if}
          {#if hotsFilter}<Button
              type="button"
              variant="outline"
              size="sm"
              class="h-7 rounded-full"
              onclick={() => {
                hotsFilter = "";
                load(1, true);
              }}>{hotsFilter === "yes" ? "HOTS" : "Non-HOTS"} ×</Button
            >{/if}
          <Button
            variant="ghost"
            size="sm"
            class="h-7 text-primary"
            onclick={clearFilters}>Bersihkan semua</Button
          >
        </div>
      {/if}
    </section>


    {#snippet questionDetailPanel(question: Question)}
      {@const options = questionOptionRows(question)}
      {@const explanation = questionExplanationHtml(question)}
      {@const stimulus = questionStimulusHtml(question)}
      <div
        class="space-y-4 rounded-2xl border border-border bg-muted/20 p-4 text-sm"
      >
        <div class="flex flex-wrap items-center gap-2">
          <Badge variant="outline" class="font-mono text-xs">
            {compactText(question.code, "Tanpa kode")}
          </Badge>
          <Badge variant="outline" class={workflowBadgeClass(question.workflow_status)}>
            {workflowLabel(question.workflow_status)}
          </Badge>
          <Badge variant="outline" class={publicationBadgeClass(question.status)}>
            {publicationLabel(question.status)}
          </Badge>
          <Badge variant="secondary" class="text-xs">
            {questionTypeLabel(question.question_type)}
          </Badge>
          {#if question.hots_flag}
            <Badge
              variant="outline"
              class="border-warning/30 bg-warning/10 text-warning"
              >HOTS</Badge
            >
          {/if}
        </div>

        {#if stimulus}
          <section class="space-y-2">
            <p
              class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >
              Stimulus
            </p>
            <div class="overflow-x-auto rounded-xl border bg-background p-3">
              <RichContent html={stimulus} class="max-w-none text-sm leading-6" />
            </div>
          </section>
        {/if}

        <section class="space-y-2">
          <p class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
            Soal
          </p>
          <div class="overflow-x-auto rounded-xl border bg-background p-3">
            <RichContent
              html={questionDetailHtml(question)}
              class="max-w-none break-words text-sm leading-6 [overflow-wrap:anywhere] [&_table]:min-w-max [&_table]:border-collapse [&_td]:border [&_td]:border-border [&_td]:px-3 [&_td]:py-2 [&_th]:border [&_th]:border-border [&_th]:px-3 [&_th]:py-2"
            />
          </div>
        </section>

        <section class="space-y-2">
          <div class="flex items-center justify-between gap-2">
            <p
              class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >
              Pilihan Jawaban
            </p>
            {#if canSeeAnswerKey && question.answer_key}
              <Badge variant="outline" class="bg-primary/10 text-primary">
                Kunci: {question.answer_key}
              </Badge>
            {/if}
          </div>
          {#if options.length > 0}
            <div class="grid gap-2 md:grid-cols-2">
              {#each options as option, index (`${option.label}-${index}`)}
                <div class="rounded-xl border bg-background p-3">
                  <div class="flex min-w-0 gap-2">
                    <span class="font-semibold text-primary">{option.label}.</span>
                    <RichContent
                      html={option.html}
                      class="min-w-0 flex-1 break-words text-sm leading-6 [overflow-wrap:anywhere]"
                    />
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <p class="rounded-xl border bg-background p-3 text-muted-foreground">
              Pilihan jawaban belum tersedia pada ringkasan daftar.
            </p>
          {/if}
        </section>

        {#if explanation}
          <section class="space-y-2">
            <p
              class="text-xs font-semibold uppercase tracking-wide text-muted-foreground"
            >
              Pembahasan
            </p>
            <div class="overflow-x-auto rounded-xl border bg-background p-3">
              <RichContent html={explanation} class="max-w-none text-sm leading-6" />
            </div>
          </section>
        {/if}

        <section
          class="grid gap-2 rounded-xl border bg-background p-3 text-xs text-muted-foreground sm:grid-cols-2 lg:grid-cols-4"
        >
          <p>
            <span class="font-medium text-foreground">Mapel:</span>
            {subjectLabel(question)}
          </p>
          <p>
            <span class="font-medium text-foreground">Tingkat:</span>
            {gradeLabel(question)}
          </p>
          <p>
            <span class="font-medium text-foreground">Materi:</span>
            {compactText(question.material_topic, "Belum diisi")}
          </p>
          <p>
            <span class="font-medium text-foreground">Level:</span>
            {compactText(question.cognitive_level, "Belum diisi")}
          </p>
          <p>
            <span class="font-medium text-foreground">Pembuat:</span>
            {displayName(
              {
                display_name: question.author_display_name,
                username: question.author_username,
              },
              "Penulis belum tercatat",
            )}
          </p>
          <p>
            <span class="font-medium text-foreground">Pemakaian:</span>
            {questionUsageText(question)}
          </p>
          <p>
            <span class="font-medium text-foreground">Kesulitan:</span>
            {difficultyLabel(question.difficulty)}
          </p>
          <p>
            <span class="font-medium text-foreground">Dibuat:</span>
            {formatDate(question.created_at)}
          </p>
        </section>

        {#if question.writer_notes || question.review_notes}
          <section
            class="space-y-1 rounded-xl border border-warning/30 bg-warning/10 p-3 text-xs text-warning"
          >
            {#if question.writer_notes}
              <p><span class="font-semibold">Catatan penulis:</span> {question.writer_notes}</p>
            {/if}
            {#if question.review_notes}
              <p><span class="font-semibold">Catatan reviewer:</span> {question.review_notes}</p>
            {/if}
          </section>
        {/if}
      </div>
    {/snippet}

    <AsyncContent
      promise={questionsPromise}
      onerror={handleQuestionsRenderError}
    >
      {#snippet pending()}
        <section class="rounded-lg border border-border bg-card p-4 shadow-sm">
          <div class="flex items-center justify-between">
            <Skeleton class="h-5 w-40" />
            <Skeleton class="h-5 w-24" />
          </div>
          <div class="mt-4 space-y-3">
            {#each Array.from( { length: 6 }, ) as _, index (`bank-soal-skeleton-${index}`)}
              <div class="rounded-lg border border-border p-4">
                <Skeleton class="h-4 w-28" />
                <Skeleton class="mt-3 h-5 w-full max-w-2xl" />
                <div class="mt-3 flex gap-2">
                  <Skeleton class="h-5 w-20" />
                  <Skeleton class="h-5 w-24" />
                  <Skeleton class="h-5 w-16" />
                </div>
              </div>
            {/each}
          </div>
        </section>
      {/snippet}

      {#snippet failed(error, reset)}
        <section class="rounded-lg border border-border bg-card p-5 shadow-sm">
          <RecoveryPanel
            title="Daftar soal belum dapat dimuat"
            message={error instanceof Error
              ? error.message
              : "Gagal memuat daftar soal"}
            onRetry={() => {
              reset?.();
              load(currentPage, true);
            }}
          />
        </section>
      {/snippet}

      {#snippet children(value)}
        {@const overview = value as BankSoalOverview}
        {@const currentQuestions = overview.questions}
        <section class="rounded-lg border border-border bg-card shadow-sm">
          {#if deleteError || revisionError || restoreError}
            <div
              role="alert"
              class="border-b border-destructive/20 bg-destructive/10 px-4 py-3 text-sm text-destructive"
            >
              {deleteError || revisionError || restoreError}
            </div>
          {/if}
          <div
            class="flex flex-col gap-2 border-b border-border p-4 md:flex-row md:items-center md:justify-between"
          >
            <div>
              <h2 class="text-base font-semibold text-foreground">
                Soal Tersedia
              </h2>
              {#if totalItems === 0}
                <p class="mt-1 text-xs text-muted-foreground">
                  Tidak ada soal pada filter ini
                </p>
              {/if}
            </div>
            <div class="flex flex-wrap gap-2">
              {#if selectedSubject}
                <Badge
                  variant="outline"
                  class="border-primary/20 bg-primary/10 text-primary"
                  >{selectedSubject.name}</Badge
                >
              {/if}
            </div>
          </div>

          {#if currentQuestions.length === 0}
            <div class="flex flex-col items-center px-6 py-12 text-center">
              <div
                class="flex size-12 items-center justify-center rounded-lg bg-primary/10 text-primary"
              >
                <FileQuestionIcon class="size-6" />
              </div>
              <h2 class="mt-4 text-lg font-semibold text-foreground">
                {hasFilters
                  ? "Soal tidak ditemukan"
                  : "Belum ada soal di Bank Soal"}
              </h2>
              <p class="mt-2 max-w-md text-sm leading-6 text-muted-foreground">
                {hasFilters
                  ? "Coba longgarkan filter atau cari dengan kode, materi, dan isi soal yang lebih umum."
                  : "Mulai dari penyusun soal untuk menulis soal pertama, atau impor file soal dari arsip lama."}
              </p>
              <div class="mt-5 flex flex-wrap justify-center gap-2">
                {#if hasFilters}
                  <Button variant="outline" onclick={clearFilters}
                    >Bersihkan filter</Button
                  >
                {/if}
                {#if canCreate}
                  <Button href={composerHref}>
                    <PlusIcon class="size-4" />
                    Tambah Soal
                  </Button>
                {/if}
                {#if canImport}
                  <Button href={importHref} variant="outline">
                    <UploadIcon class="size-4" />
                    Impor File
                  </Button>
                {/if}
              </div>
            </div>
          {:else}
            <div class="hidden overflow-x-auto md:block">
              <Table.Root class="min-w-[980px]">
                <Table.Caption class="sr-only"
                  >Daftar soal tersedia di Bank Soal</Table.Caption
                >
                <Table.Header>
                  <Table.Row class="bg-muted/50">
                    <Table.Head
                      scope="col"
                      class="w-[40%] text-muted-foreground">Soal</Table.Head
                    >
                    <Table.Head
                      scope="col"
                      class="w-[18%] text-muted-foreground"
                      >Mapel & Level</Table.Head
                    >
                    <Table.Head
                      scope="col"
                      class="w-[16%] text-muted-foreground">Status</Table.Head
                    >
                    <Table.Head
                      scope="col"
                      class="w-[12%] text-muted-foreground"
                      >Pemakaian</Table.Head
                    >
                    <Table.Head
                      scope="col"
                      class="w-[14%] text-right text-muted-foreground"
                      >Aksi</Table.Head
                    >
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {#each currentQuestions as question (question.id)}
                    <Table.Row class="align-top hover:bg-muted/50">
                      <Table.Cell>
                        <div class="space-y-2">
                          <div class="flex flex-wrap items-center gap-2">
                            <span
                              class="font-mono text-xs font-semibold text-primary"
                              >{compactText(question.code, "Tanpa kode")}</span
                            >
                            <Badge
                              variant="outline"
                              class="border-border bg-card text-muted-foreground"
                            >
                              {questionTypeLabel(question.question_type)}
                            </Badge>
                            {#if question.hots_flag}
                              <Badge
                                variant="outline"
                                class="border-warning/30 bg-warning/10 text-warning"
                                >HOTS</Badge
                              >
                            {/if}
                            {#if questionUsageLocked(question)}
                              <Badge
                                variant="outline"
                                class="border-destructive/30 bg-destructive/10 text-destructive"
                                >Terkunci</Badge
                              >
                            {/if}
                          </div>
                          <RichContent
                            html={questionPratinjauHtml(question)}
                            class="line-clamp-2 break-words text-sm leading-6 text-foreground [overflow-wrap:anywhere] [&_.katex-display]:my-1 [&_.latex-display]:my-1"
                          />
                          <div
                            class="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground"
                          >
                            <span
                              >{displayName(
                                {
                                  display_name: question.author_display_name,
                                  username: question.author_username,
                                },
                                "Penulis belum tercatat",
                              )}</span
                            >
                            <span>{difficultyLabel(question.difficulty)}</span>
                            <span
                              >{compactText(
                                question.material_topic,
                                "Materi belum diisi",
                              )}</span
                            >
                            <span>Dibuat {formatDate(question.created_at)}</span
                            >
                          </div>
                        </div>
                      </Table.Cell>
                      <Table.Cell class="text-sm text-foreground">
                        <div
                          class="break-words font-medium text-foreground [overflow-wrap:anywhere]"
                        >
                          {subjectLabel(question)}
                        </div>
                        <div class="mt-1 text-xs text-muted-foreground">
                          {gradeLabel(question)}
                        </div>
                        {#if question.cognitive_level}
                          <div class="mt-1 text-xs text-muted-foreground">
                            {question.cognitive_level}
                          </div>
                        {/if}
                      </Table.Cell>
                      <Table.Cell>
                        <div class="flex flex-col items-start gap-1.5">
                          <Badge
                            variant="outline"
                            class={workflowBadgeClass(question.workflow_status)}
                          >
                            {workflowLabel(question.workflow_status)}
                          </Badge>
                          <Badge
                            variant="outline"
                            class={publicationBadgeClass(question.status)}
                          >
                            {publicationLabel(question.status)}
                          </Badge>
                        </div>
                      </Table.Cell>
                      <Table.Cell class="text-sm text-muted-foreground">
                        {questionUsageText(question)}
                      </Table.Cell>
                      <Table.Cell class="text-right">
                        <div class="flex flex-wrap justify-end gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            type="button"
                            aria-expanded={isQuestionExpanded(question.id)}
                            aria-controls={`question-detail-${question.id}`}
                            aria-label={`${isQuestionExpanded(question.id) ? "Tutup pratinjau" : "Buka pratinjau"} soal ${compactText(question.code, "tanpa kode")}`}
                            title="Pratinjau ringkas di halaman ini"
                            onclick={() => toggleQuestionDetail(question.id)}
                          >
                            {isQuestionExpanded(question.id) ? "Tutup" : "Pratinjau"}
                          </Button>
                          <Button
                            href={questionHref(question)}
                            target="_blank"
                            rel="noopener noreferrer"
                            variant="outline"
                            size="sm"
                            title="Buka halaman soal lengkap di tab baru"
                            aria-label={`${isQuickEditable(question) ? quickEditLabel(question) : "Buka"} soal ${compactText(question.code, "tanpa kode")} di tab baru`}
                          >
                            {#if isQuickEditable(question)}
                              <PencilIcon class="size-3.5" />
                              {quickEditLabel(question)}
                            {:else}
                              <EyeIcon class="size-3.5" />
                              Buka
                            {/if}
                          </Button>
                          {#if hasSecondaryActions(question)}
                            <details
                              class="group relative inline-block text-left"
                            >
                              <summary
                                class="inline-flex h-8 cursor-pointer list-none items-center justify-center rounded-md border border-border bg-background px-3 text-xs font-medium text-muted-foreground transition hover:bg-muted hover:text-foreground focus:outline-none focus-visible:ring-2 focus-visible:ring-ring [&::-webkit-details-marker]:hidden"
                              >
                                Lainnya
                              </summary>
                              <div
                                class="absolute right-0 z-20 mt-2 w-52 space-y-1 rounded-lg border border-border bg-popover p-2 text-left shadow-lg"
                              >
                                {#if canReturnToRevision(question)}
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    class="w-full justify-start text-warning hover:bg-warning/10"
                                    disabled={revisingQuestionId ===
                                      question.id}
                                    aria-label={`Kembalikan soal ${compactText(question.code, "tanpa kode")} ke revisi`}
                                    onclick={() =>
                                      void returnToRevision(question)}
                                  >
                                    <HistoryIcon class="size-3.5" />
                                    {revisingQuestionId === question.id
                                      ? "Memproses"
                                      : "Kembalikan"}
                                  </Button>
                                {/if}
                                {#if canCreateRevision(question)}
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    class="w-full justify-start text-primary hover:bg-primary/10"
                                    disabled={revisingQuestionId ===
                                      question.id}
                                    aria-label={`Buat revisi baru untuk soal ${compactText(question.code, "tanpa kode")}`}
                                    onclick={() =>
                                      void createRevision(question)}
                                  >
                                    <SparklesIcon class="size-3.5" />
                                    {revisingQuestionId === question.id
                                      ? "Membuat"
                                      : "Revisi Baru"}
                                  </Button>
                                {/if}
                                {#if canArchive(question)}
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    class="w-full justify-start text-muted-foreground hover:bg-muted/50"
                                    disabled={archivingQuestionId ===
                                      question.id}
                                    aria-label={`Arsipkan soal ${compactText(question.code, "tanpa kode")}`}
                                    onclick={() =>
                                      void archiveQuestion(question)}
                                  >
                                    <ArchiveIcon class="size-3.5" />
                                    {archivingQuestionId === question.id
                                      ? "Mengarsipkan"
                                      : "Arsip"}
                                  </Button>
                                {/if}
                                {#if canRestoreArchive(question)}
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    class="w-full justify-start text-primary hover:bg-primary/10"
                                    disabled={restoringQuestionId ===
                                      question.id}
                                    aria-label={`Pulihkan soal arsip ${compactText(question.code, "tanpa kode")}`}
                                    onclick={() => void restoreArchive(question)}
                                  >
                                    <RefreshCcwIcon class="size-3.5" />
                                    {restoringQuestionId === question.id
                                      ? "Memulihkan"
                                      : "Pulihkan Arsip"}
                                  </Button>
                                {/if}
                                {#if canDelete && isSafeDeletable(question)}
                                  <Button
                                    variant="ghost"
                                    size="sm"
                                    class="w-full justify-start text-destructive hover:bg-destructive/10"
                                    disabled={deletingQuestionId ===
                                      question.id}
                                    aria-label={`Hapus soal ${compactText(question.code, "tanpa kode")}`}
                                    onclick={() => deleteQuestion(question)}
                                  >
                                    <Trash2Icon class="size-3.5" />
                                    {deletingQuestionId === question.id
                                      ? "Menghapus"
                                      : "Hapus"}
                                  </Button>
                                {/if}
                              </div>
                            </details>
                          {/if}
                        </div>
                      </Table.Cell>
                    </Table.Row>
                    {#if isQuestionExpanded(question.id)}
                      <Table.Row
                        id={`question-detail-${question.id}`}
                        class="bg-muted/10"
                      >
                        <Table.Cell colspan={5} class="p-4">
                          {@render questionDetailPanel(question)}
                        </Table.Cell>
                      </Table.Row>
                    {/if}
                  {/each}
                </Table.Body>
              </Table.Root>
            </div>

            <div class="divide-y divide-border md:hidden">
              {#each currentQuestions as question (question.id)}
                <article class="space-y-3 p-4">
                  <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                      <p class="font-mono text-xs font-semibold text-primary">
                        {compactText(question.code, "Tanpa kode")}
                      </p>
                      <RichContent
                        html={questionPratinjauHtml(question)}
                        class="mt-1 line-clamp-3 break-words text-sm font-semibold leading-6 text-foreground [overflow-wrap:anywhere] [&_.katex-display]:my-1 [&_.latex-display]:my-1"
                      />
                    </div>
                    <Badge
                      variant="outline"
                      class={workflowBadgeClass(question.workflow_status)}
                    >
                      {workflowLabel(question.workflow_status)}
                    </Badge>
                  </div>
                  <div class="flex flex-wrap gap-1.5">
                    <Badge
                      variant="outline"
                      class="border-border bg-card text-muted-foreground"
                      >{questionTypeLabel(question.question_type)}</Badge
                    >
                    <Badge
                      variant="outline"
                      class={publicationBadgeClass(question.status)}
                      >{publicationLabel(question.status)}</Badge
                    >
                    {#if question.hots_flag}
                      <Badge
                        variant="outline"
                        class="border-warning/30 bg-warning/10 text-warning"
                        >HOTS</Badge
                      >
                    {/if}
                  </div>
                  <dl
                    class="grid grid-cols-2 gap-2 text-xs text-muted-foreground [overflow-wrap:anywhere]"
                  >
                    <div>
                      <dt class="font-semibold text-muted-foreground">Mapel</dt>
                      <dd class="mt-0.5 break-words [overflow-wrap:anywhere]">
                        {subjectLabel(question)}
                      </dd>
                    </div>
                    <div>
                      <dt class="font-semibold text-muted-foreground">Level</dt>
                      <dd class="mt-0.5 break-words [overflow-wrap:anywhere]">
                        {gradeLabel(question)}
                      </dd>
                    </div>
                    <div>
                      <dt class="font-semibold text-muted-foreground">
                        Pemakaian
                      </dt>
                      <dd class="mt-0.5 break-words [overflow-wrap:anywhere]">
                        {questionUsageText(question)}
                      </dd>
                    </div>
                    <div>
                      <dt class="font-semibold text-muted-foreground">
                        Dibuat
                      </dt>
                      <dd class="mt-0.5">{formatDate(question.created_at)}</dd>
                    </div>
                  </dl>
                  <div class="grid gap-2 sm:grid-cols-2">
                    <Button
                      href={questionHref(question)}
                      target="_blank"
                      rel="noopener noreferrer"
                      variant="outline"
                      class="min-h-10 w-full"
                      title="Buka halaman soal lengkap di tab baru"
                    >
                      {#if isQuickEditable(question)}
                        <PencilIcon class="size-4" />
                        {quickEditLabel(question)} di tab baru
                      {:else}
                        <EyeIcon class="size-4" />
                        Buka Soal di Tab Baru
                      {/if}
                    </Button>
                    <Button
                      variant="outline"
                      class="min-h-10 w-full"
                      type="button"
                      aria-expanded={isQuestionExpanded(question.id)}
                      aria-controls={`question-mobile-detail-${question.id}`}
                      aria-label={`${isQuestionExpanded(question.id) ? "Tutup pratinjau" : "Buka pratinjau"} soal ${compactText(question.code, "tanpa kode")}`}
                      title="Pratinjau ringkas di halaman ini"
                      onclick={() => toggleQuestionDetail(question.id)}
                    >
                      {isQuestionExpanded(question.id) ? "Tutup Pratinjau" : "Pratinjau Soal"}
                    </Button>
                    {#if canReturnToRevision(question)}
                      <Button
                        variant="outline"
                        class="min-h-10 w-full border-warning/30 text-warning hover:bg-warning/10"
                        disabled={revisingQuestionId === question.id}
                        aria-label={`Kembalikan soal ${compactText(question.code, "tanpa kode")} ke revisi`}
                        onclick={() => void returnToRevision(question)}
                      >
                        <HistoryIcon class="size-4" />
                        {revisingQuestionId === question.id
                          ? "Memproses"
                          : "Kembalikan ke Revisi"}
                      </Button>
                    {/if}
                    {#if canCreateRevision(question)}
                      <Button
                        variant="outline"
                        class="min-h-10 w-full border-primary/20 text-primary hover:bg-primary/10"
                        disabled={revisingQuestionId === question.id}
                        aria-label={`Buat revisi baru untuk soal ${compactText(question.code, "tanpa kode")}`}
                        onclick={() => void createRevision(question)}
                      >
                        <SparklesIcon class="size-4" />
                        {revisingQuestionId === question.id
                          ? "Membuat"
                          : "Buat Revisi Baru"}
                      </Button>
                    {/if}
                    {#if canArchive(question)}
                      <Button
                        variant="outline"
                        class="min-h-10 w-full border-muted-foreground/30 text-muted-foreground hover:bg-muted/50"
                        disabled={archivingQuestionId === question.id}
                        aria-label={`Arsipkan soal ${compactText(question.code, "tanpa kode")}`}
                        onclick={() => void archiveQuestion(question)}
                      >
                        <ArchiveIcon class="size-4" />
                        {archivingQuestionId === question.id
                          ? "Mengarsipkan"
                          : "Arsip"}
                      </Button>
                    {/if}
                    {#if canRestoreArchive(question)}
                      <Button
                        variant="outline"
                        class="min-h-10 w-full border-primary/20 text-primary hover:bg-primary/10"
                        disabled={restoringQuestionId === question.id}
                        aria-label={`Pulihkan soal arsip ${compactText(question.code, "tanpa kode")}`}
                        onclick={() => void restoreArchive(question)}
                      >
                        <RefreshCcwIcon class="size-4" />
                        {restoringQuestionId === question.id
                          ? "Memulihkan"
                          : "Pulihkan Arsip"}
                      </Button>
                    {/if}
                    {#if canDelete && isSafeDeletable(question)}
                      <Button
                        variant="outline"
                        class="min-h-10 w-full border-destructive/30 text-destructive hover:bg-destructive/10"
                        disabled={deletingQuestionId === question.id}
                        aria-label={`Hapus soal ${compactText(question.code, "tanpa kode")}`}
                        onclick={() => deleteQuestion(question)}
                      >
                        <Trash2Icon class="size-4" />
                        {deletingQuestionId === question.id
                          ? "Menghapus"
                          : "Hapus"}
                      </Button>
                    {/if}
                  </div>
                  {#if isQuestionExpanded(question.id)}
                    <div id={`question-mobile-detail-${question.id}`} class="pt-1" role="region">
                      {@render questionDetailPanel(question)}
                    </div>
                  {/if}
                </article>
              {/each}
            </div>
          {/if}
        </section>
      {/snippet}
    </AsyncContent>

    <TablePagination
      page={currentPage}
      limit={pageSize}
      total={totalItems}
      itemLabel="soal"
      loading={refreshing}
      ariaLabel="Navigasi halaman daftar soal"
      onchange={handlePagination}
    />
  {/if}
</div>
