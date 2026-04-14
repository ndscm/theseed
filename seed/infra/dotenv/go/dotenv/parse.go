package dotenv

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func parseEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
			}
			i++
		} else {
			b.WriteByte(s[i])
		}
	}

	return b.String()
}

func ParseLine(line string) (key string, value string, err error) {
	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] == '#' {
		return "", "", nil
	}

	// Strip optional "export " prefix.
	line = strings.TrimPrefix(line, "export ")

	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return "", "", seederr.WrapErrorf("dotenv: invalid line: %q", line)
	}

	key = strings.TrimSpace(line[:eq])
	raw := strings.TrimSpace(line[eq+1:])

	if len(raw) >= 2 && raw[0] == '\'' && raw[len(raw)-1] == '\'' {
		// Single-quoted: literal value, no escape processing.
		value = raw[1 : len(raw)-1]
	} else if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		// Double-quoted: process escape sequences.
		value = parseEscapes(raw[1 : len(raw)-1])
	} else {
		// Unquoted: strip inline comments.
		if i := strings.Index(raw, " #"); i >= 0 {
			raw = strings.TrimRight(raw[:i], " ")
		}
		value = raw
	}

	return key, value, nil
}

func ParseLines(lines []string) (result map[string]string, err error) {
	result = make(map[string]string)
	for _, line := range lines {
		key, value, err := ParseLine(line)
		if err != nil {
			return nil, err
		}
		if key == "" {
			continue
		}
		result[key] = value
	}
	return result, nil
}

func Parse(text string) (result map[string]string, err error) {
	return ParseLines(strings.Split(text, "\n"))
}
