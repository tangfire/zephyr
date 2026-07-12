# Changelog

All notable changes to Peapod are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/),
and this project adheres to date-based entries.

## 2026-07-13

### Fixed

- **Product name consistency**: unified the product name to `Pedpod`
  (capitalized first letter) across `brand.ts`, `api.ts`, and all user-visible
  `main.go` copy. Go struct field names (`PedpodTriggeredBy`, etc.) and their
  JSON tags (`peapod_*`) are unchanged — they are internal identifiers and the
  API contract is preserved.

### Refactored

- **Removed dead code**: deleted the unused `productName` constant in
  `main.go` (it was never referenced and carried a wrong-case spelling).
- **Extracted `upsertTaskIntoConfig` helper**: eliminated ~25 lines of
  duplicated "load config → init repos → upsert task → save" logic that was
  repeated in `templateAction` and the `customTasks` POST branch. Net −12
  lines in `main.go`.
- **Frontend cleanup**: removed an ineffective `.replace(/Peapod/g,
  PRODUCT_NAME)` branch in `pages.tsx`'s `productText` (it matched a
  transient mis-spelling that the backend never emits). The `Zefire`/`Zephyr`
  historical-name fallbacks are retained as defensive replacements.
