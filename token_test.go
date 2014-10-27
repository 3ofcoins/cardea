package cardea

import "testing"
import . "github.com/smartystreets/goconvey/convey"

import "crypto/hmac"
import "crypto/sha256"
import "encoding/hex"
import "encoding/base64"
import "fmt"
import "strings"

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

func computeHMAC(secret []byte, data string) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func basic_auth(s string) string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(s)))
}

func TestToken(t *testing.T) {
	var secret = []byte("SWORDFISH")
	var hmac_extra = "For instance, User-Agent header"
	var token *Token
	var err error

	signed_with_glue := func(glue, unsigned string) string {
		return strings.Join(
			[]string{
				unsigned,
				computeHMAC(
					secret,
					strings.Join([]string{unsigned, b64(hmac_extra)}, glue))},
			glue)
	}

	Convey("Token deserialization", t, func() {
		Convey("Malformed cookie", func() {
			token, err = DeserializeToken(secret, hmac_extra,
				"LOREM IPSUM DOLOR SIT AMET")
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
		})

		Convey("Legacy format", func() {
			Convey("HMAC mismatch", func() {
				token, err = DeserializeToken(secret, hmac_extra,
					"bWFjaWVq,YWRtaW4,1396349947,385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			deserialize := func(serialized string) (*Token, error) {
				return DeserializeToken(
					secret,
					hmac_extra,
					signed_with_glue(",", serialized))
			}

			Convey("Well-formed cookie", func() {
				token, err = deserialize("bWFjaWVq,YWRtaW4,1396349947")
				So(err, ShouldBeNil)
				So(token.User, ShouldEqual, "maciej")
				So(token.Groups, ShouldResemble, []string{"admin"})
				So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			})

			Convey("Well-formed cookie, multiple groups", func() {
				token, err = deserialize("bWFjaWVq,Zm9vLGJhcixiYXoscXV1eA,1396349947")
				So(err, ShouldBeNil)
				So(token.User, ShouldEqual, "maciej")
				So(token.Groups, ShouldResemble, []string{"foo", "bar", "baz", "quux"})
				So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			})

			Convey("Malformed Base64 in username", func() {
				token, err = deserialize("b*WFjaWVq,YWRtaW4,1396349947")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			Convey("Malformed Base64 in groups", func() {
				token, err = deserialize("bWFjaWVq,YW*RtaW4,1396349947")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			Convey("Malformed timestamp", func() {
				token, err = deserialize("bWFjaWVq,YWRtaW4,99999999999999999999")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})
		})

		Convey("Cardea format", func() {
			Convey("HMAC mismatch", func() {
				token, err = DeserializeToken(secret, hmac_extra,
					"maciej:g=admin&t=1396349947#385291ccf3679233c2480c3a011f0fe02f7a91ba2e4d28d8023f35ac56e05dec")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			deserialize := func(serialized string) (*Token, error) {
				return DeserializeToken(
					secret,
					hmac_extra,
					signed_with_glue("#", serialized))
			}

			Convey("Well-formed cookie", func() {
				token, err = deserialize("maciej:g=admin&t=1396349947")
				So(err, ShouldBeNil)
				So(token.User, ShouldEqual, "maciej")
				So(token.Groups, ShouldResemble, []string{"admin"})
				So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			})

			Convey("Well-formed cookie, multiple groups", func() {
				token, err = deserialize("maciej:g=foo&g=bar&g=baz&g=quux&t=1396349947")
				So(err, ShouldBeNil)
				So(token.User, ShouldEqual, "maciej")
				So(token.Groups, ShouldResemble, []string{"foo", "bar", "baz", "quux"})
				So(token.Timestamp.Unix(), ShouldEqual, 1396349947)
			})

			Convey("Malformed query string", func() {
				token, err = deserialize("maciej:q=%X8")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			Convey("Malformed timestamp", func() {
				token, err = deserialize("maciej:t=dupa")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})

			Convey("Multiple timestamps", func() {
				token, err = deserialize("maciej:t=1&t=2")
				So(err, ShouldNotBeNil)
				So(token, ShouldBeNil)
			})
		})
	})

	Convey("Cookie parsing", t, func() {
		Convey("Well-escaped cookie", func() {
			token, err := ParseCookie(secret, hmac_extra, signed_with_glue("#", "maciej:t=1396349947"))
			So(err, ShouldBeNil)
			So(token.User, ShouldEqual, "maciej")
		})
		Convey("Malformed cookie", func() {
			token, err := ParseCookie(secret, hmac_extra, signed_with_glue("#", "maciej:t=1396349947%2X"))
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
		})
	})

	Convey("Authorization header parsing", t, func() {
		Convey("Well-escaped header", func() {
			token, err := ParseAuthorization(secret, hmac_extra, basic_auth(signed_with_glue("#", "maciej:t=1396349947")))
			So(err, ShouldBeNil)
			So(token.User, ShouldEqual, "maciej")
		})
		Convey("Not Basic auth", func() {
			// Change "Basic …" to "NotBasic …"
			token, err := ParseAuthorization(secret, hmac_extra, "Not"+basic_auth(signed_with_glue("#", "maciej:t=1396349947")))
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
		})
		Convey("Malformed payload", func() {
			// Malform base64 by adding invalid padding
			token, err := ParseAuthorization(secret, hmac_extra, basic_auth(signed_with_glue("#", "maciej:t=1396349947"))+"==")
			So(err, ShouldNotBeNil)
			So(token, ShouldBeNil)
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

// Some code paths can't be tested with outside calls (e.g. malformed
// hex-encoded received HMAC won't ever be seen by check_hmac, as
// regexp will reject it earlier on). Let's test the lower layers, and
// keep coverage green.
func TestCoverage(t *testing.T) {
	Convey("Cover inaccessible code paths", t, func() {
		Convey("Malformed received HMAC hex", func() {
			So(check_hmac([]byte("SWORDFISH"), "xyzzy", "whatever"), ShouldBeFalse)
		})
	})
}
