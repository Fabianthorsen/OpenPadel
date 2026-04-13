import { writable } from 'svelte/store';

export interface NumpadStoreState {
  value: string;
  fresh: boolean;
  targetPoints: number;
  shaking: boolean;
  onDigit: (d: string) => void;
  onDelete: () => void;
  onConfirm: () => void;
  onClose: () => void;
}

function createNumpadStore() {
  const { subscribe, set, update } = writable<NumpadStoreState | null>(null);

  return {
    subscribe,
    open: (state: NumpadStoreState) => set(state),
    close: () => set(null),
    update: (partial: Partial<Pick<NumpadStoreState, 'value' | 'fresh' | 'shaking'>>) =>
      update(s => s ? { ...s, ...partial } : null),
  };
}

export const numpad = createNumpadStore();
