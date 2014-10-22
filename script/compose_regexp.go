package main

import "fmt"

import . "github.com/3ofcoins/go-misc/regexp-composer"

func main() {
	fmt.Print(
		NewFile("cardea",
			MustVariable("TOKEN_RX",
				TrimAnchor(
					Capture("USERNAME", Some(Base64Char)), // Username may be base64-encoded
					Capture("PAYLOAD", // Payload (i.e. whatever's HMACed)
						Alternation(
							Sequence( // Legacy (Odin) payload
								Literal(","),
								Capture("LEGACY_GROUPS", Any(Base64Char)),
								Literal(","),
								Capture("LEGACY_TIMESTAMP", DecimalNumber),
								Literal(","),
							),
							Sequence( // Modern (Cardea URL-like) payload
								Literal(":"),
								Optional(
									// Optional "FORMAT?" prefix to indicate encryption
									// or encoding of payload
									Capture("FORMAT", Word),
									Literal("?")),
								Capture("QUERY", `[^#]+`), // URL query parameters
								Literal("#"),
							),
						)),
					Capture("HMAC", HexNumber),
				))))
}
