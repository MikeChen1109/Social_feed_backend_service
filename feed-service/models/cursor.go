package models

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Cursor identifies a position in the immutable (created_at, id) ordering.
// Scope binds comment cursors to one feed; cursors are pagination hints, not authorization.
type Cursor struct {
	Version   int       `json:"v"`
	Scope     string    `json:"scope"`
	CreatedAt time.Time `json:"createdAt"`
	ID        uint      `json:"id"`
}

func EncodeCursor(scope string, createdAt time.Time, id uint) string {
	data, _ := json.Marshal(Cursor{Version: 1, Scope: scope, CreatedAt: createdAt.UTC(), ID: id})
	return base64.RawURLEncoding.EncodeToString(data)
}
func ParsePagination(query url.Values, scope string) (*Cursor, int, error) {
	if len(query["cursor"]) > 1 || len(query["limit"]) > 1 {
		return nil, 0, errors.New("cursor and limit must appear only once")
	}
	if query.Has("page") {
		return nil, 0, errors.New("page pagination is no longer supported; use cursor")
	}
	limit := 10
	if query.Has("limit") {
		parsed, err := strconv.Atoi(query.Get("limit"))
		if err != nil || parsed < 1 || parsed > 100 {
			return nil, 0, errors.New("limit must be between 1 and 100")
		}
		limit = parsed
	}
	token := query.Get("cursor")
	if token == "" {
		return nil, limit, nil
	}
	if len(token) > 512 {
		return nil, 0, errors.New("invalid cursor")
	}
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, 0, errors.New("invalid cursor")
	}
	var cursor Cursor
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&cursor) != nil || decoder.Decode(new(any)) != io.EOF || cursor.Version != 1 || cursor.Scope != scope || cursor.ID == 0 || uint64(cursor.ID) > math.MaxInt64 || cursor.CreatedAt.IsZero() {
		return nil, 0, errors.New("invalid cursor for this list")
	}
	return &cursor, limit, nil
}
func CommentCursorScope(feedID uint) string { return fmt.Sprintf("comments:%d", feedID) }
