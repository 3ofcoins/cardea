package cardea

import "testing"
import . "github.com/smartystreets/goconvey/convey"

import "encoding/hex"
import "math"
import "regexp"
import "time"

func force_unb64(enc string) string {
	if raw, err := unb64(enc); err == nil {
		return raw
	} else {
		panic(err)
	}
}

func TestBase64(t *testing.T) {
	Convey("Stripped Base64 encoding", t, func() {
		Convey("Encoded strings are being stripped", func() {
			So(b64("foo"), ShouldEqual, "Zm9v")        // no padding
			So(b64("Odhin"), ShouldEqual, "T2RoaW4")   // single char padding
			So(b64("auth"), ShouldEqual, "YXV0aA")     // two char padding
			So(b64("dupa.8"), ShouldEqual, "ZHVwYS44") // no padding again
		})
		Convey("Strings without padding are decoded correctly", func() {
			So(force_unb64("Zm9v"), ShouldEqual, "foo")        // no padding
			So(force_unb64("T2RoaW4"), ShouldEqual, "Odhin")   // single char padding
			So(force_unb64("YXV0aA"), ShouldEqual, "auth")     // two char padding
			So(force_unb64("ZHVwYS44"), ShouldEqual, "dupa.8") // no padding again
		})
		Convey("Padded strings are also decoded correctly", func() {
			So(force_unb64("T2RoaW4="), ShouldEqual, "Odhin")
			So(force_unb64("YXV0aA=="), ShouldEqual, "auth")
		})
		Convey("Invalid base64 returns an error on decode", func() {
			_, err := unb64("dupa.8")
			So(err, ShouldNotBeNil)
		})
	})
}

func force_unhex(str string) []byte {
	if bytes, err := hex.DecodeString(str); err == nil {
		return bytes
	} else {
		panic(err)
	}
}

func TestParseCookie(t *testing.T) {
	var u, g string
	var ts uint64
	var hmac []byte
	var err error

	Convey("Defensive cookie parsing", t, func() {
		Convey("Well-formed cookie", func() {
			u, g, ts, hmac, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldBeNil)
			So(u, ShouldEqual, "maciej")
			So(g, ShouldEqual, "admin")
			So(ts, ShouldEqual, 1396349947)
			So(hmac, ShouldResemble, force_unhex("385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec"))
		})

		Convey("Malformed URL-encoding of cookie content", func() {
			u, g, ts, hmac, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2X1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "")
			So(g, ShouldEqual, "")
			So(ts, ShouldEqual, 0)
			So(len(hmac), ShouldEqual, 0)
		})

		Convey("Cookie doesn't match COOKIE_RX", func() {
			u, g, ts, hmac, err = ParseCookie("dupa.8")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "")
			So(g, ShouldEqual, "")
			So(ts, ShouldEqual, 0)
			So(len(hmac), ShouldEqual, 0)
		})

		Convey("Malformed Base64 in username", func() {
			u, g, ts, hmac, err = ParseCookie("b-WFjaWVq%2CYWRtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "")
			So(g, ShouldEqual, "")
			So(ts, ShouldEqual, 0)
			So(len(hmac), ShouldEqual, 0)
		})

		Convey("Malformed Base64 in groups", func() {
			u, g, ts, hmac, err = ParseCookie("bWFjaWVq%2CYW-RtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "maciej")
			So(g, ShouldEqual, "")
			So(ts, ShouldEqual, 0)
			So(len(hmac), ShouldEqual, 0)
		})

		Convey("Malformed timestamp", func() {
			u, g, ts, hmac, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C99999999999999999999%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "maciej")
			So(g, ShouldEqual, "admin")
			So(ts, ShouldEqual, uint64(math.MaxUint64))
			So(len(hmac), ShouldEqual, 0)
		})

		Convey("Malformed HMAC hex", func() {
			// malformed hmac hex is normally not possible because it will be
			// caught earlier by regex - we temporarily override regexp to make
			// sure that if somebody can think of invalid hex consisting of only
			// hex digits, it will get caught.
			orig_cookie_rx := COOKIE_RX
			defer func() { COOKIE_RX = orig_cookie_rx }()
			COOKIE_RX = regexp.MustCompile("^\\s*([a-zA-Z0-9_-]+),([a-zA-Z0-9_-]+),(\\d+),(.*)\\s*$")

			u, g, ts, hmac, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C1396349947%2Cdupa")
			So(err, ShouldNotBeNil)
			So(u, ShouldEqual, "maciej")
			So(g, ShouldEqual, "admin")
			So(ts, ShouldEqual, 1396349947)
			So(len(hmac), ShouldEqual, 0)
		})
	})
}

func mkSince(str string) func(time.Time) time.Duration {
	if duration, err := time.ParseDuration(str); err != nil {
		panic(err)
	} else {
		return func(_ time.Time) time.Duration {
			return duration
		}
	}
}

func TestCheckCookie(t *testing.T) {
	var c *Config
	Convey("CheckCookie", t, func() {
		c = New("swordfish")

		// We're going to override since(), let's restore it at exit
		defer func(orig_since func(time.Time) time.Duration) {
			since = orig_since
		}(since)

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
			So(res.Groups, ShouldEqual, "some,groups")
			So(res.Timestamp, ShouldEqual, 23)
			So(res.Valid, ShouldBeTrue)
			So(res.String(), ShouldEqual, "a_user[some,groups] (15m0s)")
		})

		Convey("Fails with mismatched HMAC", func() {
			since = mkSince("15m")
			res, err := c.CheckCookie("YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca", "GoConvey/different")
			So(err, ShouldNotBeNil)
			So(res.User, ShouldEqual, "a_user")
			So(res.Groups, ShouldEqual, "some,groups")
			So(res.Timestamp, ShouldEqual, 23)
			So(res.Valid, ShouldBeFalse)
			So(res.String(), ShouldEqual, "!a_user[some,groups] (15m0s)")
		})

		Convey("Fails with correct stale cookie", func() {
			since = mkSince("96h")
			res, err := c.CheckCookie("YV91c2Vy,c29tZSxncm91cHM,23,686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca", "GoConvey")
			So(err, ShouldNotBeNil)
			So(res.User, ShouldEqual, "a_user")
			So(res.Groups, ShouldEqual, "some,groups")
			So(res.Timestamp, ShouldEqual, 23)
			So(res.Valid, ShouldBeFalse)
			So(res.String(), ShouldEqual, "!a_user[some,groups] (96h0m0s)")
		})
	})
}
