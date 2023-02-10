package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	dirPerm  = 0750
	filePerm = 0666
)

var pkgPath = flag.String("pkg-path", "", "the path to a Go package containing .go files")
var outputDirPath = flag.String("out-dir", "", "the full path to the directory where the function types should be stored")
var verbose = flag.Bool("verbose", false, "show verbose log output?")

var cfg = &packages.Config{
	Mode:       packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedTypesInfo | packages.NeedTypes,
	Context:    nil,
	Logf:       nil,
	Dir:        "",
	Env:        nil,
	BuildFlags: nil,
	Fset:       nil,
	ParseFile:  nil,
	Tests:      false,
	Overlay:    nil,
}

func main() {
	flag.Parse()

	if verbose != nil && *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	if pkgPath == nil || *pkgPath == "" {
		logrus.Fatalf("--pkg-path is required")
	}

	if outputDirPath == nil || *outputDirPath == "" {
		logrus.Fatalf("--out-file is required")
	}

	pkgName := filepath.Base(*pkgPath)

	fileName, err := firstGoFileInDirectory(*pkgPath)
	if err != nil {
		logrus.Fatal(err)
	}

	filePath := path.Join(*pkgPath, fileName)
	logrus.Debugf("filePath: %s", filePath)

	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Debugf("packages loaded: %+v", pkgs)

	outputBuilder := &strings.Builder{}
	outputBuilder.WriteString(packageLine())

	if err := buildTypesInPackages(pkgs, outputBuilder); err != nil {
		logrus.Fatal(err)
	}

	outFileName := fmt.Sprintf("%s_functypes.go", pkgName)
	outFilePath := path.Join(*outputDirPath, outFileName)
	logrus.Debugf("outFilePath: %s", outFilePath)

	if err := writeOutput(outFilePath, []byte(outputBuilder.String())); err != nil {
		logrus.Fatal(err)
	}
	logrus.Infof("saved %s", outFilePath)
}

func firstGoFileInDirectory(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %v", dir, err)
	}

	logrus.Debugf("found %d entries in directory %s", len(entries), *pkgPath)

	for _, entry := range entries {
		logrus.Debugf("%s", entry.Name())

		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		return entry.Name(), nil
	}

	return "", fmt.Errorf("found no .go files in %s", dir)
}

func buildTypesInPackages(pkgs []*packages.Package, outputBuilder *strings.Builder) error {
	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		logrus.Debugf("%s scope: %v", pkg.PkgPath, scope.Names())

		for _, scopeName := range scope.Names() {
			lookupScope(scope, scopeName, outputBuilder)
		}
	}
	return nil
}

func lookupScope(scope *types.Scope, scopeName string, builder *strings.Builder) {
	obj := scope.Lookup(scopeName)

	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}

	iface, ok := named.Underlying().(*types.Interface)
	if !ok {
		return
	}

	buildInterfaceMethods(iface, builder)
}

func buildInterfaceMethods(iface *types.Interface, builder *strings.Builder) {
	for i := 0; i < iface.NumMethods(); i++ {
		method := stringifyInterfaceMethod(iface.Method(i))
		builder.WriteString(method + "\n")
		logrus.Infof("added: %s", method)
	}
}

func stringifyInterfaceMethod(meth *types.Func) string {
	sig, ok := meth.Type().Underlying().(*types.Signature)
	if !ok {
		return ""
	}
	return fmt.Sprintf("type %s %s", meth.Name(), sig.String())
}

func packageLine() string {
	return fmt.Sprintf("package functypes\n\n")
}

func writeOutput(outFilePath string, content []byte) error {
	dirPath := filepath.Dir(outFilePath)

	if err := os.MkdirAll(dirPath, dirPerm); err != nil {
		return fmt.Errorf("mkdir %s with perm %d: %w", dirPath, dirPerm, err)
	}

	if err := os.WriteFile(outFilePath, content, filePerm); err != nil {
		return fmt.Errorf("write %s with perm %d: %w", outFilePath, filePerm, err)
	}

	return nil
}
