package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/cover"
)

func main() {
	total, err := pkgOutput()
	if err != nil {
		panic(err)
	}

	level := 100.00
	var results []string
	for _, pkg := range strings.Split(os.Getenv("PACKAGES"), ",") {
		v, ok := total[pkg]
		if ok {
			cov := percent(v[0], v[1])
			if cov < level {
				level = cov
			}
			results = append(results, fmt.Sprintf("%s %.1f%%", pkg, cov))
		}
	}
	if results != nil {
		switch {
		case level >= 99:
			fmt.Printf("::notice::%s\r\n", strings.Join(results, ", "))
		case level >= 90:
			fmt.Printf("::warning::%s\r\n", strings.Join(results, ", "))
		default:
			fmt.Printf("::error::%s\r\n", strings.Join(results, ", "))
		}
	}
}

func pkgOutput() (map[string][2]int64, error) {
	profiles, err := cover.ParseProfiles("coverage.out")
	if err != nil {
		return nil, err
	}

	total := map[string][2]int64{}
	for _, profile := range profiles {
		fn := profile.FileName
		funcs, err := findFuncs(strings.TrimPrefix(fn, "github.com/eudore/eudore/"))
		if err != nil {
			return nil, err
		}

		// Now match up functions and profile blocks.
		pkg := path.Dir(fn)
		for _, f := range funcs {
			c, t := f.coverage(profile)
			total[pkg] = [2]int64{total[pkg][0] + c, total[pkg][1] + t}
			if c != t {
				fmt.Printf("%s \t%.1f%% \t%s:%d-%d\r\n",
					fn, percent(c, t),
					f.name, f.startLine, f.endLine,
				)
			}
		}
	}
	return total, nil
}

// findFuncs parses the file and returns a slice of FuncExtent descriptors.
func findFuncs(name string) ([]*FuncExtent, error) {
	fset := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fset, name, nil, 0)
	if err != nil {
		return nil, err
	}
	visitor := &FuncVisitor{
		fset:    fset,
		name:    name,
		astFile: parsedFile,
	}
	ast.Walk(visitor, visitor.astFile)
	return visitor.funcs, nil
}

// FuncExtent describes a function's extent in the source by file and position.
type FuncExtent struct {
	name      string
	startLine int
	startCol  int
	endLine   int
	endCol    int
}

// FuncVisitor implements the visitor that builds the function position list for a file.
type FuncVisitor struct {
	fset    *token.FileSet
	name    string // Name of file.
	astFile *ast.File
	funcs   []*FuncExtent
}

// Visit implements the ast.Visitor interface.
func (v *FuncVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if n.Body == nil {
			// Do not count declarations of assembly functions.
			break
		}
		start := v.fset.Position(n.Pos())
		end := v.fset.Position(n.End())
		fe := &FuncExtent{
			name:      n.Name.Name,
			startLine: start.Line,
			startCol:  start.Column,
			endLine:   end.Line,
			endCol:    end.Column,
		}
		v.funcs = append(v.funcs, fe)
	}
	return v
}

// coverage returns the fraction of the statements in the function that were covered, as a numerator and denominator.
func (f *FuncExtent) coverage(profile *cover.Profile) (num, den int64) {
	// We could avoid making this n^2 overall by doing a single scan and annotating the functions,
	// but the sizes of the data structures is never very large and the scan is almost instantaneous.
	var covered, total int64
	// The blocks are sorted, so we can stop counting as soon as we reach the end of the relevant block.
	for _, b := range profile.Blocks {
		if b.StartLine > f.endLine || (b.StartLine == f.endLine && b.StartCol >= f.endCol) {
			// Past the end of the function.
			break
		}
		if b.EndLine < f.startLine || (b.EndLine == f.startLine && b.EndCol <= f.startCol) {
			// Before the beginning of the function
			continue
		}
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	return covered, total
}

func percent(covered, total int64) float64 {
	if total == 0 {
		total = 1 // Avoid zero denominator.
	}
	return 100.0 * float64(covered) / float64(total)
}
