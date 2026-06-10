// Package fayna provides operational work execution views for the Ichizen OS monorepo.
//
// fayna owns the job lifecycle: templates (design-time), execution (runtime),
// activities (cost capture), and outcomes (Layer 7 quality assessment).
//
// From Filipino faena (Spanish origin) — labor, work, the decisive performance.
package fayna

import "embed"

// AssetsFS embeds this package's static CSS/JS so the app can copy them at boot via
// pyeza.CopyNamespacedAssets — replaces the old CopyStyles/CopyStaticAssets + runtime.Caller hack.
//
// The assets/{css,js} dirs are currently placeholder-only (real assets land as view
// modules are built); the all: prefix keeps the .gitkeep markers embeddable so the
// build stays green while CopyNamespacedAssets copies nothing — same no-op as before.
//
//go:embed all:assets
var AssetsFS embed.FS
