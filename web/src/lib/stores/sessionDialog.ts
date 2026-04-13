import { writable } from 'svelte/store';

type DialogType = 'close' | 'cancel';

interface SessionDialogState {
  type: DialogType | null;
  isOpen: boolean;
  onConfirm?: () => Promise<void>;
}

function createSessionDialogStore() {
  const { subscribe, set, update } = writable<SessionDialogState>({
    type: null,
    isOpen: false,
  });

  return {
    subscribe,
    open: (type: DialogType, onConfirm?: () => Promise<void>) => set({ type, isOpen: true, onConfirm }),
    close: () => set({ type: null, isOpen: false }),
  };
}

export const sessionDialog = createSessionDialogStore();
