package cli

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhvbnl/cambio-ssh/cmd/cli/cambioui"
)

func NewCambioGameModel() tea.Model {
	return cambioui.NewModel(navigate(screenHome))
}
