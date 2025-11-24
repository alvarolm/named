package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	generatedFileSuffix = "_named_generated.go"
	testFileSuffix      = "_test.go"
	defaultTagKey       = "json"
	directivePrefix     = "GENERATE-NAMED="
	structNameKey       = "StructName"
	tagKeyKey           = "TagKey"
)

type structInfo struct {
	name    string
	tagKey  string
	fields  []fieldInfo
	pkgName string
}

type fieldInfo struct {
	name    string
	tagName string
}

var (
	verbose bool
	clean   bool
)

func logVerbose(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, "[verbose] "+format+"\n", args...)
	}
}

// walkGoPackages recursively walks directories and calls fn for each directory
// that could be a Go package (contains .go files, not hidden, not following symlinks)
func walkGoPackages(root string, fn func(string) error) error {
	info, err := os.Lstat(root) // Use Lstat to not follow symlinks
	if err != nil {
		return err
	}

	// Don't follow symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		logVerbose("Skipping symlink: %s", root)
		return nil
	}

	if !info.IsDir() {
		return nil
	}

	// Skip hidden directories
	if strings.HasPrefix(filepath.Base(root), ".") {
		logVerbose("Skipping hidden directory: %s", root)
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	// Check if this directory has .go files (potential Go package)
	hasGoFiles := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			hasGoFiles = true
			break
		}
	}

	// Process this directory if it has Go files
	if hasGoFiles {
		if err := fn(root); err != nil {
			return err
		}
	}

	// Recurse into subdirectories
	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(root, entry.Name())
			if err := walkGoPackages(subPath, fn); err != nil {
				return err
			}
		}
	}

	return nil
}

func cleanGeneratedFiles(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		// If it's a file, check if it's a generated file and delete it
		if strings.HasSuffix(path, generatedFileSuffix) {
			logVerbose("Removing: %s", path)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("error removing %s: %v", path, err)
			}
			fmt.Printf("Removed: %s\n", path)
		}
		return nil
	}

	// If it's a directory, recursively clean all Go packages
	return walkGoPackages(path, func(dir string) error {
		logVerbose("Cleaning directory: %s", dir)

		entries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(entry.Name(), generatedFileSuffix) {
				fullPath := filepath.Join(dir, entry.Name())
				logVerbose("Removing: %s", fullPath)
				if err := os.Remove(fullPath); err != nil {
					return fmt.Errorf("error removing %s: %v", fullPath, err)
				}
				fmt.Printf("Removed: %s\n", fullPath)
			}
		}

		return nil
	})
}

func main() {
	// Define flags
	flag.BoolVar(&verbose, "v", false, "verbose mode: show detailed processing information")
	flag.BoolVar(&verbose, "verbose", false, "verbose mode: show detailed processing information")
	flag.BoolVar(&clean, "clean", false, "remove all generated *_named_generated.go files")

	// Set custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: generate-named [flags] [path...]\n\n")
		fmt.Fprintf(os.Stderr, "Generates type-safe field name accessors for Go structs.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  path    File or directory to process (default: current directory)\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  generate-named                    # Process current directory\n")
		fmt.Fprintf(os.Stderr, "  generate-named -v                 # Process with verbose output\n")
		fmt.Fprintf(os.Stderr, "  generate-named -clean             # Remove all generated files\n")
		fmt.Fprintf(os.Stderr, "  generate-named ./pkg              # Process specific directory\n")
		fmt.Fprintf(os.Stderr, "  generate-named file.go            # Process specific file\n\n")
		fmt.Fprintf(os.Stderr, "For each struct with a GENERATE-NAMED directive, creates a *_named_generated.go file\n")
		fmt.Fprintf(os.Stderr, "with methods to access field names based on struct tags.\n")
	}

	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		args = []string{"."}
	}

	// Handle clean mode
	if clean {
		for _, path := range args {
			if err := cleanGeneratedFiles(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error cleaning %s: %v\n", path, err)
				os.Exit(1)
			}
		}
		return
	}

	// Normal generation mode
	for _, path := range args {
		if err := processPath(path); err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

func processPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Recursively process all Go package directories
		return walkGoPackages(path, processDir)
	}
	return processFile(path, nil)
}

func processDir(dir string) error {
	logVerbose("Processing package directory: %s", dir)

	// Single pass: parse all Go files once, collecting both directives and AST
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	type parsedFile struct {
		path string
		node *ast.File
	}

	var parsedFiles []parsedFile
	globalDirectives := make(map[string]string)
	fset := token.NewFileSet()

	// Parse each file once, collecting directives and AST nodes
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), testFileSuffix) || strings.HasSuffix(entry.Name(), generatedFileSuffix) {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		logVerbose("Parsing file: %s", fullPath)
		node, err := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing %s: %v", fullPath, err)
		}

		// Collect directives from this file
		fileDirectives := parseGenerateComments(node)
		if len(fileDirectives) > 0 {
			logVerbose("Found %d directive(s) in %s", len(fileDirectives), entry.Name())
		}
		for structName, tagKey := range fileDirectives {
			logVerbose("  - %s (TagKey: %s)", structName, tagKey)
			// Check for conflicting directives
			if existingTagKey, exists := globalDirectives[structName]; exists {
				if existingTagKey != tagKey {
					return fmt.Errorf("conflicting GENERATE-NAMED directives for struct %s: TagKey %q vs %q",
						structName, existingTagKey, tagKey)
				}
				// Same directive, skip (idempotent)
				continue
			}
			globalDirectives[structName] = tagKey
		}

		// Store parsed file for struct extraction
		parsedFiles = append(parsedFiles, parsedFile{
			path: fullPath,
			node: node,
		})
	}

	// Process parsed files to find structs and generate code
	logVerbose("Processing %d parsed file(s) to find structs", len(parsedFiles))
	for _, pf := range parsedFiles {
		structs := findAnnotatedStructs(pf.node, globalDirectives)
		if len(structs) > 0 {
			logVerbose("Found %d struct(s) in %s", len(structs), filepath.Base(pf.path))
			for _, s := range structs {
				logVerbose("  - %s (%d fields)", s.name, len(s.fields))
			}
			if err := generateCode(pf.path, structs); err != nil {
				return err
			}
		}
	}

	return nil
}

func processFile(filename string, globalDirectives map[string]string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// If no global directives provided (single file mode), collect from this file
	if globalDirectives == nil {
		globalDirectives = parseGenerateComments(node)
	}

	structs := findAnnotatedStructs(node, globalDirectives)
	if len(structs) == 0 {
		return nil
	}

	return generateCode(filename, structs)
}

func findAnnotatedStructs(file *ast.File, structTagKeys map[string]string) []structInfo {
	var results []structInfo

	if len(structTagKeys) == 0 {
		return results
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			// Check if this struct has a GENERATE-NAMED directive
			tagKey, found := structTagKeys[typeSpec.Name.Name]
			if !found {
				continue
			}

			// Extract field information
			var fields []fieldInfo
			for _, field := range structType.Fields.List {
				// Skip unexported fields
				if len(field.Names) == 0 || !field.Names[0].IsExported() {
					continue
				}

				fieldName := field.Names[0].Name
				tagName := extractTagName(field.Tag, tagKey)

				// Skip fields with tag:"-"
				if tagName == "-" {
					continue
				}

				// Use field name if no tag specified
				if tagName == "" {
					tagName = fieldName
				}

				fields = append(fields, fieldInfo{
					name:    fieldName,
					tagName: tagName,
				})
			}

			if len(fields) > 0 {
				results = append(results, structInfo{
					name:    typeSpec.Name.Name,
					tagKey:  tagKey,
					fields:  fields,
					pkgName: file.Name.Name,
				})
			}
		}
	}

	return results
}

// parseGenerateComments scans all comments in the file for GENERATE-NAMED directives
// Returns a map of struct name to tag key
func parseGenerateComments(file *ast.File) map[string]string {
	result := make(map[string]string)

	// Collect all comment groups
	var allComments []*ast.CommentGroup
	for _, cg := range file.Comments {
		allComments = append(allComments, cg)
	}

	// Parse each comment
	for _, commentGroup := range allComments {
		for _, comment := range commentGroup.List {
			text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))

			// Check for format: GENERATE-NAMED=StructName:[name],TagKey:[key]
			if strings.HasPrefix(text, directivePrefix) {
				structName, tagKey := parseStructDirective(text)
				if structName != "" {
					result[structName] = tagKey
				}
			}
		}
	}

	return result
}

// parseStructDirective parses a directive like "GENERATE-NAMED=StructName:Foo,TagKey:db"
// Returns the struct name and tag key (uses default if not specified)
func parseStructDirective(text string) (string, string) {
	var structName string
	var tagKey string = defaultTagKey

	// Remove GENERATE-NAMED= prefix
	text = strings.TrimPrefix(text, directivePrefix)

	// Split by comma to get key-value pairs
	parts := strings.Split(text, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Split by colon
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case structNameKey:
			structName = value
		case tagKeyKey:
			tagKey = value
		}
	}

	return structName, tagKey
}

// extractTagName extracts the tag value for a given key from a struct tag
func extractTagName(tag *ast.BasicLit, key string) string {
	if tag == nil {
		return ""
	}

	// Remove backticks and use reflect.StructTag for proper parsing
	tagStr := strings.Trim(tag.Value, "`")

	// Use reflect.StructTag.Get() which properly handles:
	// - Quoted values with whitespace
	// - Multiple tag keys
	// - Proper escaping
	value := reflect.StructTag(tagStr).Get(key)

	// Extract only the name part before comma (ignore options like omitempty)
	if comma := strings.Index(value, ","); comma != -1 {
		return value[:comma]
	}
	return value
}

func generateCode(sourceFile string, structs []structInfo) error {
	if len(structs) == 0 {
		return nil
	}

	var buf bytes.Buffer

	// Write header
	fmt.Fprintf(&buf, "// Code generated by generate-named. DO NOT EDIT.\n\n")
	fmt.Fprintf(&buf, "package %s\n\n", structs[0].pkgName)

	// Generate code for each struct
	for _, s := range structs {
		if err := generateStructCode(&buf, s); err != nil {
			return err
		}
	}

	// Format the generated code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting error: %v\n%s", err, buf.String())
	}

	// Determine output filename
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	outputFile := filepath.Join(dir, nameWithoutExt+generatedFileSuffix)

	// Write to file
	if err := os.WriteFile(outputFile, formatted, 0644); err != nil {
		return err
	}

	fmt.Printf("Generated: %s\n", outputFile)
	return nil
}

func generateStructCode(buf *bytes.Buffer, s structInfo) error {
	// Validate struct name to prevent panic
	if len(s.name) == 0 {
		return fmt.Errorf("invalid struct name: empty string")
	}

	// Create private struct name (lowercase first letter) and public variable name
	privateStructName := strings.ToLower(s.name[:1]) + s.name[1:] + "Named"
	publicVarName := s.name + "Named"

	// Generate the private struct type
	fmt.Fprintf(buf, "// %s provides methods to access field names of %s\n", privateStructName, s.name)
	fmt.Fprintf(buf, "type %s struct{}\n\n", privateStructName)

	// Generate methods for each field
	for _, field := range s.fields {
		fmt.Fprintf(buf, "func (%s) %s() string {", privateStructName, field.name)
		fmt.Fprintf(buf, "\treturn %q", field.tagName)
		fmt.Fprintf(buf, "}\n")
	}

	// Generate the exported variable
	fmt.Fprintf(buf, "// %s is the exported variable for accessing %s field names\n", publicVarName, s.name)
	fmt.Fprintf(buf, "var %s %s\n\n", publicVarName, privateStructName)

	return nil
}
