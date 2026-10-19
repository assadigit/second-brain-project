// icons.js — the Lucide icon pack's frontend twin (T-72): the reference's
// gui/icons.py ICONS dict (lucide-static v0.544.0, ISC; 28 glyphs),
// verbatim, locked byte-for-byte against the reference golden by
// webui/uitheme's test (the pack is extracted from THIS file and diffed
// against testdata/icons.golden.json — a bad edit can never pass
// silently).
//
// The markup keeps the reference's '{color}' render-time tint markers;
// svgEl() swaps them for CSS currentColor at render time (the HTML twin
// of the reference's _svg() color substitution — the tint then follows
// the glyph's context color, exactly as QSvgRenderer follows the passed
// tint). Do NOT redraw these by hand: the golden locks the bytes.

// The pack — verbatim inner SVG markup, '{color}' markers and all.
// (sorted like the golden's keys; one entry per line)
const PACK = {
  'activity': `<path d="M22 12h-2.48a2 2 0 0 0-1.93 1.46l-2.35 8.36a.25.25 0 0 1-.48 0L9.24 2.18a.25.25 0 0 0-.48 0l-2.35 8.36A2 2 0 0 1 4.49 12H2" />`,
  'archive': `<rect width="20" height="5" x="2" y="3" rx="1" /> <path d="M4 8v11a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8" /> <path d="M10 12h4" />`,
  'bar-chart-3': `<path d="M3 3v16a2 2 0 0 0 2 2h16" /> <path d="M18 17V9" /> <path d="M13 17V5" /> <path d="M8 17v-3" />`,
  'bot': `<path d="M12 8V4H8" /> <rect width="16" height="12" x="4" y="8" rx="2" /> <path d="M2 14h2" /> <path d="M20 14h2" /> <path d="M15 13v2" /> <path d="M9 13v2" />`,
  'brain': `<path d="M12 18V5" /> <path d="M15 13a4.17 4.17 0 0 1-3-4 4.17 4.17 0 0 1-3 4" /> <path d="M17.598 6.5A3 3 0 1 0 12 5a3 3 0 1 0-5.598 1.5" /> <path d="M17.997 5.125a4 4 0 0 1 2.526 5.77" /> <path d="M18 18a4 4 0 0 0 2-7.464" /> <path d="M19.967 17.483A4 4 0 1 1 12 18a4 4 0 1 1-7.967-.517" /> <path d="M6 18a4 4 0 0 1-2-7.464" /> <path d="M6.003 5.125a4 4 0 0 0-2.526 5.77" />`,
  'check': `<path d="M20 6 9 17l-5-5"/>`,
  'circle': `<circle cx="12" cy="12" r="10"/>`,
  'circle-x': `<circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/>`,
  'dot': `<circle cx="12" cy="12" r="10" fill="{color}" stroke="none"/>`,
  'folder': `<path d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z" />`,
  'globe': `<circle cx="12" cy="12" r="10" /> <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" /> <path d="M2 12h20" />`,
  'inbox': `<polyline points="22 12 16 12 14 15 10 15 8 12 2 12" /> <path d="M5.45 5.11 2 12v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3.45-6.89A2 2 0 0 0 16.76 4H7.24a2 2 0 0 0-1.79 1.11z" />`,
  'info': `<circle cx="12" cy="12" r="10"/><path d="M12 16v-4"/><path d="M12 8h.01"/>`,
  'key-round': `<path d="M2.586 17.414A2 2 0 0 0 2 18.828V21a1 1 0 0 0 1 1h3a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h1a1 1 0 0 0 1-1v-1a1 1 0 0 1 1-1h.172a2 2 0 0 0 1.414-.586l.814-.814a6.5 6.5 0 1 0-4-4z" /> <circle cx="16.5" cy="7.5" r=".5" fill="{color}" stroke="none" />`,
  'layers': `<path d="M12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83z" /> <path d="M2 12a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 12" /> <path d="M2 17a1 1 0 0 0 .58.91l8.6 3.91a2 2 0 0 0 1.65 0l8.58-3.9A1 1 0 0 0 22 17" />`,
  'list': `<line x1="8" x2="21" y1="6" y2="6" /> <line x1="8" x2="21" y1="12" y2="12" /> <line x1="8" x2="21" y1="18" y2="18" /> <line x1="3" x2="3.01" y1="6" y2="6" /> <line x1="3" x2="3.01" y1="12" y2="12" /> <line x1="3" x2="3.01" y1="18" y2="18" />`,
  'loader': `<path d="M21 12a9 9 0 1 1-6.219-8.56" />`,
  'moon': `<path d="M20.985 12.486a9 9 0 1 1-9.473-9.472c.405-.022.617.46.402.803a6 6 0 0 0 8.268 8.268c.344-.215.825-.004.803.401" />`,
  'play': `<path d="M5 5a2 2 0 0 1 3.008-1.728l11.997 6.998a2 2 0 0 1 .003 3.458l-12 7A2 2 0 0 1 5 19z" fill="{color}" stroke="none"/>`,
  'refresh': `<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" /> <path d="M21 3v5h-5" /> <path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" /> <path d="M8 16H3v5" />`,
  'satellite-dish': `<path d="M4 10a7.31 7.31 0 0 0 10 10Z" /> <path d="m9 15 3-3" /> <path d="M17 13a6 6 0 0 0-6-6" /> <path d="M21 13A10 10 0 0 0 11 3" />`,
  'search': `<path d="m21 21-4.34-4.34" /> <circle cx="11" cy="11" r="8" />`,
  'settings': `<path d="M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915" /> <circle cx="12" cy="12" r="3" />`,
  'stop': `<rect width="18" height="18" x="3" y="3" rx="2" fill="{color}" stroke="none"/>`,
  'sun': `<circle cx="12" cy="12" r="4" /> <path d="M12 2v2" /> <path d="M12 20v2" /> <path d="m4.93 4.93 1.41 1.41" /> <path d="m17.66 17.66 1.41 1.41" /> <path d="M2 12h2" /> <path d="M20 12h2" /> <path d="m6.34 17.66-1.41 1.41" /> <path d="m19.07 4.93-1.41 1.41" />`,
  'trash': `<path d="M10 11v6" /> <path d="M14 11v6" /> <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6" /> <path d="M3 6h18" /> <path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />`,
  'triangle-alert': `<path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3"/><path d="M12 9v4"/><path d="M12 17h.01"/>`,
  'volume-2': `<path d="M11 4.702a.705.705 0 0 0-1.203-.498L6.413 7.587A1.4 1.4 0 0 1 5.416 8H3a1 1 0 0 0-1 1v6a1 1 0 0 0 1 1h2.416a1.4 1.4 0 0 1 .997.413l3.383 3.384A.705.705 0 0 0 11 19.298z" /> <path d="M16 9a5 5 0 0 1 0 6" /> <path d="M19.364 18.364a9 9 0 0 0 0-12.728" />`,
};

// The log desk's level map — the reference's _LOG_LEVEL_ICONS law
// (ui.py:2144): the level rides the SHAPE-coded glyph, the color on top.
export const LEVEL_ICONS = {
  success: 'check',
  warning: 'triangle-alert',
  error: 'circle-x',
  info: 'dot',
};

// svgEl builds one tinted inline SVG element for a glyph — the render
// twin of the reference's _svg()/pixmap(): '{color}' becomes CSS
// currentColor so the glyph follows its context color; the root carries
// the 2px round strokes on the 24×24 viewBox and the elements override
// per-element exactly as in the reference (the solid glyphs' own
// fill="…" stroke="none" markers win over the root, per SVG precedence).
// Unknown glyphs degrade to the empty svg (the graceful-degradation
// law — the app still runs, the button keeps its label).
export function svgEl(name, cls) {
  const wrap = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  wrap.setAttribute('viewBox', '0 0 24 24');
  if (cls) wrap.setAttribute('class', cls);
  const body = PACK[name];
  if (body === undefined) return wrap; // empty icon, never a crash
  wrap.setAttribute('fill', 'none');
  wrap.setAttribute('stroke', 'currentColor');
  wrap.setAttribute('stroke-width', '2');
  wrap.setAttribute('stroke-linecap', 'round');
  wrap.setAttribute('stroke-linejoin', 'round');
  wrap.innerHTML = body.replaceAll('{color}', 'currentColor');
  return wrap;
}
