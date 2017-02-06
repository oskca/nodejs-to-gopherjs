package main

import (
	"fmt"
	"html"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Base struct {
	TextRaw   string `json:"textRaw,omitempty"`
	Type      string `json:"type,omitempty"`
	Name      string `json:"name"`
	Desc      string `json:"desc,omitempty"`
	ShortDesc string `json:"shortDesc,omitempty"`
	// Meta      string `json:"meta,-"`
}

func basicType(typ string) string {
	switch typ {
	case "":
		return "string"
	case "Integer", "INTEGER":
		return "int64"
	case "Number", "NUMBER":
		return "float64"
	case "Function", "FUNCTION":
		return "func()"
	case "Boolean", "BOOLEAN":
		return "bool"
	}
	return "*js.Object"
}

var replacer = strings.NewReplacer(
	"`", "",
	"\"", "",
	".", "",
	"_", "",
	" ", "",
	"\t", "",
	"(", "",
	")", "",
	// ",", "",
)

func (b *Base) gosym() string {
	name := replacer.Replace(b.Name)
	letter := strings.ToUpper(string(name[0]))
	name = name[1:len(name)]
	return letter + name
}

func (b *Base) decl() string {
	return fmt.Sprintf("%s %s",
		b.gosym(),
		basicType(b.Type),
	)
}

func purify(desc string) string {
	desc = strings.Replace(desc, "\n", "\n//", -1)
	desc = html.UnescapeString(desc)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(desc))
	if err != nil {
		println(err.Error())
		return desc
	}
	return doc.Text()
}

func (b *Base) comment() string {
	ms := []string{}
	if b.TextRaw != "" {
		ms = append(ms, b.TextRaw)
	}
	if b.ShortDesc != "" {
		ms = append(ms, purify(b.ShortDesc))
	}
	if b.Desc != "" {
		ms = append(ms, purify(b.Desc))
	}
	out := ""
	if len(ms) > 0 {
		out += "\n//" + b.gosym() + " docs"
		if len(ms) > 0 {
			out += "\n//" + strings.Join(ms, "\n//")
		}
	}
	return out
}

type Property struct {
	Base
}

func (p *Property) decl() string {
	return fmt.Sprintf("%s `js:\"%s\"`",
		p.Base.decl(),
		p.Name,
	)
}

type Event struct {
	Base
}

func (e *Event) decl() string {
	return fmt.Sprintf(`Evt%s = "%s"`,
		e.gosym(),
		e.Name,
	)
}

type Param struct {
	*Base
	Optional bool `json:"optional,omitempty"`
}

func (p *Param) decl() string {
	return fmt.Sprintf("%s %s",
		p.Name,
		basicType(p.Type),
	)
}

type Return struct {
	*Base
}

type Signature struct {
	Return *Return  `json:"return,omitempty"`
	Params []*Param `json:"params,omitempty"`
}

func (s *Signature) decl() string {
	params := []string{}
	for i := 0; i < len(s.Params); i++ {
		params = append(params, s.Params[i].decl())
	}
	out := fmt.Sprintf("func(%s) ",
		strings.Join(params, ","),
	)
	if s.Return != nil {
		out += "(" + s.Return.decl() + ")"
	}
	return out
}
func (s *Signature) comment() string {
	return ""
}

type Method struct {
	Base
	Signatures []*Signature `json:"signatures,omitempty"`
}

func (m *Method) decl() string {
	if len(m.Signatures) > 1 {
		m.Signatures = []*Signature{m.Signatures[0]}
	}
	return fmt.Sprintf("%s %s `js:\"%s\"`",
		strings.Title(m.Name),
		declSlice(m.Signatures),
		m.Name,
	)
}

type Class struct {
	Base
	Methods    []*Method   `json:"methods"`
	Properties []*Property `json:"properties"`
}

func (c *Class) decl() string {
	if len(c.Methods)+len(c.Properties) == 0 {
		return ""
	}
	out := fmt.Sprintf("type %s struct {\n\t", c.gosym())
	out += "*js.Object\n\t"
	out += declSlice(c.Properties)
	out += "\n\t"
	out += declSlice(c.Methods)
	out += "\n}\n"
	return out
}

type Module struct {
	Base
	DisplayName string      `json:"displayName,omitempty"`
	Events      []*Event    `json:"events,omitempty"`
	Properties  []*Property `json:"properties,omitempty"`
	Methods     []*Method   `json:"methods,omitempty"`
	Classes     []*Class    `json:"classes,omitempty"`
	Modules     []*Module   `js:"modules,omitempty"`
}

func (m *Module) decl() string {
	childs := len(m.Events) +
		len(m.Properties) +
		len(m.Methods) +
		len(m.Classes)
	if childs == 0 {
		if len(m.Modules) == 0 {
			return ""
		} else {
			return declSlice(m.Modules)
		}
	}
	// output
	out := ""
	// modules
	if m.Modules != nil {
		out += declSlice(m.Modules)
	}
	// classes
	if m.Classes != nil {
		out += declSlice(m.Classes)
		out += "\n\t"
	}
	// events
	if m.Events != nil {
		out += "\nconst (\n\t"
		out += declSlice(m.Events)
		out += "\n)\n"
	}
	// module content
	out += fmt.Sprintf("\ntype %s struct {\n\t", m.gosym())
	out += "*js.Object\n\t"
	out += declSlice(m.Properties)
	out += "\n\t"
	out += declSlice(m.Methods)
	out += "\n}\n"
	return out
}

type ApiFile struct {
	Source  string    `js:"source"`
	Modules []*Module `js:"modules,omitempty"`
	Globals []*Module `js:"globals,omitempty"`
}

func (a *ApiFile) decl() string {
	out := "//" + a.Source
	out += "\npackage nodejs\n"
	out += "import (\n\t\"github.com/gopherjs/gopherjs/js\"\n)"
	out += declSlice(a.Modules)
	out += declSlice(a.Globals)
	return out
}
