package cardea

import "time"
import "net/url"
import "regexp"
import "errors"
import "strings"
import "encoding/base64"
import "encoding/hex"
import "crypto/hmac"
import "crypto/sha256"
import "strconv"
import "fmt"

var since = time.Since // local copy to let tests override the method

var AUTHORIZATION_RX = regexp.MustCompile("^\\s*Basic\\s+(.*)$")

type Token struct {
	User      string
	Timestamp time.Time
	Groups    []string
	url.Values
}

func DeserializeToken(secret []byte, hmac_extra, serialized string) (t *Token, err error) {
	parts := TOKEN_RX.FindStringSubmatch(serialized)
	if parts == nil {
		err = errors.New("Malformed token")
		return
	}

	t = &Token{}

	if parts[TOKEN_RX_LEGACY_TIMESTAMP] == "" {
		// Proper (Cardea-style) cookie

		if !check_hmac(secret, parts[TOKEN_RX_HMAC], "%s:%s#%s",
			parts[TOKEN_RX_USERNAME],
			parts[TOKEN_RX_QUERY],
			b64(hmac_extra)) {
			return nil, errors.New("HMAC mismatch")
		}

		t.User = parts[TOKEN_RX_USERNAME]

		// TODO: variable, future-proof format for query (to indicate
		// encrypted content, Base64 coding, etc).
		query := parts[TOKEN_RX_QUERY]

		if vals, err := url.ParseQuery(query); err != nil {
			return nil, err
		} else {
			t.Values = vals
		}

		t.Groups = t.Values["g"]
		delete(t.Values, "g")

		if ts, has_ts := t.Values["t"]; has_ts {
			if nt := len(ts); nt != 1 {
				return nil, fmt.Errorf("Confused by %d timestamps on cookie", nt)
			}
			if ts, err := strconv.ParseUint(ts[0], 10, 64); err != nil {
				return nil, err
			} else {
				t.Timestamp = time.Unix(int64(ts), 0)
			}
		}
	} else {
		// Legacy (OdinAuth-style) cookie

		if !check_hmac(secret, parts[TOKEN_RX_HMAC], "%s,%s,%s,%s",
			parts[TOKEN_RX_USERNAME],
			parts[TOKEN_RX_LEGACY_GROUPS],
			parts[TOKEN_RX_LEGACY_TIMESTAMP],
			b64(hmac_extra)) {
			return nil, errors.New("HMAC mismatch")
		}

		if user, err := unb64(parts[TOKEN_RX_USERNAME]); err != nil {
			return nil, err
		} else {
			t.User = user
		}

		if groups, err := unb64(parts[TOKEN_RX_LEGACY_GROUPS]); err != nil {
			return nil, err
		} else {
			t.Groups = strings.Split(groups, ",")
		}

		if ts, err := strconv.ParseUint(parts[TOKEN_RX_LEGACY_TIMESTAMP], 10, 64); err != nil {
			return nil, err
		} else {
			t.Timestamp = time.Unix(int64(ts), 0)
		}
	}

	return t, nil
}

func ParseCookie(secret []byte, hmac_extra, cookie string) (*Token, error) {
	if unescaped, err := url.QueryUnescape(cookie); err != nil {
		return nil, err
	} else {
		return DeserializeToken(secret, hmac_extra, unescaped)
	}
}

func ParseAuthorization(secret []byte, hmac_extra, auth string) (*Token, error) {
	if match := AUTHORIZATION_RX.FindStringSubmatch(auth); match == nil {
		return nil, errors.New("Malformed Authorization header or not basic auth")
	} else {
		if bytes, err := base64.StdEncoding.DecodeString(match[1]); err != nil {
			return nil, err
		} else {
			return DeserializeToken(secret, hmac_extra, string(bytes))
		}
	}
}

func (t *Token) Age() time.Duration {
	return since(t.Timestamp)
}

func (t *Token) String() string {
	return fmt.Sprintf("%s[%s] (%v)", t.User, strings.Join(t.Groups, ","), t.Age())
}

// Base64 helpers that trim/add `=' padding

func b64(str string) string {
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(str)), "=")
}

func unb64(str string) (string, error) {
	if n := len(str) % 4; n != 0 {
		str += strings.Repeat("=", 4-n)
	}
	if bytes, err := base64.StdEncoding.DecodeString(str); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}

func check_hmac(
	secret []byte,
	received, format string,
	params ...interface{},
) bool {
	unsigned := fmt.Sprintf(format, params...)

	h := hmac.New(sha256.New, secret)
	received_b, err := hex.DecodeString(received)
	if err != nil {
		// Malformed hex encoding, screw it.
		return false
	}
	h.Write([]byte(unsigned))
	return hmac.Equal(h.Sum(nil), received_b)
}
