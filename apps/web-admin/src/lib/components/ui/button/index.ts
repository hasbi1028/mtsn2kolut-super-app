import Root from "./button.svelte";

export type ButtonVariant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
export type ButtonSize = 'default' | 'xs' | 'sm' | 'lg' | 'icon';

export type ButtonProps = {
	variant?: ButtonVariant;
	size?: ButtonSize;
	class?: string;
	disabled?: boolean;
	children?: import('svelte').Snippet;
	href?: string;
};

export {
	Root,
	//
	Root as Button,
};
