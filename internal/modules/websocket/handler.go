package websocket

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	sharederrors "github.com/MustafaKheda/go-connect-too-backend/internal/shared/errors"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/response"
	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/security"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Handler upgrades HTTP requests to WebSocket connections.
type Handler struct {
	hub    *Hub
	tokens *security.TokenManager
	log    *slog.Logger
}

// NewHandler creates a WebSocket handler.
func NewHandler(hub *Hub, tokens *security.TokenManager, log *slog.Logger) *Handler {
	return &Handler{hub: hub, tokens: tokens, log: log}
}

// Serve handles GET /ws with JWT authentication.
func (h *Handler) Serve(w http.ResponseWriter, r *http.Request) {
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if token == "" {
		response.Error(w, http.StatusUnauthorized, "Authentication required", sharederrors.CodeUnauthorized)
		return
	}

	claims, err := h.tokens.ParseAccessToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "Invalid or expired token", sharederrors.CodeInvalidToken)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.log.Error("websocket upgrade failed", slog.String("error", err.Error()))
		return
	}

	client := &Client{
		UserID: claims.UserID,
		Send:   make(chan []byte, 16),
	}
	h.hub.Register(client)

	go h.writePump(conn, client)
	go h.readPump(conn, client)
}

func (h *Handler) readPump(conn *websocket.Conn, client *Client) {
	defer func() {
		h.hub.Unregister(client)
		_ = conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Handler) writePump(conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
