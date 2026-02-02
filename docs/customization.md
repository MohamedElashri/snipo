# Snipo CSS Customization Guide

This guide explains how to customize the appearance of Snipo using custom CSS. You can apply custom styles through Settings → Appearance → Custom CSS.

## Table of Contents

- [Getting Started](#getting-started)
- [CSS Variables Reference](#css-variables-reference)
- [Component Classes](#component-classes)
- [Common Customizations](#common-customizations)
- [Theme Examples](#theme-examples)
- [Advanced Techniques](#advanced-techniques)
- [Testing Your Custom CSS](#testing-your-custom-css)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Getting Started

### Accessing Custom CSS

1. Click the **Settings** icon (gear) in the top right
2. Navigate to the **Appearance** tab
3. Scroll to the **Custom CSS** section
4. Enter your CSS in the textarea
5. Click **Save and Apply Custom CSS**

Your custom styles apply immediately and persist across sessions.

### How It Works

Custom CSS is injected into the page as a `<style id="snipo-custom-css">` tag in the document head. It has high specificity and will override default styles. The CSS is validated before being applied to prevent security issues and common errors.

### Important: Dark & Light Mode Support

Snipo supports both dark and light themes via `[data-theme="dark"]` and `[data-theme="light"]` attributes on the `<html>` element. When writing custom CSS:

1. **Test in both modes** - Your CSS should work in both themes
2. **Use theme-specific selectors** - Target `[data-theme="dark"]` or `[data-theme="light"]` when needed
3. **Override CSS variables** - The easiest way to support both themes is to override variables in each theme context

**Example - Theme-aware customization:**
```css
/* Works in both themes by overriding CSS variables */
:root {
  --snipo-accent: #ff6b6b;  /* Applies to both themes */
}

/* Dark theme specific */
[data-theme="dark"] {
  --sidebar-bg: #000000;
  --editor-bg: #1a1a1a;
}

/* Light theme specific */
[data-theme="light"] {
  --sidebar-bg: #f0f0f0;
  --editor-bg: #ffffff;
}
```

### Validation

The app validates your CSS for:
- **Security**: Blocks `<script>` tags and warns about external `@import` URLs
- **Syntax**: Checks for balanced braces `{}`
- **Size**: Warns if CSS exceeds 50KB (performance consideration)

## CSS Variables Reference

Snipo uses CSS custom properties (variables) for theming. Overriding these is the easiest way to customize the interface and ensure compatibility with both dark and light modes.

### Core Application Variables

These variables control the overall application appearance:

```css
:root {
  /* Primary theme colors */
  --snipo-primary: #6366f1;        /* Main brand color */
  --snipo-accent: #8b5cf6;         /* Accent/highlight color */
  
  /* Background colors (theme-specific) */
  --pico-background-color: ...;    /* Main app background */
  --pico-card-background-color: ...; /* Cards and panels */
  
  /* Text colors (theme-specific) */
  --pico-color: ...;               /* Primary text */
  --pico-muted-color: ...;         /* Secondary text */
  
  /* Borders (theme-specific) */
  --pico-muted-border-color: ...;  /* Border color */
}
```

### Sidebar Variables (Theme-Specific)

```css
/* Light mode */
[data-theme="light"] {
  --sidebar-bg: #f1f5f9;
  --sidebar-text: #1e293b;
  --sidebar-hover: #e2e8f0;
  --sidebar-border: #cbd5e1;
}

/* Dark mode */
[data-theme="dark"] {
  --sidebar-bg: #0d1117;
  --sidebar-text: #c9d1d9;
  --sidebar-hover: #161b22;
  --sidebar-border: #30363d;
}
```

### Editor Variables (Theme-Specific)

```css
/* Light mode */
[data-theme="light"] {
  --editor-bg: #f8fafc;
  --editor-text: #1e293b;
  --editor-line-numbers: #94a3b8;
}

/* Dark mode */
[data-theme="dark"] {
  --editor-bg: #1e1e1e;
  --editor-text: #d4d4d4;
  --editor-line-numbers: #6e7681;
}
```

### Icon Variables (Theme-Specific)

```css
/* Light mode */
[data-theme="light"] {
  --icon-color: #475569;
  --icon-hover: #1e293b;
  --icon-active: var(--snipo-primary);
  --icon-favorite: #eab308;
  --icon-danger: #dc2626;
}

/* Dark mode */
[data-theme="dark"] {
  --icon-color: #8b949e;
  --icon-hover: #c9d1d9;
  --icon-active: #818cf8;
  --icon-favorite: #f0c14b;
  --icon-danger: #f85149;
}
```

### Typography Variables

```css
:root {
  /* Font families */
  --font-sans: 'Inter', system-ui, -apple-system, sans-serif;
  --font-mono: 'Fira Code', 'Monaco', 'Courier New', monospace;
  
  /* Font sizes */
  --font-size-xs: 0.75rem;    /* 12px */
  --font-size-sm: 0.875rem;   /* 14px */
  --font-size-base: 1rem;     /* 16px */
  --font-size-lg: 1.125rem;   /* 18px */
  --font-size-xl: 1.25rem;    /* 20px */
  
  /* Markdown preview font size (user-configurable) */
  --markdown-font-size: 14px;
}
```

### Spacing Variables

```css
:root {
  --spacing-xs: 0.25rem;   /* 4px */
  --spacing-sm: 0.5rem;    /* 8px */
  --spacing-md: 1rem;      /* 16px */
  --spacing-lg: 1.5rem;    /* 24px */
  --spacing-xl: 2rem;      /* 32px */
}
```

## Component Classes

### Main Layout

```css
.app-container             /* Entire app container */
.main-content              /* Main content area (right of sidebar) */
.main-content.expanded     /* Main content when sidebar is collapsed */
```

### Sidebar Components

```css
.sidebar                    /* Main sidebar container */
.sidebar-header            /* Sidebar header section */
.sidebar-section           /* Section within sidebar */
.snippet-item              /* Individual snippet in list */
.snippet-item.active       /* Currently selected snippet */
.snippet-item:hover        /* Snippet hover state */
.folder-item               /* Folder in sidebar */
.tag-item                  /* Tag in sidebar */
```

### Snippets List (Main Content Area)

```css
.snippets-list             /* Container for snippets grid/list */
.snippet-card              /* Individual snippet card in main area */
.snippet-card:hover        /* Snippet card hover state */
.snippet-header            /* Snippet card header */
.snippet-content           /* Snippet card content area */
.snippet-footer            /* Snippet card footer */
```

### Editor Components

```css
.editor-container          /* Main editor container */
.editor-header             /* Editor header (title, language, etc.) */
.editor-toolbar            /* Toolbar with action buttons */
.ace_editor                /* ACE editor instance */
.ace_gutter                /* Line numbers gutter */
.ace_active-line           /* Currently active line */
.editor-tabs               /* File tabs in multi-file mode */
.editor-tab                /* Individual file tab */
```

### Preview Components

```css
.preview-container         /* Preview pane container */
.preview-content           /* Preview content area */
.markdown-preview          /* Markdown preview specific */
```

### Modal Components

```css
.snipo-modal-backdrop    /* Modal background overlay */
.snipo-modal-overlay     /* Alternative modal overlay (used in newer modals) */
.modal                     /* Modal container */
.modal-header              /* Modal header */
.modal-body                /* Modal body content */
.modal-footer              /* Modal footer with actions */
.settings-modal            /* Settings modal specific */
.settings-tabs             /* Settings tab navigation */
.settings-tab              /* Individual settings tab */

/* Search Help Modal (new) */
.search-help-dialog        /* Search help modal container */
.search-help-header        /* Search help modal header */
.search-help-body          /* Search help modal body */
.search-help-grid          /* Grid layout for help cards */
.help-card                 /* Individual help card */
.help-card-header          /* Help card header with icon */
.help-item                 /* Help item/example within card */
```

### Button Components

```css
.btn                       /* Base button */
.btn-primary               /* Primary action button */
.btn-secondary             /* Secondary action button */
.btn-danger                /* Danger/delete button */
.btn-icon                  /* Icon-only button */
```

### Form Components

```css
.editor-field              /* Form field wrapper */
.editor-field label        /* Form field label */
.editor-field input        /* Form text input */
.editor-field select       /* Form select dropdown */
.editor-field textarea     /* Form textarea */
```

## Common Customizations

### Complete App Color Change

Change the accent color throughout the entire app (both themes):

```css
:root {
  --snipo-accent: #ff6b6b;  /* Applies to both themes */
}
```

### Entire App Background

Customize the main background and all content areas:

```css
[data-theme="dark"] {
  --pico-background-color: #0a0a0a;
  --pico-card-background-color: #1a1a1a;
  --sidebar-bg: #0f0f0f;
  --editor-bg: #0a0a0a;
}

[data-theme="light"] {
  --pico-background-color: #fafafa;
  --pico-card-background-color: #ffffff;
  --sidebar-bg: #f5f5f5;
  --editor-bg: #ffffff;
}
```

### Custom Sidebar AND Main Content

Style both the sidebar and main content area:

```css
/* Sidebar */
[data-theme="dark"] .sidebar {
  background: linear-gradient(180deg, #1a1a2e 0%, #16213e 100%);
}

[data-theme="light"] .sidebar {
  background: linear-gradient(180deg, #e8eaf6 0%, #c5cae9 100%);
}

/* Main content area */
[data-theme="dark"] .main-content {
  background: #0f0f1a;
}

[data-theme="light"] .main-content {
  background: #f5f7fa;
}

/* Snippet cards in main area */
.snippet-card {
  border-radius: 12px;
  transition: all 0.2s ease;
}

.snippet-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
}
```

### Complete App Styling Example

Here's a comprehensive example that customizes the entire application (sidebar, main content, editor, and all components) with theme support:

```css
/* Global accent color */
:root {
  --snipo-primary: #ff6b6b;
  --snipo-accent: #ff8787;
}

/* Dark theme */
[data-theme="dark"] {
  --pico-background-color: #0a0a0f;
  --pico-card-background-color: #1a1a1f;
  --sidebar-bg: #0f0f14;
  --sidebar-hover: #1a1a1f;
  --editor-bg: #0f0f14;
  --pico-color: #e5e5e5;
  --sidebar-text: #d5d5d5;
}

/* Light theme */
[data-theme="light"] {
  --pico-background-color: #fafafa;
  --pico-card-background-color: #ffffff;
  --sidebar-bg: #f5f5f5;
  --sidebar-hover: #eeeeee;
  --editor-bg: #ffffff;
  --pico-color: #1a1a1a;
  --sidebar-text: #2a2a2a;
}

/* Style all major components */
.sidebar {
  border-right: 2px solid var(--snipo-accent);
}

.main-content {
  background: var(--pico-background-color);
}

.snippet-item,
.snippet-card {
  border-left: 3px solid transparent;
  transition: all 0.2s ease;
}

.snippet-item:hover,
.snippet-card:hover {
  border-left-color: var(--snipo-accent);
  background: var(--sidebar-hover);
  transform: translateX(4px);
}

.snippet-item.active,
.snippet-card.active {
  border-left-color: var(--snipo-primary);
}

.editor-container,
.preview-container {
  background: var(--editor-bg);
  border-radius: 12px;
}

.btn-primary {
  background: var(--snipo-primary);
  border-color: var(--snipo-primary);
}

.btn-primary:hover {
  background: var(--snipo-accent);
}
```

### Rounded Corners

```css
.snippet-item,
.modal,
.btn,
.editor-container {
  border-radius: 12px;
}
```

### Custom Snippet Hover Effect

```css
.snippet-item:hover {
  background: rgba(139, 92, 246, 0.1);
  border-left: 3px solid var(--snipo-accent);
  transform: translateX(4px);
  transition: all 0.2s ease;
}
```

### Custom Scrollbar

```css
::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

::-webkit-scrollbar-track {
  background: var(--snipo-bg-primary);
}

::-webkit-scrollbar-thumb {
  background: var(--snipo-accent);
  border-radius: 5px;
}

::-webkit-scrollbar-thumb:hover {
  background: var(--snipo-primary);
}
```

## Theme Examples

**Important:** These examples are designed to work in both dark and light modes by overriding variables in theme-specific contexts. Always test your custom CSS in both modes.

### Purple Dream Theme

Vibrant purple and pink color scheme for the entire app.

```css
/* Override accent colors globally */
:root {
  --snipo-primary: #8b5cf6;
  --snipo-accent: #ec4899;
}

/* Dark mode */
[data-theme="dark"] {
  --sidebar-bg: #1e1b4b;
  --sidebar-hover: #312e81;
  --editor-bg: #1e1b4b;
  --pico-background-color: #1a1625;
  --pico-card-background-color: #1e1b4b;
}

/* Light mode */
[data-theme="light"] {
  --sidebar-bg: #f3e8ff;
  --sidebar-hover: #e9d5ff;
  --editor-bg: #faf5ff;
  --pico-background-color: #fdf4ff;
  --pico-card-background-color: #faf5ff;
}

/* Apply to both themes */
.snippet-item:hover,
.snippet-card:hover {
  background: rgba(236, 72, 153, 0.15);
  border-left: 3px solid var(--snipo-accent);
  transform: translateX(4px);
  transition: all 0.2s ease;
}

.editor-header {
  background: linear-gradient(90deg, 
    rgba(139, 92, 246, 0.1) 0%, 
    rgba(236, 72, 153, 0.05) 100%
  );
}
```

### Ocean Blue Theme

Calming blue tones for the entire interface.

```css
:root {
  --snipo-primary: #0ea5e9;
  --snipo-accent: #06b6d4;
}

[data-theme="dark"] {
  --sidebar-bg: #0c1e2e;
  --sidebar-hover: #1a3a52;
  --sidebar-border: #2d5f7f;
  --editor-bg: #0c1e2e;
  --pico-background-color: #071420;
  --pico-card-background-color: #0c1e2e;
}

[data-theme="light"] {
  --sidebar-bg: #e0f2fe;
  --sidebar-hover: #bae6fd;
  --sidebar-border: #7dd3fc;
  --editor-bg: #f0f9ff;
  --pico-background-color: #f8fcff;
  --pico-card-background-color: #f0f9ff;
}

.sidebar {
  border-right: 2px solid var(--sidebar-border);
}

.snippet-item.active,
.snippet-card.active {
  background: rgba(6, 182, 212, 0.15);
  border-left: 3px solid var(--snipo-accent);
}

.main-content {
  background: var(--pico-background-color);
}
```

### Forest Green Theme

Natural green palette for the entire app.

```css
:root {
  --snipo-primary: #10b981;
  --snipo-accent: #34d399;
}

[data-theme="dark"] {
  --sidebar-bg: #0f1f14;
  --sidebar-hover: #1a2f20;
  --editor-bg: #0f1f14;
  --pico-background-color: #0a140d;
  --pico-card-background-color: #0f1f14;
}

[data-theme="light"] {
  --sidebar-bg: #dcfce7;
  --sidebar-hover: #bbf7d0;
  --editor-bg: #f0fdf4;
  --pico-background-color: #f7fef9;
  --pico-card-background-color: #f0fdf4;
}

.editor-header,
.snippet-header {
  background: linear-gradient(90deg, 
    rgba(16, 185, 129, 0.1) 0%, 
    rgba(52, 211, 153, 0.05) 100%
  );
}

.snippet-card {
  border: 1px solid rgba(16, 185, 129, 0.2);
}
```

### Sunset Orange Theme

Warm orange and amber for the entire interface.

```css
:root {
  --snipo-primary: #f97316;
  --snipo-accent: #fb923c;
}

[data-theme="dark"] {
  --sidebar-bg: #1f1410;
  --sidebar-hover: #2d1f1a;
  --editor-bg: #1f1410;
  --pico-background-color: #160d08;
  --pico-card-background-color: #1f1410;
  --pico-color: #fff7ed;
}

[data-theme="light"] {
  --sidebar-bg: #ffedd5;
  --sidebar-hover: #fed7aa;
  --editor-bg: #fff7ed;
  --pico-background-color: #fffbf5;
  --pico-card-background-color: #fff7ed;
}

.snippet-item:hover,
.snippet-card:hover {
  background: rgba(251, 146, 60, 0.1);
  box-shadow: 0 4px 12px rgba(249, 115, 22, 0.15);
}

.main-content,
.editor-container {
  background: var(--pico-background-color);
}
```

### Cyberpunk Neon Theme

Futuristic neon glow effects - dark mode optimized.

```css
:root {
  --snipo-primary: #ff00ff;
  --snipo-accent: #00ffff;
}

/* Optimized for dark mode */
[data-theme="dark"] {
  --sidebar-bg: #0a0a0f;
  --sidebar-hover: #1a1a2e;
  --editor-bg: #0f0f14;
  --pico-background-color: #050508;
  --pico-card-background-color: #0a0a0f;
}

/* Light mode fallback */
[data-theme="light"] {
  --sidebar-bg: #1a1a2e;
  --sidebar-hover: #2a2a3e;
  --editor-bg: #1a1a2e;
  --pico-background-color: #0f0f1e;
  --pico-card-background-color: #1a1a2e;
  --pico-color: #e0e0ff;
}

.btn-primary {
  background: var(--snipo-accent);
  color: #000;
  box-shadow: 0 0 10px var(--snipo-accent),
              0 0 20px var(--snipo-accent);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.snippet-item.active,
.snippet-card {
  border-left: 3px solid var(--snipo-accent);
  box-shadow: 0 0 15px rgba(0, 255, 255, 0.3);
}

.editor-header {
  border-bottom: 2px solid var(--snipo-accent);
  box-shadow: 0 2px 10px rgba(0, 255, 255, 0.2);
}

.main-content {
  background: var(--pico-background-color);
}
```

### Minimalist Monochrome

Clean black and white - works in both modes with automatic inversion.

```css
[data-theme="dark"] {
  --sidebar-bg: #000000;
  --sidebar-hover: #1a1a1a;
  --sidebar-border: #333333;
  --editor-bg: #0a0a0a;
  --pico-background-color: #000000;
  --pico-card-background-color: #0a0a0a;
  --pico-color: #ffffff;
  --snipo-accent: #ffffff;
}

[data-theme="light"] {
  --sidebar-bg: #f5f5f5;
  --sidebar-hover: #e5e5e5;
  --sidebar-border: #cccccc;
  --editor-bg: #ffffff;
  --pico-background-color: #fafafa;
  --pico-card-background-color: #ffffff;
  --pico-color: #1a1a1a;
  --snipo-accent: #000000;
}

.sidebar {
  border-right: 1px solid var(--sidebar-border);
}

.snippet-item,
.snippet-card {
  border-left: 2px solid transparent;
}

.snippet-item:hover,
.snippet-card:hover {
  border-left-color: var(--snipo-accent);
}

.snippet-item.active,
.snippet-card.active {
  border-left-color: var(--snipo-accent);
}

.modal {
  border: 1px solid var(--sidebar-border);
}
```

### Retro Terminal Theme

Classic green-on-black terminal look - dark mode only.

```css
[data-theme="dark"] {
  --snipo-primary: #00ff00;
  --snipo-accent: #33ff33;
  --sidebar-bg: #000000;
  --sidebar-hover: rgba(0, 255, 0, 0.05);
  --sidebar-border: #003300;
  --editor-bg: #000000;
  --pico-background-color: #000000;
  --pico-card-background-color: #001100;
  --pico-color: #00ff00;
  --icon-color: #00aa00;
}

[data-theme="light"] {
  --snipo-primary: #00ff00;
  --snipo-accent: #33ff33;
  --sidebar-bg: #000000;
  --sidebar-hover: rgba(0, 255, 0, 0.05);
  --sidebar-border: #003300;
  --editor-bg: #000000;
  --pico-background-color: #000000;
  --pico-card-background-color: #001100;
  --pico-color: #00ff00;
}

* {
  font-family: 'Courier New', monospace !important;
}

.sidebar {
  border-right: 1px solid #003300;
}

.snippet-item,
.snippet-card {
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.5);
}

.snippet-item:hover,
.snippet-card:hover {
  background: rgba(0, 255, 0, 0.05);
  box-shadow: 0 0 10px rgba(0, 255, 0, 0.3);
}

.btn-primary {
  background: #001100;
  color: #00ff00;
  border: 1px solid #00ff00;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.7);
}

.editor-container,
.main-content {
  background: #000000;
  border: 1px solid #003300;
}
```

### Minimal Professional

Clean and simple for professional use - optimized for both modes.

```css
:root {
  --snipo-accent: #2563eb;
}

[data-theme="dark"] {
  --sidebar-bg: #1a1a1a;
  --sidebar-hover: #2a2a2a;
  --editor-bg: #1a1a1a;
  --pico-background-color: #0a0a0a;
  --pico-card-background-color: #1a1a1a;
}

[data-theme="light"] {
  --sidebar-bg: #ffffff;
  --sidebar-hover: #f3f4f6;
  --sidebar-border: #e5e7eb;
  --editor-bg: #ffffff;
  --pico-background-color: #fafafa;
  --pico-card-background-color: #ffffff;
}

.sidebar {
  border-right: 1px solid var(--sidebar-border);
}

.snippet-item,
.snippet-card {
  border-radius: 8px;
  margin-bottom: 4px;
}

.snippet-item:hover,
.snippet-card:hover {
  background: var(--sidebar-hover);
}

.main-content {
  background: var(--pico-background-color);
}
```

### Compact Interface

Reduce spacing for more content on screen.

```css
.sidebar {
  width: 240px;
  font-size: 0.875rem;
}

.snippet-item {
  padding: 0.5rem 0.75rem;
  min-height: auto;
}

.snippet-title {
  font-size: 0.875rem;
}

.snippet-card {
  padding: 0.75rem;
}

.editor-header {
  padding: 0.5rem 1rem;
}

.modal {
  max-width: 700px;
}

.editor-field {
  margin-bottom: 0.75rem;
}
```

### Rounded Everything

Increase border radius for softer look across the entire app.

```css
.snippet-item,
.snippet-card,
.modal,
.btn,
.editor-container,
.ace_editor,
.preview-container,
input,
select,
textarea {
  border-radius: 16px !important;
}

.btn-icon {
  border-radius: 50% !important;
}

.tag-badge {
  border-radius: 9999px !important;
}

.main-content {
  border-radius: 16px;
}
```

### High Contrast Mode

Improved visibility for accessibility.

```css
:root {
  --snipo-bg-primary: #000000;
  --snipo-bg-secondary: #1a1a1a;
  --snipo-text-primary: #ffffff;
  --snipo-text-secondary: #e5e5e5;
  --snipo-border: #ffffff;
  --snipo-accent: #ffff00;
}

.snippet-item {
  border: 2px solid var(--snipo-border);
}

.snippet-item:hover {
  background: #1a1a1a;
  border-color: var(--snipo-accent);
}

.btn-primary {
  background: #ffffff;
  color: #000000;
  border: 3px solid #ffffff;
  font-weight: 700;
}
```

## Advanced Techniques

### Glassmorphism Effect

Create a frosted glass effect for modals.

```css
.modal {
  background: rgba(30, 41, 59, 0.7);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
}

.sidebar {
  background: rgba(15, 23, 42, 0.8);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-right: 1px solid rgba(255, 255, 255, 0.05);
}

.snippet-item:hover {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(5px);
}
```

### Animated Gradient Background

Add a subtle animated gradient.

```css
body {
  background: linear-gradient(45deg, 
    #1a1a2e, 
    #16213e, 
    #0f3460, 
    #16213e, 
    #1a1a2e
  );
  background-size: 400% 400%;
  animation: gradientFlow 20s ease infinite;
}

@keyframes gradientFlow {
  0%, 100% { background-position: 0% 50%; }
  50% { background-position: 100% 50%; }
}
```

### Card-Style Snippets

Make snippets look like cards with shadows.

```css
.snippet-item {
  background: var(--snipo-bg-secondary);
  border-radius: 12px;
  margin: 8px 12px;
  padding: 1rem;
  border: 1px solid var(--snipo-border);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.snippet-item:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  border-color: var(--snipo-accent);
}

.snippet-item.active {
  border-color: var(--snipo-primary);
  box-shadow: 0 8px 16px rgba(99, 102, 241, 0.3);
}
```

### Gradient Sidebar

Colorful gradient background for sidebar.

```css
.sidebar {
  background: linear-gradient(180deg, 
    #667eea 0%, 
    #764ba2 50%, 
    #f093fb 100%
  );
}

.sidebar-header {
  background: rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(5px);
}

.snippet-item {
  background: rgba(255, 255, 255, 0.1);
  margin: 4px 8px;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.snippet-item:hover {
  background: rgba(255, 255, 255, 0.2);
  transform: translateX(4px);
}

.snippet-item.active {
  background: rgba(255, 255, 255, 0.25);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
```

### Neon Glow Effect

Add a cyberpunk-style glow.

```css
.btn-primary {
  background: var(--snipo-accent);
  box-shadow: 0 0 10px var(--snipo-accent),
              0 0 20px var(--snipo-accent),
              0 0 30px var(--snipo-accent);
  transition: all 0.3s ease;
}

.btn-primary:hover {
  box-shadow: 0 0 20px var(--snipo-accent),
              0 0 40px var(--snipo-accent),
              0 0 60px var(--snipo-accent);
}
```

### Theme-Specific Customization

Customize colors for specific themes.

```css
/* Dark theme only */
[data-theme="dark"] {
  --snipo-bg-primary: #000000;
  --snipo-accent: #00ff88;
}

/* Light theme only */
[data-theme="light"] {
  --snipo-bg-primary: #ffffff;
  --snipo-text-primary: #1a1a1a;
  --snipo-accent: #ff3366;
}
```

### Rounded Everything

Increase border radius for softer look.

```css
.snippet-item,
.modal,
.btn,
.editor-container,
.ace_editor,
input,
select,
textarea {
  border-radius: 16px !important;
}

.btn-icon {
  border-radius: 50% !important;
}

.tag-badge {
  border-radius: 9999px !important;
}
```

### Customize Search Help Modal

Style the new search help dialog with custom colors and effects.

```css
/* Larger, more prominent search help modal */
.search-help-dialog {
  max-width: 1000px;
  border: 2px solid var(--snipo-primary);
}

/* Custom header styling */
.search-help-header {
  background: linear-gradient(135deg, var(--snipo-primary), var(--snipo-accent));
  color: white;
}

.search-help-title svg,
.search-help-title h2 {
  color: white;
}

/* Help card hover effects */
.help-card {
  transition: all 0.3s ease;
}

.help-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 16px rgba(0, 0, 0, 0.2);
  border-color: var(--snipo-accent);
}

/* Custom code example styling */
.help-item {
  background: linear-gradient(135deg, 
    var(--pico-code-background-color), 
    rgba(var(--snipo-primary-rgb), 0.05));
}

.help-item code {
  font-weight: 600;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

/* Keyboard shortcuts styling */
.help-item kbd {
  background: var(--snipo-primary);
  color: white;
  border-color: var(--snipo-accent);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}
```

## Testing Your Custom CSS

### ⚠️ Critical: Test in BOTH Dark and Light Modes

**Always verify your custom CSS in both themes:**

1. Apply your custom CSS and save
2. **Test in Dark Mode**: Settings → Appearance → Theme → Dark
   - Check sidebar styling
   - Check main content area
   - Check editor and preview panels
   - Check snippet cards and lists
3. **Test in Light Mode**: Switch to Light theme
   - Verify all areas still look good
   - Ensure text is readable
   - Check for color contrast issues

### Quick Test

1. Open Settings → Appearance
2. Paste this test CSS:
   ```css
   /* This should work in both themes */
   [data-theme="dark"] {
     --snipo-accent: #ff6b6b;
   }
   
   [data-theme="light"] {
     --snipo-accent: #d63031;
   }
   ```
3. Click "Save and Apply Custom CSS"
4. Toggle between dark/light themes
5. Verify accent color changes appropriately in both modes

### Complete Application Test

Ensure your CSS styles the **entire application**, not just the sidebar:

**Areas to verify:**
- **Sidebar** - Navigation, folders, tags
- **Main Content Area** - Background, overall layout
- **Snippets List** - Snippet cards, hover states
- **Editor Panel** - Editor background, text, line numbers
- **Preview Panel** - Preview background, rendered content
- **Header** - Top navigation, buttons
- **Modals** - Settings, create/edit dialogs

**Test CSS for whole app:**
```css
[data-theme="dark"] {
  --pico-background-color: #1a1a2e;  /* Main app background */
  --sidebar-bg: #16213e;              /* Sidebar */
  --editor-bg: #0f3460;               /* Editor area */
}

[data-theme="light"] {
  --pico-background-color: #f8f9fa;  /* Main app background */
  --sidebar-bg: #e9ecef;              /* Sidebar */
  --editor-bg: #ffffff;               /* Editor area */
}
```

Apply this and verify **all areas** change color in both themes.

### Security Validation Tests

**Test 1: Script tag (should be blocked)**
```css
<script>alert('test')</script>
.sidebar { background: red; }
```
Expected: Error message, CSS not saved

**Test 2: External import (should warn)**
```css
@import url('https://example.com/style.css');
```
Expected: Warning dialog with security concern

**Test 3: Unmatched braces (should warn)**
```css
.sidebar {
  background: red;
/* Missing closing brace */
```
Expected: Warning about unmatched braces

### Browser DevTools Testing

1. Press F12 to open DevTools
2. Go to Elements/Inspector tab
3. Look for `<style id="snipo-custom-css">` in `<head>`
4. Verify your CSS is inside
5. Inspect elements to see your styles applied

### Persistence Test

1. Apply custom CSS and save
2. Refresh the page (F5)
3. Verify CSS still applied
4. Log out and back in
5. Verify CSS persists

## Best Practices

### 1. Always Support Both Themes

**Don't do this** (only works in one theme):
```css
.sidebar {
  background: #000000;  /* Hard to see in light mode */
  color: #ffffff;       /* Invisible text in light mode */
}
```

**Do this instead** (works in both themes):
```css
[data-theme="dark"] .sidebar {
  background: #000000;
  color: #ffffff;
}

[data-theme="light"] .sidebar {
  background: #f5f5f5;
  color: #1a1a1a;
}
```

Or better yet, use CSS variables:
```css
.sidebar {
  background: var(--sidebar-bg);  /* Already theme-aware */
  color: var(--sidebar-text);     /* Already theme-aware */
}
```

### 2. Test in Both Modes

After applying custom CSS:
1. Click your user icon → Toggle theme
2. Verify all your customizations look good in both modes
3. Check text is readable in both themes
4. Ensure hover states work in both themes

### 3. Use CSS Variables

Prefer overriding CSS variables instead of specific classes.

```css
/* Good - Uses variables, works in both themes */
:root {
  --snipo-accent: #ff6b6b;
}

/* Less ideal - Hardcodes values, may break themes */
.btn-primary {
  background: #ff6b6b;
}
```

### 2. Avoid !important

Use `!important` sparingly. Instead, increase specificity.

```css
/* Avoid */
.snippet-item {
  background: red !important;
}

/* Better */
.sidebar .snippet-item {
  background: red;
}
```

### 3. Test Both Themes

If you support both light and dark themes, test in both.

```css
[data-theme="dark"] .custom-element {
  /* Dark theme styles */
}

[data-theme="light"] .custom-element {
  /* Light theme styles */
}
```

### 4. Use Browser DevTools

Press F12 to open DevTools and inspect elements to find the right classes and current styles.

### 5. Keep It Organized

Add comments to organize your CSS.

```css
/* ============================================
   SIDEBAR CUSTOMIZATION
   ============================================ */

.sidebar {
  /* styles */
}

/* ============================================
   EDITOR CUSTOMIZATION
   ============================================ */

.editor-container {
  /* styles */
}
```

### 6. Performance Considerations

- Avoid overly complex selectors
- Limit expensive properties like `box-shadow` and `filter`
- Use `transform` and `opacity` for animations (GPU-accelerated)
- Keep CSS under 50KB when possible

### 7. Validate Your CSS

- Ensure braces are balanced `{ }`
- Check for typos in property names
- Test before deploying

## Troubleshooting

### CSS Not Applying

1. **Check for syntax errors**: Ensure braces are balanced and properties are valid
2. **Increase specificity**: Your selector might not be specific enough
3. **Clear cache**: Try Ctrl+Shift+R (Cmd+Shift+R on Mac)
4. **Check DevTools**: Look for errors in browser console (F12)

### Styles Look Different Than Expected

1. **Inspect the element**: Use DevTools to see which styles are applied
2. **Check theme**: Some styles may be theme-specific
3. **CSS cascade**: Later rules override earlier ones

### Custom CSS Resets on Page Reload

Make sure you clicked "Save and Apply Custom CSS" button. Custom CSS is stored in the database and should persist.

### Performance Issues

If the app becomes slow after adding custom CSS:

1. **Reduce complexity**: Simplify selectors and remove expensive properties
2. **Limit animations**: Too many animations can impact performance
3. **Check CSS size**: Very large CSS files (>50KB) may cause issues

### Validation Warnings

- **Unmatched braces**: Count your `{` and `}` - they should be equal
- **External imports**: External `@import` URLs are flagged as security risks
- **Large CSS**: Files over 50KB may impact performance

## Additional Resources

- **CSS Variables Reference**: [MDN CSS Custom Properties](https://developer.mozilla.org/en-US/docs/Web/CSS/Using_CSS_custom_properties)
- **Flexbox Guide**: [CSS-Tricks Flexbox Guide](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)
- **CSS Grid Guide**: [CSS-Tricks Grid Guide](https://css-tricks.com/snippets/css/complete-guide-grid/)
- **Color Palette Generators**: [Coolors.co](https://coolors.co/), [ColorHunt](https://colorhunt.co/)

## Support

If you encounter issues or have questions:

1. Check this documentation first
2. Use browser DevTools to inspect elements
3. Open an issue on GitHub with:
   - Your custom CSS
   - Screenshots of the issue
   - Browser and version information

