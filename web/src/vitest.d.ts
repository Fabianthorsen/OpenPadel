// Makes @testing-library/jest-dom's custom matchers (toBeInTheDocument,
// toHaveAttribute, …) visible to TypeScript. The runtime registration lives in
// vitest-setup.ts; this only pulls in the vitest `expect` augmentation for the
// type checker.
import '@testing-library/jest-dom/vitest';
