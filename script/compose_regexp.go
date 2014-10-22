package main

import "bytes"
import "fmt"
import "regexp"
import "regexp/syntax"
import "strings"
import "text/template"

// these imports are for main() only
import "io/ioutil"
import "os"

type Element string

func rxf(format string, a ...interface{}) Element {
	return Element(fmt.Sprintf(format, a...))
}

// Wrap rx in a capturing group
func Capture(name, rx Element) Element {
	return rxf("(?P<%s>%s)", name, rx)
}

// Wrap rx in a non-capturing group
func Group(rx Element) Element {
	return rxf("(?:%s)", rx)
}

// Wrap all rxs in capturing groups
func join(glue string, rxs []Element) Element {
	wrapped := make([]string, len(rxs))
	for i, rx := range rxs {
		wrapped[i] = string(Group(rx))
	}
	return rxf("(?:%s)", strings.Join(wrapped, glue))
}

// Regexp alternation
func Alternation(rxs ...Element) Element {
	return Group(join("|", rxs))
}

// Sequence of regexps
func Sequence(rxs ...Element) Element {
	return Group(join("", rxs))
}

// Modify (with "*", "*?", "+", "?" etc)
func Mod(rx Element, modifier string) Element {
	return Group(rx) + Element(modifier)
}

func Optional(rxs ...Element) Element {
	return Mod(Sequence(rxs...), "?")
}

func Any(rxs ...Element) Element {
	return Mod(Sequence(rxs...), "*")
}

func Some(rxs ...Element) Element {
	return Mod(Sequence(rxs...), "+")
}

func Literal(str string) Element {
	return Element(regexp.QuoteMeta(str))
}

type RxVar struct {
	Name string
	*syntax.Regexp
}

func NewRxVar(name string, elements ...Element) (*RxVar, error) {
	rx, err := syntax.Parse(string(Sequence(elements...)), syntax.Perl)
	if err != nil {
		return nil, err
	}
	return &RxVar{name, rx.Simplify()}, nil
}

func MustRxVar(name string, elements ...Element) *RxVar {
	if rxv, err := NewRxVar(name, elements...); err != nil {
		panic(err)
	} else {
		return rxv
	}
}

var rxvTemplate = template.Must(template.New("RxVar").Parse(`var {{.Name}} = regexp.MustCompile({{.Regexp|printf "%q"}})
{{if gt .MaxCap 0}}
const ({{$prefix := .Name}}
{{range .CapNames}}  {{if eq . ""}}_{{else}}{{$prefix}}_{{.}}{{end}} = iota
{{end}})
{{end}}`))

func (rxv *RxVar) String() string {
	buf := bytes.NewBuffer(nil)
	if err := rxvTemplate.Execute(buf, rxv); err != nil {
		panic(err)
	}
	return buf.String()
}

func RxFile(pkg string, rxvs ...*RxVar) string {
	pieces := make([]string, len(rxvs)+1)
	pieces[0] = fmt.Sprintf("package %s\n\nimport \"regexp\"\n", pkg)
	for i, rxv := range rxvs {
		pieces[i+1] = rxv.String()
	}
	return strings.Join(pieces, "\n")
}

var (
	HexDigit        = Element(`[0-9a-f]`)
	HexNumber       = Some(HexDigit)
	Word            = Element(`\w+`)
	DecimalNumber   = Element(`\d+`)
	Base64Char      = Element(`[./a-zA-Z0-9_-]`)
	Base64          = Sequence(Some(Base64Char), Any(`=`))
	Beginning       = Element(`^`)
	End             = Element(`$`)
	AllowWhitespace = Element(`\s*`)
)

func main() {
	b64 := Some(Base64Char) // Base64 stripped of `=` padding
	src := RxFile("cardea",
		MustRxVar("TOKEN_RX",
			Beginning,
			AllowWhitespace,
			Capture("USERNAME", b64), // Capture username (may be base64)
			Capture("PAYLOAD",
				Alternation(
					Sequence(
						// Legacy (Odin) payload
						Literal(","),
						Optional(Capture("LEGACY_GROUPS", b64)),
						Literal(","),
						Capture("LEGACY_TIMESTAMP", DecimalNumber),
						Literal(","),
					),
					Sequence(
						// Modern (Cardea URL-like) payload
						Literal(":"),
						Optional(
							Capture("FORMAT", Word),
							Literal("?")),
						Capture("QUERY", `[^#]+`),
						Literal("#"),
					),
				)),
			Capture("HMAC", HexNumber),
			AllowWhitespace,
			End,
		))

	if len(os.Args) < 2 {
		fmt.Println(src)
	} else {
		if err := ioutil.WriteFile(os.Args[1]+".go", []byte(src), os.ModePerm); err != nil {
			panic(err)
		}
	}
}
