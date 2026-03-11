package cli

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dhvbnl/cambio-ssh/cmd/cli/cambioui"
)

func NewCambioGameModel(username string) tea.Model {
	clientID := fmt.Sprintf("%s-%d", username, time.Now().UnixNano())
	return cambioui.NewModel(navigate(screenHome), username, clientID)
}
