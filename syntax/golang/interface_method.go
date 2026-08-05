package golang

import (
	"go/ast"
	"go/types"
)

// IsInterfaceMethod reports whether the given FuncDecl implements a method
// required by an interface declared in the same package. Requires type info
// from go/packages (type-aware mode).
//
// The function scans the package scope for interface types and checks two
// conditions for each: (1) the interface has a method with the same name as
// fn.Name, and (2) the receiver type fully implements that interface (checked
// for both value and pointer receivers). If both hold, the method is
// interface-driven boilerplate, not copy-paste.
//
// Only same-package interfaces are checked. Stdlib interfaces (fmt.Stringer,
// io.Reader, etc.) are covered by the static name list in the actionability
// layer as a complementary fallback.
func IsInterfaceMethod(info *types.Info, fn *ast.FuncDecl) bool {
	if info == nil || fn == nil || fn.Recv == nil || fn.Name == nil {
		return false
	}

	// Resolve the method object from type info.
	obj, ok := info.Defs[fn.Name]
	if !ok || obj == nil {
		return false
	}

	funcObj, ok := obj.(*types.Func)
	if !ok {
		return false
	}

	// Extract the receiver type from the method signature.
	sig, ok := funcObj.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}

	recvType := sig.Recv().Type()

	pkg := funcObj.Pkg()
	if pkg == nil {
		return false
	}

	methodName := fn.Name.Name
	pointerRecv := types.NewPointer(recvType)

	// Scan the package scope for interface types.
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		scopeObj := scope.Lookup(name)
		if scopeObj == nil {
			continue
		}

		typeName, ok := scopeObj.(*types.TypeName)
		if !ok {
			continue
		}

		iface, ok := typeName.Type().Underlying().(*types.Interface)
		if !ok {
			continue
		}

		// Quick check: does the interface declare a method with this name?
		if !interfaceHasMethod(iface, methodName) {
			continue
		}

		// Full check: does the receiver type implement the entire interface?
		if types.Implements(recvType, iface) || types.Implements(pointerRecv, iface) {
			return true
		}
	}

	return false
}

// interfaceHasMethod reports whether the interface declares a method with the
// given name (including embedded methods, which go/types flattens).
func interfaceHasMethod(iface *types.Interface, name string) bool {
	for i := range iface.NumMethods() {
		if iface.Method(i).Name() == name {
			return true
		}
	}

	return false
}
