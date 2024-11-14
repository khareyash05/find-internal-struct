package main

import (
  "bufio"
  "fmt"
  "go/ast"
  "go/parser"
  "go/token"
  "os"
  "path/filepath"
  "strings"
)

func main() {
  // Specify the main file you are analyzing
  mainFilePath := "/home/runner/ThinUntimelyNanocad/utils/h.go"
  projectRoot := "/home/runner/ThinUntimelyNanocad"
  outputFilePath := "./structs_log.txt"
  moduleName, err := getModuleName(filepath.Join(projectRoot, "go.mod"))
  if err != nil {
    fmt.Println("Error reading go.mod:", err)
    return
  }

  // Open or create the output file
  logFile, err := os.Create(outputFilePath)
  if err != nil {
    fmt.Println("Error creating log file:", err)
    return
  }
  defer logFile.Close()

  // Parse imports in the main file
  imports, err := getImports(mainFilePath)
  if err != nil {
    fmt.Println("Error getting imports:", err)
    return
  }

  // Filter imports to only include internal packages (those starting with the module name)
  internalImports := filterInternalImports(imports, moduleName)

  // For each internal import, locate and parse files to find structs
  for _, imp := range internalImports {
    logFile.WriteString(fmt.Sprintf("Structs in internal package: %s\n", imp))
    err := findStructsInPackage(projectRoot, imp, moduleName, logFile)
    if err != nil {
      fmt.Println("Error parsing package:", err)
    }
  }
}

// getModuleName reads the module name from the go.mod file
func getModuleName(goModPath string) (string, error) {
  file, err := os.Open(goModPath)
  if err != nil {
    return "", err
  }
  defer file.Close()

  scanner := bufio.NewScanner(file)
  for scanner.Scan() {
    line := scanner.Text()
    if strings.HasPrefix(line, "module ") {
      return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
    }
  }

  return "", fmt.Errorf("module name not found in go.mod")
}

// getImports parses the main file to list all imported packages
func getImports(filepath string) ([]string, error) {
  fset := token.NewFileSet()
  node, err := parser.ParseFile(fset, filepath, nil, parser.ImportsOnly)
  if err != nil {
    return nil, err
  }

  var imports []string
  for _, imp := range node.Imports {
    importPath := strings.Trim(imp.Path.Value, `"`)
    imports = append(imports, importPath)
  }
  return imports, nil
}

// filterInternalImports returns only the imports that start with the module name
func filterInternalImports(imports []string, moduleName string) []string {
  var internalImports []string
  for _, imp := range imports {
    if strings.HasPrefix(imp, moduleName) {
      internalImports = append(internalImports, imp)
    }
  }
  return internalImports
}

// findStructsInPackage locates the package files and logs all struct definitions with their fields
func findStructsInPackage(projectRoot, packagePath, moduleName string, logFile *os.File) error {
  // Convert package path to file path, removing the internal package portion based on module name
  relativePath := strings.TrimPrefix(packagePath, moduleName+"/")
  packageDir := filepath.Join(projectRoot, filepath.FromSlash(relativePath))

  err := filepath.Walk(packageDir, func(path string, info os.FileInfo, err error) error {
    if err != nil || filepath.Ext(path) != ".go" {
      return nil
    }

    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
    if err != nil {
      return err
    }

    // Log file path
    logFile.WriteString(fmt.Sprintf("File: %s\n", path))
    foundStruct := false // Track if any structs are found in the file

    // Inspect the file's AST for struct definitions
    ast.Inspect(node, func(n ast.Node) bool {
      ts, ok := n.(*ast.TypeSpec)
      if ok {
        if structType, isStruct := ts.Type.(*ast.StructType); isStruct {
          foundStruct = true
          logFile.WriteString(fmt.Sprintf("  Struct: %s\n", ts.Name.Name))
          logStructFields(structType, logFile)
        }
      }
      return true
    })

    // If no structs were found, note it in the log file
    if !foundStruct {
      logFile.WriteString("  No structs found in this file.\n")
    }

    return nil
  })

  return err
}

// logStructFields logs each field's name and type in the struct
func logStructFields(structType *ast.StructType, logFile *os.File) {
  for _, field := range structType.Fields.List {
    // Field names
    var fieldNames []string
    for _, name := range field.Names {
      fieldNames = append(fieldNames, name.Name)
    }

    // Get field type as string
    fieldType := exprToString(field.Type)

    // Log field names and types
    logFile.WriteString(fmt.Sprintf("    Field: %s, Type: %s\n", strings.Join(fieldNames, ", "), fieldType))
  }
}

// exprToString converts an expression to a readable string representation
func exprToString(expr ast.Expr) string {
  switch v := expr.(type) {
  case *ast.Ident:
    return v.Name
  case *ast.SelectorExpr:
    return fmt.Sprintf("%s.%s", exprToString(v.X), v.Sel.Name)
  case *ast.StarExpr:
    return "*" + exprToString(v.X)
  case *ast.ArrayType:
    return "[]" + exprToString(v.Elt)
  case *ast.MapType:
    return fmt.Sprintf("map[%s]%s", exprToString(v.Key), exprToString(v.Value))
  case *ast.StructType:
    return "struct{...}"
  default:
    return fmt.Sprintf("%T", expr) // Fallback to the type
  }
}
