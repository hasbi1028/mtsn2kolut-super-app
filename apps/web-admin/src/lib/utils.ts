import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import type { HTMLAttributes } from "svelte/elements";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & {
	ref?: U | null;
};

export type WithoutChildren<T> = Omit<T, "children">;
export type WithoutChild<T> = Omit<T, "child">;
export type WithoutChildrenOrChild<T> = Omit<T, "children" | "child">;
export type WithChild<T, U = Record<never, never>> = T & {
	child?: import("svelte").Snippet<[U]>;
};
export type WithChildren<T = Record<never, never>> = T & {
	children?: import("svelte").Snippet;
};
export type AnyExceptNull = Record<string, unknown> | unknown[] | string | number | boolean;
export type TransitionConfig = {
	delay?: number;
	duration?: number;
	easing?: (t: number) => number;
	css?: (t: number, u: number) => string;
	tick?: (t: number, u: number) => void;
};

export type WithAttr<T extends keyof HTMLAttributes<HTMLElement>> = Pick<
	HTMLAttributes<HTMLElement>,
	T
>;
