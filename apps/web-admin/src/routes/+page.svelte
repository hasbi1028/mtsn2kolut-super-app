<script lang="ts">
  import { onMount } from 'svelte';
  import * as Card from '$lib/components/ui/card';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';

  let { data }: { data: { user?: { role?: string; employee_id?: string } } } = $props();

  interface AcademicStats {
    total_students: number; total_classes: number;
    total_subjects: number; total_years: number;
  }
  interface GuruStats {
    active_sessions: number; ungraded_essays: number;
    my_students: number; my_subjects: number;
  }

  let academicStats = $state<AcademicStats | null>(null);
  let guruStats     = $state<GuruStats | null>(null);

  const isGuru = $derived(data.user?.role === 'guru');

  async function load() {
    try {
      if (isGuru) {
        const [sessRes, essRes, stuRes] = await Promise.all([
          fetch('/api/cbt/sessions?limit=100'),
          fetch('/api/cbt/sessions/my-essays'),
          fetch('/api/students?active=1'),
        ]);
        const sessions = await sessRes.json().catch(() => []);
        const essays   = await essRes.json().catch(() => []);
        const students = await stuRes.json().catch(() => []);

        const activeSessions = Array.isArray(sessions)
          ? sessions.filter((s: any) => s.status === 'active' || s.status === 'scheduled')
          : [];
        const subjectsSet = new Set(activeSessions.map((s: any) => s.package_title));

        guruStats = {
          active_sessions: activeSessions.length,
          ungraded_essays: Array.isArray(essays) ? essays.length : 0,
          my_students:     Array.isArray(students) ? students.length : 0,
          my_subjects:     subjectsSet.size,
        };
        return;
      }

      const aRes = await fetch('/api/academic/stats');
      const a    = await aRes.json().catch(() => null);
      academicStats = a;
    } catch { /* silent */ }
  }

  onMount(() => { load(); });
</script>

<svelte:head><title>Dashboard — MTSN 2 Kolut</title></svelte:head>

<div class="space-y-6">

  <!-- Header -->
  <div>
    {#if isGuru}
      <h1 class="text-2xl font-semibold text-slate-800">Dashboard Guru</h1>
      <p class="text-sm text-muted-foreground mt-1">Mata pelajaran yang Anda ampu dan aktivitas CBT</p>
    {:else}
      <h1 class="text-2xl font-semibold text-slate-800">Dashboard</h1>
      <p class="text-sm text-muted-foreground mt-1">Ringkasan operasional MTs Negeri 2 Kolaka Utara</p>
    {/if}
  </div>

  <!-- Guru Stats -->
  {#if isGuru && guruStats}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Sesi Ujian Aktif</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{guruStats.active_sessions}</p>
          <p class="text-xs text-muted-foreground mt-1">Mata pelajaran saya</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-amber-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Essay Belum Dikoreksi</p>
          <p class="text-3xl font-bold text-amber-700 mt-1">{guruStats.ungraded_essays}</p>
          <p class="text-xs text-muted-foreground mt-1">Perlu penilaian manual</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Siswa Saya</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{guruStats.my_students}</p>
          <p class="text-xs text-muted-foreground mt-1">Di kelas yang saya ajar</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Mata Pelajaran</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{guruStats.my_subjects}</p>
          <p class="text-xs text-muted-foreground mt-1">Yang saya ampu</p>
        </Card.Content>
      </Card.Root>
    </div>
  {/if}

  <!-- Academic Stats (admin only) -->
  {#if !isGuru}
    <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Total Siswa</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{academicStats?.total_students ?? '—'}</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Kelas Aktif</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{academicStats?.total_classes ?? '—'}</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Mata Pelajaran</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{academicStats?.total_subjects ?? '—'}</p>
        </Card.Content>
      </Card.Root>
      <Card.Root class="border-green-100">
        <Card.Content class="pt-4">
          <p class="text-xs font-medium text-muted-foreground uppercase tracking-wider">Tahun Ajaran</p>
          <p class="text-3xl font-bold text-green-800 mt-1">{academicStats?.total_years ?? '—'}</p>
        </Card.Content>
      </Card.Root>
    </div>

    <!-- PUSAKA quick access -->
    <Card.Root class="border-slate-200">
      <Card.Header class="pb-3">
        <div class="flex items-center justify-between">
          <div>
            <Card.Title class="text-base">Integrasi PUSAKA Kemenag</Card.Title>
            <Card.Description>Sinkronisasi data kehadiran pegawai dari sistem pemerintah</Card.Description>
          </div>
          <Badge variant="outline" class="text-xs">Eksternal</Badge>
        </div>
      </Card.Header>
      <Card.Content>
        <div class="flex flex-wrap gap-2">
          <Button variant="default" size="sm" href="/pusaka">Kontrol & Monitor</Button>
          <Button variant="outline" size="sm" href="/pusaka/kehadiran">Data Kehadiran</Button>
          <Button variant="outline" size="sm" href="/pusaka/summary">Ringkasan</Button>
          <Button variant="outline" size="sm" href="/pusaka/antrian">Antrian Job</Button>
        </div>
      </Card.Content>
    </Card.Root>
  {/if}

</div>
