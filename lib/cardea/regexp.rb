module Cardea
  TOKEN_RX = Regexp.compile("\\A[\\t-\\n\\f-\\r ]*(?<USERNAME>[\\--9A-Z_a-z]+)(?<PAYLOAD>,(?<LEGACY_GROUPS>[\\--9A-Z_a-z]*),(?<LEGACY_TIMESTAMP>[0-9]+),|:(?:(?<FORMAT>[0-9A-Z_a-z]+)\\?)?(?<QUERY>[^#]+)#)(?<HMAC>[0-9a-f]+|[\\--9A-Z_a-z]+)[\\t-\\n\\f-\\r ]*(?-m:$)");
end
