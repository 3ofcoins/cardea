package cardea

import "testing"
import . "github.com/smartystreets/goconvey/convey"

import "bytes"
import "log"
import "net/http"
import "net/http/httptest"
import "os"
import "time"

func mkSince(str string) func(time.Time) time.Duration {
	if duration, err := time.ParseDuration(str); err != nil {
		panic(err)
	} else {
		return func(_ time.Time) time.Duration {
			return duration
		}
	}
}

func tapLog(inner func()) string {
	// Redirect log output, assume we've been logging to stderr
	logb := bytes.NewBuffer(nil)
	log.SetOutput(logb)
	defer func() { log.SetOutput(os.Stderr) }()
	inner()
	return logb.String()
}

func TestConfig(t *testing.T) {
	var c *Config

	Convey("cardea.Config", t, func() {
		c = NewConfig("swordfish")

		// We're going to override since(), let's restore it at exit
		defer func(orig_since func(time.Time) time.Duration) {
			since = orig_since
		}(since)

		Convey("CheckCookie", func() {
			Convey("Fails with invalid cookie", func() {
				res, err := c.CheckCookie("dupa.8", "GoConvey")
				So(res, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("Succeeds with correct, fresh cookie", func() {
				since = mkSince("15m")
				res, err := c.CheckCookie("YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca", "GoConvey")
				So(err, ShouldBeNil)
				So(res.User, ShouldEqual, "a_user")
				So(res.Groups, ShouldResemble, []string{"some", "groups"})
				So(res.Timestamp.Unix(), ShouldEqual, 23)
				So(res.String(), ShouldEqual, "a_user[some,groups] (15m0s)")
			})

			Convey("Fails with mismatched HMAC", func() {
				since = mkSince("15m")
				res, err := c.CheckCookie("YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca", "GoConvey/different")
				So(err, ShouldNotBeNil)
				So(res, ShouldBeNil)
			})

			Convey("Fails with correct stale cookie", func() {
				since = mkSince("96h")
				res, err := c.CheckCookie("YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca", "GoConvey")
				So(err, ShouldNotBeNil)
				So(res.User, ShouldEqual, "a_user")
				So(res.Groups, ShouldResemble, []string{"some", "groups"})
				So(res.Timestamp.Unix(), ShouldEqual, 23)
				So(res.String(), ShouldEqual, "a_user[some,groups] (96h0m0s)")
			})
		})

		Convey("ServeHTTP", func() {
			req, err := http.NewRequest("GET", "http://127.0.0.1:9999/", nil)
			if err != nil {
				panic(err)
			}
			res := httptest.NewRecorder()

			Convey("No cookie", func() {
				tapLog(func() { c.ServeHTTP(res, req) })
				So(res.Code, ShouldEqual, http.StatusForbidden)
			})

			Convey("Good cookie", func() {
				since = mkSince("15m")
				req.AddCookie(&http.Cookie{Name: "ca", Value: "YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca"})
				req.Header.Add("X-Cardea-HMAC-Extra", "GoConvey")
				tapLog(func() { c.ServeHTTP(res, req) })
				So(res.Code, ShouldEqual, http.StatusOK)
			})
		})
	})
}
