package converter

import (
	"fmt"
	"strings"

	"github.com/therecipe/qt/internal/binding/parser"
)

func GoHeaderName(f *parser.Function) (o string) {

	if f.SignalMode == "callback" {
		var tmp = fmt.Sprintf("callback%v%v", f.Class(), strings.Replace(strings.Title(f.Name), "~", "Destroy", -1))
		if f.Overload {
			tmp += f.OverloadNumber
		}
		return tmp
	}

	if f.Static {
		o += fmt.Sprintf("%v_", strings.Split(f.Fullname, "::")[0])
	}

	switch f.Meta {
	case "constructor":
		{
			o += "New"
		}

	case "destructor":
		{
			o += "Destroy"
		}
	}

	o += f.SignalMode

	o += strings.Title(f.Name)

	o += f.TemplateMode

	if f.Overload {
		o += f.OverloadNumber
	}

	if strings.ContainsAny(o, "&<>=/!()[]{}-^|*+-") || strings.Contains(o, "Operator") {
		f.Access = "unsupported_GoHeaderName"
		return f.Access
	}

	return strings.Replace(o, "~", "", -1)
}

func CppHeaderName(f *parser.Function) string {
	return fmt.Sprintf("%v_%v", strings.Split(f.Fullname, "::")[0], GoHeaderName(f))
}

func GoHeaderOutput(f *parser.Function) string {

	var value = f.Output

	if parser.ClassMap[f.Class()].Module == "main" && f.TempOutput != "void" && f.SignalMode == "callback" {
		return cgoType(f, f.TempOutput)
	}

	if f.Meta == "constructor" && f.Output == "" {
		value = f.Name
	}

	var o = goType(f, value)
	if isClass(o) {
		return fmt.Sprintf("*%v", o)
	}
	return o
}

func CppHeaderOutput(f *parser.Function) string {

	var value = f.Output

	if f.Meta == "constructor" && f.Output == "" {
		value = f.Name
	}

	return cppType(f, value)
}

func GoHeaderInput(f *parser.Function) (o string) {

	if f.SignalMode == "callback" {
		o += "ptr unsafe.Pointer, ptrName *C.char"
		for _, p := range f.Parameters {
			if v := cgoType(f, p.Value); v != "" {
				o += fmt.Sprintf(", %v %v", cleanName(p.Name), v)
			}
		}
		return strings.TrimSuffix(o, ", ")
	}

	if (f.Meta == "signal" || strings.Contains(f.Virtual, "impure")) && f.SignalMode == "Connect" {
		o += "f func ("
	}

	if (f.Meta == "signal" || strings.Contains(f.Virtual, "impure")) && f.SignalMode != "Connect" {
		if strings.Contains(f.Virtual, "impure") && f.SignalMode == "" {
		} else {
			return
		}
	}

	for _, p := range f.Parameters {
		if v := goType(f, p.Value); v != "" {
			if isClass(v) {
				if (f.Meta == "signal" || strings.Contains(f.Virtual, "impure")) && f.SignalMode == "Connect" {
					o += fmt.Sprintf("%v *%v, ", cleanName(p.Name), v)
				} else {
					o += fmt.Sprintf("%v %v_ITF, ", cleanName(p.Name), v)
				}
			} else {
				o += fmt.Sprintf("%v %v, ", cleanName(p.Name), v)
			}
		} else {
			f.Access = "unsupported_GoHeaderInput"
			return f.Access
		}
	}

	o = strings.TrimSuffix(o, ", ")

	if (f.Meta == "signal" || strings.Contains(f.Virtual, "impure")) && f.SignalMode == "Connect" {
		o += ")"
	}

	if parser.ClassMap[f.Class()].Module == "main" && f.TempOutput != "void" && f.SignalMode == "Connect" {
		if isClass(goType(f, f.TempOutput)) {
			o += " *" + goType(f, f.TempOutput)
		} else {
			o += " " + goType(f, f.TempOutput)
		}
	}

	return
}

func GoHeaderInputSignalFunction(f *parser.Function) (o string) {

	o += "func ("

	for _, p := range f.Parameters {
		if v := goType(f, p.Value); v != "" {
			if isClass(v) {
				o += fmt.Sprintf("*%v, ", v)
			} else {
				o += fmt.Sprintf("%v, ", v)
			}
		} else {
			f.Access = "unsupported_GoHeaderInputSignalFunction"
			return f.Access
		}
	}

	o = strings.TrimSuffix(o, ", ")

	o += ")"

	if parser.ClassMap[f.Class()].Module == "main" && f.TempOutput != "void" && f.SignalMode == "callback" {
		if isClass(goType(f, f.TempOutput)) {
			o += " *" + goType(f, f.TempOutput)
		} else {
			o += " " + goType(f, f.TempOutput)
		}
	}

	return
}

func CppHeaderInput(f *parser.Function) (o string) {
	if !(f.Static || f.Meta == "constructor") {
		o += "void* ptr, "
	}

	if f.Meta == "signal" {
		return strings.TrimSuffix(o, ", ")
	}

	for _, p := range f.Parameters {
		if v := cppType(f, p.Value); !(v == "") {
			o += fmt.Sprintf("%v %v, ", v, cleanName(p.Name))
		} else {
			f.Access = "unsupported_CppHeaderInput"
			return f.Access
		}
	}

	return strings.TrimSuffix(o, ", ")
}