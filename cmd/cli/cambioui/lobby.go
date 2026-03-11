package cambioui

import (
	"fmt"
	"sync"
	"time"

	internalcambio "github.com/dhvbnl/cambio-ssh/cmd/internal/cambio"
)

type lobbySession struct {
	id         string
	owner      string
	ownerID    string
	guest      string
	guestID    string
	game       *internalcambio.Game
	ownerReady bool
	guestReady bool
}

type lobbySummary struct {
	id       string
	hostName string
}

type lobbyRegistry struct {
	mu       sync.RWMutex
	sessions map[string]*lobbySession
}

var sharedLobbies = &lobbyRegistry{sessions: make(map[string]*lobbySession)}

func (r *lobbyRegistry) createLobby(owner string, ownerID string) (*lobbySession, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing := r.findSessionByClientLocked(ownerID); existing != nil {
		return existing, r.playerIndexForClientLocked(existing, ownerID), nil
	}

	id := fmt.Sprintf("%s-%d", ownerID, time.Now().UnixNano())
	session := &lobbySession{
		id:      id,
		owner:   owner,
		ownerID: ownerID,
		game:    internalcambio.NewGame(2),
	}
	r.sessions[id] = session
	return session, 0, nil
}

func (r *lobbyRegistry) listJoinableLobbies(requestorID string) []lobbySummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]lobbySummary, 0, len(r.sessions))
	for _, session := range r.sessions {
		if session.ownerID == requestorID {
			continue
		}
		if session.guest != "" {
			continue
		}
		out = append(out, lobbySummary{id: session.id, hostName: session.owner})
	}
	return out
}

func (r *lobbyRegistry) joinLobby(id string, guest string, guestID string) (*lobbySession, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing := r.findSessionByClientLocked(guestID); existing != nil {
		return existing, r.playerIndexForClientLocked(existing, guestID), nil
	}

	session, ok := r.sessions[id]
	if !ok {
		return nil, -1, fmt.Errorf("selected game no longer exists")
	}
	if session.ownerID == guestID {
		return nil, -1, fmt.Errorf("cannot join your own game")
	}
	if session.guest != "" {
		return nil, -1, fmt.Errorf("game is already full")
	}

	session.guest = guest
	session.guestID = guestID
	session.guestReady = false
	return session, 1, nil
}

func (r *lobbyRegistry) markReady(sessionID string, clientID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return false, fmt.Errorf("game lobby no longer exists")
	}

	switch clientID {
	case session.ownerID:
		session.ownerReady = true
	case session.guestID:
		session.guestReady = true
	default:
		return false, fmt.Errorf("you are not part of this game")
	}

	bothReady := session.ownerReady && session.guestReady && session.guest != ""
	if bothReady && session.game.GetGameState() == internalcambio.GameStart {
		session.game.StartGame()
	}
	return bothReady, nil
}

func (r *lobbyRegistry) getSession(id string) (*lobbySession, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[id]
	return session, ok
}

func (r *lobbyRegistry) getPlayerNames(id string) ([]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return nil, false
	}

	names := []string{session.owner, session.guest}
	if names[1] == "" {
		names[1] = "Waiting for player"
	}
	return names, true
}

func (r *lobbyRegistry) hasGuest(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[id]
	if !ok {
		return false
	}
	return session.guest != ""
}

func (r *lobbyRegistry) playerIndexForClient(sessionID string, clientID string) (int, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return -1, false
	}

	if session.ownerID == clientID {
		return 0, true
	}
	if session.guestID == clientID {
		return 1, true
	}
	return -1, false
}

func (r *lobbyRegistry) leave(sessionID string, clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, ok := r.sessions[sessionID]
	if !ok {
		return
	}

	switch clientID {
	case session.ownerID:
		delete(r.sessions, sessionID)
	case session.guestID:
		session.guest = ""
		session.guestID = ""
		session.guestReady = false
	}
}

func (r *lobbyRegistry) findSessionByClientLocked(clientID string) *lobbySession {
	for _, session := range r.sessions {
		if session.ownerID == clientID || session.guestID == clientID {
			return session
		}
	}
	return nil
}

func (r *lobbyRegistry) playerIndexForClientLocked(session *lobbySession, clientID string) int {
	if session.ownerID == clientID {
		return 0
	}
	if session.guestID == clientID {
		return 1
	}
	return -1
}
