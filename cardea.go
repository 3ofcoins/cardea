package cardea

import "crypto/hmac"
import "crypto/sha256"
import "encoding/base64"
import "encoding/hex"
import "errors"
import "fmt"
import "log"
import "net/http"
import "net/url"
import "regexp"
import "strconv"
import "strings"
import "time"

var (
	COOKIE_RX              = regexp.MustCompile("^\\s*([a-zA-Z0-9_-]+),([a-zA-Z0-9_-]+),(\\d+),([0-9a-f]+)\\s*$")
	DEFAULT_COOKIE_NAME    = "ca"
	DEFAULT_EXPIRATION_SEC = uint64(36 * 3600)
)

type Config struct {
	Secret        string
	Cookie        string
	ExpirationSec uint64
}

type CheckResult struct {
	User      string
	Groups    string
	Timestamp uint64
	Age       time.Duration
	Valid     bool
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

func ParseCookie(cookie string) (user, groups string, ts uint64, hmac []byte, err error) {
	cookie, err = url.QueryUnescape(cookie)
	if err != nil {
		return
	}

	_pieces := COOKIE_RX.FindStringSubmatch(cookie)
	if _pieces == nil {
		err = errors.New("Malformed cookie")
		return
	}

	user, err = unb64(_pieces[1])
	if err != nil {
		return
	}

	groups, err = unb64(_pieces[2])
	if err != nil {
		return
	}

	ts, err = strconv.ParseUint(_pieces[3], 10, 64)
	if err != nil {
		return
	}

	hmac, err = hex.DecodeString(_pieces[4])
	if err != nil {
		// CAN'T HAPPEN: regexp will catch a malformed value, but it's
		// better to err on the safe side.
		return
	}

	return
}

func HMAC(secret, user, groups string, ts uint64, ua string) []byte {
	_hmac := hmac.New(sha256.New, []byte(secret))
	_hmac.Write([]byte(fmt.Sprintf("%s,%s,%v,%s",
		b64(user), b64(groups), ts, b64(ua))))
	return _hmac.Sum(nil)
}

func New(secret string) *Config {
	return &Config{secret,
		DEFAULT_COOKIE_NAME,
		DEFAULT_EXPIRATION_SEC,
	}
}

func (cr *CheckResult) String() string {
	var invalid_flag = ""
	if !cr.Valid {
		invalid_flag = "!"
	}
	return fmt.Sprintf("%s%s[%s] (%v)", invalid_flag, cr.User, cr.Groups, cr.Age)
}

var since = time.Since // to be mocked in tests

func (c *Config) CheckCookie(cookie string, ua string) (result *CheckResult, err error) {
	user, groups, ts, received_hmac, err := ParseCookie(cookie)
	if err != nil {
		return
	}

	result = &CheckResult{user, groups, ts, since(time.Unix(int64(ts), 0)), false}

	computed_hmac := HMAC(c.Secret, user, groups, ts, ua)

	if !hmac.Equal(computed_hmac, received_hmac) {
		err = errors.New("HMAC mismatch")
		return
	}

	if uint64(result.Age.Seconds()) > c.ExpirationSec {
		err = errors.New("Expired cookie")
		return
	}

	result.Valid = true
	return
}

func (c *Config) CheckRequest(r *http.Request) (result *CheckResult, err error) {
	_cookie, err := r.Cookie(c.Cookie)
	if err != nil {
		return
	}

	return c.CheckCookie(_cookie.Value, strings.Join(r.Header["User-Agent"], "\n"))
}

func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()

	if res, err := c.CheckRequest(r); err != nil {
		log.Printf("%v DENY %s (%s)", r.Header["X-Cardea-RequestInfo"], res, err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Denied\n"))
	} else {
		log.Printf("%v ALLOW %s", r.Header["X-Cardea-RequestInfo"], res)
		hdr["X-Cardea-User"] = []string{res.User}
		hdr["X-Cardea-Groups"] = []string{res.Groups}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	}
}
