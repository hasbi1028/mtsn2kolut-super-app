<script lang="ts">
	import { page } from '$app/state';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb';
	import { buildAdminBreadcrumbs, compactAdminBreadcrumbs, type AdminBreadcrumbCrumb } from './admin-breadcrumb';

	let {
		user
	}: {
		user?: { role?: string; roles?: string[]; permissions?: string[] } | null;
	} = $props();

	const roles = $derived(user?.roles ?? (user?.role ? [user.role] : []));
	const permissions = $derived(user?.permissions ?? []);
	const crumbs = $derived(buildAdminBreadcrumbs(page.url.pathname, roles, permissions));
	const compactCrumbs = $derived(compactAdminBreadcrumbs(crumbs));

	function crumbTitle(crumb: AdminBreadcrumbCrumb) {
		return crumb.section ? `${crumb.section} ${crumb.label}` : crumb.label;
	}
</script>

{#if crumbs.length > 1}
	<Breadcrumb.Breadcrumb class="mb-3 rounded-lg border border-border bg-card/90 px-3 py-2 shadow-sm">
		<Breadcrumb.BreadcrumbList class="flex sm:hidden">
			{#each compactCrumbs as crumb, index (index)}
				{#if index === 1 && crumbs.length > 3}
					<Breadcrumb.BreadcrumbItem>
						<span class="text-muted-foreground/70">...</span>
					</Breadcrumb.BreadcrumbItem>
					<Breadcrumb.BreadcrumbSeparator />
				{/if}
				<Breadcrumb.BreadcrumbItem>
					{#if crumb.href}
						<Breadcrumb.BreadcrumbLink href={crumb.href} title={crumbTitle(crumb)} class="max-w-[8.5rem]">
							{#if crumb.section}
								<span class="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.68rem] font-semibold leading-none text-muted-foreground">
									{crumb.section}
								</span>
							{/if}
							<span class="truncate">{crumb.label}</span>
						</Breadcrumb.BreadcrumbLink>
					{:else}
						<Breadcrumb.BreadcrumbPage title={crumbTitle(crumb)} class="max-w-[11rem]">
							{#if crumb.section}
								<span class="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[0.68rem] font-semibold leading-none text-primary">
									{crumb.section}
								</span>
							{/if}
							<span class="truncate">{crumb.label}</span>
						</Breadcrumb.BreadcrumbPage>
					{/if}
				</Breadcrumb.BreadcrumbItem>
				{#if index < compactCrumbs.length - 1}
					<Breadcrumb.BreadcrumbSeparator />
				{/if}
			{/each}
		</Breadcrumb.BreadcrumbList>

		<Breadcrumb.BreadcrumbList class="hidden sm:flex">
			{#each crumbs as crumb, index (index)}
				<Breadcrumb.BreadcrumbItem>
					{#if crumb.href}
						<Breadcrumb.BreadcrumbLink href={crumb.href} title={crumbTitle(crumb)}>
							{#if crumb.section}
								<span class="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.68rem] font-semibold leading-none text-muted-foreground">
									{crumb.section}
								</span>
							{/if}
							<span class="truncate">{crumb.label}</span>
						</Breadcrumb.BreadcrumbLink>
					{:else}
						<Breadcrumb.BreadcrumbPage title={crumbTitle(crumb)}>
							{#if crumb.section}
								<span class="rounded bg-primary/10 px-1.5 py-0.5 font-mono text-[0.68rem] font-semibold leading-none text-primary">
									{crumb.section}
								</span>
							{/if}
							<span class="truncate">{crumb.label}</span>
						</Breadcrumb.BreadcrumbPage>
					{/if}
				</Breadcrumb.BreadcrumbItem>
				{#if index < crumbs.length - 1}
					<Breadcrumb.BreadcrumbSeparator />
				{/if}
			{/each}
		</Breadcrumb.BreadcrumbList>
	</Breadcrumb.Breadcrumb>
{/if}
