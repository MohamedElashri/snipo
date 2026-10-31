# CSS customization

Snipo uses local vanilla CSS for its interface. Customize it under **Settings →
Appearance → Custom CSS**, then click **Save and Apply Custom CSS**. Saving
applies the stylesheet immediately; clear the field and save again to remove it.

Custom CSS is stored in the instance's shared settings. Other browsers load the
saved stylesheet when they next load the main snippets app. The login and public
snippet pages use the built-in styles but do not load saved custom CSS.

Custom CSS is injected after the built-in styles. Normal CSS specificity and
inheritance still apply; inline styles and `!important` rules can take precedence.

> Custom CSS is trusted owner-supplied content. The browser's validation checks
> for a few common mistakes; it is not a CSS parser, sanitizer, or security boundary.

## Start with variables

Variables are the easiest way to customize colors without replacing component
rules. This example changes branded buttons and accents in both themes:

```css
:root {
  --snipo-primary: #7c3aed;
  --snipo-primary-hover: #6d28d9;
}
```

Links and native controls have their own color variables. Changing the brand
color does not recolor every interactive element. To give them the same accent:

```css
:root[data-theme="light"],
:root[data-theme="dark"] {
  --snipo-link-color: #7c3aed;
  --snipo-link-hover: #6d28d9;
  --snipo-link-underline: rgba(124, 58, 237, 0.5);
  --snipo-button-bg: #7c3aed;
  --snipo-button-hover-bg: #6d28d9;
  --snipo-focus-color: rgba(124, 58, 237, 0.4);
}
```

Use the theme toggle to check both modes. The UI uses explicit `light` and `dark`
themes, remembers the choice in browser local storage, and defaults to dark.
The editor's **Auto (Follow App Theme)** option controls its syntax theme; it
is separate from the UI theme toggle.

## Theme-specific palettes

Put palette changes on the document's active theme selector. For example:

```css
:root[data-theme="light"] {
  --snipo-background: #ffffff;
  --snipo-card-background: #f8fafc;
  --snipo-color: #111827;
  --snipo-text-muted: #4b5563;
  --snipo-border-muted: #d1d5db;
}

:root[data-theme="dark"] {
  --snipo-background: #09090b;
  --snipo-card-background: #18181b;
  --snipo-color: #fafafa;
  --snipo-text-muted: #a1a1aa;
  --snipo-border-muted: #3f3f46;
}
```

These variables control the main backgrounds, cards, dialogs, and general text.
The sidebar, heading levels, form controls, icons, and preview code have separate
colors. Some components also use `--snipo-text-primary`,
`--snipo-text-secondary`, `--snipo-bg-secondary`, and `--snipo-border` directly;
update those separately when completing a palette.

`--snipo-color` and `--snipo-background` are also set locally on links, buttons,
and inputs. Headings set a local `--snipo-color` from their heading color. A root
value therefore does not replace every element's foreground or background.
Use the dedicated variables below or a scoped component rule for those elements.

## Variable reference

Defaults are defined in
[`variables.css`](../internal/web/static/css/components/variables.css).
Shared HTML defaults are in
[`base.css`](../internal/web/static/css/components/base.css), and
[`main.css`](../internal/web/static/css/main.css) lists the component stylesheets
in their import order.

| Area | Variables and purpose |
|---|---|
| Brand | `--snipo-primary`, `--snipo-primary-hover`: branded buttons and accents; `--snipo-secondary`: secondary status color |
| Main surfaces | `--snipo-background`, `--snipo-card-background`, `--snipo-border-muted`: page, card/dialog, and general border colors |
| General text | `--snipo-color`, `--snipo-text-muted`: default and muted text |
| Component colors | `--snipo-text-primary`, `--snipo-text-secondary`, `--snipo-bg-secondary`, `--snipo-border`: colors used directly by some components |
| Links | `--snipo-link-color`, `--snipo-link-hover`, `--snipo-link-underline`: link colors and underline |
| Default buttons | `--snipo-button-bg`, `--snipo-button-hover-bg`, `--snipo-button-secondary-bg`, `--snipo-button-secondary-hover-bg`: native button and file-picker defaults |
| Inputs | `--snipo-input-bg`, `--snipo-input-border`, `--snipo-input-color`, `--snipo-input-placeholder`, `--snipo-input-active-bg`, `--snipo-input-selected-bg`: native input, select, and textarea defaults |
| Focus and selection | `--snipo-focus-color`: shared focus ring; `--snipo-search-focus-rgb`: search focus tint as comma-separated RGB channels; `--snipo-selection-bg`: selected text background |
| Headings | `--snipo-h1-color` through `--snipo-h6-color`: colors for each heading level |
| Inline code | `--snipo-code-bg`, `--snipo-code-color`: base code styling; previews can override these |
| Sidebar | `--sidebar-width`, `--sidebar-bg`, `--sidebar-text`, `--sidebar-hover`, `--sidebar-border` |
| Code previews | `--card-preview-bg`, `--editor-bg`, `--editor-text`: snippet-card and preview colors |
| Icons | `--icon-color`, `--icon-hover`, `--icon-favorite`, `--icon-danger` |
| Status | `--snipo-success`, `--snipo-warning`, `--snipo-danger`, `--snipo-danger-rgb` |
| Typography | `--font-sans`: body and controls; `--font-system`: base headings; `--font-mono`: app code/filename text; `--font-code`: base code elements; `--font-arabic`: Arabic text styling |
| Other HTML styling | `--snipo-contrast`, `--snipo-mark-bg`, `--snipo-mark-color`, `--snipo-ins-color`, `--snipo-del-color` |
| Cards | `--snipo-card-shadow`: default snippet-card shadow; `--snipo-shadow-rgb`: its RGB channels |

Some existing components have explicit colors or their own focus rules. Use
browser developer tools to check which variable or declaration controls a
particular element. Prism syntax highlighting and Ace editor themes have their
own color rules. For the code editor's font size, syntax theme, and layout
options, use **Settings → Editor**. For rendered Markdown size, use **Settings →
Appearance → Markdown Preview Font Size**; the app sets `--markdown-font-size`
inline from that setting.

A few older variables remain defined for compatibility with custom styles:
`--snipo-bg-primary`, `--snipo-primary-rgb`, `--editor-line-numbers`, and
`--icon-active`. The built-in styles currently do not read them. In particular,
use `--snipo-background` to change the main page background.

## Component examples

Use current class names from the DOM. Component selectors may change between
releases, and more specific built-in rules can override a broad selector.

```css
/* Round cards and the outer modal dialog. */
.snippet-card,
.modal {
  border-radius: 12px;
}

/* Make list-view cards more compact. */
.snippet-list-view .snippet-card {
  padding-block: 0.35rem;
}

/* Give the dark sidebar a gradient. */
[data-theme="dark"] .sidebar {
  background: linear-gradient(180deg, #111827, #1f2937);
}
```

To change sidebar width, start with:

```css
:root {
  --sidebar-width: 240px;
}
```

Dragging the sidebar's resize handle writes an inline `--sidebar-width` value and
saves it as `sidebarWidth` in browser local storage. That inline value takes
precedence over the CSS example. To restore CSS-controlled width in that browser,
remove the `sidebarWidth` local-storage entry in developer tools and reload.

Avoid broad selectors such as `button`, `input`, or `*` unless you intend to
change every matching element. Keep keyboard focus, disabled states, and busy
indicators visible when styling controls.

## Migrating custom CSS from Pico

Pico CSS is no longer loaded. Saved custom CSS is not automatically rewritten,
and `--pico-*` variables no longer control Snipo's built-in styles. Update both
variable declarations and `var(...)` references using the replacements below.
If your stylesheet defines and reads a Pico variable itself, that private
variable can still work, but the built-in interface will not read it.

| Former variable | Current variable |
|---|---|
| `--pico-background-color` | `--snipo-background` |
| `--pico-card-background-color` | `--snipo-card-background` |
| `--pico-color` | `--snipo-color` |
| `--pico-muted-color` | `--snipo-text-muted` |
| `--pico-muted-border-color` | `--snipo-border-muted` |
| `--pico-primary` | `--snipo-link-color` |
| `--pico-primary-hover` | `--snipo-link-hover` |
| `--pico-form-element-background-color` | `--snipo-input-bg` |
| `--pico-form-element-border-color` | `--snipo-input-border` |
| `--pico-code-background-color` | `--snipo-code-bg` |
| `--pico-secondary-background` | `--snipo-button-secondary-bg` |
| `--pico-contrast` | `--snipo-contrast` |

This table lists common replacements. It is not a compatibility
layer for the full Pico framework. Pico layout classes and other framework
variables need to be replaced with your own CSS.

## RTL customization

Snipo marks Arabic content with `.rtl` classes and, in some editor paths,
`dir="rtl"` attributes. Match both when customizing text:

```css
.editor-title-input.rtl,
.description-textarea.rtl,
.preview-markdown-scroll.rtl,
.public-title.rtl,
[dir="rtl"] {
  font-family: var(--font-arabic);
}

.preview-markdown-scroll.rtl pre,
.preview-markdown-scroll.rtl code {
  direction: ltr;
  text-align: left;
  font-family: var(--font-mono);
}
```

Code inside RTL Markdown previews stays left-to-right by default. Public-page
selectors only apply if you change the built-in stylesheet; saved Custom CSS is
loaded by the main snippets app only.

## Validation and troubleshooting

Before saving, check both themes at desktop and mobile widths, including grid
and list views, Markdown preview, editing, settings, and dialogs. Check hover,
keyboard focus, checked, disabled, and busy states as well.

The browser blocks text containing `<script`. External `@import url(...)`
references, unmatched brace counts, and CSS longer than 50,000 JavaScript string
code units trigger warnings that you can choose to override. The size warning
is not a hard 50 KiB limit, and brace counting does not establish that the CSS is
valid. Prefer the bundled fonts and local assets over external imports, which
can expose requests to third parties or be blocked by the content security policy.

If a rule does not apply:

1. Inspect the element and its computed styles to find the winning declaration.
2. Check the active `data-theme`, selector specificity, and any inline styles.
3. Check that Custom CSS was saved successfully and the
   `<style id="snipo-custom-css">` element is present in the page's `<head>`.
4. Remember that new CSS is loaded after settings arrive, not before the first
   page render. Reload other open browsers to pick up the saved version.

To recover from a broken layout, remove the `snipo-custom-css` style element in
browser developer tools to restore the built-in appearance temporarily. Then
open Settings, clear or fix the saved CSS, and save it again. Removing the style
element alone does not erase the stored stylesheet.

## Local development checks

Run `npm install` from the repository root to install the workspace dependencies.
Then, with Go available:

```sh
npx playwright install chromium
npm run test:web
```

The browser checks build an isolated server and cover login, settings controls,
Markdown, editing, RTL text, and public-page styling at 1440px and 390px in light
and dark themes. Public-page styling uses fixture responses to isolate the
checks from the existing public URL parsing issue. The test database and server
log remain in the temporary directory printed by the command.
