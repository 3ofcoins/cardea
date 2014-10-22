package cardea

import "regexp"


var TOKEN_RX = regexp.MustCompile("\\A[\\t-\\n\\f-\\r ]*(?P<USERNAME>[\\--9A-Z_a-z]+)(?P<PAYLOAD>,(?P<LEGACY_GROUPS>[\\--9A-Z_a-z]+)?,(?P<LEGACY_TIMESTAMP>[0-9]+),|:(?:(?P<FORMAT>[0-9A-Z_a-z]+)\\?)?(?P<QUERY>[^#]+)#)(?P<HMAC>[0-9a-f]+)[\\t-\\n\\f-\\r ]*(?-m:$)")

const (
  _ = iota
  TOKEN_RX_USERNAME = iota
  TOKEN_RX_PAYLOAD = iota
  TOKEN_RX_LEGACY_GROUPS = iota
  TOKEN_RX_LEGACY_TIMESTAMP = iota
  TOKEN_RX_FORMAT = iota
  TOKEN_RX_QUERY = iota
  TOKEN_RX_HMAC = iota
)
