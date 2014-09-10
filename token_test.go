package cardea

import "testing"
import . "github.com/smartystreets/goconvey/convey"

import "encoding/hex"
import "regexp"

func force_unb64(enc string) string {
	if raw, err := unb64(enc); err == nil {
		return raw
	} else {
		panic(err)
	}
}

func force_unhex(str string) []byte {
	if bytes, err := hex.DecodeString(str); err == nil {
		return bytes
	} else {
		panic(err)
	}
}

func TestParseCookie(t *testing.T) {
	var token *Token
	var err error

	Convey("Defensive cookie parsing", t, func() {
		Convey("Well-formed cookie", func() {
			token, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldBeNil)
			So(token.User, ShouldEqual, "maciej")
			So(token.Groups, ShouldEqual, "admin")
			So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			So(token.HMAC, ShouldResemble, force_unhex("385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec"))
		})

		Convey("Malformed URL-encoding of cookie content", func() {
			token, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2X1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
		})

		Convey("Cookie doesn't match COOKIE_RX", func() {
			token, err = ParseCookie("dupa.8")
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
		})

		Convey("Malformed Base64 in username", func() {
			token, err = ParseCookie("b-WFjaWVq%2CYWRtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(token.User, ShouldEqual, "")
			So(token.Groups, ShouldEqual, "")
			So(token.Timestamp.IsZero(), ShouldBeTrue)
			So(len(token.HMAC), ShouldEqual, 0)
		})

		Convey("Malformed Base64 in groups", func() {
			token, err = ParseCookie("bWFjaWVq%2CYW-RtaW4%2C1396349947%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(token.User, ShouldEqual, "maciej")
			So(token.Groups, ShouldEqual, "")
			So(token.Timestamp.IsZero(), ShouldBeTrue)
			So(len(token.HMAC), ShouldEqual, 0)
		})

		Convey("Malformed timestamp", func() {
			token, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C99999999999999999999%2C385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
			So(err, ShouldNotBeNil)
			So(token.User, ShouldEqual, "maciej")
			So(token.Groups, ShouldEqual, "admin")
			So(token.Timestamp.IsZero(), ShouldBeTrue)
			So(len(token.HMAC), ShouldEqual, 0)
		})

		Convey("Malformed HMAC hex", func() {
			// malformed hmac hex is normally not possible because it will be
			// caught earlier by regex - we temporarily override regexp to make
			// sure that if somebody can think of invalid hex consisting of only
			// hex digits, it will get caught.
			orig_cookie_rx := COOKIE_RX
			defer func() { COOKIE_RX = orig_cookie_rx }()
			COOKIE_RX = regexp.MustCompile("^\\s*([a-zA-Z0-9_-]+),([a-zA-Z0-9_-]+),(\\d+),(.*)\\s*$")

			token, err = ParseCookie("bWFjaWVq%2CYWRtaW4%2C1396349947%2Cdupa")
			So(err, ShouldNotBeNil)
			So(token.User, ShouldEqual, "maciej")
			So(token.Groups, ShouldEqual, "admin")
			So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			So(len(token.HMAC), ShouldEqual, 0)
		})
	})
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
