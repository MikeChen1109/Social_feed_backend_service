package models

import (
	"github.com/stretchr/testify/require"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestCursorRoundTripAndValidation(t *testing.T) {
	stamp := time.Now().UTC().Truncate(time.Microsecond)
	token := EncodeCursor("feeds", stamp, 42)
	cursor, limit, err := ParsePagination(url.Values{"cursor": {token}, "limit": {"2"}}, "feeds")
	require.NoError(t, err)
	require.Equal(t, 2, limit)
	require.Equal(t, uint(42), cursor.ID)
	require.True(t, stamp.Equal(cursor.CreatedAt))
	for _, q := range []url.Values{{"cursor": {token}}, {"cursor": {strings.Repeat("a", 513)}}, {"cursor": {"!"}}, {"cursor": {EncodeCursor("comments:1", stamp, 0)}}, {"page": {"1"}}, {"limit": {"101"}}} {
		_, _, err := ParsePagination(q, "comments:1")
		require.Error(t, err)
	}
}
