package main

import "fmt"

import rx "github.com/3ofcoins/go-misc/regexp-composer"

func main() {
	b64 := rx.Some(rx.Base64Char) // Base64 stripped of `=` padding
	fmt.Print(
		rx.NewFile("cardea",
			rx.MustVariable("TOKEN_RX",
				rx.Beginning,
				rx.AnyWhitespace,
				rx.Capture("USERNAME", b64), // Capture username (may be base64)
				rx.Capture("PAYLOAD",
					rx.Alternation(
						rx.Sequence(
							// Legacy (Odin) payload
							rx.Literal(","),
							rx.Optional(rx.Capture("LEGACY_GROUPS", b64)),
							rx.Literal(","),
							rx.Capture("LEGACY_TIMESTAMP", rx.DecimalNumber),
							rx.Literal(","),
						),
						rx.Sequence(
							// Modern (Cardea URL-like) payload
							rx.Literal(":"),
							rx.Optional(
								rx.Capture("FORMAT", rx.Word),
								rx.Literal("?")),
							rx.Capture("QUERY", `[^#]+`),
							rx.Literal("#"),
						),
					)),
				rx.Capture("HMAC", rx.HexNumber),
				rx.AnyWhitespace,
				rx.End,
			)))
}
