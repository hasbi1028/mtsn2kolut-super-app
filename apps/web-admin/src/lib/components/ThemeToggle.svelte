<script lang="ts">
	import { onMount } from 'svelte';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import MoonIcon from '@lucide/svelte/icons/moon';
	import SunIcon from '@lucide/svelte/icons/sun';
	import { Button, type ButtonSize, type ButtonVariant } from '$lib/components/ui/button';
	import {
		THEME_QUERY,
		THEME_STORAGE_KEY,
		applyThemePreference,
		effectiveThemeLabel,
		nextThemePreference,
		normalizeThemePreference,
		readThemePreference,
		themePreferenceLabel,
		writeThemePreference,
		type EffectiveTheme,
		type ThemePreference
	} from '$lib/client/theme';

	let {
		expanded = true,
		variant = 'outline',
		size = 'sm',
		class: className = ''
	}: {
		expanded?: boolean;
		variant?: ButtonVariant;
		size?: ButtonSize;
		class?: string;
	} = $props();

	let preference = $state<ThemePreference>('system');
	let effectiveTheme = $state<EffectiveTheme>('light');
	let mounted = $state(false);

	const preferenceLabel = $derived(themePreferenceLabel(preference));
	const effectiveLabel = $derived(effectiveThemeLabel(effectiveTheme));
	const buttonLabel = $derived(
		preference === 'system'
			? `Tema: ${preferenceLabel} (${effectiveLabel})`
			: `Tema: ${preferenceLabel}`
	);

	function setPreference(nextPreference: ThemePreference) {
		preference = writeThemePreference(nextPreference);
		effectiveTheme = applyThemePreference(preference);
	}

	function cycleTheme() {
		setPreference(nextThemePreference(preference));
	}

	function listenToSystemTheme(media: MediaQueryList, listener: (event: MediaQueryListEvent) => void) {
		const legacyMedia = media as MediaQueryList & {
			addListener?: (callback: (event: MediaQueryListEvent) => void) => void;
			removeListener?: (callback: (event: MediaQueryListEvent) => void) => void;
		};

		if (typeof media.addEventListener === 'function') {
			media.addEventListener('change', listener);
			return () => media.removeEventListener('change', listener);
		}

		legacyMedia.addListener?.(listener);
		return () => legacyMedia.removeListener?.(listener);
	}

	onMount(() => {
		mounted = true;
		preference = readThemePreference();
		effectiveTheme = applyThemePreference(preference);

		const media = window.matchMedia(THEME_QUERY);
		const handleMediaChange = () => {
			if (preference === 'system') {
				effectiveTheme = applyThemePreference(preference);
			}
		};
		const handleStorage = (event: StorageEvent) => {
			if (event.key !== THEME_STORAGE_KEY) return;
			preference = normalizeThemePreference(event.newValue);
			effectiveTheme = applyThemePreference(preference);
		};

		const stopMediaListener = listenToSystemTheme(media, handleMediaChange);
		window.addEventListener('storage', handleStorage);
		return () => {
			stopMediaListener();
			window.removeEventListener('storage', handleStorage);
		};
	});
</script>

<Button
	{variant}
	{size}
	class={className}
	onclick={cycleTheme}
	aria-label={`Ubah tema. Saat ini ${mounted ? buttonLabel : 'Tema: Sistem'}`}
	title={mounted ? buttonLabel : 'Tema'}
>
	{#if preference === 'system'}
		<MonitorIcon class="size-4" />
	{:else if effectiveTheme === 'dark'}
		<MoonIcon class="size-4" />
	{:else}
		<SunIcon class="size-4" />
	{/if}
	<span class={expanded ? 'inline' : 'sr-only'}>{mounted ? buttonLabel : 'Tema'}</span>
</Button>
