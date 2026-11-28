package repository

// SQLite recognizes RFC 3339 values directly. The SQLite driver also writes
// time.Time.String() values with an unpunctuated offset and a zone name.
const snippetTimeZone = `max(instr(s.expires_at, ' +'), instr(s.expires_at, ' -'))`

const unexpiredSnippet = `(s.expires_at IS NULL OR COALESCE(
	julianday(s.expires_at),
	julianday(substr(s.expires_at, 1, ` + snippetTimeZone + ` - 1) ||
		substr(s.expires_at, ` + snippetTimeZone + ` + 1, 3) || ':' ||
		substr(s.expires_at, ` + snippetTimeZone + ` + 4, 2))
) > julianday('now'))`
