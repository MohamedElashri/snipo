package ui

import (
	"fmt"
	"strings"

	"github.com/MohamedElashri/snipo/tui/internal/api"
	"github.com/MohamedElashri/snipo/tui/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewMode int

const (
	ViewList ViewMode = iota
	ViewDetail
	ViewCreate
	ViewEdit
	ViewSearch
	ViewSettings
	ViewHelp
)

type Model struct {
	client  *api.Client
	config  *config.Config
	mode    ViewMode
	width   int
	height  int
	err     error
	message string

	snippets     []api.Snippet
	selectedIdx  int
	currentPage  int
	totalPages   int
	searchQuery  string
	filterTags   []int
	showFavorite bool

	detailSnippet   *api.Snippet
	detailScroll    int
	selectedFileIdx int

	tags    []api.Tag
	folders []api.Folder

	inputs       []textinput.Model
	focusedInput int
	formData     map[string]interface{}

	quitting bool
}

type errMsg struct{ err error }
type successMsg struct{ message string }
type snippetsLoadedMsg struct {
	snippets   []api.Snippet
	pagination *api.Pagination
}
type snippetLoadedMsg struct{ snippet *api.Snippet }
type tagsLoadedMsg struct{ tags []api.Tag }
type foldersLoadedMsg struct{ folders []api.Folder }

func (e errMsg) Error() string { return e.err.Error() }

func NewModel(cfg *config.Config) Model {
	client := api.NewClient(cfg.ServerURL, cfg.APIKey)

	return Model{
		client:      client,
		config:      cfg,
		mode:        ViewList,
		snippets:    []api.Snippet{},
		currentPage: 1,
		formData:    make(map[string]interface{}),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadSnippets(m.client, 1, 20, "", nil, nil, "", nil, nil),
		loadTags(m.client),
		loadFolders(m.client),
	)
}

func loadSnippets(client *api.Client, page, limit int, query string, tagIDs, folderIDs []int, language string, favorite, archived *bool) tea.Cmd {
	return func() tea.Msg {
		snippets, pagination, err := client.ListSnippets(page, limit, query, tagIDs, folderIDs, language, favorite, archived)
		if err != nil {
			return errMsg{err}
		}
		return snippetsLoadedMsg{snippets: snippets, pagination: pagination}
	}
}

func loadSnippet(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		snippet, err := client.GetSnippet(id)
		if err != nil {
			return errMsg{err}
		}
		return snippetLoadedMsg{snippet: snippet}
	}
}

func loadTags(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		tags, err := client.ListTags()
		if err != nil {
			return errMsg{err}
		}
		return tagsLoadedMsg{tags: tags}
	}
}

func loadFolders(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		folders, err := client.ListFolders()
		if err != nil {
			return errMsg{err}
		}
		return foldersLoadedMsg{folders: folders}
	}
}

func createSnippet(client *api.Client, input api.SnippetInput) tea.Cmd {
	return func() tea.Msg {
		snippet, err := client.CreateSnippet(input)
		if err != nil {
			return errMsg{err}
		}
		return successMsg{message: fmt.Sprintf("Created snippet: %s", snippet.Title)}
	}
}

func updateSnippet(client *api.Client, id string, input api.SnippetInput) tea.Cmd {
	return func() tea.Msg {
		snippet, err := client.UpdateSnippet(id, input)
		if err != nil {
			return errMsg{err}
		}
		return successMsg{message: fmt.Sprintf("Updated snippet: %s", snippet.Title)}
	}
}

func deleteSnippet(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.DeleteSnippet(id)
		if err != nil {
			return errMsg{err}
		}
		return successMsg{message: "Snippet deleted successfully"}
	}
}

func toggleFavorite(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		snippet, err := client.ToggleFavorite(id)
		if err != nil {
			return errMsg{err}
		}
		return snippetLoadedMsg{snippet: snippet}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.mode == ViewList || m.mode == ViewHelp {
				m.quitting = true
				return m, tea.Quit
			}
			m.mode = ViewList
			m.err = nil
			m.message = ""
			return m, nil

		case "?":
			if m.mode != ViewHelp {
				m.mode = ViewHelp
			} else {
				m.mode = ViewList
			}
			return m, nil
		}

		switch m.mode {
		case ViewList:
			return m.updateList(msg)
		case ViewDetail:
			return m.updateDetail(msg)
		case ViewCreate, ViewEdit:
			return m.updateForm(msg)
		case ViewSearch:
			return m.updateSearch(msg)
		case ViewSettings:
			return m.updateSettings(msg)
		case ViewHelp:
			return m, nil
		}

	case snippetsLoadedMsg:
		m.snippets = msg.snippets
		if msg.pagination != nil {
			m.currentPage = msg.pagination.Page
			m.totalPages = msg.pagination.TotalPages
		}
		m.selectedIdx = 0
		m.detailSnippet = nil // Clear detail snippet when loading list

	case snippetLoadedMsg:
		m.detailSnippet = msg.snippet
		m.detailScroll = 0    // Reset scroll when loading new snippet
		m.selectedFileIdx = 0 // Reset file selection
		if m.mode == ViewList {
			for i, s := range m.snippets {
				if s.ID == msg.snippet.ID {
					m.snippets[i] = *msg.snippet
					break
				}
			}
		}

	case tagsLoadedMsg:
		m.tags = msg.tags

	case foldersLoadedMsg:
		m.folders = msg.folders

	case successMsg:
		m.message = msg.message
		m.mode = ViewList
		cmds = append(cmds, loadSnippets(m.client, m.currentPage, 20, m.searchQuery, m.filterTags, nil, "", nil, nil))

	case errMsg:
		m.err = msg.err
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}

	case "down", "j":
		if m.selectedIdx < len(m.snippets)-1 {
			m.selectedIdx++
		}

	case "s":
		m.mode = ViewSettings
		m.initSettingsForm()
		return m, nil

	case "enter":
		if len(m.snippets) > 0 {
			m.mode = ViewDetail
			return m, loadSnippet(m.client, m.snippets[m.selectedIdx].ID)
		}

	case "/":
		m.mode = ViewSearch
		m.initSearchForm()

	case "r":
		return m, loadSnippets(m.client, m.currentPage, 20, m.searchQuery, m.filterTags, nil, "", nil, nil)

	case "right", "l":
		if m.currentPage < m.totalPages {
			m.currentPage++
			return m, loadSnippets(m.client, m.currentPage, 20, m.searchQuery, m.filterTags, nil, "", nil, nil)
		}

	case "left", "h":
		if m.currentPage > 1 {
			m.currentPage--
			return m, loadSnippets(m.client, m.currentPage, 20, m.searchQuery, m.filterTags, nil, "", nil, nil)
		}
	}

	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace":
		m.mode = ViewList
		m.detailSnippet = nil
		m.detailScroll = 0

	case "up", "k":
		if m.detailScroll > 0 {
			m.detailScroll--
		}

	case "down", "j":
		m.detailScroll++

	case "left", "h":
		if m.detailSnippet != nil && len(m.detailSnippet.Files) > 1 {
			if m.selectedFileIdx > 0 {
				m.selectedFileIdx--
				m.detailScroll = 0
			}
		}

	case "right", "l":
		if m.detailSnippet != nil && len(m.detailSnippet.Files) > 1 {
			if m.selectedFileIdx < len(m.detailSnippet.Files)-1 {
				m.selectedFileIdx++
				m.detailScroll = 0
			}
		}

	case "c":
		if m.detailSnippet != nil {
			return m, copyToClipboard(m.detailSnippet.Content)
		}
	}

	return m, nil
}

func (m *Model) initCreateForm() {
	m.inputs = make([]textinput.Model, 3)

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Snippet Title"
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 200

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Language (e.g., go, python, javascript)"
	m.inputs[1].CharLimit = 50

	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Description (optional)"
	m.inputs[2].CharLimit = 1000

	m.focusedInput = 0
	m.formData = make(map[string]interface{})
}

func (m *Model) initEditForm(snippet *api.Snippet) {
	m.inputs = make([]textinput.Model, 3)

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Snippet Title"
	m.inputs[0].SetValue(snippet.Title)
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 200

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "Language"
	m.inputs[1].SetValue(snippet.Language)
	m.inputs[1].CharLimit = 50

	m.inputs[2] = textinput.New()
	m.inputs[2].Placeholder = "Description"
	m.inputs[2].SetValue(snippet.Description)
	m.inputs[2].CharLimit = 1000

	m.focusedInput = 0
	m.formData = map[string]interface{}{
		"content": snippet.Content,
	}
}

func (m *Model) initSearchForm() {
	m.inputs = make([]textinput.Model, 1)

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Search snippets..."
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 200

	m.focusedInput = 0
}

func (m *Model) initSettingsForm() {
	m.inputs = make([]textinput.Model, 2)

	m.inputs[0] = textinput.New()
	m.inputs[0].Placeholder = "Server URL"
	m.inputs[0].SetValue(m.config.ServerURL)
	m.inputs[0].Focus()
	m.inputs[0].CharLimit = 200

	m.inputs[1] = textinput.New()
	m.inputs[1].Placeholder = "API Key"
	m.inputs[1].SetValue(m.config.APIKey)
	m.inputs[1].CharLimit = 200
	m.inputs[1].EchoMode = textinput.EchoPassword
	m.inputs[1].EchoCharacter = '•'

	m.focusedInput = 0
}

func (m Model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.mode = ViewList
		return m, nil

	case "tab", "shift+tab":
		if msg.String() == "tab" {
			m.focusedInput++
		} else {
			m.focusedInput--
		}

		if m.focusedInput >= len(m.inputs) {
			m.focusedInput = 0
		} else if m.focusedInput < 0 {
			m.focusedInput = len(m.inputs) - 1
		}

		for i := range m.inputs {
			if i == m.focusedInput {
				m.inputs[i].Focus()
			} else {
				m.inputs[i].Blur()
			}
		}

		return m, nil

	case "ctrl+s":
		return m.submitForm()
	}

	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

func (m Model) submitForm() (tea.Model, tea.Cmd) {
	if len(m.inputs) < 2 {
		return m, nil
	}

	title := strings.TrimSpace(m.inputs[0].Value())
	language := strings.TrimSpace(m.inputs[1].Value())
	description := ""
	if len(m.inputs) > 2 {
		description = strings.TrimSpace(m.inputs[2].Value())
	}

	if title == "" {
		m.err = fmt.Errorf("title is required")
		return m, nil
	}

	content := ""
	if val, ok := m.formData["content"]; ok {
		if str, ok := val.(string); ok {
			content = str
		}
	}

	input := api.SnippetInput{
		Title:       title,
		Description: description,
		Language:    language,
		Content:     content,
	}

	if m.mode == ViewCreate {
		return m, createSnippet(m.client, input)
	} else if m.mode == ViewEdit && m.detailSnippet != nil {
		return m, updateSnippet(m.client, m.detailSnippet.ID, input)
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.mode = ViewList
		return m, nil

	case "enter":
		m.searchQuery = strings.TrimSpace(m.inputs[0].Value())
		m.mode = ViewList
		m.currentPage = 1
		return m, loadSnippets(m.client, 1, 20, m.searchQuery, m.filterTags, nil, "", nil, nil)
	}

	m.inputs[0], cmd = m.inputs[0].Update(msg)
	return m, cmd
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "esc":
		m.mode = ViewList
		return m, nil

	case "tab", "shift+tab":
		if msg.String() == "tab" {
			m.focusedInput++
		} else {
			m.focusedInput--
		}

		if m.focusedInput >= len(m.inputs) {
			m.focusedInput = 0
		} else if m.focusedInput < 0 {
			m.focusedInput = len(m.inputs) - 1
		}

		for i := range m.inputs {
			if i == m.focusedInput {
				m.inputs[i].Focus()
			} else {
				m.inputs[i].Blur()
			}
		}

		return m, nil

	case "ctrl+s":
		return m.saveSettings()
	}

	m.inputs[m.focusedInput], cmd = m.inputs[m.focusedInput].Update(msg)
	return m, cmd
}

func (m Model) saveSettings() (tea.Model, tea.Cmd) {
	if len(m.inputs) < 2 {
		return m, nil
	}

	serverURL := strings.TrimSpace(m.inputs[0].Value())
	apiKey := strings.TrimSpace(m.inputs[1].Value())

	if serverURL == "" || apiKey == "" {
		m.err = fmt.Errorf("server URL and API key are required")
		return m, nil
	}

	m.config.ServerURL = serverURL
	m.config.APIKey = apiKey

	if err := m.config.Save(); err != nil {
		m.err = fmt.Errorf("failed to save settings: %w", err)
		return m, nil
	}

	// Recreate client with new settings
	m.client = api.NewClient(m.config.ServerURL, m.config.APIKey)
	m.message = "Settings saved successfully"
	m.mode = ViewList

	return m, loadSnippets(m.client, 1, 20, "", nil, nil, "", nil, nil)
}

func copyToClipboard(content string) tea.Cmd {
	return func() tea.Msg {
		return successMsg{message: "Content copied to clipboard (feature requires clipboard package)"}
	}
}

func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	var s strings.Builder

	// Only show header in list view, not in detail view
	if m.mode == ViewList || m.mode == ViewSearch || m.mode == ViewSettings || m.mode == ViewHelp {
		s.WriteString(titleStyle.Render("🔖 Snipo TUI"))
		s.WriteString("\n")
		s.WriteString(subtitleStyle.Render(fmt.Sprintf("Connected to: %s", m.config.ServerURL)))
		s.WriteString("\n\n")
	}

	if m.err != nil {
		s.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.err)))
		s.WriteString("\n\n")
	}

	if m.message != "" {
		s.WriteString(successStyle.Render(m.message))
		s.WriteString("\n\n")
	}

	switch m.mode {
	case ViewList:
		s.WriteString(m.viewList())
	case ViewDetail:
		s.WriteString(m.viewDetail())
	case ViewCreate:
		s.WriteString(m.viewCreateForm())
	case ViewEdit:
		s.WriteString(m.viewEditForm())
	case ViewSearch:
		s.WriteString(m.viewSearchForm())
	case ViewHelp:
		s.WriteString(m.viewHelp())
	case ViewSettings:
		s.WriteString(m.viewSettings())
	}

	return s.String()
}

func (m Model) viewList() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render(fmt.Sprintf("Snippets (Page %d/%d)", m.currentPage, m.totalPages)))
	s.WriteString("\n\n")

	if len(m.snippets) == 0 {
		if m.searchQuery != "" {
			s.WriteString(dimmedStyle.Render("No snippets found matching your search. Press 'r' to refresh or '/' to search again."))
		} else {
			s.WriteString(dimmedStyle.Render("No snippets found. Press 'r' to refresh."))
		}
		s.WriteString("\n")
	}

	for i, snippet := range m.snippets {
		cursor := "  "
		style := normalItemStyle
		if i == m.selectedIdx {
			cursor = "▶ "
			style = selectedItemStyle
		}

		favorite := ""
		if snippet.IsFavorite {
			favorite = favoriteStyle.Render("★ ")
		}

		tags := ""
		if len(snippet.Tags) > 0 {
			var tagStrs []string
			for _, tag := range snippet.Tags {
				tagStrs = append(tagStrs, tagStyle.Render(tag.Name))
			}
			tags = " " + strings.Join(tagStrs, "")
		}

		lang := ""
		if snippet.Language != "" {
			lang = " " + languageStyle.Render("["+snippet.Language+"]")
		}

		line := fmt.Sprintf("%s%s%s%s%s", cursor, favorite, snippet.Title, lang, tags)
		s.WriteString(style.Render(line))
		s.WriteString("\n")
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/k up • ↓/j down • ←/h prev page • →/l next page • enter view • / search • s settings • r refresh • q quit • ? help"))

	return s.String()
}

func (m Model) viewDetail() string {
	if m.detailSnippet == nil {
		return dimmedStyle.Render("Loading...")
	}

	var s strings.Builder

	// Show snippet title prominently
	favorite := ""
	if m.detailSnippet.IsFavorite {
		favorite = favoriteStyle.Render(" ★")
	}

	s.WriteString(titleStyle.Render("🔖 " + m.detailSnippet.Title + favorite))
	s.WriteString("\n")

	// Show metadata in a compact format
	var metadata []string

	if m.detailSnippet.Language != "" {
		metadata = append(metadata, languageStyle.Render("Language: "+m.detailSnippet.Language))
	}

	if len(m.detailSnippet.Tags) > 0 {
		var tagStrs []string
		for _, tag := range m.detailSnippet.Tags {
			tagStrs = append(tagStrs, tagStyle.Render(tag.Name))
		}
		metadata = append(metadata, "Tags: "+strings.Join(tagStrs, " "))
	}

	if m.detailSnippet.IsPublic {
		metadata = append(metadata, dimmedStyle.Render("Public"))
	}

	if len(metadata) > 0 {
		s.WriteString(dimmedStyle.Render(strings.Join(metadata, " • ")))
		s.WriteString("\n")
	}

	if m.detailSnippet.Description != "" {
		s.WriteString("\n")
		s.WriteString(dimmedStyle.Render(m.detailSnippet.Description))
		s.WriteString("\n")
	}

	s.WriteString("\n")

	// Multi-file snippet support
	var content string
	var currentFilename string

	if len(m.detailSnippet.Files) > 0 {
		// Multi-file snippet - show file tabs with clear separator
		for i, file := range m.detailSnippet.Files {
			fileStyle := dimmedStyle
			if i == m.selectedFileIdx {
				fileStyle = selectedItemStyle.Copy().Underline(true)
			}
			s.WriteString(fileStyle.Render(fmt.Sprintf(" %s ", file.Filename)))
			s.WriteString(" ")
		}
		s.WriteString("\n")
		s.WriteString(dimmedStyle.Render(strings.Repeat("─", 60)))
		s.WriteString("\n\n")

		if m.selectedFileIdx < len(m.detailSnippet.Files) {
			content = m.detailSnippet.Files[m.selectedFileIdx].Content
			currentFilename = m.detailSnippet.Files[m.selectedFileIdx].Filename
		}
	} else {
		// Single-file snippet
		content = m.detailSnippet.Content
	}

	// Handle scrolling for large content
	contentLines := strings.Split(content, "\n")
	availableHeight := m.height - 18 // Reserve more space for file tabs

	if availableHeight < 5 {
		availableHeight = 5
	}

	// Calculate max line width for consistent rendering
	maxLineWidth := 0
	for _, line := range contentLines {
		if len(line) > maxLineWidth {
			maxLineWidth = len(line)
		}
	}

	// Ensure scroll position is within bounds
	maxScroll := len(contentLines) - availableHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.detailScroll > maxScroll {
		m.detailScroll = maxScroll
	}

	// Get visible content window and pad lines to max width
	startLine := m.detailScroll
	endLine := m.detailScroll + availableHeight
	if endLine > len(contentLines) {
		endLine = len(contentLines)
	}

	// Pad each visible line to the maximum width for consistent rendering
	var paddedLines []string
	for i := startLine; i < endLine; i++ {
		line := contentLines[i]
		if len(line) < maxLineWidth {
			line = line + strings.Repeat(" ", maxLineWidth-len(line))
		}
		paddedLines = append(paddedLines, line)
	}

	visibleContent := strings.Join(paddedLines, "\n")
	s.WriteString(codeBlockStyle.Render(visibleContent))

	// Show scroll indicator if content is larger than viewport
	if len(contentLines) > availableHeight {
		scrollInfo := fmt.Sprintf(" [%d-%d/%d lines]", startLine+1, endLine, len(contentLines))
		if currentFilename != "" {
			scrollInfo = fmt.Sprintf(" %s %s", currentFilename, scrollInfo)
		}
		s.WriteString("\n")
		s.WriteString(dimmedStyle.Render(scrollInfo))
	}

	s.WriteString("\n\n")

	helpText := "↑/k up • ↓/j down • esc back • c copy • q quit"
	if len(m.detailSnippet.Files) > 1 {
		helpText = "←/h prev file • →/l next file • " + helpText
	}
	s.WriteString(helpStyle.Render(helpText))

	return s.String()
}

func (m Model) viewCreateForm() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Create New Snippet"))
	s.WriteString("\n\n")

	for i, input := range m.inputs {
		s.WriteString(input.View())
		s.WriteString("\n")
		if i < len(m.inputs)-1 {
			s.WriteString("\n")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimmedStyle.Render("Note: Content editing in external editor coming soon"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("tab next field • ctrl+s save • esc cancel"))

	return s.String()
}

func (m Model) viewEditForm() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Edit Snippet"))
	s.WriteString("\n\n")

	for i, input := range m.inputs {
		s.WriteString(input.View())
		s.WriteString("\n")
		if i < len(m.inputs)-1 {
			s.WriteString("\n")
		}
	}

	s.WriteString("\n\n")
	s.WriteString(dimmedStyle.Render("Note: Content editing in external editor coming soon"))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("tab next field • ctrl+s save • esc cancel"))

	return s.String()
}

func (m Model) viewSearchForm() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Search Snippets"))
	s.WriteString("\n\n")

	s.WriteString(m.inputs[0].View())
	s.WriteString("\n\n")

	s.WriteString(helpStyle.Render("enter search • esc cancel"))

	return s.String()
}

func (m Model) viewSettings() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("⚙️  Settings"))
	s.WriteString("\n\n")

	s.WriteString(dimmedStyle.Render("Configure your Snipo server connection"))
	s.WriteString("\n\n")

	// Server URL field
	s.WriteString(normalItemStyle.Render("Server URL"))
	s.WriteString("\n")
	s.WriteString(dimmedStyle.Render("The URL of your Snipo server (e.g., http://localhost:8081)"))
	s.WriteString("\n")
	s.WriteString(m.inputs[0].View())
	s.WriteString("\n\n")

	// API Key field
	s.WriteString(normalItemStyle.Render("API Key"))
	s.WriteString("\n")
	s.WriteString(dimmedStyle.Render("Your personal API key for authentication"))
	s.WriteString("\n")
	s.WriteString(m.inputs[1].View())
	s.WriteString("\n\n")

	// Current connection status
	s.WriteString(dimmedStyle.Render("─────────────────────────────────────────────────────────"))
	s.WriteString("\n")
	s.WriteString(dimmedStyle.Render("Current connection: "))
	s.WriteString(successStyle.Render(m.config.ServerURL))
	s.WriteString("\n")
	s.WriteString(dimmedStyle.Render("API Key: "))
	if m.config.APIKey != "" {
		s.WriteString(successStyle.Render("••••••••••••••••"))
	} else {
		s.WriteString(errorStyle.Render("Not set"))
	}
	s.WriteString("\n")
	s.WriteString(dimmedStyle.Render("─────────────────────────────────────────────────────────"))

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("tab/shift+tab navigate • ctrl+s save • esc cancel"))

	return s.String()
}

func (m Model) viewHelp() string {
	var s strings.Builder

	s.WriteString(headerStyle.Render("Snipo TUI - Help"))
	s.WriteString("\n\n")

	help := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move up in list"},
		{"↓/j", "Move down in list"},
		{"←/h", "Previous page / Previous file (in detail view)"},
		{"→/l", "Next page / Next file (in detail view)"},
		{"enter", "View selected snippet"},
		{"/", "Search snippets"},
		{"s", "Settings (change server/API key)"},
		{"r", "Refresh list"},
		{"c", "Copy content to clipboard (in detail view)"},
		{"esc", "Go back / Cancel"},
		{"?", "Toggle this help screen"},
		{"q", "Quit application"},
	}

	for _, h := range help {
		s.WriteString(fmt.Sprintf("  %s  %s\n",
			selectedItemStyle.Render(h.key),
			normalItemStyle.Render(h.desc)))
	}

	s.WriteString("\n")
	s.WriteString(helpStyle.Render("Press ? to close help"))

	return s.String()
}
