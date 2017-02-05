package main

import (
	"fmt"
	"reflect"
	"strings"
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

func (b *Base) comment() string {
	ms := []string{}
	if b.TextRaw != "" {
		ms = append(ms, b.TextRaw)
	}
	if b.ShortDesc != "" {
		ms = append(ms, b.ShortDesc)
	}
	if b.Desc != "" {
		ms = append(ms, b.Desc)
	}
	out := "//" + strings.Title(b.Name) + "docs"
	if len(ms) > 0 {
		out += "\n//" + strings.Join(ms, "\n//")
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
		out += s.Return.decl()
	}
	return out
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

type decler interface {
	decl() string
}

func declSlice(ar interface{}) string {
	v := reflect.ValueOf(ar)
	if v.IsNil() {
		return ""
	}
	var t []decler
	ret := reflect.MakeSlice(reflect.TypeOf(t), 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		ret = reflect.Append(ret, v.Index(i))
	}
	// declers to string
	ds := ret.Interface().([]decler)
	ss := []string{}
	for i := 0; i < len(ds); i++ {
		ss = append(ss, ds[i].decl())
	}
	return strings.Join(ss, "\n\t")
}

func (c *Class) decl() string {
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
	out := ""
	// events
	if m.Events != nil {
		out += "\nconst (\n\t"
		out += declSlice(m.Events)
		out += "\n)\n"
	}
	// module
	out += fmt.Sprintf("\ntype %s struct {\n\t", m.gosym())
	out += "*js.Object\n\t"
	out += declSlice(m.Properties)
	out += "\n\t"
	out += declSlice(m.Methods)
	out += "\n}\n"
	// classes
	if m.Classes != nil {
		out += declSlice(m.Classes)
		out += "\n\t"
	}
	// modules
	if m.Modules != nil {
		out += declSlice(m.Modules)
	}
	return out
}

type ApiFile struct {
	Source  string    `js:"source"`
	Modules []*Module `js:"modules,omitempty"`
	Globals []*Module `js:"globals,omitempty"`
}

func (a *ApiFile) decl() string {
	out := "//" + a.Source
	out += declSlice(a.Modules)
	out += declSlice(a.Globals)
	return out
}
