---
name: Scope reset — April 2026
description: Fabian asked for a clean-slate, simpler plan on 2026-04-21 after a phase of building complex Timed Americano UI
type: project
---

Fabian explicitly asked for a fresh, simplified plan on 2026-04-21. The Timed Americano slice shipped (PR 1/2/3 all marked Done in ROADMAP) but the associated UI surface area grew: BufferTimer, NextRoundPreview, RoundIntervalCountdown, RoundTimer, interval_between_rounds_minutes config, drift correction, separate A/B scoring flow inside a 768-line ActiveSession.svelte. His read is that the complexity is outpacing the user value.

**Why:** Personal/family-use PWA (per PROJECT.md). Target is courtside phones on flaky WiFi. The product wedge is "fast, one-tap, boring" — not configurable timers or predictive UI.

**How to apply:**
- When a Timed Americano enhancement is proposed, default to "can we cut this?" before "can we build it?"
- Prefer collapsing the three timer components (RoundTimer, BufferTimer, RoundIntervalCountdown) into one surface with states
- Treat interval_between_rounds_minutes as a candidate for removal unless usage data says otherwise
- Weigh any new config knob against the cost on CreateDrawer (already 451 lines) and ActiveSession (768 lines)
