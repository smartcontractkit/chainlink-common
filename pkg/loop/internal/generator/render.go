package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func RenderAll(protoOut, goOutDir string, svc *Service) error {
	if err := renderFile(templatesProto, protoOut, svc); err != nil {
		return err
	}
	if err := renderFile(templatesWiring, filepath.Join(goOutDir, "rpc.go"), svc); err != nil {
		return err
	}
	if err := renderFile(templatesRPCTest, filepath.Join(goOutDir, "rpc_test.go"), svc); err != nil {
		return err
	}
	return nil
}

func renderFile(tmplSrc, outPath string, data any) error {
	// Build alias lookup for conditional helpers/imports
	aliasMap := map[string]string{}
	if svc, ok := data.(*Service); ok {
		for _, ia := range svc.ImportAliases {
			aliasMap[ia.Path] = ia.Alias
		}
	}
	aliasFor := func(p string) string {
		if a := aliasMap[p]; a != "" {
			return a
		}
		if i := strings.LastIndex(p, "/"); i >= 0 {
			return p[i+1:]
		}
		return p
	}

	t, err := template.New("x").Funcs(template.FuncMap{
		"pbField":           pbFieldName,
		"isMsg":             isMessage,
		"toPB":              toPBFuncName,
		"fromPB":            fromPBFuncName,
		"domainFieldGoName": domainFieldGoName,
		"add":               add,
		"lower":             lower,

		// Flags from generator.go
		"elemIsArray":     func(f Field) bool { return f.DomainIsArray },
		"elemGo":          func(f Field) string { return f.DomainElemGoType },
		"elemArrayIsByte": func(f Field) bool { return f.DomainElemIsByte },
		"isBytesType":     func(pt ProtoType) bool { return pt.Name == "bytes" },

		// Element type from Go type string (safe fallback)
		"sliceElem": func(f Field) string { return chooseElemType(f) },

		// String-based helpers
		"isSliceType": func(s string) bool {
			s = strings.TrimSpace(s)
			return strings.HasPrefix(s, "[]")
		},

		// Import/alias helpers
		"aliasFor": aliasFor,
	}).Parse(tmplSrc)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return err
	}
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func chooseElemType(f Field) string {
	if f.DomainElemGoType != "" {
		return f.DomainElemGoType
	}
	return elemFromGoType(f.GoType)
}

func elemFromGoType(goType string) string {
	s := strings.TrimSpace(goType)
	if s == "" {
		return s
	}
	// Strip leading slice brackets "[]"
	for strings.HasPrefix(s, "[]") {
		s = s[2:]
	}
	// If array form "[N]T", strip "[N]" and return "T"
	if strings.HasPrefix(s, "[") {
		if i := strings.Index(s, "]"); i > 0 && i+1 < len(s) {
			return strings.TrimSpace(s[i+1:])
		}
	}
	return s
}

func domainFieldGoName(proto string) string {
	// reverse of pbFieldName
	if strings.Contains(proto, "_") {
		parts := strings.Split(proto, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, "")
	}
	return strings.ToUpper(proto[:1]) + proto[1:]
}
func isMessage(pt ProtoType) bool {
	switch pt.Name {
	case "bytes", "string", "bool", "uint64", "uint32", "int64", "int32", "double", "float", "sint64", "sint32", "fixed64", "fixed32":
		return false
	default:
		// anything not a scalar/bytes is considered a message reference
		return true
	}
}

func toPBFuncName(typeName string) string   { return "toPB_" + typeName }
func fromPBFuncName(typeName string) string { return "fromPB_" + typeName }
func pbFieldName(s string) string {
	if s == "" {
		return ""
	}
	// Allow both snake_case and lowerCamel in source model
	if strings.Contains(s, "_") {
		parts := strings.Split(s, "_")
		for i, p := range parts {
			if p == "" {
				continue
			}
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
		return strings.Join(parts, "")
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
func add(a, b int) int { return a + b }

func lower(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

var templatesProto = `syntax = "proto3";
option go_package = "{{.OptionGoPackage}}";
package {{.ProtoPkg}};
{{- range .ProtoImports }}
import "{{ . }}";
{{- end }}

{{/* 1) Domain struct messages */}}
{{- range .UserMessages }}
message {{ .Name }} {
{{- range .Fields }}
{{- if .Type.IsRepeated }}
repeated {{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
{{- else }}
{{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
{{- end }}
{{- end }}
}
{{- end }}

{{ range .InterfaceOneofs }}
message {{ .Name }} {
  oneof kind {
    {{- range $i, $c := .Cases }}
    {{ $c.TypeName }} {{ lower $c.TypeName }} = {{ add $i 1 }};
    {{- end }}
  }
}
{{ end }}

service {{ .ServiceName }} {
  {{- range .Methods }}
  rpc {{ .Name }}({{ .Req.Name }}) returns ({{ .Rep.Name }});
  {{- end }}
}

{{/* 2) Per-method request/reply */}}
{{- range .Methods }}
message {{ .Req.Name }} {
  {{- range .Req.Fields }}
  {{- if .Type.IsRepeated }}
  repeated {{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
  {{- else }}
  {{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
  {{- end }}
  {{- end }}
}
message {{ .Rep.Name }} {
  {{- range .Rep.Fields }}
  {{- if .Type.IsRepeated }}
  repeated {{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
  {{- else }}
  {{ .Type.Name }} {{ .Name }} = {{ .ProtoTag }};
  {{- end }}
  {{- end }}
}
{{- end }}`

var templatesServer = `
// Code generated by genrpc; DO NOT EDIT.
package {{ .ServiceName | lower }}

import (
  "context"
  {{- if aliasFor "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm" }}
  "fmt"
  {{- end }}

  pb "{{ .OptionGoPackage }}"
  {{- range .ImportAliases }}
  {{ .Alias }} "{{ .Path }}"
  {{- end }}
)


`

var templatesWiring = `
// Code generated by genrpc; DO NOT EDIT.
package {{ .ServiceName | lower }}

import (
  "context"
  {{- if aliasFor "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm" }}
  "fmt"
  {{- end }}

  pb "{{ .OptionGoPackage }}"
  {{- range .ImportAliases }}
  {{ .Alias }} "{{ .Path }}"
  {{- end }}
)

type Client struct{ rpc pb.{{ .ServiceName }}Client }

func NewClient(r pb.{{ .ServiceName }}Client) *Client { return &Client{rpc: r} }

{{ range .Methods }}
func (c *Client) {{ .Name }}(
  ctx context.Context{{ range .Req.Fields }}, {{ .Name }} {{ .GoType }}{{ end }},
) ({{ if .Rep.Fields }}{{ (index .Rep.Fields 0).GoType }}{{ else }}struct{}{{ end }}, error) {

  req := &pb.{{ .Req.Name }}{
    {{- range .Req.Fields }}
    {{- if and (isMsg .Type) .Type.IsRepeated }}
    {{ pbField .Name }}: func(in {{ .GoType }}) []*pb.{{ .Type.Name }} {
      if in == nil { return nil }
      out := make([]*pb.{{ .Type.Name }}, len(in))
      for i, v := range in { out[i] = {{ toPB .Type.Name }}(v) }
      return out
    }({{ .Name }}),
    {{- else if isMsg .Type }}
    {{ pbField .Name }}: {{ toPB ( .Type.Name ) }}({{ .Name }}),
    {{- else if .Type.IsRepeated }}
      {{- if eq .Type.Name "bytes" }}
    // domain []T -> pb [][]byte (T may be a fixed [N]byte)
    {{ pbField .Name }}: func(in {{ .GoType }}) [][]byte {
      if in == nil { return nil }
      out := make([][]byte, len(in))
      for i := range in {
        {{- if isSliceType (sliceElem .) }}
        out[i] = []byte(in[i])
        {{- else }}
        out[i] = in[i][:]
        {{- end }}
      }
      return out
    }({{ .Name }}),
      {{- else }}
    // domain []scalar -> pb []scalar
    {{ pbField .Name }}: func(in {{ .GoType }}) []{{ .Type.Name }} {
      if in == nil { return nil }
      out := make([]{{ .Type.Name }}, len(in))
      for i := range in { out[i] = {{ .Type.Name }}(in[i]) }
      return out
    }({{ .Name }}),
      {{- end }}
    {{- else if eq .Type.Name "bytes" }}
      {{- if isSliceType .GoType }}
    {{ pbField .Name }}: {{ .Name }},
      {{- else }}
    // domain [N]byte -> pb []byte
    {{ pbField .Name }}: {{ .Name }}[:],
      {{- end }}
    {{- else }}
    {{ pbField .Name }}: {{ .Type.Name }}({{ .Name }}),
    {{- end }}
    {{- end }}
  }

  rep, err := c.rpc.{{ .Name }}(ctx, req)
  if err != nil {
    {{- if .Rep.Fields }} var zero {{ (index .Rep.Fields 0).GoType }}; return zero, err {{- else }} return struct{}{}, err {{- end }}
  }

  {{- if .Rep.Fields }}
  {{- $rf := (index .Rep.Fields 0) }}
  {{- if isMsg $rf.Type }}
    {{- if $rf.Type.IsRepeated }}
  if rep.{{ pbField $rf.Name }} == nil { var zero {{ $rf.GoType }}; return zero, nil }
  out := make({{ $rf.GoType }}, len(rep.{{ pbField $rf.Name }}))
  for i, v := range rep.{{ pbField $rf.Name }} { out[i] = {{ fromPB $rf.Type.Name }}(v) }
  return out, nil
    {{- else }}
  return {{ fromPB ( $rf.Type.Name ) }}(rep.{{ pbField $rf.Name }}), nil
    {{- end }}
  {{- else }}
    {{- if $rf.Type.IsRepeated }}
      {{- if $rf.DomainIsArray }}
  // pb []scalar -> domain [N]T
  var out {{ $rf.GoType }}
  if rep.{{ pbField $rf.Name }} != nil {
    for i := range out {
      if i < len(rep.{{ pbField $rf.Name }}) {
        out[i] = {{ sliceElem $rf }}(rep.{{ pbField $rf.Name }}[i])
      }
    }
  }
  return out, nil
      {{- else if eq $rf.Type.Name "bytes" }}
  // pb [][]byte -> domain []T (T may be fixed [N]byte)
  if rep.{{ pbField $rf.Name }} == nil { var zero {{ $rf.GoType }}; return zero, nil }
  out := make({{ $rf.GoType }}, len(rep.{{ pbField $rf.Name }}))
  for i := range rep.{{ pbField $rf.Name }} {
    {{- if isSliceType (sliceElem $rf) }}
    out[i] = {{ sliceElem $rf }}(rep.{{ pbField $rf.Name }}[i])
    {{- else }}
    var e {{ sliceElem $rf }}
    if len(rep.{{ pbField $rf.Name }}[i]) != len(e) { var zero {{ $rf.GoType }}; return zero, fmt.Errorf("invalid length for {{ pbField $rf.Name }}[%d]: got %d want %d", i, len(rep.{{ pbField $rf.Name }}[i]), len(e)) }
    copy(e[:], rep.{{ pbField $rf.Name }}[i])
    out[i] = e
    {{- end }}
  }
  return out, nil
      {{- else }}
  // pb []scalar -> domain []scalar (cast if needed)
  if rep.{{ pbField $rf.Name }} == nil { var zero {{ $rf.GoType }}; return zero, nil }
  out := make({{ $rf.GoType }}, len(rep.{{ pbField $rf.Name }}))
  for i := range rep.{{ pbField $rf.Name }} {
    out[i] = {{ sliceElem $rf }}(rep.{{ pbField $rf.Name }}[i])
  }
  return out, nil
      {{- end }}
    {{- else if eq $rf.Type.Name "bytes" }}
      {{- if isSliceType $rf.GoType }}
  // pb []byte -> domain []byte
  return rep.{{ pbField $rf.Name }}, nil
      {{- else }}
  // pb []byte -> domain [N]byte
  var out {{ $rf.GoType }}
  if rep.{{ pbField $rf.Name }} != nil {
    if len(rep.{{ pbField $rf.Name }}) != len(out) { var zero {{ $rf.GoType }}; return zero, fmt.Errorf("invalid length for {{ pbField $rf.Name }}: got %d want %d", len(rep.{{ pbField $rf.Name }}), len(out)) }
    copy(out[:], rep.{{ pbField $rf.Name }})
  }
  return out, nil
      {{- end }}
    {{- else }}
  return {{ $rf.GoType }}(rep.{{ pbField $rf.Name }}), nil
    {{- end }}
  {{- end }}
  {{- else }}
  return struct{}{}, nil
  {{- end }}
}
{{ end }}

// ---- pb<->domain converters for user messages ----
{{- range .UserMessages }}
func {{ toPB .Name }}(in {{ .DomainGo }}) *pb.{{ .Name }} {
  out := &pb.{{ .Name }}{}
  {{- range .Fields }}
  {{- if and (isMsg .Type) .Type.IsRepeated }}
  if in.{{ domainFieldGoName .Name }} != nil {
    out.{{ pbField .Name }} = make([]*pb.{{ .Type.Name }}, len(in.{{ domainFieldGoName .Name }}))
    for i, v := range in.{{ domainFieldGoName .Name }} { out.{{ pbField .Name }}[i] = {{ toPB .Type.Name }}(v) }
  }
  {{- else if isMsg .Type }}
  out.{{ pbField .Name }} = {{ toPB .Type.Name }}(in.{{ domainFieldGoName .Name }})
  {{- else if .Type.IsRepeated }}
    {{- if eq .Type.Name "bytes" }}
  // domain []T -> pb [][]byte (T may be fixed [N]byte)
  if in.{{ domainFieldGoName .Name }} != nil {
    out.{{ pbField .Name }} = make([][]byte, len(in.{{ domainFieldGoName .Name }}))
    for i := range in.{{ domainFieldGoName .Name }} {
      {{- if isSliceType (sliceElem .) }}
      out.{{ pbField .Name }}[i] = []byte(in.{{ domainFieldGoName .Name }}[i])
      {{- else }}
      out.{{ pbField .Name }}[i] = in.{{ domainFieldGoName .Name }}[i][:]
      {{- end }}
    }
  }
    {{- else if .DomainIsArray }}
  // domain [N]T -> pb []scalar
  {
    src := in.{{ domainFieldGoName .Name }}
    out.{{ pbField .Name }} = make([]{{ .Type.Name }}, len(src))
    for i := range src { out.{{ pbField .Name }}[i] = {{ .Type.Name }}(src[i]) }
  }
    {{- else }}
  // domain []T -> pb []scalar
  if in.{{ domainFieldGoName .Name }} != nil {
    out.{{ pbField .Name }} = make([]{{ .Type.Name }}, len(in.{{ domainFieldGoName .Name }}))
    for i := range in.{{ domainFieldGoName .Name }} {
      out.{{ pbField .Name }}[i] = {{ .Type.Name }}(in.{{ domainFieldGoName .Name }}[i])
    }
  }
    {{- end }}
  {{- else if eq .Type.Name "bytes" }}
    {{- if isSliceType .GoType }}
  out.{{ pbField .Name }} = in.{{ domainFieldGoName .Name }}
    {{- else }}
  // domain [N]byte/uint8 -> pb []byte
  out.{{ pbField .Name }} = in.{{ domainFieldGoName .Name }}[:]
    {{- end }}
  {{- else }}
  out.{{ pbField .Name }} = {{ .Type.Name }}(in.{{ domainFieldGoName .Name }})
  {{- end }}
  {{- end }}
  return out
}

func {{ fromPB .Name }}(in *pb.{{ .Name }}) {{ .DomainGo }} {
  var out {{ .DomainGo }}
  if in == nil { return out }
  {{- range .Fields }}
  {{- if and (isMsg .Type) .Type.IsRepeated }}
  if in.{{ pbField .Name }} != nil {
    out.{{ domainFieldGoName .Name }} = make([]{{ sliceElem . }}, len(in.{{ pbField .Name }}))
    for i, v := range in.{{ pbField .Name }} {
      out.{{ domainFieldGoName .Name }}[i] = {{ fromPB .Type.Name }}(v)
    }
  }
  {{- else if isMsg .Type }}
  out.{{ domainFieldGoName .Name }} = {{ fromPB .Type.Name }}(in.{{ pbField .Name }})
  {{- else if .Type.IsRepeated }}
    {{- if eq .Type.Name "bytes" }}
  // pb [][]byte -> domain []T (T may be fixed [N]byte)
  if in.{{ pbField .Name }} != nil {
    out.{{ domainFieldGoName .Name }} = make([]{{ sliceElem . }}, len(in.{{ pbField .Name }}))
    for i := range in.{{ pbField .Name }} {
      {{- if isSliceType (sliceElem .) }}
      out.{{ domainFieldGoName .Name }}[i] = {{ sliceElem . }}(in.{{ pbField .Name }}[i])
      {{- else }}
      var e {{ sliceElem . }}
      // Length-check omitted here to keep converter pure; client/server paths enforce it and can return error.
      copy(e[:], in.{{ pbField .Name }}[i])
      out.{{ domainFieldGoName .Name }}[i] = e
      {{- end }}
    }
  }
    {{- else if .DomainIsArray }}
  // pb []scalar -> domain [N]T
  for i := range out.{{ domainFieldGoName .Name }} {
    if in.{{ pbField .Name }} != nil && i < len(in.{{ pbField .Name }}) {
      out.{{ domainFieldGoName .Name }}[i] = {{ sliceElem . }}(in.{{ pbField .Name }}[i])
    }
  }
    {{- else }}
  // pb []scalar -> domain []T
  if in.{{ pbField .Name }} != nil {
    out.{{ domainFieldGoName .Name }} = make([]{{ sliceElem . }}, len(in.{{ pbField .Name }}))
    for i := range in.{{ pbField .Name }} {
      out.{{ domainFieldGoName .Name }}[i] = {{ sliceElem . }}(in.{{ pbField .Name }}[i])
    }
  }
    {{- end }}
  {{- else if eq .Type.Name "bytes" }}
    {{- if isSliceType .GoType }}
  // pb []byte -> domain []byte
  out.{{ domainFieldGoName .Name }} = in.{{ pbField .Name }}
    {{- else }}
  // pb []byte -> domain [N]byte/uint8
  if in.{{ pbField .Name }} != nil {
    copy(out.{{ domainFieldGoName .Name }}[:], in.{{ pbField .Name }})
  }
    {{- end }}
  {{- else }}
  out.{{ domainFieldGoName .Name }} = {{ .GoType }}(in.{{ pbField .Name }})
  {{- end }}
  {{- end }}
  return out
}
{{- end }}


{{- range $io := .InterfaceOneofs }}
func toPB_{{$io.Name}}(in {{$io.GoInterface}}) *pb.{{$io.Name}} {
  if in == nil { return nil }
  switch v := in.(type) {
  {{- range $c := $io.Cases }}
  case *{{$c.DomainGo}}:
    return &pb.{{$io.Name}}{
      Kind: &pb.{{$io.Name}}_{{$c.TypeName}}{ {{$c.TypeName}}: toPB_{{$c.TypeName}}(*v) },
    }
  {{- end }}
  default:
    return nil
  }
}

func fromPB_{{$io.Name}}(in *pb.{{$io.Name}}) {{$io.GoInterface}} {
  if in == nil { return nil }
  switch k := in.Kind.(type) {
  {{- range $c := $io.Cases }}
  case *pb.{{$io.Name}}_{{$c.TypeName}}:
    {
      v := fromPB_{{$c.TypeName}}(k.{{$c.TypeName}})
      return &v
    }
  {{- end }}
  default:
    return nil
  }
}
{{- end }}

// ---- Optional helpers for fixed-size arrays (with length checks) ----
{{- $evm := aliasFor "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm" -}}
{{- if $evm }}

// toAddress converts a byte slice into an {{$evm}}.Address with length check.
func toAddress(b []byte) ({{$evm}}.Address, error) {
  var out {{$evm}}.Address
  if len(b) != len(out) { return out, fmt.Errorf("invalid evm.Address length: got %d want %d", len(b), len(out)) }
  copy(out[:], b)
  return out, nil
}

// toHash converts a byte slice into an {{$evm}}.Hash with length check.
func toHash(b []byte) ({{$evm}}.Hash, error) {
  var out {{$evm}}.Hash
  if len(b) != len(out) { return out, fmt.Errorf("invalid evm.Hash length: got %d want %d", len(b), len(out)) }
  copy(out[:], b)
  return out, nil
}
{{- end }}

// server
type Server struct {
  pb.Unimplemented{{ .ServiceName }}Server
  impl interface{
    {{- range .Methods }}
    {{ .Name }}(ctx context.Context{{ range .Req.Fields }}, {{ .Name }} {{ .GoType }}{{ end }}) ({{ if .Rep.Fields }}{{ (index .Rep.Fields 0).GoType }}{{ else }}struct{}{{ end }}, error)
    {{- end }}
  }
}

func NewServer(impl any) *Server {
  return &Server{impl: impl.(interface{
    {{- range .Methods }}
    {{ .Name }}(ctx context.Context{{ range .Req.Fields }}, {{ .Name }} {{ .GoType }}{{ end }}) ({{ if .Rep.Fields }}{{ (index .Rep.Fields 0).GoType }}{{ else }}struct{}{{ end }}, error)
    {{- end }}
  })}
}

{{ range .Methods }}
func (s *Server) {{ .Name }}(ctx context.Context, req *pb.{{ .Req.Name }}) (*pb.{{ .Rep.Name }}, error) {
  {{- /* pb -> domain params */}}
  {{- range .Req.Fields }}
  {{- if and (isMsg .Type) .Type.IsRepeated }}
  var {{ .Name }} {{ .GoType }}
  if req.{{ pbField .Name }} != nil {
    {{ .Name }} = make({{ .GoType }}, len(req.{{ pbField .Name }}))
    for i, v := range req.{{ pbField .Name }} {
      {{ .Name }}[i] = {{ fromPB .Type.Name }}(v)
    }
  }
  {{- else if isMsg .Type }}
  var {{ .Name }} {{ .GoType }} = {{ fromPB .Type.Name }}(req.{{ pbField .Name }})
  {{- else if and .Type.IsRepeated (eq .Type.Name "bytes") }}
  // pb [][]byte -> domain []T (T may be fixed [N]byte)
  var {{ .Name }} {{ .GoType }}
  if req.{{ pbField .Name }} != nil {
    {{ .Name }} = make({{ .GoType }}, len(req.{{ pbField .Name }}))
    for i := range req.{{ pbField .Name }} {
      {{- if isSliceType (sliceElem .) }}
      {{ .Name }}[i] = {{ sliceElem . }}(req.{{ pbField .Name }}[i])
      {{- else }}
      var e {{ sliceElem . }}
      if len(req.{{ pbField .Name }}[i]) != len(e) { return nil, fmt.Errorf("invalid length for {{ .Name }}[%d]: got %d want %d", i, len(req.{{ pbField .Name }}[i]), len(e)) }
      copy(e[:], req.{{ pbField .Name }}[i])
      {{ .Name }}[i] = e
      {{- end }}
    }
  }
  {{- else if .Type.IsRepeated }}
  // scalar repeated: assign directly (proto-shaped types)
  var {{ .Name }} {{ .GoType }} = req.{{ pbField .Name }}
  {{- else if eq .Type.Name "bytes" }}
    {{- if isSliceType .GoType }}
  var {{ .Name }} {{ .GoType }} = req.{{ pbField .Name }}
    {{- else }}
  // pb []byte -> domain [N]byte
  var {{ .Name }} {{ .GoType }}
  if req.{{ pbField .Name }} != nil {
    if len(req.{{ pbField .Name }}) != len({{ .Name }}) { return nil, fmt.Errorf("invalid length for {{ .Name }}: got %d want %d", len(req.{{ pbField .Name }}), len({{ .Name }})) }
    copy({{ .Name }}[:], req.{{ pbField .Name }})
  }
    {{- end }}
  {{- else }}
  // scalar single
  var {{ .Name }} {{ .GoType }} = {{ .GoType }}(req.{{ pbField .Name }})
  {{- end }}
  {{- end }}

  {{- if .Rep.Fields }}
  res, err := s.impl.{{ .Name }}(ctx{{ range .Req.Fields }}, {{ .Name }}{{ end }})
  if err != nil { return nil, err }
  rep := &pb.{{ .Rep.Name }}{}
  {{- $rf := (index .Rep.Fields 0) }}
  {{- if isMsg $rf.Type }}
    {{- if $rf.Type.IsRepeated }}
  if res != nil {
    rep.{{ pbField $rf.Name }} = make([]*pb.{{ $rf.Type.Name }}, len(res))
    for i, v := range res { rep.{{ pbField $rf.Name }}[i] = {{ toPB $rf.Type.Name }}(v) }
  }
    {{- else }}
  rep.{{ pbField $rf.Name }} = {{ toPB $rf.Type.Name }}(res)
    {{- end }}
  {{- else }}
    {{- if $rf.Type.IsRepeated }}
      {{- if eq $rf.Type.Name "bytes" }}
  // domain []T -> pb [][]byte (T may be fixed [N]byte)
  if res != nil {
    rep.{{ pbField $rf.Name }} = make([][]byte, len(res))
    for i := range res {
      {{- if isSliceType (sliceElem $rf) }}
      rep.{{ pbField $rf.Name }}[i] = []byte(res[i])
      {{- else }}
      rep.{{ pbField $rf.Name }}[i] = res[i][:]
      {{- end }}
    }
  }
      {{- else if $rf.DomainIsArray }}
  // domain [N]T -> pb []scalar (array cannot be nil)
  rep.{{ pbField $rf.Name }} = make([]{{ $rf.Type.Name }}, len(res))
  for i := range res { rep.{{ pbField $rf.Name }}[i] = {{ $rf.Type.Name }}(res[i]) }
      {{- else }}
  // domain slice -> pb slice (scalar)
  if res != nil {
    rep.{{ pbField $rf.Name }} = make([]{{ $rf.Type.Name }}, len(res))
    for i := range res { rep.{{ pbField $rf.Name }}[i] = {{ $rf.Type.Name }}(res[i]) }
  }
      {{- end }}
    {{- else if eq $rf.Type.Name "bytes" }}
      {{- if isSliceType $rf.GoType }}
  rep.{{ pbField $rf.Name }} = res
      {{- else }}
  // domain [N]byte/uint8 -> pb []byte
  rep.{{ pbField $rf.Name }} = res[:]
      {{- end }}
    {{- else }}
  rep.{{ pbField $rf.Name }} = {{ $rf.Type.Name }}(res)
    {{- end }}
  {{- end }}
  return rep, nil
  {{- else }}
  _, err := s.impl.{{ .Name }}(ctx{{ range .Req.Fields }}, {{ .Name }}{{ end }})
  if err != nil { return nil, err }
  return &pb.{{ .Rep.Name }}{}, nil
  {{- end }}
}
{{ end }}

`

var templatesRPCTest = `// Code generated by genrpc tests; DO NOT EDIT.
package {{ .ServiceName | lower }}

import (
  "context"
  "net"
  "reflect"
  "testing"

  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials/insecure"
  "google.golang.org/grpc/test/bufconn"

  pb "{{ .OptionGoPackage }}"
  {{- range .ImportAliases }}
  {{ .Alias }} "{{ .Path }}"
  {{- end }}
)

const bufSize = 1024 * 1024

// ---- fake impl that returns preloaded values per method ----
type fakeImpl struct {
  {{- range .Methods }}
  ret_{{ .Name }} {{ if .Rep.Fields }}{{ (index .Rep.Fields 0).GoType }}{{ else }}struct{}{{ end }}
  {{- end }}
}
{{- range .Methods }}
func (f *fakeImpl) {{ .Name }}(ctx context.Context{{ range .Req.Fields }}, {{ .Name }} {{ .GoType }}{{ end }}) ({{ if .Rep.Fields }}{{ (index .Rep.Fields 0).GoType }}{{ else }}struct{}{{ end }}, error) {
  return f.ret_{{ .Name }}, nil
}
{{- end }}

// ---- fixtures (depth-limited) for user-defined messages ----
{{- range .UserMessages }}

// Public, depth-1 convenience wrapper
func fixture_{{ .Name }}() {{ .DomainGo }} { return fixture_{{ .Name }}_depth(3) }

func fixture_{{ .Name }}_depth(d int) {{ .DomainGo }} {
  var out {{ .DomainGo }}
  if d <= 0 { return out }
  {{- range .Fields }}
  {{- if and (isMsg .Type) .Type.IsRepeated }}
  out.{{ domainFieldGoName .Name }} = make([]{{ sliceElem . }}, 2)
  out.{{ domainFieldGoName .Name }}[0] = fixture_{{ .Type.Name }}_depth(d-1)
  out.{{ domainFieldGoName .Name }}[1] = fixture_{{ .Type.Name }}_depth(d-1)
  {{- else if isMsg .Type }}
  out.{{ domainFieldGoName .Name }} = fixture_{{ .Type.Name }}_depth(d-1)
  {{- else if .Type.IsRepeated }}
    {{- if eq .Type.Name "bytes" }}
  // []bytes (element may be []byte or fixed-size [N]byte-like)
  out.{{ domainFieldGoName .Name }} = make([]{{ sliceElem . }}, 2)
    {{- if isSliceType (sliceElem .) }}
  out.{{ domainFieldGoName .Name }}[0] = []byte{1,2,3}
  out.{{ domainFieldGoName .Name }}[1] = []byte{4,5,6}
    {{- else }}
  {
    var e0 {{ sliceElem . }}; copy(e0[:], []byte{1,2,3,4})
    var e1 {{ sliceElem . }}; copy(e1[:], []byte{5,6,7,8})
    out.{{ domainFieldGoName .Name }}[0], out.{{ domainFieldGoName .Name }}[1] = e0, e1
  }
    {{- end }}
    {{- else }}
  // []scalar
  out.{{ domainFieldGoName .Name }} = []{{ sliceElem . }}{ {{ sliceElem . }}(1), {{ sliceElem . }}(2) }
    {{- end }}
  {{- else if eq .Type.Name "bytes" }}
    {{- if isSliceType .GoType }}
  // []byte
  out.{{ domainFieldGoName .Name }} = []byte{9,8,7,6}
    {{- else }}
  // fixed-size [N]byte-like (named array type)
  {
    var e {{ .GoType }}
    copy(e[:], []byte{9,8,7,6})
    out.{{ domainFieldGoName .Name }} = e
  }
    {{- end }}
  {{- else if eq .Type.Name "string" }}
  out.{{ domainFieldGoName .Name }} = {{ .GoType }}("fixture")
  {{- else if eq .Type.Name "bool" }}
  out.{{ domainFieldGoName .Name }} = {{ .GoType }}(true)
  {{- else if or (eq .Type.Name "double") (eq .Type.Name "float") }}
  out.{{ domainFieldGoName .Name }} = {{ .GoType }}(1.5)
  {{- else }}
  // numeric scalar (cast)
  out.{{ domainFieldGoName .Name }} = {{ .GoType }}(2)
  {{- end }}
  {{- end }}
  return out
}
{{- end }}

// ---- fixtures for interface oneofs (pick first case), depth-limited ----
{{- range $io := .InterfaceOneofs }}

// Public, depth-1 convenience wrapper
func fixture_{{$io.Name}}() {{$io.GoInterface}} { return fixture_{{$io.Name}}_depth(1) }

func fixture_{{$io.Name}}_depth(d int) {{$io.GoInterface}} {
  if d <= 0 { return nil }
  {{- if gt (len $io.Cases) 0 }}
  {
    v := fixture_{{ (index $io.Cases 0).TypeName }}_depth(d-1)
    return &v
  }
  {{- else }}
  return nil
  {{- end }}
}
{{- end }}

// ---- single suite with subtests (one server/client for all) ----
func Test_{{ .ServiceName }}_DomainRoundtrip(t *testing.T) {
  impl := &fakeImpl{}

  lis := bufconn.Listen(bufSize)
  s := grpc.NewServer()
  pb.Register{{ .ServiceName }}Server(s, NewServer(impl))
  go func() { _ = s.Serve(lis) }()
  t.Cleanup(func() { s.Stop(); _ = lis.Close() })

  dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
  conn, err := grpc.DialContext(context.Background(), "bufnet",
    grpc.WithContextDialer(dialer),
    grpc.WithTransportCredentials(insecure.NewCredentials()))
  if err != nil { t.Fatalf("dial: %v", err) }
  t.Cleanup(func() { _ = conn.Close() })

  c := NewClient(pb.New{{ .ServiceName }}Client(conn))

  {{- range .Methods }}
  t.Run("{{ .Name }} happy path", func(t *testing.T) {
    ctx := context.Background()
    {{- if .Rep.Fields }}
      {{- $rf := (index .Rep.Fields 0) }}
      {{- if and (isMsg $rf.Type) $rf.Type.IsRepeated }}
    // want: []Message
    want := make({{ $rf.GoType }}, 2)
    want[0] = fixture_{{ $rf.Type.Name }}()
    want[1] = fixture_{{ $rf.Type.Name }}()
      {{- else if isMsg $rf.Type }}
    // want: Message
    want := fixture_{{ $rf.Type.Name }}()
      {{- else if $rf.DomainIsArray }}
    // want: [N]T (fixed-size array)
    var want {{ $rf.GoType }}
    for i := range want { want[i] = {{ sliceElem $rf }}(i+1) }
      {{- else if and $rf.Type.IsRepeated (eq $rf.Type.Name "bytes") }}
    // want: [][]byte (or []<[N]byte]-like)
    want := make({{ $rf.GoType }}, 2)
      {{- if isSliceType (sliceElem $rf) }}
    want[0] = []byte{10,11}
    want[1] = []byte{12,13}
      {{- else }}
    {
      var e0 {{ sliceElem $rf }}; copy(e0[:], []byte{1,2,3,4})
      var e1 {{ sliceElem $rf }}; copy(e1[:], []byte{5,6,7,8})
      want[0], want[1] = e0, e1
    }
      {{- end }}
      {{- else if $rf.Type.IsRepeated }}
    // want: []scalar
    want := {{ $rf.GoType }}{ {{ sliceElem $rf }}(1), {{ sliceElem $rf }}(2), {{ sliceElem $rf }}(3) }
      {{- else if eq $rf.Type.Name "bytes" }}
        {{- if isSliceType $rf.GoType }}
    // want: []byte
    want := []byte{10,11}
        {{- else }}
    // want: [N]byte-like
    var want {{ $rf.GoType }}; copy(want[:], []byte{10,11,12,13})
        {{- end }}
      {{- else if eq $rf.Type.Name "string" }}
    want := {{ $rf.GoType }}("want")
      {{- else if eq $rf.Type.Name "bool" }}
    want := {{ $rf.GoType }}(true)
      {{- else if or (eq $rf.Type.Name "double") (eq $rf.Type.Name "float") }}
    want := {{ $rf.GoType }}(2.5)
      {{- else }}
    want := {{ $rf.GoType }}(2)
      {{- end }}
    impl.ret_{{ .Name }} = want
    {{- end }}

    {{- range .Req.Fields }}
      {{- if and (isMsg .Type) .Type.IsRepeated }}
    var {{ .Name }} {{ .GoType }}
    {{ .Name }} = make({{ .GoType }}, 2)
    {{ .Name }}[0] = fixture_{{ .Type.Name }}()
    {{ .Name }}[1] = fixture_{{ .Type.Name }}()
      {{- else if isMsg .Type }}
    var {{ .Name }} {{ .GoType }} = fixture_{{ .Type.Name }}()
      {{- else if and .Type.IsRepeated (eq .Type.Name "bytes") }}
    var {{ .Name }} {{ .GoType }}
      {{- if isSliceType (sliceElem .) }}
    {{ .Name }} = [][]byte{ []byte{4,5,6}, []byte{7,8,9} }
      {{- else }}
    {
      var e0 {{ sliceElem . }}; copy(e0[:], []byte{1,2,3,4})
      var e1 {{ sliceElem . }}; copy(e1[:], []byte{5,6,7,8})
      {{ .Name }} = []{{ sliceElem . }}{ e0, e1 }
    }
      {{- end }}
      {{- else if .Type.IsRepeated }}
    var {{ .Name }} {{ .GoType }} = {{ .GoType }}{ {{ sliceElem . }}(1), {{ sliceElem . }}(2) }
      {{- else if eq .Type.Name "bytes" }}
        {{- if isSliceType .GoType }}
    var {{ .Name }} {{ .GoType }} = []byte{4,5,6}
        {{- else }}
    var {{ .Name }} {{ .GoType }}; copy({{ .Name }}[:], []byte{9,8,7,6})
        {{- end }}
      {{- else if eq .Type.Name "string" }}
    var {{ .Name }} {{ .GoType }} = {{ .GoType }}("in")
      {{- else if eq .Type.Name "bool" }}
    var {{ .Name }} {{ .GoType }} = {{ .GoType }}(true)
      {{- else if or (eq .Type.Name "double") (eq .Type.Name "float") }}
    var {{ .Name }} {{ .GoType }} = {{ .GoType }}(1.25)
      {{- else }}
    var {{ .Name }} {{ .GoType }} = {{ .GoType }}(3)
      {{- end }}
    {{- end }}

    {{- if .Rep.Fields }}
    got, err := c.{{ .Name }}(ctx{{ range .Req.Fields }}, {{ .Name }}{{ end }})
    if err != nil { t.Fatalf("rpc error: %v", err) }
    if !reflect.DeepEqual(got, impl.ret_{{ .Name }}) {
      t.Fatalf("result mismatch:\n got  = %#v\n want = %#v", got, impl.ret_{{ .Name }})
    }
    {{- else }}
    _, err := c.{{ .Name }}(ctx{{ range .Req.Fields }}, {{ .Name }}{{ end }})
    if err != nil { t.Fatalf("rpc error: %v", err) }
    {{- end }}
  })
  {{- end }}
}
`
