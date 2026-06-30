# CSS customization

Custom CSS is available under Settings → Appearance. It is stored in Snipo's
shared settings and injected after the built-in styles, so normal CSS cascade
rules apply.

> Custom CSS is trusted owner-supplied content. Browser-side validation catches
> a few mistakes but is not a sanitizer or security boundary.

## Start with variables

Variables are more stable than component class names:

```css
:root {
  --snipo-primary: #7c3aed;
  --snipo-primary-hover: #6d28d9;
  --snipo-primary-rgb: 124, 58, 237;
  --sidebar-width: 240px;
}
```

Theme-specific values belong under the active theme selector:

```css
[data-theme="light"] {
  --snipo-bg-primary: #ffffff;
  --snipo-bg-secondary: #f8fafc;
  --snipo-text-primary: #111827;
  --snipo-text-secondary: #4b5563;
  --snipo-border: #d1d5db;
}

[data-theme="dark"] {
  --snipo-bg-primary: #09090b;
  --snipo-bg-secondary: #18181b;
  --snipo-text-primary: #fafafa;
  --snipo-text-secondary: #a1a1aa;
  --snipo-border: #27272a;
}
```

## Supported variables

The authoritative defaults are in
[`variables.css`](../internal/web/static/css/components/variables.css).
The main customization points are:

| Group | Variables |
|---|---|
| Brand | `--snipo-primary`, `--snipo-primary-hover`, `--snipo-primary-rgb`, `--snipo-secondary` |
| Status | `--snipo-success`, `--snipo-warning`, `--snipo-danger`, `--snipo-danger-rgb` |
| Surfaces | `--snipo-bg-primary`, `--snipo-bg-secondary`, `--snipo-border`, `--card-preview-bg` |
| Text | `--snipo-text-primary`, `--snipo-text-secondary` |
| Sidebar | `--sidebar-width`, `--sidebar-bg`, `--sidebar-text`, `--sidebar-hover`, `--sidebar-border` |
| Editor | `--editor-bg`, `--editor-text`, `--editor-line-numbers` |
| Icons | `--icon-color`, `--icon-hover`, `--icon-active`, `--icon-favorite`, `--icon-danger` |
| Fonts | `--font-sans`, `--font-mono`, `--font-arabic` |

Pico CSS variables such as `--pico-background-color` also work, but Snipo's own
variables are preferred where available.

## Component examples

Use browser developer tools to inspect current class names. Component selectors
may change between releases.

```css
/* Rounded cards */
.snippet-item,
.editor-container,
.modal-content {
  border-radius: 12px;
}

/* Compact snippet list */
.snippet-item {
  padding-block: 0.5rem;
}

/* Sidebar treatment */
.sidebar {
  background: linear-gradient(180deg, #111827, #1f2937);
}
```

Avoid broad selectors such as `button`, `input`, or `*` unless you intend to
change every matching element.

## Light and dark themes

Test both explicit themes and the automatic theme. A rule on `:root` affects
both modes; a rule under `[data-theme="light"]` or `[data-theme="dark"]` affects
only that mode.

If a theme-specific value does not apply, check selector specificity before
adding `!important`. Excessive `!important` rules make upgrades harder.

## RTL customization

Snipo assigns direction based on content. Scope RTL changes with the standard
direction attribute:

```css
[dir="rtl"] {
  font-family: var(--font-arabic);
}

[dir="rtl"] pre,
[dir="rtl"] code {
  direction: ltr;
  text-align: left;
  font-family: var(--font-mono);
}
```

## Validation and troubleshooting

Before saving:

1. Check that braces are balanced.
2. Avoid external `@import` rules; they leak requests to third parties and may
   conflict with the content security policy.
3. Keep CSS below 50 KiB.
4. Test list, preview, edit, settings, and modal views at desktop and mobile
   widths.
5. Test light, dark, and automatic themes.

If styles do not apply, inspect the element and computed styles, verify the
selector, and check the browser console. To recover from a broken layout, clear
the Custom CSS field and save it.
