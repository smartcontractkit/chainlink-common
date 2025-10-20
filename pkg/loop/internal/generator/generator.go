package generator

import (
	"fmt"
	"go/types"
	"os"
	"path"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

type Service struct {
	PkgPath         string
	IfaceName       string
	ServiceName     string
	Methods         []Method
	ProtoImports    []string
	GoImports       []string
	ProtoPkg        string
	OptionGoPackage string
	WrapGoPackage   string
	UserMessages    []UserMessage
	IfacePkgName    string
	ImportAliases   []ImportAlias
	InterfaceOneofs []InterfaceOneof
}

type UserMessage struct {
	Name     string
	Fields   []Field
	PkgPath  string
	DomainGo string
}

type Method struct {
	Name       string
	Req        Message
	Rep        Message
	HasContext bool
}

type Message struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name     string
	ProtoTag int
	Type     ProtoType
	GoType   string

	DomainIsArray    bool   // true if domain is [N]T
	DomainElemGoType string // element type of array/slice on the domain side
	DomainElemIsByte bool   // true if element is byte/uint8
}

type ProtoType struct {
	Name       string
	IsRepeated bool
}

type ExternalMap struct {
	GoType      string `yaml:"go_type"`
	ProtoType   string `yaml:"proto_type"`
	ToProto     string `yaml:"to_proto"`
	FromProto   string `yaml:"from_proto"`
	Import      string `yaml:"import"`
	ProtoImport string `yaml:"proto_import"`
}

type EnumMap struct {
	GoType    string   `yaml:"go_type"`
	ProtoType string   `yaml:"proto_type"`
	Values    []string `yaml:"values"`
}

type Config struct {
	Externals  []ExternalMap  `yaml:"externals"`
	Enums      []EnumMap      `yaml:"enums"`
	Interfaces []InterfaceMap `yaml:"interfaces"`
}

type InterfaceMap struct {
	GoType         string          `yaml:"go_type"`
	Strategy       string          `yaml:"strategy"`
	ProtoContainer string          `yaml:"proto_container"`
	Cases          []InterfaceCase `yaml:"cases"`
}

type InterfaceCase struct {
	GoType string `yaml:"go_type"`
}

type OneofCase struct {
	PkgPath  string
	TypeName string
	DomainGo string
}

type InterfaceOneof struct {
	Name          string
	InterfacePkg  string
	InterfaceName string
	GoInterface   string
	Cases         []OneofCase
}

type ImportAlias struct {
	Path  string
	Alias string
}

func LoadConfig(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ParseInterface(pkgPath, iface string, cfg *Config) (*Service, error) {
	svc := &Service{PkgPath: pkgPath, IfaceName: iface}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedName}, pkgPath)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 || pkgs[0].Types == nil {
		return nil, fmt.Errorf("package not found: %s", pkgPath)
	}

	pkg := pkgs[0].Types
	obj := pkg.Scope().Lookup(iface)
	if obj == nil {
		return nil, fmt.Errorf("interface %s not found in %s", iface, pkgPath)
	}

	svc.IfacePkgName = pkg.Name()
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s is not a named type", iface)
	}
	itf, ok := named.Underlying().(*types.Interface)
	if !ok {
		return nil, fmt.Errorf("%s is not an interface", iface)
	}
	itf = itf.Complete()

	reg := map[string]*UserMessage{}
	oneofReg := map[string]*InterfaceOneof{}
	domainPkgs := map[string]string{}

	for i := 0; i < itf.NumMethods(); i++ {
		m := itf.Method(i)
		sig := m.Type().(*types.Signature)
		meth, err := buildMethod(m.Name(), sig, cfg, reg, domainPkgs, oneofReg)
		if err != nil {
			return nil, fmt.Errorf("method %s: %w", m.Name(), err)
		}
		svc.Methods = append(svc.Methods, *meth)
	}

	for _, um := range reg {
		svc.UserMessages = append(svc.UserMessages, *um)
	}
	sort.Slice(svc.UserMessages, func(i, j int) bool { return svc.UserMessages[i].Name < svc.UserMessages[j].Name })

	for _, io := range oneofReg {
		svc.InterfaceOneofs = append(svc.InterfaceOneofs, *io)
	}
	sort.Slice(svc.InterfaceOneofs, func(i, j int) bool { return svc.InterfaceOneofs[i].Name < svc.InterfaceOneofs[j].Name })

	aliasMap, aliases := assignAliases(domainPkgs)
	svc.ImportAliases = aliases

	// Qualify method field types
	for mi := range svc.Methods {
		for fi := range svc.Methods[mi].Req.Fields {
			f := &svc.Methods[mi].Req.Fields[fi]
			f.GoType = goTypeStringWithAliases(f.GoType, aliasMap)
			if f.DomainElemGoType != "" {
				f.DomainElemGoType = goTypeStringWithAliases(f.DomainElemGoType, aliasMap)
			}
		}
		for fi := range svc.Methods[mi].Rep.Fields {
			f := &svc.Methods[mi].Rep.Fields[fi]
			f.GoType = goTypeStringWithAliases(f.GoType, aliasMap)
			if f.DomainElemGoType != "" {
				f.DomainElemGoType = goTypeStringWithAliases(f.DomainElemGoType, aliasMap)
			}
		}
	}

	// Qualify user messages + field types
	for i := range svc.UserMessages {
		alias := aliasMap[svc.UserMessages[i].PkgPath]
		if alias == "" {
			alias = path.Base(svc.UserMessages[i].PkgPath)
		}
		svc.UserMessages[i].DomainGo = alias + "." + svc.UserMessages[i].Name
		for fi := range svc.UserMessages[i].Fields {
			f := &svc.UserMessages[i].Fields[fi]
			f.GoType = goTypeStringWithAliases(f.GoType, aliasMap)
			if f.DomainElemGoType != "" {
				f.DomainElemGoType = goTypeStringWithAliases(f.DomainElemGoType, aliasMap)
			}
		}
	}

	// Qualify oneofs
	for i := range svc.InterfaceOneofs {
		ialias := aliasMap[svc.InterfaceOneofs[i].InterfacePkg]
		if ialias == "" {
			ialias = path.Base(svc.InterfaceOneofs[i].InterfacePkg)
		}
		svc.InterfaceOneofs[i].GoInterface = ialias + "." + svc.InterfaceOneofs[i].InterfaceName
		for j := range svc.InterfaceOneofs[i].Cases {
			c := &svc.InterfaceOneofs[i].Cases[j]
			alias := aliasMap[c.PkgPath]
			if alias == "" {
				alias = path.Base(c.PkgPath)
			}
			c.DomainGo = alias + "." + c.TypeName
		}
	}

	return svc, nil
}

// Replace full package prefixes with aliases, longest path first
func goTypeStringWithAliases(s string, aliasMap map[string]string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	paths := make([]string, 0, len(aliasMap))
	for p := range aliasMap {
		paths = append(paths, p)
	}
	sort.Slice(paths, func(i, j int) bool { return len(paths[i]) > len(paths[j]) })
	out := s
	for _, p := range paths {
		alias := aliasMap[p]
		out = strings.ReplaceAll(out, p+".", alias+".")
	}
	return out
}

func assignAliases(domainPkgs map[string]string) (map[string]string, []ImportAlias) {
	aliasMap := make(map[string]string)
	used := make(map[string]int)

	paths := make([]string, 0, len(domainPkgs))
	for p := range domainPkgs {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	for _, p := range paths {
		base := domainPkgs[p]
		if base == "" {
			base = path.Base(p)
		}
		alias := base
		if c := used[base]; c > 0 {
			alias = fmt.Sprintf("%s%d", base, c+1)
		}
		used[base]++
		aliasMap[p] = alias
	}

	var out []ImportAlias
	for _, p := range paths {
		out = append(out, ImportAlias{Path: p, Alias: aliasMap[p]})
	}
	return aliasMap, out
}

func buildMethod(name string, sig *types.Signature, cfg *Config, reg map[string]*UserMessage, domainImports map[string]string, oneofReg map[string]*InterfaceOneof) (*Method, error) {
	m := &Method{Name: name}
	params := sig.Params()
	results := sig.Results()

	recordPkg := func(t types.Type) {
		if n, ok := t.(*types.Named); ok {
			if p := n.Obj().Pkg(); p != nil {
				domainImports[p.Path()] = p.Name()
			}
		} else {
			recordPkgFromTypeString(goTypeName(t), domainImports)
		}
	}

	if results.Len() == 0 {
		return nil, fmt.Errorf("no results; need (T, error)")
	}
	if !isErrorType(results.At(results.Len() - 1).Type()) {
		return nil, fmt.Errorf("last result must be error")
	}

	start := 0
	if params.Len() > 0 && isContext(params.At(0).Type()) {
		m.HasContext = true
		start = 1
	}

	req := Message{Name: name + "Request"}
	for i := start; i < params.Len(); i++ {
		pv := params.At(i)
		recordPkg(pv.Type())
		pt, err := mapGoToProto(pv.Type(), cfg, reg, domainImports, oneofReg)
		if err != nil {
			return nil, fmt.Errorf("param %d: %w", i, err)
		}

		elem := arrayElemViaUnderlying(pv.Type())
		isArr := elem != nil
		if !isArr && pt.Name == "bytes" { // fixed [N]byte alias case
			if yes, e := underlyingByteArray(pv.Type()); yes {
				isArr = true
				elem = e
			}
		}
		// make sure we record any package from element too (e.g. []evm.Hash)
		recordPkg(elem)

		req.Fields = append(req.Fields, Field{
			Name:             exported(pv.Name(), i-start),
			ProtoTag:         len(req.Fields) + 1,
			Type:             pt,
			GoType:           pv.Type().String(),
			DomainIsArray:    isArr,
			DomainElemGoType: goTypeOrEmpty(elem),
			DomainElemIsByte: isByteOrUint8(elem),
		})
	}
	m.Req = req

	if results.Len() == 1 {
		m.Rep = Message{Name: name + "Reply"}
	} else {
		rt := results.At(0).Type()
		recordPkg(rt)
		pt, err := mapGoToProto(rt, cfg, reg, domainImports, oneofReg)
		if err != nil {
			return nil, fmt.Errorf("result: %w", err)
		}

		elem := arrayElemViaUnderlying(rt)
		isArr := elem != nil
		if !isArr && pt.Name == "bytes" {
			if yes, e := underlyingByteArray(rt); yes {
				isArr = true
				elem = e
			}
		}
		recordPkg(elem)

		m.Rep = Message{
			Name: name + "Reply",
			Fields: []Field{{
				Name:             "result",
				ProtoTag:         1,
				Type:             pt,
				GoType:           rt.String(),
				DomainIsArray:    isArr,
				DomainElemGoType: goTypeOrEmpty(elem),
				DomainElemIsByte: isByteOrUint8(elem),
			}},
		}
	}
	return m, nil
}

func isContext(t types.Type) bool {
	p, ok := t.(*types.Named)
	if !ok {
		return false
	}
	return p.Obj().Pkg() != nil && p.Obj().Pkg().Path() == "context" && p.Obj().Name() == "Context"
}

func isErrorType(t types.Type) bool {
	ni, ok := t.Underlying().(*types.Interface)
	return ok && ni.NumMethods() == 1 && ni.Method(0).Name() == "Error"
}

func exported(name string, idx int) string {
	if name == "" {
		name = fmt.Sprintf("arg%d", idx+1)
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func mapGoToProto(t types.Type, cfg *Config, reg map[string]*UserMessage, domainImports map[string]string, oneofReg map[string]*InterfaceOneof) (ProtoType, error) {
	switch u := t.(type) {
	case *types.Named:
		qn := qualifiedName(u)

		// ensure import
		if p := u.Obj().Pkg(); p != nil {
			domainImports[p.Path()] = p.Name()
		}

		// interface -> oneof container
		if _, isIface := u.Underlying().(*types.Interface); isIface {
			im, ok := findInterfaceCfg(qn, cfg)
			if !ok {
				return ProtoType{}, fmt.Errorf("interface %q in RPC signature has no mapping (config.interfaces)", qn)
			}
			if _, exists := oneofReg[qn]; !exists {
				io := &InterfaceOneof{
					Name:          im.ProtoContainer,
					InterfacePkg:  u.Obj().Pkg().Path(),
					InterfaceName: u.Obj().Name(),
				}
				for _, c := range im.Cases {
					nc, err := lookupNamed(c.GoType)
					if err != nil {
						return ProtoType{}, err
					}
					if p := nc.Obj().Pkg(); p != nil {
						domainImports[p.Path()] = p.Name()
					}
					st, ok := nc.Underlying().(*types.Struct)
					if !ok {
						return ProtoType{}, fmt.Errorf("interface case %q is not a named struct", c.GoType)
					}
					ensureUserMessage(nc, st, cfg, reg, domainImports, oneofReg)
					io.Cases = append(io.Cases, OneofCase{
						PkgPath:  nc.Obj().Pkg().Path(),
						TypeName: nc.Obj().Name(),
					})
				}
				oneofReg[qn] = io
			}
			return ProtoType{Name: im.ProtoContainer}, nil
		}

		// externals / enums
		for _, ex := range cfg.Externals {
			if ex.GoType == qn {
				return ProtoType{Name: ex.ProtoType}, nil
			}
		}
		for _, em := range cfg.Enums {
			if em.GoType == qn {
				return ProtoType{Name: em.ProtoType}, nil
			}
		}

		// struct -> user message
		if st, ok := u.Underlying().(*types.Struct); ok {
			ensureUserMessage(u, st, cfg, reg, domainImports, oneofReg)
			return ProtoType{Name: u.Obj().Name()}, nil
		}

		// alias: delegate to underlying so arrays/slices/scalars map correctly
		return mapGoToProto(u.Underlying(), cfg, reg, domainImports, oneofReg)

	case *types.Pointer:
		return mapGoToProto(u.Elem(), cfg, reg, domainImports, oneofReg)

	case *types.Slice:
		if isUint8(u.Elem()) {
			return ProtoType{Name: "bytes"}, nil
		}
		pt, err := mapGoToProto(u.Elem(), cfg, reg, domainImports, oneofReg)
		if err != nil {
			return ProtoType{}, err
		}
		pt.IsRepeated = true
		return pt, nil

	default:
		return mapPrimitiveToProto(t, cfg)
	}
}

func mapPrimitiveToProto(t types.Type, cfg *Config) (ProtoType, error) {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.Bool:
			return ProtoType{Name: "bool"}, nil
		case types.String:
			return ProtoType{Name: "string"}, nil
		case types.Int8, types.Int16, types.Int32:
			return ProtoType{Name: "int32"}, nil
		case types.Int, types.Int64:
			return ProtoType{Name: "int64"}, nil
		case types.Uint8, types.Uint16, types.Uint32:
			return ProtoType{Name: "uint32"}, nil
		case types.Uint, types.Uint64, types.Uintptr:
			return ProtoType{Name: "uint64"}, nil
		case types.UnsafePointer:
			return ProtoType{}, fmt.Errorf("unsafe.Pointer unsupported")
		default:
			return ProtoType{}, fmt.Errorf("unsupported basic kind: %v", u.Kind())
		}

	case *types.Slice:
		pt, err := mapPrimitiveToProto(u.Elem(), cfg)
		if err != nil {
			return ProtoType{}, err
		}
		if pt.Name == "uint32" && isUint8(u.Elem()) { // []byte
			return ProtoType{Name: "bytes"}, nil
		}
		pt.IsRepeated = true
		return pt, nil

	case *types.Array:
		if isUint8(u.Elem()) {
			return ProtoType{Name: "bytes"}, nil
		}
		pt, err := mapPrimitiveToProto(u.Elem(), cfg)
		if err != nil {
			return ProtoType{}, err
		}
		pt.IsRepeated = true
		return pt, nil

	case *types.Pointer:
		return mapPrimitiveToProto(u.Elem(), cfg)

	case *types.Named:
		qn := qualifiedName(u)
		for _, ex := range cfg.Externals {
			if ex.GoType == qn {
				return ProtoType{Name: ex.ProtoType}, nil
			}
		}
		for _, em := range cfg.Enums {
			if em.GoType == qn {
				return ProtoType{Name: em.ProtoType}, nil
			}
		}
		// unwrap alias so arrays/scalars map correctly
		return mapPrimitiveToProto(u.Underlying(), cfg)

	default:
		return ProtoType{}, fmt.Errorf("unsupported type: %T", u)
	}
}

func ensureUserMessage(n *types.Named, st *types.Struct, cfg *Config, reg map[string]*UserMessage, domainImports map[string]string, oneofReg map[string]*InterfaceOneof) {
	qn := qualifiedName(n)
	if _, exists := reg[qn]; exists {
		return
	}

	um := &UserMessage{
		Name:    n.Obj().Name(),
		PkgPath: n.Obj().Pkg().Path(),
	}
	reg[qn] = um

	tag := 1
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !f.Exported() {
			continue
		}

		// record package even if alias (e.g. evm.Address)
		if p := f.Pkg(); p != nil {
			domainImports[p.Path()] = p.Name()
		}
		recordPkgFromTypeString(goTypeName(f.Type()), domainImports)

		ft, err := mapGoToProto(f.Type(), cfg, reg, domainImports, oneofReg)
		if err != nil {
			continue
		}

		elem := arrayElemViaUnderlying(f.Type())
		isArr := elem != nil
		if !isArr && ft.Name == "bytes" {
			if yes, e := underlyingByteArray(f.Type()); yes {
				isArr = true
				elem = e
			}
		}
		// also record the element's package (e.g. []evm.Hash)
		recordPkgFromType(elem, domainImports)

		pname := toSnake(f.Name())

		um.Fields = append(um.Fields, Field{
			Name:             pname,
			ProtoTag:         tag,
			Type:             ft,
			GoType:           f.Type().String(),
			DomainIsArray:    isArr,
			DomainElemGoType: goTypeOrEmpty(elem),
			DomainElemIsByte: isByteOrUint8(elem),
		})
		tag++
	}

	reg[qn] = um
}

func splitQ(q string) (pkgPath, typeName string) {
	i := strings.LastIndex(q, ".")
	if i < 0 {
		return "", q
	}
	return q[:i], q[i+1:]
}

func lookupNamed(q string) (*types.Named, error) {
	p, tn := splitQ(q)
	if p == "" || tn == "" {
		return nil, fmt.Errorf("invalid qualified name: %q", q)
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedName}, p)
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 || pkgs[0].Types == nil {
		return nil, fmt.Errorf("package not found: %s", p)
	}
	obj := pkgs[0].Types.Scope().Lookup(tn)
	if obj == nil {
		return nil, fmt.Errorf("type %s not found in %s", tn, p)
	}
	n, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s in %s is not a named type", tn, p)
	}
	return n, nil
}

func findInterfaceCfg(qn string, cfg *Config) (InterfaceMap, bool) {
	for _, im := range cfg.Interfaces {
		if im.GoType == qn && strings.EqualFold(im.Strategy, "oneof") {
			return im, true
		}
	}
	return InterfaceMap{}, false
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(toLower(r))
	}
	return b.String()
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

func isUint8(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.Uint8
}

func qualifiedName(n *types.Named) string {
	pkg := n.Obj().Pkg()
	if pkg == nil {
		return n.Obj().Name()
	}
	return pkg.Path() + "." + n.Obj().Name()
}

func arrayInfo(t types.Type) (isArray bool, elem types.Type) {
	switch u := t.(type) {
	case *types.Array:
		return true, u.Elem()
	case *types.Named:
		return arrayInfo(u.Underlying())
	case *types.Pointer:
		return arrayInfo(u.Elem())
	default:
		return false, nil
	}
}

// returns array element type after unwrapping named/pointer; nil if not an array
func arrayElemViaUnderlying(t types.Type) types.Type {
	for {
		switch u := t.(type) {
		case *types.Array:
			return u.Elem()
		case *types.Named:
			t = u.Underlying()
		case *types.Pointer:
			t = u.Elem()
		default:
			return nil
		}
	}
}

func elemTypeOf(t types.Type) types.Type {
	switch u := t.(type) {
	case *types.Slice:
		return u.Elem()
	case *types.Array:
		return u.Elem()
	case *types.Pointer:
		return elemTypeOf(u.Elem())
	case *types.Named:
		return elemTypeOf(u.Underlying())
	default:
		return nil
	}
}

func isByteOrUint8(t types.Type) bool {
	for {
		switch u := t.(type) {
		case *types.Named:
			t = u.Underlying()
		case *types.Pointer:
			t = u.Elem()
		default:
			goto EXIT
		}
	}
EXIT:
	b, ok := t.(*types.Basic)
	if !ok {
		return false
	}
	return b.Kind() == types.Byte || b.Kind() == types.Uint8
}

func goTypeName(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p == nil {
			return ""
		}
		return p.Path()
	})
}

func goTypeOrEmpty(t types.Type) string {
	if t == nil {
		return ""
	}
	return goTypeName(t)
}

func underlyingByteArray(t types.Type) (bool, types.Type) {
	for {
		switch u := t.(type) {
		case *types.Pointer:
			t = u.Elem()
		case *types.Named:
			t = u.Underlying()
		case *types.Array:
			if isByteOrUint8(u.Elem()) {
				return true, u.Elem()
			}
			return false, nil
		default:
			return false, nil
		}
	}
}

// record packages for named types or, if alias erased, try to recover from type string
func recordPkgFromType(t types.Type, domainImports map[string]string) {
	if t == nil {
		return
	}
	if n, ok := t.(*types.Named); ok {
		if p := n.Obj().Pkg(); p != nil {
			domainImports[p.Path()] = p.Name()
			return
		}
	}
	recordPkgFromTypeString(goTypeName(t), domainImports)
}

func recordPkgFromTypeString(s string, domainImports map[string]string) {
	// strip common wrappers
	trim := func(x string) string {
		for {
			switch {
			case strings.HasPrefix(x, "[]"):
				x = x[2:]
			case strings.HasPrefix(x, "*"):
				x = x[1:]
			case strings.HasPrefix(x, "map["):
				// drop until ']'
				if i := strings.IndexByte(x, ']'); i >= 0 {
					x = x[i+1:]
				} else {
					return x
				}
			case strings.HasPrefix(x, "chan "):
				x = x[5:]
			default:
				return x
			}
		}
	}
	x := trim(s)
	if i := strings.LastIndex(x, "."); i > 0 {
		p := x[:i]
		// still may include wrappers; ensure it looks like a path
		if strings.Contains(p, "/") {
			domainImports[p] = path.Base(p)
		}
	}
}
