package ui

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/MohamedElashri/snipo/tui/internal/api"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const templateDraftFormRows = 21

type templateWorkflow struct {
	items       []api.TemplateSummary
	selected    int
	search      textinput.Model
	record      *api.Template
	presetIndex int
	field       int
	value       textarea.Model
	values      map[string]string
	draft       *api.SnippetInput
	file        int
	folder      int
	loading     bool
	generation  int
	unsupported bool
}

type templatesLoaded struct {
	generation int
	items      []api.TemplateSummary
	err        error
}

type templateOpened struct {
	generation int
	record     *api.Template
	err        error
}

type templateRendered struct {
	generation int
	draft      *api.SnippetInput
	err        error
}

type templateSaved struct {
	title string
	err   error
}

func (m Model) openTemplatePicker() (tea.Model, tea.Cmd) {
	if m.templates.unsupported {
		m.err = api.ErrTemplatesUnavailable
		return m, nil
	}
	m.mode, m.err, m.message = ViewTemplatePicker, nil, ""
	generation := m.templates.generation + 1
	m.templates = templateWorkflow{loading: true, generation: generation, search: textinput.New()}
	m.templates.search.Placeholder = "Search name, purpose, or filename"
	m.templates.search.Focus()
	return m, func() tea.Msg {
		items, err := m.client.ListTemplates()
		return templatesLoaded{generation, items, err}
	}
}

func (m Model) updateTemplateMessage(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case templatesLoaded:
		if m.mode != ViewTemplatePicker || m.templates.generation != msg.generation {
			return m, nil, true
		}
		m.templates.items, m.templates.loading, m.err = msg.items, false, msg.err
		m.templates.unsupported = errors.Is(msg.err, api.ErrTemplatesUnavailable)
		if m.templates.unsupported {
			m.mode = ViewList
		}
	case templateOpened:
		if m.mode != ViewTemplatePicker || m.templates.generation != msg.generation {
			return m, nil, true
		}
		m.templates.record, m.templates.loading, m.err = msg.record, false, msg.err
		if msg.err != nil {
			return m, nil, true
		}
		if len(msg.record.Definition.Presets) > 0 {
			m.mode = ViewTemplatePreset
			return m, nil, true
		}
		cmd := m.startTemplateFields(-1)
		return m, cmd, true
	case templateRendered:
		if m.mode != ViewTemplateFields || m.templates.generation != msg.generation {
			return m, nil, true
		}
		m.templates.loading, m.err = false, msg.err
		if msg.err == nil {
			m.openTemplateDraft(msg.draft)
		}
	case templateSaved:
		m.templates.loading, m.err = false, msg.err
		if msg.err == nil {
			m.templates.draft = nil
			m.mode = ViewList
			m.message = "Created snippet: " + msg.title
			return m, loadSnippets(m.client, m.currentPage, 20, m.searchQuery, m.filterTags, nil, "", nil, nil), true
		}
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m Model) filteredTemplates() []api.TemplateSummary {
	query := strings.ToLower(strings.TrimSpace(m.templates.search.Value()))
	var result []api.TemplateSummary
	for _, item := range m.templates.items {
		text := item.Name + " " + item.Purpose + " " + strings.Join(item.Filenames, " ")
		if strings.Contains(strings.ToLower(text), query) {
			result = append(result, item)
		}
	}
	return result
}

func (m Model) updateTemplateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode, m.err = ViewList, nil
		m.templates.generation++
		return m, nil
	}
	if m.templates.loading {
		return m, nil
	}

	switch m.mode {
	case ViewTemplatePicker:
		items := m.filteredTemplates()
		switch msg.String() {
		case "up":
			m.templates.selected = max(0, m.templates.selected-1)
		case "down":
			m.templates.selected = max(0, min(len(items)-1, m.templates.selected+1))
		case "enter":
			if len(items) != 0 {
				item := items[m.templates.selected]
				m.templates.loading, m.err = true, nil
				return m, func() tea.Msg {
					record, err := m.client.GetTemplate(item.ID, item.Revision)
					return templateOpened{m.templates.generation, record, err}
				}
			}
		default:
			var cmd tea.Cmd
			m.templates.search, cmd = m.templates.search.Update(msg)
			m.templates.selected = 0
			return m, cmd
		}
	case ViewTemplatePreset:
		switch msg.String() {
		case "up":
			m.templates.presetIndex = max(0, m.templates.presetIndex-1)
		case "down":
			m.templates.presetIndex = min(len(m.templates.record.Definition.Presets), m.templates.presetIndex+1)
		case "enter":
			cmd := m.startTemplateFields(m.templates.presetIndex - 1)
			return m, cmd
		}
	case ViewTemplateFields:
		return m.updateTemplateField(msg)
	}
	return m, nil
}

func (m *Model) startTemplateFields(preset int) tea.Cmd {
	m.mode = ViewTemplateFields
	m.templates.values = make(map[string]string)
	def := m.templates.record.Definition
	for _, field := range def.Fields {
		if field.Default != nil {
			m.templates.values[field.Key] = *field.Default
		}
		if preset >= 0 {
			if value, ok := def.Presets[preset].Values[field.Key]; ok {
				m.templates.values[field.Key] = value
			}
		}
	}
	if len(def.Fields) == 0 {
		return m.renderTemplate()
	}
	return m.focusTemplateField(0)
}

func (m *Model) focusTemplateField(index int) tea.Cmd {
	m.templates.field = index
	field := m.templates.record.Definition.Fields[index]
	m.templates.value = textarea.New()
	m.templates.value.SetWidth(max(18, m.width-10))
	m.templates.value.SetHeight(1)
	if field.Type == "multiline" {
		m.templates.value.SetHeight(6)
	}
	m.templates.value.ShowLineNumbers = false
	m.templates.value.CharLimit = 0
	m.templates.value.MaxHeight = 0
	m.templates.value.SetValue(m.templates.values[field.Key])
	return m.templates.value.Focus()
}

func (m *Model) renderTemplate() tea.Cmd {
	m.templates.loading, m.err = true, nil
	id, revision, values := m.templates.record.ID, m.templates.record.Revision, m.templates.values
	generation := m.templates.generation
	return func() tea.Msg {
		draft, err := m.client.RenderTemplate(id, revision, values)
		return templateRendered{generation, draft, err}
	}
}

func (m Model) updateTemplateField(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	fields := m.templates.record.Definition.Fields
	if len(fields) == 0 {
		if msg.String() == "enter" || msg.String() == "ctrl+r" {
			cmd := m.renderTemplate()
			return m, cmd
		}
		return m, nil
	}
	field := fields[m.templates.field]
	key := msg.String()
	if key == "tab" || key == "shift+tab" || key == "ctrl+r" || (key == "enter" && field.Type != "multiline") {
		m.templates.values[field.Key] = m.templates.value.Value()
		if key == "ctrl+r" || (key == "enter" && m.templates.field == len(fields)-1) {
			cmd := m.renderTemplate()
			return m, cmd
		}
		index := (m.templates.field + 1) % len(fields)
		if key == "shift+tab" {
			index = (m.templates.field + len(fields) - 1) % len(fields)
		}
		cmd := m.focusTemplateField(index)
		return m, cmd
	}
	if field.Type == "choice" || field.Type == "toggle" {
		if key == "up" || key == "down" {
			choices := field.Choices
			if field.Type == "toggle" {
				choices = []string{"true", "false"}
			}
			if !field.Required {
				choices = append([]string{""}, choices...)
			}
			index := slices.Index(choices, m.templates.value.Value())
			if key == "down" {
				index++
			} else {
				index--
			}
			if len(choices) > 0 {
				m.templates.value.SetValue(choices[(index+2*len(choices))%len(choices)])
			}
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.templates.value, cmd = m.templates.value.Update(msg)
	return m, cmd
}

func (m *Model) openTemplateDraft(draft *api.SnippetInput) {
	m.mode = ViewCreate
	m.templates.draft, m.templates.file, m.templates.folder = draft, 0, 0
	m.initCreateForm()
	m.textarea.CharLimit = 0
	m.inputs[0].SetValue(draft.Title)
	m.inputs[2].SetValue(strings.Join(draft.Tags, ", "))
	m.inputs[3].SetValue(draft.Description)
	filename, folder := textinput.New(), textinput.New()
	filename.Prompt, filename.CharLimit = "Filename:    ", 255
	folder.Prompt = "Folder:      "
	m.inputs = append(m.inputs, filename, folder)
	m.showTemplateFolder()
	m.textarea.SetHeight(max(3, m.height-templateDraftFormRows))
	m.showTemplateDraftFile(0)
}

func (m *Model) showTemplateFolder() {
	m.templates.folder = 0
	m.inputs[5].SetValue("No folder")
	if m.templates.draft.FolderID == nil {
		return
	}
	for i, item := range m.folders {
		if item.ID == *m.templates.draft.FolderID {
			m.templates.folder = i + 1
			m.inputs[5].SetValue(item.Name)
			return
		}
	}
	m.inputs[5].SetValue(fmt.Sprintf("Folder #%d", *m.templates.draft.FolderID))
}

func (m *Model) showTemplateDraftFile(index int) {
	m.templates.file = index
	file := m.templates.draft.Files[index]
	m.inputs[1].SetValue(file.Language)
	m.inputs[4].SetValue(file.Filename)
	m.textarea.SetValue(file.Content)
}

func (m *Model) stashTemplateDraftFile() {
	file := &m.templates.draft.Files[m.templates.file]
	file.Filename = m.inputs[4].Value()
	file.Language = strings.ToLower(strings.TrimSpace(m.inputs[1].Value()))
	file.Content = m.textarea.Value()
}

func (m *Model) setTemplateFolder() {
	if m.templates.folder == 0 {
		m.templates.draft.FolderID = nil
		m.inputs[5].SetValue("No folder")
		return
	}
	folder := m.folders[m.templates.folder-1]
	m.templates.draft.FolderID = &folder.ID
	m.inputs[5].SetValue(folder.Name)
}

func (m Model) updateTemplateDraftKey(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	if m.mode != ViewCreate || m.templates.draft == nil {
		return m, nil, false
	}
	if m.templates.loading {
		return m, nil, true
	}
	switch msg.String() {
	case "esc":
		m.templates.draft = nil
		return m, nil, false
	case "f2", "f3":
		m.stashTemplateDraftFile()
		step := 1
		if msg.String() == "f2" {
			step = -1
		}
		count := len(m.templates.draft.Files)
		m.showTemplateDraftFile((m.templates.file + step + count) % count)
		return m, nil, true
	case "f4":
		return m, copyToClipboard(m.textarea.Value()), true
	}
	if m.focusedInput == 5 {
		switch msg.String() {
		case "up", "down":
			step := 1
			if msg.String() == "up" {
				step = -1
			}
			count := len(m.folders) + 1
			m.templates.folder = (m.templates.folder + step + count) % count
			m.setTemplateFolder()
			return m, nil, true
		case "tab", "shift+tab":
			return m, nil, false
		case "ctrl+s":
			return m, nil, false
		default:
			return m, nil, true
		}
	}
	return m, nil, false
}

func (m Model) submitTemplateDraft(input api.SnippetInput) (tea.Model, tea.Cmd) {
	m.stashTemplateDraftFile()
	draft := m.templates.draft
	input.Files = slices.Clone(draft.Files)
	input.Content = ""
	input.Language = input.Files[0].Language
	input.TemplateOrigin, input.FolderID = draft.TemplateOrigin, draft.FolderID
	m.templates.loading = true
	return m, func() tea.Msg {
		result, err := m.client.CreateSnippet(input)
		if err != nil {
			return templateSaved{err: err}
		}
		return templateSaved{title: result.Title}
	}
}

func (m Model) viewTemplates() string {
	var out strings.Builder
	switch m.mode {
	case ViewTemplatePicker:
		out.WriteString(headerStyle.Render("New from template") + "\n\n")
		out.WriteString(m.templates.search.View() + "\n\n")
		items := m.filteredTemplates()
		visible := max(2, (m.height-9)/3)
		start := max(0, m.templates.selected-visible+1)
		for i := start; i < min(len(items), start+visible); i++ {
			item := items[i]
			marker := "  "
			if i == m.templates.selected {
				marker = "> "
			}
			if item.Favorite {
				marker += "★ "
			}
			out.WriteString(marker + terminalText(item.Name) + "\n")
			if item.Purpose != "" {
				out.WriteString("  " + dimmedStyle.Render(terminalText(item.Purpose)) + "\n")
			}
			out.WriteString("  " + dimmedStyle.Render(terminalText(strings.Join(item.Filenames, ", "))) + "\n")
		}
		if !m.templates.loading && len(items) == 0 {
			out.WriteString("No matching templates. Manage templates in the Snipo web library.\n")
		}
		out.WriteString("\n↑/↓ Select · Enter Continue · Esc Cancel")
	case ViewTemplatePreset:
		out.WriteString(headerStyle.Render("Choose a preset") + "\n\n")
		names := []string{"Template defaults"}
		for _, preset := range m.templates.record.Definition.Presets {
			names = append(names, preset.Name)
		}
		for i, name := range names {
			marker := "  "
			if i == m.templates.presetIndex {
				marker = "> "
			}
			out.WriteString(marker + terminalText(name) + "\n")
		}
		out.WriteString("\n↑/↓ Select · Enter Continue · Esc Cancel")
	case ViewTemplateFields:
		out.WriteString(headerStyle.Render(terminalText(m.templates.record.Definition.Name)) + "\n\n")
		fields := m.templates.record.Definition.Fields
		if len(fields) > 0 {
			field := fields[m.templates.field]
			label := terminalText(field.Label)
			if field.Required {
				label += " *"
			}
			fmt.Fprintf(&out, "%s (%d/%d)\n", label, m.templates.field+1, len(fields))
			if field.Help != "" {
				out.WriteString(terminalText(field.Help) + "\n")
			}
			out.WriteString(m.templates.value.View() + "\n\n")
			if field.Type == "choice" || field.Type == "toggle" {
				out.WriteString("↑/↓ Choose · ")
			}
			out.WriteString("Tab/Shift+Tab Fields · Ctrl+R Preview · Esc Cancel")
		} else if !m.templates.loading {
			out.WriteString("Enter Retry · Esc Cancel")
		}
	}
	if m.templates.loading {
		out.WriteString("\n\nLoading…")
	}
	return out.String()
}

func (m Model) viewTemplateDraftHeader() string {
	file := m.templates.draft.Files[m.templates.file]
	return fmt.Sprintf("Template draft · File %d/%d · %s\nF2 Previous file · F3 Next file · F4 Copy · ↑/↓ Choose folder\n\n",
		m.templates.file+1, len(m.templates.draft.Files), terminalText(file.Filename))
}
