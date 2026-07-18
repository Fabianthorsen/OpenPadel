// The 1–5 self-set Rating scale from ADR 0006. The concrete labels and
// milestone descriptions live in i18n under `rating_<n>_name` / `rating_<n>_desc`
// so RatingPicker and the lobby render one source of truth.
export const RATING_LEVELS = [1, 2, 3, 4, 5] as const;
