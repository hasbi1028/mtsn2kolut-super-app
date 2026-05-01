<script lang="ts">
  import { tick } from 'svelte';

  let { open = $bindable(false), children } = $props();
  let overlayRef: HTMLDivElement | null = $state(null);

  function onOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) open = false;
  }

  function focusableElements() {
    if (!overlayRef) return [];
    return Array.from(
      overlayRef.querySelectorAll<HTMLElement>(
        'a[href], button:not([disabled]), textarea:not([disabled]), input:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])'
      )
    ).filter((element) => !element.hasAttribute('aria-hidden'));
  }

  function onKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      open = false;
      return;
    }

    if (event.key !== 'Tab') return;
    const focusable = focusableElements();
    if (focusable.length === 0) {
      event.preventDefault();
      overlayRef?.focus();
      return;
    }

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
      return;
    }
    if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  }

  $effect(() => {
    if (!open || typeof document === 'undefined') return;
    const previousActive = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    tick().then(() => {
      const [first] = focusableElements();
      (first ?? overlayRef)?.focus();
    });

    return () => {
      document.body.style.overflow = previousOverflow;
      previousActive?.focus();
    };
  });
</script>

{#if open}
  <div
    bind:this={overlayRef}
    class="overlay"
    onclick={onOverlayClick}
    onkeydown={onKeydown}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    {@render children?.()}
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 50;
    padding: 1rem;
  }
</style>
