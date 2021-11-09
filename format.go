package formatting

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
)

func Format(v interface{}) string {
	buf := &bytes.Buffer{}
	_ = ReflectiveFormatToWriter(reflect.ValueOf(v), "    ", 0, buf)
	return buf.String()
}

func ReflectiveFormatToWriter(refV reflect.Value, prefix string, indent int, target io.Writer) error {
	refType := refV.Type()

	switch refV.Kind() {
	case reflect.Invalid:
		return fprint(target, "<invalid>")

	case reflect.Bool:
		return fprintf(target, "bool{%t}", refV.Bool())

	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return fprintf(target, "%s{%v}", refType.Name(), refV.Int())

	case
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr:
		return fprintf(target, "%s{%v}", refType.Name(), refV.Uint())

	case
		reflect.Float32,
		reflect.Float64:
		return fprintf(target, "%s{%v}", refType.Name(), refV.Float())

	case reflect.String:
		return fprintf(target, "string{%q}", refV.String())

	case reflect.Complex64, reflect.Complex128:
		return fprintf(target, "complex{%v}", refV.Complex())

	case reflect.Interface:
		return ReflectiveFormatInterface(refV, prefix, indent, target, refType)

	case reflect.Ptr:
		return ReflectiveFormatPointer(refV, prefix, indent, target)

	case reflect.Array, reflect.Slice:
		return ReflectiveFormatArrayOrSlice(refV, prefix, indent, target)

	case reflect.Map:
		return ReflectiveFormatMap(refV, prefix, indent, target)

	case reflect.Func:
		if refV.IsNil() {
			return fprint(target, "("+refType.String()+")(<nil>)")
		} else {
			return ReflectiveFormatFunc(refV, refType, target)
		}

	case reflect.Chan:
		return ReflectiveFormatChan(refV, target, refType)

	case reflect.Struct:
		return ReflectiveFormatStruct(refV, prefix, indent, target, refType)

	case reflect.UnsafePointer:
		return fprintf(target, "Pointer(%x)", refV.Pointer())

	default:
		return fprintf(target, "UnknownKind(%v)", refV.Kind())
	}
}

func ReflectiveFormatStruct(refV reflect.Value, prefix string, indent int, target io.Writer, refType reflect.Type) error {
	numFields := refType.NumField()

	fes := make([]func() error, 0, 4+5*numFields)
	fes = append(fes, func() error { return fprint(target, refType.String()) })
	fes = append(fes, func() error { return fprintln(target, "{") })

	for i := 0; i < numFields; i++ {
		var idxField = i
		fes = append(fes, func() error { return doIndent(target, prefix, indent+1) })
		fes = append(fes, func() error { return fprint(target, refType.Field(idxField).Name) })
		fes = append(fes, func() error { return fprint(target, ": ") })
		fes = append(fes, func() error { return ReflectiveFormatToWriter(refV.Field(idxField), prefix, indent+1, target) })
		fes = append(fes, func() error { return fprintln(target, ",") })
	}

	fes = append(fes, func() error { return doIndent(target, prefix, indent) })
	fes = append(fes, func() error { return fprint(target, "}") })

	return RunAllBeforeFirstError(fes...)
}

func ReflectiveFormatChan(refV reflect.Value, target io.Writer, refType reflect.Type) error {
	fes := make([]func() error, 0, 7)
	fes = append(fes, func() error { return fprint(target, "chan ") })
	fes = append(fes, func() error { return fprint(target, refType.Elem().String()) })
	fes = append(fes, func() error { return fprint(target, "{len=") })
	fes = append(fes, func() error { return fprint(target, refV.Len()) })
	fes = append(fes, func() error { return fprint(target, ",cap=") })
	fes = append(fes, func() error { return fprint(target, refV.Cap()) })
	fes = append(fes, func() error { return fprint(target, "}") })

	return RunAllBeforeFirstError(fes...)
}

func ReflectiveFormatFunc(refV reflect.Value, refType reflect.Type, target io.Writer) error {
	fes := make([]func() error, 0, 2*refType.NumIn()+2*refType.NumOut()+7)

	fes = append(fes, func() error { return fprint(target, "func ") })
	fes = append(fes, func() error { return fprint(target, runtime.FuncForPC(refV.Pointer()).Name()) })
	fes = append(fes, func() error { return fprint(target, "(") })

	for i := 0; i < refType.NumIn()-1; i++ {
		idx := i
		fes = append(fes, func() error { return fprint(target, refType.In(idx).String()) })
		fes = append(fes, func() error { return fprint(target, ", ") })
	}
	if refType.NumIn() > 0 {
		if refType.IsVariadic() {
			fes = append(fes, func() error {
				return fprint(target, "...")
			})
			fes = append(fes, func() error {
				return fprint(target, refType.In(refType.NumIn()-1).Elem().String())
			})
		} else {
			return fprint(target, refType.In(refType.NumIn()-1).String())
		}
	}

	fes = append(fes, func() error { return fprint(target, ")") })

	if refType.NumOut() > 0 {
		fes = append(fes, func() error { return fprint(target, " ") })
		if refType.NumOut() == 1 {
			fes = append(fes, func() error { return fprint(target, refType.Out(0).String()) })
		} else {
			fes = append(fes, func() error { return fprint(target, "(") })
			for i := 0; i < refType.NumOut(); i++ {
				idx := i
				fes = append(fes, func() error { return fprint(target, refType.Out(idx).String()) })
				if idx < refType.NumOut()-1 {
					fes = append(fes, func() error { return fprint(target, ", ") })
				}
			}
			fes = append(fes, func() error { return fprint(target, ")") })
		}
	}
	return RunAllBeforeFirstError(fes...)
}

func ReflectiveFormatPointer(refV reflect.Value, prefix string, indent int, target io.Writer) error {
	fes := make([]func() error, 0, 2)
	if refV.IsNil() {
		fes = append(fes, func() error {
			return fprint(target, "(*"+refV.Type().Elem().String()+")(<nil>)")
		})
	} else {
		fes = append(fes, func() error { return fprint(target, "&") })
		fes = append(fes, func() error { return ReflectiveFormatToWriter(refV.Elem(), prefix, indent, target) })
	}
	return RunAllBeforeFirstError(fes...)
}

func ReflectiveFormatMap(refV reflect.Value, prefix string, indent int, target io.Writer) error {
	if refV.IsNil() {
		return fprint(target, "map(<nil>)")
	} else {
		fes := make([]func() error, 0, 3+refV.Len()*5)

		fes = append(fes, func() error { return fprint(target, "map{\n") })
		it := refV.MapRange()
		for it.Next() {
			k := it.Key()
			v := it.Value()
			fes = append(fes, func() error { return doIndent(target, prefix, indent+1) })
			fes = append(fes, func() error { return ReflectiveFormatToWriter(k, prefix, indent+1, target) })
			fes = append(fes, func() error { return fprint(target, ": ") })
			fes = append(fes, func() error { return ReflectiveFormatToWriter(v, prefix, indent+1, target) })
			fes = append(fes, func() error { return fprint(target, ",\n") })
		}
		fes = append(fes, func() error { return doIndent(target, prefix, indent) })
		fes = append(fes, func() error { return fprint(target, "}") })

		return RunAllBeforeFirstError(fes...)
	}
}

func ReflectiveFormatInterface(refV reflect.Value, prefix string, indent int, target io.Writer, refType reflect.Type) error {
	fes := make([]func() error, 0, 7)

	fes = append(fes, func() error { return fprint(target, refV.Type().String()) })
	fes = append(fes, func() error { return fprint(target, "(\n") })
	fes = append(fes, func() error { return doIndent(target, prefix, indent+1) })
	fes = append(fes, func() error { return ReflectiveFormatToWriter(refV.Elem(), prefix, indent+1, target) })

	fes = append(fes, func() error { return fprint(target, "\n") })
	fes = append(fes, func() error { return doIndent(target, prefix, indent) })
	fes = append(fes, func() error { return fprint(target, ")") })

	return RunAllBeforeFirstError(fes...)
}

func ReflectiveFormatArrayOrSlice(refV reflect.Value, prefix string, indent int, target io.Writer) error {
	l := refV.Len()
	fes := make([]func() error, 0, 2*l+3)
	fes = append(fes, func() error { return fprintln(target, "[") })

	for i := 0; i < l; i++ {
		idx := i
		fes = append(fes, func() error { return doIndent(target, prefix, indent+1) })
		fes = append(fes, func() error { return ReflectiveFormatToWriter(refV.Index(idx), prefix, indent+1, target) })
		fes = append(fes, func() error { return fprintln(target, ",") })
	}
	fes = append(fes, func() error { return doIndent(target, prefix, indent) })
	fes = append(fes, func() error { return fprint(target, "]") })

	return RunAllBeforeFirstError(fes...)
}

func fprint(writer io.Writer, val ...interface{}) error {
	_, err := fmt.Fprint(writer, val...)
	return err
}

func fprintf(writer io.Writer, format string, val ...interface{}) error {
	_, err := fmt.Fprintf(writer, format, val...)
	return err
}

func fprintln(writer io.Writer, val ...interface{}) error {
	_, err := fmt.Fprintln(writer, val...)
	return err
}

func RunAllBeforeFirstError(funcs ...func() error) error {
	for _, f := range funcs {
		if f == nil {
			continue
		}
		err := f()
		if err != nil {
			return err
		}
	}
	return nil
}

func doIndent(writer io.Writer, prefix string, i int) error {
	_, err := fmt.Fprint(writer, strings.Repeat(prefix, i))
	return err
}
