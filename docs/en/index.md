---
layout: home

hero:
  name: "frp-panel v2"
  text: "A secure, cross-platform FRP control plane"
  tagline: Linux controller with Linux, macOS, and Windows node Agents
  actions:
    - theme: brand
      text: Deployment guide
      link: /en/deployment
    - theme: alt
      text: Quick start
      link: /en/quick-start

features:
  - title: Native system layout
    details: systemd, launchd, or Windows SCM owns the Agent; the bootstrap directory stays clean.
  - title: Secure defaults
    details: One-time enrollment, protected credentials, Argon2id passwords, same-origin WebSockets, and TLS verification.
  - title: Reproducible delivery
    details: Six Agent targets, two multi-architecture images, SHA-256 checksums, SBOMs, provenance, and pinned CI actions.
---
