# frp-panel web

The console is a Vue 3 / Vite single-page application. The UI primitives are
adapted from the open-source sub2api frontend (LGPL-3.0); see
`src/components/vendor/NOTICE` for attribution.

```bash
pnpm install --frozen-lockfile
pnpm test
pnpm build
```

The production build is written to `cmd/frpp/out` for embedding in the Master binary.
