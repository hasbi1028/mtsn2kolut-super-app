export type EditableCellTone = 'default' | 'warning' | 'danger' | 'success';

export type EditableOption = {
	value: string;
	label: string;
	description?: string;
	disabled?: boolean;
};

export type EditableCellCommit<T = string> = {
	value: T;
	previousValue: T;
};

export type EditableCellCancel<T = string> = {
	value: T;
	originalValue: T;
};

export type DirtyChangeBarAction = 'save' | 'discard';

