package field

import (
	"bytes"
	"encoding/json"
	"errors"
	"unicode/utf8"
	_ "unsafe"
)

//go:linkname htmlSafeSet encoding/json.htmlSafeSet
var htmlSafeSet [utf8.RuneSelf]bool

//go:linkname safeSet encoding/json.safeSet
var safeSet [utf8.RuneSelf]bool

var hex = "0123456789abcdef"

func appendString[Bytes []byte | string](dst []byte, src Bytes, escapeHTML bool) []byte {
	dst = append(dst, '"')
	start := 0
	for i := 0; i < len(src); {
		if b := src[i]; b < utf8.RuneSelf {
			if htmlSafeSet[b] || (!escapeHTML && safeSet[b]) {
				i++
				continue
			}
			dst = append(dst, src[start:i]...)
			switch b {
			case '\\', '"':
				dst = append(dst, '\\', b)
			case '\n':
				dst = append(dst, '\\', 'n')
			case '\r':
				dst = append(dst, '\\', 'r')
			case '\t':
				dst = append(dst, '\\', 't')
			default:
				// This encodes bytes < 0x20 except for \t, \n and \r.
				// If escapeHTML is set, it also escapes <, >, and &
				// because they can lead to security holes when
				// user-controlled strings are rendered into JSON
				// and served to some browsers.
				dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		// TODO(https://go.dev/issue/56948): Use generic utf8 functionality.
		// For now, cast only a small portion of byte slices to a string
		// so that it can be stack allocated. This slows down []byte slightly
		// due to the extra copy, but keeps string performance roughly the same.
		n := len(src) - i
		if n > utf8.UTFMax {
			n = utf8.UTFMax
		}
		c, size := utf8.DecodeRuneInString(string(src[i : i+n]))
		if c == utf8.RuneError && size == 1 {
			dst = append(dst, src[start:i]...)
			dst = append(dst, `\ufffd`...)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			dst = append(dst, src[start:i]...)
			dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	dst = append(dst, src[start:]...)
	dst = append(dst, '"')
	return dst
}

func appendJsonStringBuf(w Buffer, s string) (err error) {
	var _, wrote = w.Write(appendString[string](nil, s, false))
	return wrote
}

type jsonErr interface {
	error
	json.Marshaler
}

func asJsonErrMarshaler(err error) (jsonErr jsonErr) {
	if err == nil {
		return nil
	}
	if errors.As(err, &jsonErr) {
		return jsonErr
	}
	return jsonErrMarshaler{err}
}

type jsonErrMarshaler struct{ error }

func (e jsonErrMarshaler) MarshalJSON() ([]byte, error) {
	if e.error == nil {
		return []byte("null"), nil
	}
	var buffer bytes.Buffer
	if err := appendJsonStringBuf(&buffer, e.Error()); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
