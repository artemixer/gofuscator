package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/printer"
	"os"
	"strings"
	"strconv"
	"math/rand"
	"math"
	"time"
)
var input_file = flag.String("i", "", "the path to the input file")
var output_file = flag.String("o", "", "the path to the output file")

var names_dictionary map[string]string = make(map[string]string)
var unicode_chars = []rune("аa")

func main() {
	flag.Parse()
	if (len(*input_file) < 1) {
		fmt.Println("Please provide an input file with '--i'")
	}
	if (len(*output_file) < 1) {
		fmt.Println("Please provide an output file with '--i'")
	}

	// Parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, *input_file, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		os.Exit(1)
	}

	// Replace all consts with var
	ast.Inspect(file, func(n ast.Node) bool {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			genDecl.Tok = token.VAR
		}
		return true
	})

	/*
	file = addFunction(file, "test_func", 
	`fmt.Println("Line 1 from the new function!")
	fmt.Println("Line 2 from the new function!")
	fmt.Println("Line 3 from the new function!")`, strings.Split("input_str", " "), strings.Split("string", " "), strings.Split("output_int", " "), strings.Split("int", " "))

	file = addGlobalVar(file, "test_key", "\"3n84f38yedj\"")
	*/

	// Add "math" import if it doesn't exist
	file = addImport(file, fset, "math")

	// Find variable names
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Check if it is a variable (not a type, function, etc.)
			if ident.Obj != nil && (ident.Obj.Kind == ast.Var) && ident.Name != "_" {
				ident.Name = obfuscateVariableName(ident.Name)
			}
		}
		return true
	})

	// Write import paths to array
	var importPaths []string
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			// Found an import declaration
			for _, spec := range genDecl.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					// Extract the import path
					importPath := removeChar(importSpec.Path.Value, "\""[0])
					importPaths = append(importPaths, importPath)
				}
			}
		}
	}

	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if ident.Obj == nil && (ident.Name == "true" || ident.Name == "false") {
				ident.Name = obfuscateBool(ident.Name)
			}
		}
		return true
	})

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Check if it is the old function name and declared in the current file
			if node.Name.Name != "main" && node.Recv == nil && node.Name.Obj != nil && node.Name.Obj.Pos().IsValid() && fset.Position(node.Name.Obj.Pos()).Filename == *output_file {
				node.Name.Name = obfuscateFunctionName(node.Name.Name)
			}

		case *ast.CallExpr:
			// Check if it is a function call
			if ident, ok := node.Fun.(*ast.Ident); ok {
				// Check if it is the old function name and declared in the current file
				if ident.Name != "main" && ident.Obj != nil && ident.Obj.Pos().IsValid() && fset.Position(ident.Obj.Pos()).Filename == *output_file {
					ident.Name = obfuscateFunctionName(ident.Name)
				}
			}
		case *ast.BasicLit:
			// Check if it is a string literal
			if node.Kind == token.STRING && !isInArray(strings.Trim(node.Value, "\""), importPaths){
				node.Value = obfuscateString(strings.Trim(node.Value, "\""))
			}
			
		}
		return true
	})

	writeToOutputFile(*output_file, file, fset)
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, *output_file, nil, parser.ParseComments)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {

		case *ast.BasicLit:
			if node.Kind == token.INT {
				integer_value, _ := strconv.Atoi(node.Value)
				node.Value = obfuscateIntFloat(float64(integer_value))
			}
			if node.Kind == token.FLOAT {
				float_value, _ := strconv.ParseFloat(node.Value, 64)
				node.Value = obfuscateIntFloat(float64(float_value))
			}
			
		}
		return true
	})


	writeToOutputFile(*output_file, file, fset)

}



func obfuscateIntFloat(real_value float64) string {
	var terms_array []string
	var terms_value_array []float64
	var operations_array []string
	
	terms_amount := rand.Intn(3) + 3
	
	possible_operations_array := []string{"*", "/"}
	possible_modifiers_array := []string{"Sqrt", "Sin", "Cos", "Log", "Tan", "Frexp", "Hypot", "Cbrt"}
	possible_reversible_modifiers_array := []string{"Tan", "Frexp", "Cbrt"}

	// Generate random numbers and operations
	i := 0
	for {
		i = i + 1
		if (i > terms_amount) {
			break
		}

		rand.Seed(time.Now().UnixNano())

		var value float64 = float64(rand.Intn(100000000000000000) + 1) / 10000000000000000
		terms_value_array = append(terms_value_array, float64(value))
		terms_array = append(terms_array, strconv.FormatFloat(value, 'f', -1, 64))
		operations_array = append(operations_array, possible_operations_array[rand.Intn(len(possible_operations_array))])

	}
	operations_array = operations_array[:len(operations_array)-1]

	i = 0
	for {
		if (i >= terms_amount) {
			break
		}

		rand.Seed(time.Now().UnixNano())

		// If the term is not the last in the string, just select a random modifier for it
		if (i+1 < terms_amount) {
			modifier := possible_modifiers_array[rand.Intn(len(possible_modifiers_array))]
			if (modifier == "Sqrt") {
				terms_value_array[i] = math.Sqrt(terms_value_array[i])
				terms_array[i] = "math.Sqrt(" + terms_array[i] + ")"
			} else if (modifier == "Sin") {	
				terms_value_array[i] = math.Sin(terms_value_array[i])
				terms_array[i] = "math.Sin(" + terms_array[i] + ")"
			} else if (modifier == "Cos") {
				terms_value_array[i] = math.Cos(terms_value_array[i])
				terms_array[i] = "math.Cos(" + terms_array[i] + ")"
			} else if (modifier == "Log") {
				terms_value_array[i] = math.Log(terms_value_array[i])
				terms_array[i] = "math.Log(" + terms_array[i] + ")"
			} else if (modifier == "Tan") {
				terms_value_array[i] = math.Tan(terms_value_array[i])
				terms_array[i] = "math.Tan(" + terms_array[i] + ")"
			} else if (modifier == "Frexp") {
				exponent := float64((rand.Intn(100000000000000000) + 1) - 50000000000000000) / 10000000000000000
				terms_value_array[i] = terms_value_array[i]*math.Pow(2, float64(exponent))
				terms_array[i] = "(" +terms_array[i] + "*math.Pow(2, float64(" + strconv.FormatFloat(exponent, 'f', -1, 64) + ")))"
			} else if (modifier == "Hypot") {
				exponent := float64((rand.Intn(100000000000000000) + 1) - 50000000000000000) / 10000000000000000
				terms_value_array[i] = math.Hypot(terms_value_array[i], exponent)
				terms_array[i] = "math.Hypot(" + terms_array[i] + ", " + strconv.FormatFloat(exponent, 'f', -1, 64) + ")"
			} else if (modifier == "Cbrt") {
				terms_value_array[i] = math.Cbrt(terms_value_array[i])
				terms_array[i] = "math.Cbrt(" + terms_array[i] + ")"
			}

		} else {
			// If the term is last, find the value needed to bring the output to the real value
			x := 0
			total := terms_value_array[0]
			for {
				if (x+2 >= terms_amount) {
					break
				}

				if (operations_array[x] == "*") {
					total = total * terms_value_array[x+1]
				} else if (operations_array[x] == "/") {
					total = total / terms_value_array[x+1]
				}

				x = x + 1
			}

			target_num := 0.0
			if (operations_array[len(operations_array)-1] == "*") {
				target_num = real_value / total
			} else if (operations_array[len(operations_array)-1] == "/") {
				target_num = total / real_value
			}

			fmt.Println("total:")
			fmt.Println(total)
			fmt.Println("operator:")
			fmt.Println(operations_array[len(operations_array)-1])
			fmt.Println("real:")
			fmt.Println(real_value)
			fmt.Println("target:")
			fmt.Println(target_num)
			fmt.Println()
			fmt.Println()

			//target_num := float64(real_value) - total
			//target_log := math.Atan(target_num)

			modifier := possible_reversible_modifiers_array[rand.Intn(len(possible_reversible_modifiers_array))]
			if (modifier == "Tan") {
				target_modified_num := math.Atan(target_num)
				terms_value_array[i] = target_num
				terms_array[i] = "math.Tan(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + ")"
			} else if (modifier == "Frexp") {
				target_modified_num, exponent := math.Frexp(target_num)
				terms_value_array[i] = target_num
				terms_array[i] = "(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + "*math.Pow(2, float64(" + strconv.Itoa(exponent) + ")))"
			} else if (modifier == "Cbrt") {
				target_modified_num := math.Pow(target_num, 3)
				terms_value_array[i] = target_num
				terms_array[i] = "math.Cbrt(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + ")"
			} 

			//operations_array[len(operations_array)-1] = "+"
			//terms_value_array[len(terms_value_array)-1] = target_num
			//terms_array[len(terms_array)-1] = "math.Tan(" + strconv.FormatFloat(target_log, 'f', -1, 64) + ")"
			
		}

		i = i + 1
	}

	/*
	fmt.Println(terms_array)
	fmt.Println(terms_value_array)
	fmt.Println(operations_array)
	*/
	

	// Append the arrays to form the output string
	result_string := ""
	x := 0
	for {
		if (x >= terms_amount) {
			break
		}

		result_string = result_string + terms_array[x]
		if (x + 1 < terms_amount) {
			result_string = result_string + operations_array[x]
		}

		x = x + 1
	}

	// Find the decimal places of the input and round the output to those decimal places
	str := strconv.FormatFloat(real_value, 'f', -1, 64)
    parts := strings.Split(str, ".")
	decimal_places := 0
    if len(parts) == 2 {
    	decimal_places = len(parts[1])
    } else {
        decimal_places = 0
    }
	divider := int(math.Pow(float64(10), float64(decimal_places)))

	if (divider > 1) {
		result_string = "(int(math.Round((" + result_string + "))*" +  strconv.Itoa(divider) + ")/" + strconv.Itoa(divider) + ")"
	} else {
		result_string = "(int(math.Round(" + result_string + ")))"

	}


	//fmt.Println(result_string)
	return result_string
}

func obfuscateVariableName(real_value string) string {
	if _, exists := names_dictionary[real_value]; !exists {
		rand.Seed(time.Now().UnixNano())
		var result []rune
		for i := 0; i < 20; i++ {
			result = append(result, unicode_chars[rand.Intn(len(unicode_chars))])
		}
		if valueExists(names_dictionary, string(result)) {
			return obfuscateFunctionName(real_value)
		}
		names_dictionary[real_value] = string(result)
	}
	return names_dictionary[real_value]
}

func obfuscateFunctionName(real_value string) string {
	if _, exists := names_dictionary[real_value]; !exists {
		rand.Seed(time.Now().UnixNano())
		var result []rune
		for i := 0; i < 20; i++ {
			result = append(result, unicode_chars[rand.Intn(len(unicode_chars))])
		}
		if valueExists(names_dictionary, string(result)) {
			return obfuscateFunctionName(real_value)
		}
		names_dictionary[real_value] = string(result)
	}
	return names_dictionary[real_value]
}

func obfuscateString(real_value string) string {
	byte_array := []byte(real_value)
	result_string := "" 
	i := 0

	if (len(byte_array) != len(real_value)) {
		// TODO Add support for multiple-byte encodings
		return real_value
	}

	for _, b := range byte_array {
        // Convert byte to int
        int_value := int(b)

		result_string = result_string + "string(" + strconv.Itoa(int_value) + ")"
		
		if (i != len(byte_array)-1) {
			result_string = result_string + "+"
		}
		
		i = i + 1
	}

	result_string = "(" + result_string + ")"
	return result_string
}

func obfuscateBool(real_value string) string {
	rand.Seed(time.Now().UnixNano())
	int1 := rand.Intn(10000)

	rand.Seed(time.Now().UnixNano())
	int2 := rand.Intn(10000)
	operator := ""

	if (real_value == "true") {
		if (int1 > int2) {
			operator = ">"
		} else {
			operator = "<="
		}
	} else {
		if (int1 > int2) {
			operator = "<"
		} else {
			operator = ">="
		}
	}
	
	result_string := strconv.Itoa(int1) + operator + strconv.Itoa(int2)
	result_string = "(" + result_string + ")"
	return result_string
}

func isInArray(target string, arr []string) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

func removeChar(input string, charToRemove byte) string {
	result := ""
	for i := 0; i < len(input); i++ {
		if input[i] != charToRemove {
			result += string(input[i])
		}
	}
	return result
}

func debug(str interface{}) {
	fmt.Printf("[?] ")
	fmt.Println(str)
}

func hasImport(file *ast.File, importPath string) bool {
	for _, imp := range file.Imports {
		if imp.Path != nil && imp.Path.Value == fmt.Sprintf(`"%s"`, importPath) {
			return true
		}
	}
	return false
}

func addImport(file *ast.File, fset *token.FileSet, import_str string) (*ast.File) {
	// Add the imports
	for i := 0; i < len(file.Decls); i++ {
		d := file.Decls[i]

		switch d.(type) {
		case *ast.FuncDecl:
			// No action
		case *ast.GenDecl:
			dd := d.(*ast.GenDecl)

			// IMPORT Declarations
			if dd.Tok == token.IMPORT {
				// Add the new import
				iSpec := &ast.ImportSpec{Path: &ast.BasicLit{Value: strconv.Quote(import_str)}}
				dd.Specs = append(dd.Specs, iSpec)
			}
		}
	}

	// Sort the imports
	ast.SortImports(fset, file)

	return file
}

func valueExists(dict map[string]string, value string) bool {
    for _, v := range dict {
        if v == value {
            return true
        }
    }
    return false
}

func writeToOutputFile(file string, contents *ast.File, fset *token.FileSet) {
	os.Remove(file)
	outputFile, err := os.Create(file)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	err = printer.Fprint(outputFile, fset, contents)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		os.Exit(1)
	}
}

func addFunction(file *ast.File, function_name string, function_content string, inputs []string, input_types []string, outputs []string, output_types []string) *ast.File {
	
	inputs_parsed := parseFieldList(inputs, input_types)
	outputs_parsed := parseFieldList(outputs, output_types)

	newFunc := &ast.FuncDecl{
		Name: ast.NewIdent(function_name),
		Type: &ast.FuncType{
			Params:  inputs_parsed,
			Results: outputs_parsed,
		},
		Body: &ast.BlockStmt{},
	}

	// Insert lines of code into the body of the new function
	lines := strings.Split(function_content, "\n")
	for _, line := range lines {
		stmt, err := parser.ParseExpr(line)
		if err != nil {
			fmt.Println("Error parsing line:", err)
			continue
		}
		newFunc.Body.List = append(newFunc.Body.List, &ast.ExprStmt{X: stmt})
	}

	// Add the new function to the file's declarations
	file.Decls = append(file.Decls, newFunc)
	return file
}

func addGlobalVar(file *ast.File, var_name string, var_content string) *ast.File {
	newVar := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{
			&ast.ValueSpec{
				Names:  []*ast.Ident{ast.NewIdent(var_name)},
				Values: []ast.Expr{ast.NewIdent(var_content)},
			},
		},
	}

	// Insert the new variable declaration at the beginning of the file's declarations
	file.Decls = append([]ast.Decl{newVar}, file.Decls...)

	return file
}

func parseFieldList(fields []string, field_types []string) *ast.FieldList {
	fields_parsed := []*ast.Field{}
	i := 0
	for _, name := range fields {
		fields_parsed = append(fields_parsed, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  ast.NewIdent(field_types[i]),
		})
		i = i + 1
	}
	return &ast.FieldList{List: fields_parsed}
}

