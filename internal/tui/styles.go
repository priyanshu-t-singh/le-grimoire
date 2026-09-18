package tui

import "charm.land/lipgloss/v2"

var (
	// HeaderStyle renders the top bar with inverted colours.
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("15")).
			Padding(0, 1)

	// GapFillStyle fills the space between the left and right header pieces.
	// Same colours as HeaderStyle but ZERO padding so the gap width is exact —
	// HeaderStyle.Render() adds 2 extra chars (left+right pad) which would push
	// the right-side text off screen.
	GapFillStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("0")).
			Foreground(lipgloss.Color("15"))

	// SelectedStyle highlights the currently focused list item.
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("2"))

	// NormalStyle renders unfocused list items.
	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	// FooterStyle renders the key-hint bar at the bottom.
	FooterStyle = lipgloss.NewStyle().
			Faint(true).
			Padding(0, 1)

	// ErrorStyle renders error messages in red.
	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("1"))

	// LoadingStyle renders the loading spinner/text.
	LoadingStyle = lipgloss.NewStyle().
			Faint(true).
			Italic(true)

	// TitleStyle renders section titles in the reader.
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12"))

	// ReaderStyle renders normal body text in the reader.
	ReaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15"))

	// PlaceholderStyle renders image-page placeholder notices.
	PlaceholderStyle = lipgloss.NewStyle().
				Faint(true).
				Italic(true).
				Foreground(lipgloss.Color("3"))
)
