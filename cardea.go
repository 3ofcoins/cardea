package cardea

//go:generate make generate

import "errors"
import "log"
import "net/http"
import "strings"

var (
	DEFAULT_COOKIE_NAME    = "ca"
	DEFAULT_EXPIRATION_SEC = uint64(36 * 3600)
)

type Config struct {
	Secret        []byte
	Cookie        string
	ExpirationSec uint64
}

func NewConfig(secret string) *Config {
	return &Config{[]byte(secret),
		DEFAULT_COOKIE_NAME,
		DEFAULT_EXPIRATION_SEC,
	}
}

func (c *Config) CheckToken(t *Token, ua string) (err error) {
	if uint64(t.Age().Seconds()) > c.ExpirationSec {
		return errors.New("Expired cookie")
	}

	return nil
}

func (c *Config) CheckCookie(cookie string, ua string) (t *Token, err error) {
	t, err = ParseCookie(c.Secret, ua, cookie)
	if err != nil {
		return
	}

	err = c.CheckToken(t, ua)
	return
}

func (c *Config) CheckRequest(r *http.Request) (t *Token, err error) {
	_cookie, err := r.Cookie(c.Cookie)
	if err != nil {
		return
	}

	return c.CheckCookie(_cookie.Value, strings.Join(r.Header["User-Agent"], "\n"))
}

func (c *Config) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hdr := w.Header()

	if t, err := c.CheckRequest(r); err != nil {
		log.Printf("%v DENY %s (%s)", r.Header["X-Cardea-Requestinfo"], t, err)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Denied\n"))
	} else {
		log.Printf("%v ALLOW %s", r.Header["X-Cardea-Requestinfo"], t)
		hdr["X-Cardea-User"] = []string{t.User}
		hdr["X-Cardea-Groups"] = t.Values["g"]
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	}
}
