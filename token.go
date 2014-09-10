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

// Cookie format: urlencode( base64(username) "," base64(groups) "," timestamp (time_t) "," HMAC )
var COOKIE_RX = regexp.MustCompile("^\\s*([a-zA-Z0-9_-]+),([a-zA-Z0-9_-]+),(\\d+),([0-9a-f]+)\\s*$")

type Token struct {
	User      string
	Groups    string
	Timestamp time.Time
	HMAC      []byte
}

func ParseCookie(cookie string) (t *Token, err error) {
	// (user, groups string, ts uint64, hmac []byte, err error)
	cookie, err = url.QueryUnescape(cookie)
	if err != nil {
		return
	}

	_pieces := COOKIE_RX.FindStringSubmatch(cookie)
	if _pieces == nil {
		err = errors.New("Malformed cookie")
		return
	}

	t = &Token{}

	t.User, err = unb64(_pieces[1])
	if err != nil {
		return
	}

	t.Groups, err = unb64(_pieces[2])
	if err != nil {
		return
	}

	if ts, err := strconv.ParseUint(_pieces[3], 10, 64); err != nil {
		return t, err
	} else {
		t.Timestamp = time.Unix(int64(ts), 0)
	}

	t.HMAC, err = hex.DecodeString(_pieces[4])
	if err != nil {
		// CAN'T HAPPEN: regexp will catch a malformed value, but it's
		// better to err on the safe side.
		return
	}

	return
}

func (t *Token) Age() time.Duration {
	return since(t.Timestamp)
}

func (t *Token) ComputeHMAC(secret, ua string) []byte {
	_hmac := hmac.New(sha256.New, []byte(secret))
	_hmac.Write([]byte(fmt.Sprintf("%s,%s,%v,%s",
		b64(t.User), b64(t.Groups), t.Timestamp.Unix(), b64(ua))))
	return _hmac.Sum(nil)
}

func (t *Token) IsValid(secret, ua string) bool {
	return hmac.Equal(t.HMAC, t.ComputeHMAC(secret, ua))
}

func (t *Token) String() string {
	return fmt.Sprintf("%s[%s] (%v)", t.User, t.Groups, t.Age())
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
