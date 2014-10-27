Cardea
======

> She is the goddess of the hinge: by her divine power she opens what
> is closed, and closes what is open. — Ovid, *Fasti*

Cookie Format
-------------

The cookie format is described in an Augmented Backus-Naur Form
(ABNF), as described in
[rfc4234](http://www.ietf.org/rfc/rfc4234.txt), with two additions:

 - The terminal expression notation `/REGEXP/` means a string matching
   a Perl-compatible regular expression. The whole string should match
   the expression, so the anchors are ommitted for clarity; the actual
   regular expression would be `/^(?:REGEXP)$/`. Modificators
   (case-insensitiveness, multi-line match, etc) are also allowed
   (e.g. `/ReGeXp/i`).
 - Encoded values are defined in a "function notation",
   e.g. `urlencode(X) = <any string X, percent-encoded as specified in
   RFC3986>`

Some base token definitions that will be used later on:

    WORD           = /[a-z_][a-z0-9_.-]*/i ; alphanumeric word, with reasonable punctuation allowed
    NUMBER         = /\d+/
    CRLF           = "\r\n"
    ANY            = <arbitrary string, used as input for encoding or HMAC>
    URL(ANY)       = <arbitrary string, percent-escaped as per RFC3986 section 2.1>
    B64(ANY)       = <arbitrary string, escaped as URL-safe Base64 without padding, as per RFC4648 section 5>
    HEX(ANY)       = <arbitrary string, likely binary, hex-encoded>

    HMAC(HMAC_K,HMAC_M) = <HMAC-SHA256 of message HMAC_M, using key HMAC_K>
    HMAC_K         = ANY
    HMAC_M         = ANY

To compute the HMAC, a token is combined with **HMAC Extras**. It is an
out-of-band value, based on request parameters, and designed to
prevent cookie stealing / replay attacks. Originally it was value of
the *User-Agent* HTTP header, postprocessed in a very stupid way. In
Cardea, parameters used are open to interpretation: it can include
User-Agent, client IP, "browser signature" based on headers, or an
actual out-of-base value, such as a time-based token. Multiple values
can be used as extras; if this is the case, they are joined together
by CRLF (`"\r\n"`):

    hmac_extras    = B64( hmac_extra *( CRLF hmac_extra ) )
    hmac_extra     = ANY

HMAC is computed on a **payload**, which is cookie's text combined
with the extras:
    
    secret         = ANY ; preshared secret, provided in configuration
    hmac           = HEX(HMAC(secret, payload))

There are two cookie formats, "Modern" (Cardea's), and optional
"Legacy" (Odin Authenticator's). The two formats imply slightly
different HMAC payload structure. As legacy format is optional, the
base definition includes only modern format:

    cookie         = modern_cookie
    payload        = modern_payload

The default and recommended format for Cookie is the modern format. It
is designed to resemble an URI (and utilise existing parsing code),
and to be compatible with HTTP Basic authentication header. There is
also an optional legacy format, which will be introduced later on.

    modern_cookie  = modern_token "#" hmac
    modern_payload = modern_token "#" hmac_extras
    modern_token   = username ":" [ format "?" ] query
    username       = WORD
    format         = WORD

    ; The "query" is a set of key-value pairs like in HTTP query
    ; string, which seems to lack a formal definition, so we
    ; (re)define it here.
    query          = kv_pair ( & kv_pair )*
    kv_pair        = key "=" value
    key            = WORD
    value          = URL(ANY)

### Legacy Format

Cardea also supports legacy format, used by her predecessor, Odin
Authenticator. It should be used only in environment that mixes Odin
Authenticator and Cardea. It is not flexible or extensible, and
User-Agent header handling is, to put it mild, wrong. This format is
disabled by default.

    cookie        /= legacy_cookie
    payload       /= legacy_payload
    
    legacy_cookie  = legacy_token "," hmac
    legacy_payload = legacy_token "," hmac_extras
    legacy_token   = B64(username) "," B64(legacy_groups) "," legacy_timestamp
    legacy_timestamp = NUMBER ; UNIX time as decimal number

Odin Authenticator, which aimed at (at least partial) compatibility
with GodAuth script, performs following transformation on User-Agent
header before using it as a HMAC extra:

 - If the header includes substring "AppleWebKit", a hard-coded value
   "StupidAppleWebkitHacksGRRR" is used instead (don't ask)
 - `s/ FirePHP\/\d+\.\d+//`
