package cardea

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

func (c *Config) CheckToken(t *Token, hmac_extra string) (err error) {
	if uint64(t.Age().Seconds()) > c.ExpirationSec {
		return errors.New("Expired cookie")
	}

	return nil
}

func (c *Config) CheckCookie(cookie string, hmac_extra string) (t *Token, err error) {
	t, err = ParseCookie(c.Secret, hmac_extra, cookie)
	if err != nil {
		return
	}

	err = c.CheckToken(t, hmac_extra)
	return
}

func (c *Config) CheckAuthorization(auth string, hmac_extra string) (t *Token, err error) {
	t, err = ParseAuthorization(c.Secret, hmac_extra, auth)
	if err != nil {
		return
	}

	err = c.CheckToken(t, hmac_extra)
	return
}

func (c *Config) CheckRequest(r *http.Request) (t *Token, err error) {
	switch cookie, err := r.Cookie(c.Cookie); err {
	case nil:
		return c.CheckCookie(cookie.Value,
			strings.Join(r.Header["X-Cardea-Hmac-Extra"], "\n"))
	case http.ErrNoCookie:
		// Try to parse basic auth

		auth := r.Header["Authorization"]
		switch len(auth) {
		case 0:
			return nil, err
		case 1: // we're good
		default:
			return nil, errors.New("More than one Authorization: headers")
		}
		return c.CheckAuthorization(string(auth[0]),
			strings.Join(r.Header["X-Cardea-Hmac-Extra"], "\n"))
	default:
		return nil, err
	}

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
		hdr["X-Cardea-Groups"] = t.Groups
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	}
}
