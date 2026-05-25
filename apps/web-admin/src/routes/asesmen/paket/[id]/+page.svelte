<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import {
    clientApiPath,
    clientApiPathWithQuery,
    readClientApiData,
  } from "$lib/client/api";
  import * as Card from "$lib/components/ui/card";
  import * as Dialog from "$lib/components/ui/dialog";
  import * as Table from "$lib/components/ui/table";
  import { Badge } from "$lib/components/ui/badge";
  import { Button } from "$lib/components/ui/button";
  import { Input } from "$lib/components/ui/input";
  import { Textarea } from "$lib/components/ui/textarea";
  import LoadingButton from "$lib/components/LoadingButton.svelte";
  import MicroActionTable from "$lib/components/ops/MicroActionTable.svelte";
  import AsyncContent from "$lib/components/AsyncContent.svelte";
  import { TablePagination } from "$lib/components/ui/pagination";
  import { toast } from "$lib/components/ui/sonner";
  import {
    DEFAULT_PAGE_SIZE_OPTIONS,
    clampPage,
    paginateItems,
    type PaginationChange,
  } from "$lib/utils/pagination";

  type PackageRow = {
    id: string;
    event_id?: string | null;
    subject_id: string;
    subject_name?: string;
    subject_code?: string;
    title: string;
    description: string;
    duration_minutes: number;
    randomize_questions: boolean;
    randomize_options?: boolean;
    source_mode?: string;
    draw_pg_count?: number;
    draw_essay_count?: number;
    random_seed?: string;
    is_active: boolean;
    locked_at?: string | null;
    lock_reason?: string | null;
    snapshot_version?: number;
    session_count?: number;
  };
  type PackageQuestion = {
    question_id: string;
    position: number;
    points: number;
    question_code?: string;
    question_text?: string;
    question_type?: string;
    difficulty?: string;
    status?: string;
    workflow_status?: string;
    target_level?: string;
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    material_topic?: string;
    cognitive_level?: string;
    hots_flag?: boolean;
  };
  type Readiness = {
    status: string;
    target_pg_count: number;
    target_essay_count: number;
    missing_pg_count: number;
    missing_essay_count: number;
    question_count: number;
    pg_count: number;
    essay_count: number;
    total_points: number;
    published_count: number;
    unpublished_count: number;
    metadata_gap_count: number;
    session_count: number;
    locked: boolean;
    ready: boolean;
  };
  type DetailPayload = {
    package: PackageRow;
    questions: PackageQuestion[];
    readiness: Readiness;
  };
  type PoolQuestion = {
    id: string;
    subject_id: string;
    code?: string;
    question_text?: string;
    question_type?: string;
    status?: string;
    event_id?: string | null;
    target_level?: string;
    difficulty?: string;
    workflow_status?: string;
    cp_ref?: string;
    tp_ref?: string;
    kd_ref?: string;
    material_topic?: string;
    cognitive_level?: string;
    hots_flag?: boolean;
    author_username?: string;
    author_display_name?: string;
    created_at?: string;
    updated_at?: string;
    package_count?: number;
  };
  type QuestionListPayload = {
    items?: PoolQuestion[];
    meta?: { total?: number };
  };

  const packageQuestionColumns = [
    { key: "number", label: "No", headClass: "w-12", class: "w-12" },
    { key: "question", label: "Soal", class: "min-w-[280px]" },
    { key: "type", label: "Bentuk", class: "min-w-[100px]" },
    { key: "points", label: "Bobot", class: "w-24" },
    { key: "quality", label: "Mutu", class: "min-w-[120px]" },
  ];

  const packageId = page.params.id ?? "";
  let detailPromise = $state<Promise<DetailPayload> | null>(null);
  let detail = $state<DetailPayload | null>(null);
  let pool = $state<PoolQuestion[]>([]);
  let selectedPool = $state(new Set<string>());
  let busy = $state("");
  type PackageTab = "questions" | "pool" | "blueprint" | "lock";
  let activeTab = $state<PackageTab>("questions");

  function tabFromHash(hash: string): PackageTab | null {
    const value = hash.replace(/^#/, "").toLowerCase();
    if (value === "soal" || value === "questions" || value === "isi-soal")
      return "questions";
    if (value === "tambah-soal" || value === "bank-soal" || value === "pool")
      return "pool";
    if (value === "blueprint" || value === "kisi-kisi" || value === "mutu")
      return "blueprint";
    if (value === "lock" || value === "kunci" || value === "revisi")
      return "lock";
    return null;
  }

  function syncTabFromHash() {
    const tab = tabFromHash(window.location.hash);
    if (tab) activeTab = tab;
  }

  function selectTab(tab: PackageTab) {
    activeTab = tab;
    const hashByTab: Record<PackageTab, string> = {
      questions: "#soal",
      pool: "#tambah-soal",
      blueprint: "#blueprint",
      lock: "#kunci",
    };
    window.history.replaceState(null, "", hashByTab[tab]);
  }

  let title = $state("");
  let description = $state("");
  let duration = $state(60);
  let randomizeQuestions = $state(false);
  let randomizeOptions = $state(false);
  let active = $state(true);
  let targetPg = $state(20);
  let targetEssay = $state(5);
  let rows = $state<PackageQuestion[]>([]);
  let poolSearch = $state("");
  let poolLevel = $state("all");
  let poolType = $state("all");
  let poolStatus = $state("published");
  let poolCognitive = $state("all");
  let poolHots = $state("all");
  let poolDifficulty = $state("all");
  let poolMetadata = $state("all");
  let poolTopic = $state("all");
  let poolCurriculumSearch = $state("");
  let poolAuthor = $state("all");
  let poolCreatedFrom = $state("");
  let poolCreatedTo = $state("");
  let poolUsage = $state("all");
  let poolSort = $state("metadata_first");
  let poolPage = $state(1);
  let poolPageSize = $state<number>(DEFAULT_PAGE_SIZE_OPTIONS[0]);
  let cloneDialogOpen = $state(false);
  let cloneTitle = $state("");
  let lockDialogOpen = $state(false);
  let lockReason = $state("");

  let isLocked = $derived(
    Boolean(detail?.package.locked_at || detail?.readiness.locked),
  );
  let poolForSubject = $derived(
    pool.filter((q) => q.subject_id === detail?.package.subject_id),
  );
  let availablePool = $derived(
    sortPoolQuestions(
      poolForSubject
        .filter(matchesPoolFilters)
        .filter((q) => !rows.some((row) => row.question_id === q.id)),
    ),
  );
  let poolAuthorOptions = $derived(
    Array.from(
      new Map(
        poolForSubject
          .map(
            (q) =>
              [
                String(q.author_username ?? "").trim(),
                String(q.author_display_name ?? q.author_username ?? "").trim(),
              ] as const,
          )
          .filter(([username]) => Boolean(username)),
      ).entries(),
    ).sort((a, b) => a[1].localeCompare(b[1])),
  );
  let selectedQuestions = $derived(
    availablePool.filter((q) => selectedPool.has(q.id)),
  );
  let safePoolPage = $derived(
    clampPage(poolPage, availablePool.length, poolPageSize),
  );
  let paginatedPool = $derived(
    paginateItems(availablePool, safePoolPage, poolPageSize),
  );
  let missingLabel = $derived(
    detail
      ? `${Math.max(0, targetPg - countType(rows, "multiple_choice"))} PG + ${Math.max(0, targetEssay - countType(rows, "essay"))} Essay kurang`
      : "",
  );

  function typeLabel(value?: string) {
    const map: Record<string, string> = {
      multiple_choice: "PG",
      essay: "Essay",
      true_false: "Benar/Salah",
      short_answer: "Isian",
    };
    return map[value ?? ""] ?? (value || "Lainnya");
  }

  function countType(items: PackageQuestion[], type: string) {
    return items.filter((item) => item.question_type === type).length;
  }

  function hasMetadataGap(item: PackageQuestion | PoolQuestion) {
    return (
      !String(item.target_level ?? "").trim() ||
      !String(item.cp_ref ?? "").trim() ||
      !(String(item.tp_ref ?? "").trim() || String(item.kd_ref ?? "").trim()) ||
      !String(item.cognitive_level ?? "").trim()
    );
  }

  function inferLevelFromText(text?: string) {
    const upper = String(text ?? "").toUpperCase();
    if (/\bVII\b/.test(upper)) return "VII";
    if (/\bVIII\b/.test(upper)) return "VIII";
    if (/\bIX\b/.test(upper)) return "IX";
    return "all";
  }

  function difficultyLabel(value?: string) {
    return (
      { easy: "Mudah", medium: "Sedang", hard: "Sulit" }[value ?? ""] ??
      (value || "Belum")
    );
  }

  function statusLabel(value?: string) {
    return (
      {
        published: "Terbit",
        draft: "Konsep",
        review: "Review",
        archived: "Arsip",
        rejected: "Ditolak",
      }[value ?? ""] ??
      (value || "Belum")
    );
  }

  function userDisplayName(
    item: Pick<PoolQuestion, "author_display_name" | "author_username">,
  ) {
    return String(
      item.author_display_name || item.author_username || "",
    ).trim();
  }

  function parseDateOnly(value: string, endOfDay = false) {
    if (!value) return null;
    const parsed = new Date(
      `${value}T${endOfDay ? "23:59:59.999" : "00:00:00.000"}`,
    );
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  function dateLabel(value?: string) {
    if (!value) return "Tanggal kosong";
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return value;
    return new Intl.DateTimeFormat("id-ID", { dateStyle: "medium" }).format(
      parsed,
    );
  }

  function matchesPoolFilters(item: PoolQuestion) {
    const term = poolSearch.trim().toLowerCase();
    if (
      term &&
      !`${item.code ?? ""} ${item.question_text ?? ""} ${item.material_topic ?? ""}`
        .toLowerCase()
        .includes(term)
    )
      return false;
    if (poolLevel !== "all" && item.target_level !== poolLevel) return false;
    if (poolType !== "all" && item.question_type !== poolType) return false;
    if (poolStatus !== "all" && item.status !== poolStatus) return false;
    if (poolCognitive !== "all" && item.cognitive_level !== poolCognitive)
      return false;
    if (poolHots === "hots" && !item.hots_flag) return false;
    if (poolHots === "non_hots" && item.hots_flag) return false;
    if (poolDifficulty !== "all" && item.difficulty !== poolDifficulty)
      return false;
    if (poolMetadata === "complete" && hasMetadataGap(item)) return false;
    if (poolMetadata === "gap" && !hasMetadataGap(item)) return false;
    if (poolTopic !== "all" && String(item.material_topic ?? "") !== poolTopic)
      return false;
    if (
      poolAuthor !== "all" &&
      String(item.author_username ?? "") !== poolAuthor
    )
      return false;
    if (poolUsage === "unused" && Number(item.package_count ?? 0) > 0)
      return false;
    if (poolUsage === "used" && Number(item.package_count ?? 0) === 0)
      return false;
    const createdAt = item.created_at ? new Date(item.created_at) : null;
    const from = parseDateOnly(poolCreatedFrom);
    const to = parseDateOnly(poolCreatedTo, true);
    if (from && (!createdAt || createdAt < from)) return false;
    if (to && (!createdAt || createdAt > to)) return false;
    const cur = poolCurriculumSearch.trim().toLowerCase();
    if (
      cur &&
      !`${item.cp_ref ?? ""} ${item.tp_ref ?? ""} ${item.kd_ref ?? ""}`
        .toLowerCase()
        .includes(cur)
    )
      return false;
    return true;
  }

  function sortPoolQuestions(items: PoolQuestion[]) {
    return [...items].sort((a, b) => {
      if (poolSort === "newest")
        return (
          new Date(b.created_at ?? 0).getTime() -
          new Date(a.created_at ?? 0).getTime()
        );
      if (poolSort === "oldest")
        return (
          new Date(a.created_at ?? 0).getTime() -
          new Date(b.created_at ?? 0).getTime()
        );
      if (poolSort === "author")
        return (
          userDisplayName(a).localeCompare(userDisplayName(b)) ||
          String(a.code ?? "").localeCompare(String(b.code ?? ""))
        );
      if (poolSort === "unused_first")
        return (
          Number(a.package_count ?? 0) - Number(b.package_count ?? 0) ||
          String(a.code ?? "").localeCompare(String(b.code ?? ""))
        );
      if (poolSort === "hots_first")
        return Number(Boolean(b.hots_flag)) - Number(Boolean(a.hots_flag));
      if (poolSort === "difficulty")
        return (
          String(a.difficulty ?? "").localeCompare(
            String(b.difficulty ?? ""),
          ) || String(a.code ?? "").localeCompare(String(b.code ?? ""))
        );
      if (poolSort === "type")
        return (
          String(a.question_type ?? "").localeCompare(
            String(b.question_type ?? ""),
          ) || String(a.code ?? "").localeCompare(String(b.code ?? ""))
        );
      if (poolSort === "code")
        return String(a.code ?? "").localeCompare(String(b.code ?? ""));
      return (
        Number(hasMetadataGap(a)) - Number(hasMetadataGap(b)) ||
        String(a.code ?? "").localeCompare(String(b.code ?? ""))
      );
    });
  }

  function syncForm(payload: DetailPayload) {
    detail = payload;
    title = payload.package.title;
    description = payload.package.description ?? "";
    duration = payload.package.duration_minutes || 60;
    randomizeQuestions = Boolean(payload.package.randomize_questions);
    randomizeOptions = Boolean(payload.package.randomize_options);
    active = Boolean(payload.package.is_active);
    targetPg = payload.package.draw_pg_count || 20;
    targetEssay = payload.package.draw_essay_count || 5;
    rows = [...payload.questions].sort((a, b) => a.position - b.position);
    if (poolLevel === "all")
      poolLevel = inferLevelFromText(payload.package.title);
    selectedPool = new Set();
  }

  async function loadDetail() {
    const payload = await fetch(
      clientApiPath`/api/asesmen/packages/${packageId}`,
    ).then((res) => readClientApiData<DetailPayload>(res));
    syncForm(payload);
    return payload;
  }

  async function fetchQuestionPool(payload: DetailPayload) {
    const params = new URLSearchParams({
      limit: "1000",
      offset: "0",
      scope: payload.package.event_id ? "event_pool" : "global",
    });
    params.set("subject_id", payload.package.subject_id);
    if (payload.package.event_id)
      params.set("event_id", payload.package.event_id);
    const q = await fetch(
      clientApiPathWithQuery("/api/bank-soal/questions", params),
    ).then((res) =>
      readClientApiData<QuestionListPayload | PoolQuestion[]>(res),
    );
    pool = Array.isArray(q) ? q : (q.items ?? []);
  }

  async function refresh() {
    detailPromise = loadDetail();
    const payload = await detailPromise;
    await fetchQuestionPool(payload);
    return payload;
  }

  onMount(() => {
    syncTabFromHash();
    window.addEventListener("hashchange", syncTabFromHash);
    void refresh();
    return () => window.removeEventListener("hashchange", syncTabFromHash);
  });

  async function saveMetadata() {
    if (!detail || isLocked) return;
    busy = "metadata";
    try {
      const payload = await fetch(
        clientApiPath`/api/asesmen/packages/${packageId}`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            title,
            description,
            duration_minutes: duration,
            randomize_questions: randomizeQuestions,
            randomize_options: randomizeOptions,
            source_mode: detail.package.source_mode || "teacher_class",
            draw_pg_count: targetPg,
            draw_essay_count: targetEssay,
            random_seed: detail.package.random_seed || "",
            is_active: active,
          }),
        },
      ).then((res) => readClientApiData<DetailPayload>(res));
      syncForm(payload);
      toast.success("Metadata paket tersimpan");
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      busy = "";
    }
  }

  async function saveQuestions() {
    if (isLocked) return;
    busy = "questions";
    try {
      const payload = await fetch(
        clientApiPath`/api/asesmen/packages/${packageId}/questions`,
        {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            questions: rows.map((row) => ({
              question_id: row.question_id,
              points: row.points || 1,
            })),
          }),
        },
      ).then((res) => readClientApiData<DetailPayload>(res));
      syncForm(payload);
      await fetchQuestionPool(payload);
      toast.success("Isi soal paket tersimpan");
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      busy = "";
    }
  }

  function addSelected() {
    const next = [...rows];
    for (const question of selectedQuestions) {
      next.push({
        question_id: question.id,
        position: next.length + 1,
        points: 1,
        question_code: question.code,
        question_text: question.question_text,
        question_type: question.question_type,
        status: question.status,
        cp_ref: question.cp_ref,
        tp_ref: question.tp_ref,
        kd_ref: question.kd_ref,
        material_topic: question.material_topic,
        cognitive_level: question.cognitive_level,
        hots_flag: question.hots_flag,
        target_level: question.target_level,
        difficulty: question.difficulty,
      });
    }
    rows = next;
    selectedPool = new Set();
    activeTab = "questions";
  }

  function autofillTarget() {
    const next = [...rows];
    const needPg = Math.max(0, targetPg - countType(next, "multiple_choice"));
    const needEssay = Math.max(0, targetEssay - countType(next, "essay"));
    const picked = [
      ...availablePool
        .filter((q) => q.question_type === "multiple_choice")
        .slice(0, needPg),
      ...availablePool
        .filter((q) => q.question_type === "essay")
        .slice(0, needEssay),
    ];
    for (const q of picked)
      next.push({
        question_id: q.id,
        position: next.length + 1,
        points: 1,
        question_code: q.code,
        question_text: q.question_text,
        question_type: q.question_type,
        status: q.status,
        cp_ref: q.cp_ref,
        tp_ref: q.tp_ref,
        kd_ref: q.kd_ref,
        material_topic: q.material_topic,
        cognitive_level: q.cognitive_level,
        hots_flag: q.hots_flag,
        target_level: q.target_level,
        difficulty: q.difficulty,
      });
    rows = next;
    toast.info(`${picked.length} soal ditambahkan ke draft paket`);
  }

  function removeRow(index: number) {
    rows = rows
      .filter((_, i) => i !== index)
      .map((row, i) => ({ ...row, position: i + 1 }));
  }
  function moveRow(index: number, direction: -1 | 1) {
    const target = index + direction;
    if (target < 0 || target >= rows.length) return;
    const copy = [...rows];
    [copy[index], copy[target]] = [copy[target], copy[index]];
    rows = copy.map((row, i) => ({ ...row, position: i + 1 }));
  }
  function setPoints(index: number, value: number) {
    rows = rows.map((row, i) =>
      i === index
        ? { ...row, points: Math.max(1, Math.min(100, Math.round(value || 1))) }
        : row,
    );
  }
  let allSelected = $derived(
    paginatedPool.length > 0 &&
      paginatedPool.every((q) => selectedPool.has(q.id)),
  );
  let someSelected = $derived(
    paginatedPool.some((q) => selectedPool.has(q.id)) && !allSelected,
  );

  function togglePool(id: string) {
    const next = new Set(selectedPool);
    next.has(id) ? next.delete(id) : next.add(id);
    selectedPool = next;
  }
  function toggleSelectAll() {
    if (allSelected) {
      selectedPool = new Set([...selectedPool].filter((id) => !paginatedPool.some((q) => q.id === id)));
      return;
    }
    const next = new Set(selectedPool);
    for (const q of paginatedPool) {
      next.add(q.id);
    }
    selectedPool = next;
  }

  function resetPoolPage() {
    poolPage = 1;
  }

  function handlePoolPagination(change: PaginationChange) {
    poolPage = change.reason === "limit" ? 1 : change.page;
    poolPageSize = change.limit;
  }

  function openCloneDialog() {
    if (!detail) return;
    cloneTitle = `${detail.package.title} - Revisi`;
    cloneDialogOpen = true;
  }

  async function submitClonePackage() {
    if (!detail) return;
    const newTitle = cloneTitle.trim();
    if (!newTitle) return;
    busy = "clone";
    try {
      const payload = await fetch(
        clientApiPath`/api/asesmen/packages/${packageId}/clone`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ title: newTitle }),
        },
      ).then((res) => readClientApiData<DetailPayload>(res));
      toast.success("Paket revisi dibuat");
      cloneDialogOpen = false;
      await goto(`/asesmen/paket/${payload.package.id}`);
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      busy = "";
    }
  }

  function openLockDialog() {
    if (!detail || isLocked) return;
    lockReason = detail.package.lock_reason?.trim() || "Dikunci dari Penyusunan Paket";
    lockDialogOpen = true;
  }

  async function submitLockPackage() {
    if (!detail || isLocked) return;
    const reason = lockReason.trim();
    if (!reason) return;
    busy = "lock";
    try {
      await fetch(clientApiPath`/api/asesmen/packages/${packageId}/lock`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reason }),
      }).then((res) => readClientApiData(res));
      toast.success("Paket terkunci dan snapshot dibuat");
      lockDialogOpen = false;
      await refresh();
    } catch (e) {
      toast.error((e as Error).message);
    } finally {
      busy = "";
    }
  }
</script>

<svelte:head><title>Penyusunan Paket — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-5 p-4 md:p-6">
  <AsyncContent promise={detailPromise}>
    {#snippet children()}
      {#if detail}
        <div
          class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
        >
          <div>
            <p
              class="text-xs font-semibold uppercase tracking-[0.18em] text-primary"
            >
              Penyusunan Paket
            </p>
            <h1 class="text-2xl font-semibold tracking-tight">
              {detail.package.title}
            </h1>
            <p
              class="text-sm text-muted-foreground truncate max-w-[90vw] sm:max-w-md"
              title={`${detail.package.subject_name ?? detail.package.subject_code} · ${detail.readiness.session_count} sesi memakai paket ini`}
            >
              {detail.package.subject_name ?? detail.package.subject_code} · {detail
                .readiness.session_count} sesi memakai paket ini
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Badge
              class={detail.readiness.ready
                ? "bg-primary/15 text-primary border-primary/20"
                : "bg-warning/10 text-warning border-warning/30"}
              >{detail.readiness.status}</Badge
            >
            {#if isLocked}<Badge
                class="bg-destructive/10 text-destructive border-destructive/30"
                >Terkunci v{detail.package.snapshot_version ?? 1}</Badge
              >{/if}
            <a
              class="rounded-md border px-3 py-2 text-sm hover:bg-muted"
              href="/asesmen/paket">Kembali</a
            >
          </div>
        </div>

        <div class="grid gap-4 md:grid-cols-4">
          <Card.Root
            ><Card.Content class="p-4"
              ><p class="text-xs text-muted-foreground">PG</p>
              <p class="text-2xl font-semibold">
                {countType(rows, "multiple_choice")}/{targetPg}
              </p></Card.Content
            ></Card.Root
          >
          <Card.Root
            ><Card.Content class="p-4"
              ><p class="text-xs text-muted-foreground">Essay</p>
              <p class="text-2xl font-semibold">
                {countType(rows, "essay")}/{targetEssay}
              </p></Card.Content
            ></Card.Root
          >
          <Card.Root
            ><Card.Content class="p-4"
              ><p class="text-xs text-muted-foreground">Total Poin</p>
              <p class="text-2xl font-semibold">
                {rows.reduce((s, r) => s + (r.points || 0), 0)}
              </p></Card.Content
            ></Card.Root
          >
          <Card.Root
            ><Card.Content class="p-4"
              ><p class="text-xs text-muted-foreground">Gap</p>
              <p class="text-sm font-semibold">{missingLabel}</p></Card.Content
            ></Card.Root
          >
        </div>

        <Card.Root>
          <Card.Header
            ><Card.Title class="text-base">Identitas Paket</Card.Title
            ><Card.Description
              >{isLocked
                ? "Paket terkunci. Buat revisi jika perlu mengubah."
                : "Edit identitas, durasi, pengacakan, dan target kisi-kisi."}</Card.Description
            ></Card.Header
          >
          <Card.Content class="grid gap-3 md:grid-cols-3">
            <label class="space-y-1 md:col-span-2"
              ><span class="text-xs text-muted-foreground">Nama Paket</span
              ><Input bind:value={title} disabled={isLocked} /></label
            >
            <label class="space-y-1"
              ><span class="text-xs text-muted-foreground">Durasi menit</span
              ><Input
                type="number"
                bind:value={duration}
                min="1"
                max="360"
                disabled={isLocked}
              /></label
            >
            <label class="space-y-1 md:col-span-3"
              ><span class="text-xs text-muted-foreground">Deskripsi</span
              ><Textarea bind:value={description} disabled={isLocked} /></label
            >
            <label class="space-y-1"
              ><span class="text-xs text-muted-foreground">Target PG</span
              ><Input
                type="number"
                bind:value={targetPg}
                min="0"
                disabled={isLocked}
              /></label
            >
            <label class="space-y-1"
              ><span class="text-xs text-muted-foreground">Target Essay</span
              ><Input
                type="number"
                bind:value={targetEssay}
                min="0"
                disabled={isLocked}
              /></label
            >
            <div class="flex flex-wrap items-center gap-4 text-sm">
              <label
                ><input
                  type="checkbox"
                  bind:checked={randomizeQuestions}
                  disabled={isLocked}
                /> Acak soal</label
              >
              <label
                ><input
                  type="checkbox"
                  bind:checked={randomizeOptions}
                  disabled={isLocked}
                /> Acak opsi</label
              >
              <label
                ><input
                  type="checkbox"
                  bind:checked={active}
                  disabled={isLocked}
                /> Aktif</label
              >
            </div>
            <div class="md:col-span-3 flex flex-wrap gap-2">
              <LoadingButton
                onclick={saveMetadata}
                loading={busy === "metadata"}
                disabled={isLocked}>Simpan Identitas</LoadingButton
              >
              <LoadingButton
                variant="outline"
                onclick={openCloneDialog}
                loading={busy === "clone"}>Buat Revisi/Salinan</LoadingButton
              >
            </div>
          </Card.Content>
        </Card.Root>

        <div class="flex flex-wrap gap-2">
          {#each [["questions", "Soal Dalam Paket"], ["pool", "Tambah dari Bank Soal"], ["blueprint", "Kisi-kisi & Mutu"], ["lock", "Kunci & Salinan"]] as tab}
            <button
              class={`rounded-md border px-3 py-2 text-sm ${activeTab === tab[0] ? "bg-primary text-primary-foreground" : "hover:bg-muted"}`}
              onclick={() => selectTab(tab[0] as PackageTab)}>{tab[1]}</button
            >
          {/each}
        </div>

        {#snippet packageQuestionCell(rowValue: unknown, column: { key: string }, i: number)}
                {@const row = rowValue as PackageQuestion}
                {#if column.key === "number"}
                  <span class="font-semibold text-muted-foreground">#{i + 1}</span>
                {:else if column.key === "question"}
                  <div class="min-w-0">
                    <p class="font-mono text-xs font-semibold text-primary">
                      {row.question_code || row.question_id}
                    </p>
                    <p class="mt-1 line-clamp-2 break-words text-xs text-muted-foreground [overflow-wrap:anywhere]">
                      {row.question_text}
                    </p>
                  </div>
                {:else if column.key === "type"}
                  <Badge variant="outline" class="text-xs">{typeLabel(row.question_type)}</Badge>
                {:else if column.key === "points"}
                  <Input
                    class="h-8 w-20 text-right text-xs"
                    type="number"
                    value={row.points}
                    min="1"
                    max="100"
                    disabled={isLocked}
                    oninput={(e) => setPoints(i, Number((e.currentTarget as HTMLInputElement).value))}
                  />
                {:else if column.key === "quality"}
                  {#if hasMetadataGap(row)}
                    <Badge class="bg-warning/10 text-warning border-warning/30 text-xs">Metadata kurang</Badge>
                  {:else}
                    <Badge variant="outline" class="text-xs">OK</Badge>
                  {/if}
                {/if}
        {/snippet}

        {#snippet packageQuestionActions(rowValue: unknown, i: number)}
                <button
                  class="rounded border px-2 py-1 text-xs"
                  disabled={isLocked || i === 0}
                  onclick={() => moveRow(i, -1)}>↑</button
                ><button
                  class="rounded border px-2 py-1 text-xs"
                  disabled={isLocked || i === rows.length - 1}
                  onclick={() => moveRow(i, 1)}>↓</button
                ><button
                  class="rounded border px-2 py-1 text-xs text-destructive"
                  disabled={isLocked}
                  onclick={() => removeRow(i)}>Hapus</button
                >
        {/snippet}

        {#snippet packageQuestionMobile(rowValue: unknown, i: number)}
                {@const row = rowValue as PackageQuestion}
                <div class="space-y-2 text-xs">
                  <div class="flex items-start justify-between gap-2">
                    <div class="min-w-0">
                      {@render packageQuestionCell(row, packageQuestionColumns[0], i)}
                      {@render packageQuestionCell(row, packageQuestionColumns[1], i)}
                    </div>
                    {@render packageQuestionCell(row, packageQuestionColumns[4], i)}
                  </div>
                  <div class="flex flex-wrap items-center gap-2">
                    {@render packageQuestionCell(row, packageQuestionColumns[2], i)}
                    <span class="text-muted-foreground">Bobot</span>
                    {@render packageQuestionCell(row, packageQuestionColumns[3], i)}
                  </div>
                  <div class="flex flex-wrap gap-1.5 pt-1">{@render packageQuestionActions(row, i)}</div>
                </div>
        {/snippet}


        {#if activeTab === "questions"}
          <Card.Root
            ><Card.Header
              ><Card.Title class="text-base"
                >Soal Dalam Paket ({rows.length})</Card.Title
              ></Card.Header
            ><Card.Content class="space-y-3">
              <div class="flex flex-wrap gap-2">
                <LoadingButton
                  onclick={saveQuestions}
                  loading={busy === "questions"}
                  disabled={isLocked}>Simpan Isi Soal</LoadingButton
                ><button
                  class="rounded-md border px-3 py-2 text-sm"
                  disabled={isLocked}
                  onclick={autofillTarget}
                  >Autofill Target dari Soal Terbit</button
                >
              </div>
              <MicroActionTable
                rows={rows}
                columns={packageQuestionColumns}
                rowKey={(row) => (row as PackageQuestion).question_id}
                cell={packageQuestionCell}
                actions={packageQuestionActions}
                mobile={packageQuestionMobile}
                tableClass="min-w-[760px]"
                emptyTitle="Belum ada soal dalam paket."
                emptyDescription="Tambahkan soal dari Bank Soal atau gunakan autofill target."
              />
            </Card.Content></Card.Root
          >
        {:else if activeTab === "pool"}
          <Card.Root>
            <Card.Header>
              <Card.Title class="text-base">Tambah dari Bank Soal</Card.Title>
              <Card.Description
                >{availablePool.length} soal tersedia sesuai filter untuk mapel ini.
                Default filter status = Terbit agar aman untuk paket resmi.</Card.Description
              >
            </Card.Header>
            <Card.Content class="space-y-4">
              <div class="rounded-2xl border bg-muted/30 p-4 space-y-4">
                <!-- Baris 1: Pencarian Teks -->
                <div class="grid gap-3 md:grid-cols-2">
                  <Input
                    placeholder="Cari kode, teks soal, materi..."
                    bind:value={poolSearch}
                    oninput={resetPoolPage}
                  />
                  <Input
                    placeholder="Cari CP/TP/KD..."
                    bind:value={poolCurriculumSearch}
                    oninput={resetPoolPage}
                  />
                </div>
                <div class="border-t border-border/40"></div>
                <!-- Baris 2: Kategori Soal -->
                <div
                  class="grid gap-3 max-sm:grid-cols-2 sm:grid-cols-3 lg:grid-cols-5"
                >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Tingkat</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolLevel}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua tingkat</option><option
                        value="VII">VII</option
                      ><option value="VIII">VIII</option><option value="IX"
                        >IX</option
                      >
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Jenis Soal</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolType}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua jenis</option><option
                        value="multiple_choice">Pilihan Ganda</option
                      ><option value="essay">Essay</option><option
                        value="true_false">Benar/Salah</option
                      ><option value="short_answer">Isian</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Status</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolStatus}
                      onchange={resetPoolPage}
                    >
                      <option value="published">Terbit saja</option><option
                        value="all">Semua status</option
                      ><option value="draft">Konsep</option><option
                        value="review">Verifikasi</option
                      ><option value="archived">Arsip</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Kesulitan</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolDifficulty}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua kesulitan</option><option
                        value="easy">Mudah</option
                      ><option value="medium">Sedang</option><option
                        value="hard">Sulit</option
                      >
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Pembuat</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolAuthor}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua pembuat</option>
                      {#each poolAuthorOptions as [username, displayName]}<option
                          value={username}>{displayName}</option
                        >{/each}
                    </select></label
                  >
                </div>
                <div class="border-t border-border/40"></div>
                <!-- Baris 3: Kualitas & Filter Lanjutan -->
                <div
                  class="grid gap-3 max-sm:grid-cols-2 sm:grid-cols-3 lg:grid-cols-5"
                >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Level Kognitif</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolCognitive}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua level</option><option value="C1"
                        >C1</option
                      ><option value="C2">C2</option><option value="C3"
                        >C3</option
                      ><option value="C4">C4</option><option value="C5"
                        >C5</option
                      ><option value="C6">C6</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >HOTS</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolHots}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua HOTS</option><option
                        value="hots">HOTS</option
                      ><option value="non_hots">Non-HOTS</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Identitas</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolMetadata}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua identitas</option><option
                        value="complete">Identitas lengkap</option
                      ><option value="gap">Identitas kurang</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Pemakaian</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolUsage}
                      onchange={resetPoolPage}
                    >
                      <option value="all">Semua pemakaian</option><option
                        value="unused">Belum dipakai paket</option
                      ><option value="used">Sudah dipakai paket</option>
                    </select></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Urutkan</span
                    >
                    <select
                      class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                      bind:value={poolSort}
                      onchange={resetPoolPage}
                    >
                      <option value="metadata_first"
                        >Identitas lengkap dulu</option
                      ><option value="newest">Terbaru</option><option
                        value="oldest">Terlama</option
                      ><option value="author">Pembuat A-Z</option><option
                        value="unused_first">Belum dipakai</option
                      ><option value="hots_first">HOTS dulu</option><option
                        value="difficulty">Kesulitan</option
                      ><option value="type">Jenis soal</option><option
                        value="code">Kode A-Z</option
                      >
                    </select></label
                  >
                </div>
                <div class="border-t border-border/40"></div>
                <!-- Baris 4: Rentang Tanggal -->
                <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Dari tanggal dibuat</span
                    ><Input type="date" bind:value={poolCreatedFrom} oninput={resetPoolPage} /></label
                  >
                  <label class="space-y-1.5"
                    ><span class="text-xs font-medium text-muted-foreground"
                      >Sampai tanggal dibuat</span
                    ><Input type="date" bind:value={poolCreatedTo} oninput={resetPoolPage} /></label
                  >
                </div>
                {#if poolStatus !== "published"}<p
                    class="mt-2 text-xs text-warning"
                  >
                    Mode pemeriksaan: soal belum Terbit bisa dilihat, tetapi
                    layanan sistem tetap menolak jika dimasukkan ke paket resmi.
                  </p>{/if}
              </div>
              <button
                class="rounded-md bg-primary px-3 py-2 text-sm text-primary-foreground disabled:opacity-50"
                disabled={isLocked || selectedQuestions.length === 0}
                onclick={addSelected}
                >Tambah {selectedQuestions.length} Soal</button
              >
              <div class="overflow-x-auto rounded-lg border">
                <Table.Root class="min-w-[700px] table-auto">
                  <Table.Header>
                    <Table.Row>
                      <Table.Head class="w-10"
                        ><input
                          type="checkbox"
                          checked={allSelected}
                          indeterminate={someSelected}
                          disabled={isLocked}
                          onchange={toggleSelectAll}
                        /></Table.Head
                      >
                      <Table.Head class="min-w-[160px]">Kode Soal</Table.Head>
                      <Table.Head class="w-20">Bentuk</Table.Head>
                      <Table.Head class="w-20">Tingkat</Table.Head>
                      <Table.Head class="w-28">Pembuat</Table.Head>
                      <Table.Head class="w-20">Status</Table.Head>
                      <Table.Head class="w-16">Mutu</Table.Head>
                      <Table.Head class="w-24">Pakai</Table.Head>
                    </Table.Row>
                  </Table.Header>
                  <Table.Body>
                    {#each paginatedPool as q (q.id)}
                      <Table.Row
                        class={selectedPool.has(q.id) ? "bg-muted/50" : ""}
                      >
                        <Table.Cell class="p-2"
                          ><input
                            type="checkbox"
                            checked={selectedPool.has(q.id)}
                            disabled={isLocked || q.status !== "published"}
                            onchange={() => togglePool(q.id)}
                          /></Table.Cell
                        >
                        <Table.Cell class="p-2">
                          <p class="font-medium text-sm truncate max-w-[200px]">
                            {q.code || "\u2014"}
                          </p>
                          <p
                            class="truncate text-xs text-muted-foreground max-w-[200px]"
                          >
                            {q.question_text}
                          </p>
                        </Table.Cell>
                        <Table.Cell class="p-2 text-xs whitespace-nowrap"
                          >{typeLabel(q.question_type)}</Table.Cell
                        >
                        <Table.Cell class="p-2 text-xs whitespace-nowrap"
                          >{#if q.target_level}<Badge
                              variant="outline"
                              class="text-xs">{q.target_level}</Badge
                            >{:else}<Badge
                              class="bg-warning/10 text-warning border-warning/30 text-xs whitespace-nowrap"
                              >Kosong</Badge
                            >{/if}</Table.Cell
                        >
                        <Table.Cell class="p-2 text-xs truncate max-w-[120px]"
                          >{userDisplayName(q)}</Table.Cell
                        >
                        <Table.Cell class="p-2 text-xs whitespace-nowrap"
                          ><Badge
                            variant="outline"
                            class="text-xs whitespace-nowrap"
                            >{statusLabel(q.status)}</Badge
                          ></Table.Cell
                        >
                        <Table.Cell class="p-2 text-xs whitespace-nowrap"
                          >{#if hasMetadataGap(q)}<Badge
                              class="bg-warning/10 text-warning border-warning/30 text-xs whitespace-nowrap"
                              >Gap</Badge
                            >{:else}<Badge
                              variant="outline"
                              class="text-xs whitespace-nowrap">OK</Badge
                            >{/if}</Table.Cell
                        >
                        <Table.Cell class="p-2 text-xs whitespace-nowrap"
                          >{Number(q.package_count ?? 0) > 0
                            ? `Dipakai ${q.package_count}`
                            : "Baru"}</Table.Cell
                        >
                      </Table.Row>
                    {:else}
                      <Table.Row>
                        <Table.Cell
                          colspan={8}
                          class="py-8 text-center text-muted-foreground"
                          >Belum ada soal sesuai filter yang bisa ditambahkan.</Table.Cell
                        >
                      </Table.Row>
                    {/each}
                  </Table.Body>
                </Table.Root>
              </div>
              <TablePagination
                page={safePoolPage}
                limit={poolPageSize}
                total={availablePool.length}
                itemLabel="soal"
                ariaLabel="Navigasi halaman pool soal paket"
                onchange={handlePoolPagination}
              />
              <p class="mt-2 text-xs text-muted-foreground">
                {selectedPool.size} dari {availablePool.length} soal dipilih.
              </p>
            </Card.Content>
          </Card.Root>
        {:else if activeTab === "blueprint"}
          <Card.Root
            ><Card.Header
              ><Card.Title class="text-base">Kisi-kisi & Mutu</Card.Title
              ></Card.Header
            ><Card.Content class="grid gap-3 md:grid-cols-3"
              ><div class="rounded-xl border p-4">
                <p class="text-xs text-muted-foreground">Status</p>
                <p class="text-lg font-semibold">{detail.readiness.status}</p>
              </div>
              <div class="rounded-xl border p-4">
                <p class="text-xs text-muted-foreground">Identitas kurang</p>
                <p class="text-lg font-semibold">
                  {rows.filter(hasMetadataGap).length}
                </p>
              </div>
              <div class="rounded-xl border p-4">
                <p class="text-xs text-muted-foreground">HOTS</p>
                <p class="text-lg font-semibold">
                  {rows.filter((r) => r.hots_flag).length}
                </p>
              </div>
              <div class="md:col-span-3 text-sm text-muted-foreground">
                Distribusi: PG {countType(rows, "multiple_choice")}, Essay {countType(
                  rows,
                  "essay",
                )}. Lengkapi CP/TP/KD dan level kognitif di Bank Soal untuk
                menutup kekurangan identitas soal.
              </div></Card.Content
            ></Card.Root
          >
        {:else}
          <Card.Root
            ><Card.Header
              ><Card.Title class="text-base"
                >Kunci / Salinan Kondisi / Revisi</Card.Title
              ><Card.Description
                >Paket terkunci tidak bisa diedit langsung; gunakan salin/revisi
                agar riwayat perubahan tetap aman.</Card.Description
              ></Card.Header
            ><Card.Content class="space-y-3"
              ><p class="text-sm">
                Status: {isLocked
                  ? `Terkunci (${detail.package.locked_at})`
                  : "Belum terkunci"}
              </p>
              {#if detail.package.lock_reason}<p
                  class="text-sm text-muted-foreground"
                >
                  Alasan: {detail.package.lock_reason}
                </p>{/if}
              <div class="flex gap-2">
                <LoadingButton
                  onclick={openLockDialog}
                  loading={busy === "lock"}
                  disabled={isLocked}
                  >Kunci + Simpan Salinan Kondisi</LoadingButton
                ><LoadingButton
                  variant="outline"
                  onclick={openCloneDialog}
                  loading={busy === "clone"}>Buat Revisi/Salinan</LoadingButton
                >
              </div></Card.Content
            ></Card.Root
          >
        {/if}
      {/if}
    {/snippet}
  </AsyncContent>

  <Dialog.Root bind:open={cloneDialogOpen}>
    <Dialog.Content>
      <form
        class="space-y-4"
        onsubmit={(event) => {
          event.preventDefault();
          void submitClonePackage();
        }}
      >
        <Dialog.Header>
          <Dialog.Title>Buat Revisi atau Salinan Paket</Dialog.Title>
          <Dialog.Description>
            Salinan baru akan dibuat dari isi paket saat ini. Paket asal tetap
            tidak berubah, dan Anda akan diarahkan ke paket hasil revisi.
          </Dialog.Description>
        </Dialog.Header>

        <label class="space-y-2 block">
          <span class="text-sm font-medium">Nama Paket Baru</span>
          <Input
            bind:value={cloneTitle}
            placeholder="Masukkan nama paket revisi atau salinan"
            disabled={busy === "clone"}
            autofocus
          />
        </label>

        <Dialog.Footer>
          <Button
            type="button"
            variant="outline"
            disabled={busy === "clone"}
            onclick={() => (cloneDialogOpen = false)}>Batal</Button
          >
          <LoadingButton
            type="submit"
            loading={busy === "clone"}
            disabled={!cloneTitle.trim()}>Buat Paket</LoadingButton
          >
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>

  <Dialog.Root bind:open={lockDialogOpen}>
    <Dialog.Content>
      <form
        class="space-y-4"
        onsubmit={(event) => {
          event.preventDefault();
          void submitLockPackage();
        }}
      >
        <Dialog.Header>
          <Dialog.Title>Kunci Paket dan Simpan Salinan Kondisi</Dialog.Title>
          <Dialog.Description>
            Setelah dikunci, paket tidak dapat diedit langsung. Sistem akan
            menyimpan snapshot untuk menjaga konsistensi sesi asesmen yang
            memakai paket ini.
          </Dialog.Description>
        </Dialog.Header>

        {#if detail}
          <div class="rounded-xl border bg-muted/30 p-4 text-sm space-y-2">
            <p><span class="font-medium">Paket:</span> {detail.package.title}</p>
            <p>
              <span class="font-medium">Dampak:</span> identitas dan isi soal
              terkunci; perubahan berikutnya harus dilakukan melalui revisi atau
              salinan baru.
            </p>
            <p>
              <span class="font-medium">Pemakaian:</span>
              {detail.readiness.session_count} sesi memakai paket ini.
            </p>
          </div>
        {/if}

        <label class="space-y-2 block">
          <span class="text-sm font-medium">Alasan Penguncian</span>
          <Textarea
            bind:value={lockReason}
            placeholder="Tuliskan alasan penguncian paket"
            disabled={busy === "lock"}
            rows={4}
          />
          <span class="text-xs text-muted-foreground">
            Alasan ini akan dikirim sebagai catatan penguncian paket.
          </span>
        </label>

        <Dialog.Footer>
          <Button
            type="button"
            variant="outline"
            disabled={busy === "lock"}
            onclick={() => (lockDialogOpen = false)}>Batal</Button
          >
          <LoadingButton
            type="submit"
            loading={busy === "lock"}
            disabled={!lockReason.trim()}>Kunci Paket</LoadingButton
          >
        </Dialog.Footer>
      </form>
    </Dialog.Content>
  </Dialog.Root>
</div>
