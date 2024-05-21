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
	"io/ioutil"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"unicode"
	"hash/fnv"
)
var input_file = flag.String("i", "", "the path to the input file")
var output_file = flag.String("o", "", "the path to the output file")
var seed = flag.String("seed", "", "seed to use for code generation")

var ignore_ints_bool = flag.Bool("no-ints", false, "disables int/float obfuscation")
var ignore_strings_obfuscation_bool = flag.Bool("no-strings-obf", false, "disables string obfuscation")
var ignore_strings_encryption_bool = flag.Bool("no-strings-enc", false, "disables string encryption")
var ignore_vars_bool = flag.Bool("no-vars", false, "disables variable name obfuscation")
var ignore_functions_bool = flag.Bool("no-functions", false, "disables function name/call obfuscation")
var ignore_bools_bool = flag.Bool("no-bools", false, "disables bool obfuscation")
var ignore_hexes_bool = flag.Bool("no-hexes", false, "disables hex value obfuscation")
var ignore_imports_bool = flag.Bool("no-imports", false, "disables import obfuscation")


var names_dictionary map[string]string = make(map[string]string)
var unicode_chars = []rune("аa")
var int_seed = int64(0)

var aes_key_obf string
var iv_obf string

// Workflow
//	Replace 'const' with 'var'
//	Write and read
//	Add import 'math'
//	Write and read
// 	Obfuscate variable names
// 	Get imports list
// 	Obfuscate bools
// 	Obfuscate function decls
// 	Obfuscate function calls
// 	Obfuscate strings
//	Write and read
// 	Obfuscate ints
// 	Obfuscate floats
// 	Obfuscate import aliases
//	Write and read
//	Add math operations array
//	Replace referrences to math operations
//	Replace import refferences

func main() {
	flag.Parse()
	if (len(*input_file) < 1) {
		fmt.Println("Please provide an input file with '--i'")
	}
	if (len(*output_file) < 1) {
		fmt.Println("Please provide an output file with '--i'")
	}

	string_seed := *seed
	if (len(string_seed) > 0) {
		int_seed = int64(hashString(string_seed))
	} else {
		int_seed = time.Now().UnixNano()
	}
	rand.Seed(int_seed)

	aes_key_obf_byte := make([]byte, 16)
	_, err := rand.Read(aes_key_obf_byte)
	aes_key_obf = hex.EncodeToString(aes_key_obf_byte)

	iv_obf_byte := make([]byte, 8)
	_, err = rand.Read(iv_obf_byte)
	iv_obf = hex.EncodeToString(iv_obf_byte)


	// Parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, *input_file, nil, 0)
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

	writeToOutputFile(*output_file, file, fset)
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, *output_file, nil, parser.ParseComments)

	// Adding AES functions
	if (!*ignore_strings_encryption_bool) { 
		file, fset = addAESFunctions(file, fset)
		file, fset = addImport(file, fset, "crypto/aes")
		file, fset = addImport(file, fset, "crypto/cipher")
		file, fset = addImport(file, fset, "encoding/base64")
	}

	if (!*ignore_ints_bool) { 
		file, fset = addImport(file, fset, "math") 
		file, fset = addImport(file, fset, "reflect") 
	}

	writeToOutputFile(*output_file, file, fset)
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, *output_file, nil, parser.ParseComments)

	// Find variable names
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Check if it is a variable (not a type, function, etc.)
			if ident.Obj != nil && (ident.Obj.Kind == ast.Var) && ident.Name != "_" && !*ignore_vars_bool {
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
			if ident.Obj == nil && (ident.Name == "true" || ident.Name == "false") && !*ignore_bools_bool {
				ident.Name = obfuscateBool(ident.Name)
			}
		}
		return true
	})

	var functions_list []string
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Check if it is the old function name and declared in the current file
			if node.Name.Name != "main" && node.Recv == nil && node.Name.Obj != nil && node.Name.Obj.Pos().IsValid() && fset.Position(node.Name.Obj.Pos()).Filename == *output_file && !*ignore_functions_bool {
				functions_list = append(functions_list, node.Name.Name)
				node.Name.Name = obfuscateFunctionName(node.Name.Name)
			}

		case *ast.CallExpr:
			// Check if it is a function call
			if ident, ok := node.Fun.(*ast.Ident); ok {
				// Check if it is the old function name and declared in the current file
				if ((ident.Name != "main" && ident.Obj != nil && ident.Obj.Pos().IsValid() && fset.Position(ident.Obj.Pos()).Filename == *output_file) || isInArray(ident.Name, functions_list)) && !*ignore_functions_bool {
					ident.Name = obfuscateFunctionName(ident.Name)
				}
			}
		case *ast.BasicLit:
			// Check if it is a string literal
			if node.Kind == token.STRING && !isInArray(trimFirstLastChars(node.Value), importPaths) {
				if (trimFirstLastChars(node.Value) != aes_key_obf && trimFirstLastChars(node.Value) != string(iv_obf) && !*ignore_strings_encryption_bool) {
					node.Value = string(obfuscateFunctionName("aesDecrypt") + "(" + obfuscateString(aesEncrypt(trimFirstLastChars(node.Value))) + ")")
				} else {
					node.Value = obfuscateString(trimFirstLastChars(node.Value))
				}
			}
			
		}
		return true
	})

	debug(functions_list)


	writeToOutputFile(*output_file, file, fset)
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, *output_file, nil, parser.ParseComments)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {

		case *ast.BasicLit:
			if node.Kind == token.INT {
				if (containsNonNumericChars(node.Value)) {
					node.Value = obfuscateHex(node.Value)
				} else {
					integer_value, _ := strconv.Atoi(node.Value)
					node.Value = obfuscateIntFloat(float64(integer_value))
				}
			}
			if node.Kind == token.FLOAT {
				float_value, _ := strconv.ParseFloat(node.Value, 64)
				node.Value = obfuscateIntFloat(float64(float_value))
				//debug(node.Value)
			}
			
		}
		return true
	})

	var imports_array []string
	for _, decl := range file.Decls {
        genDecl, ok := decl.(*ast.GenDecl)
        if !ok || genDecl.Tok != token.IMPORT || *ignore_imports_bool {
            continue
        }

        for _, spec := range genDecl.Specs {
            importSpec, ok := spec.(*ast.ImportSpec)
            if !ok {
                continue
            }

            importSpec.Name = &ast.Ident{Name: obfuscateFunctionName(strings.Split(strings.Trim(importSpec.Path.Value, "\""), "/")[len(strings.Split(strings.Trim(importSpec.Path.Value, "\""), "/"))-1])}
			imports_array = append(imports_array, strings.Split(strings.Trim(importSpec.Path.Value, "\""), "/")[len(strings.Split(strings.Trim(importSpec.Path.Value, "\""), "/"))-1])
        }
    }
	

	writeToOutputFile(*output_file, file, fset)
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, *output_file, nil, parser.ParseComments)

	// Adding math operation array
	if (!*ignore_ints_bool) {
		file = addGlobalVar(file, obfuscateVariableName("operations_array_obf"), "string", token.STRING, "operations_array_here")
	}

	writeToOutputFile(*output_file, file, fset)

	// Read the file contents
	content, err := ioutil.ReadFile(*output_file)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}
	
	modifiedContent := string(content)
	for i := 0; i < len(imports_array); i++ {
		modifiedContent = strings.ReplaceAll(modifiedContent, imports_array[i] + ".", obfuscateFunctionName(imports_array[i]) + ".")
	}

	operations_str := []interface{}{
		"math.Sqrt",
		"math.Sin", 
		"math.Cos", 
		"math.Log", 
		"math.Tan",
		"math.Cbrt", 
	}
	operations_str = shuffle(operations_str)

	if (!*ignore_ints_bool) {
		operations_array_str := "[]func(x float64)(float64){"
		for i := 0; i < len(operations_str); i++ {
			operations_array_str = operations_array_str + operations_str[i].(string) + ","
		}
		operations_array_str = operations_array_str + "}"
		if (!*ignore_imports_bool) {
			operations_array_str = strings.ReplaceAll(operations_array_str, `math`, obfuscateFunctionName("math"))
		}

		modifiedContent = strings.ReplaceAll(modifiedContent, `string = operations_array_here`, `= ` + operations_array_str)
		
		for i := 0; i < len(operations_str); i++ {
			real_operation := ""
			if (!*ignore_imports_bool) {
				real_operation = strings.ReplaceAll(operations_str[i].(string), `math`, obfuscateFunctionName("math"))
			} else {
				real_operation = operations_str[i].(string)
			}
			modifiedContent = strings.ReplaceAll(modifiedContent, real_operation + "(", obfuscateVariableName("operations_array_obf") + "[" + obfuscateIntFloat(float64(i)) + "](")
			modifiedContent = strings.ReplaceAll(modifiedContent, real_operation + ")", obfuscateVariableName("operations_array_obf") + "[" + obfuscateIntFloat(float64(i)) + "])")
		}

		if (!*ignore_imports_bool) {
			modifiedContent = strings.ReplaceAll(modifiedContent, `math.`, obfuscateFunctionName("math") + ".")
			modifiedContent = strings.ReplaceAll(modifiedContent, `reflect.`, obfuscateFunctionName("reflect") + ".")
		}
	}

	// Write the modified content back to the file
	err = ioutil.WriteFile(*output_file, []byte(modifiedContent), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		os.Exit(1)
	}
	


}



func obfuscateIntFloat(real_value float64) string {
	if (*ignore_ints_bool) {
		return strconv.FormatFloat(real_value, 'f', -1, 64)
	}

	var terms_array []string
	var terms_value_array []float64
	var operations_array []string
	
	terms_amount := rand.Intn(1) + 2
	
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

		// If the term is not the last in the string, just select a random modifier for it
		// TODO Add randomisation between reflect and normal
		if (i+1 < terms_amount) {
			modifier := possible_modifiers_array[rand.Intn(len(possible_modifiers_array))]
			if (modifier == "Sqrt") {
				terms_value_array[i] = math.Sqrt(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Sqrt).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Sin") {	
				terms_value_array[i] = math.Sin(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Sin).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Cos") {
				terms_value_array[i] = math.Cos(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Cos).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Log") {
				terms_value_array[i] = math.Log(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Log).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Tan") {
				terms_value_array[i] = math.Tan(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Tan).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Cbrt") {
				terms_value_array[i] = math.Cbrt(terms_value_array[i])
				terms_array[i] = "reflect.ValueOf(math.Cbrt).Call([]reflect.Value{reflect.ValueOf(" + terms_array[i] + ")})[0].Interface().(float64)"
			} else if (modifier == "Frexp") {
				exponent := float64((rand.Intn(100000000000000000) + 1) - 50000000000000000) / 10000000000000000
				terms_value_array[i] = terms_value_array[i]*math.Pow(2, float64(exponent))
				terms_array[i] = "(" +terms_array[i] + "*reflect.ValueOf(math.Pow).Call([]reflect.Value{reflect.ValueOf(float64(2)), reflect.ValueOf(float64(" + strconv.FormatFloat(exponent, 'f', -1, 64) + ") )})[0].Interface().(float64))"
			} else if (modifier == "Hypot") {
				exponent := float64((rand.Intn(100000000000000000) + 1) - 50000000000000000) / 10000000000000000
				terms_value_array[i] = math.Hypot(terms_value_array[i], exponent)
				terms_array[i] = "(reflect.ValueOf(math.Hypot).Call([]reflect.Value{reflect.ValueOf(float64(" + terms_array[i] + ")), reflect.ValueOf(float64(" + strconv.FormatFloat(exponent, 'f', -1, 64) + ") )})[0].Interface().(float64))"
			}

		} else {
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

			modifier := possible_reversible_modifiers_array[rand.Intn(len(possible_reversible_modifiers_array))]
			var exponent int
			var target_modified_num float64
			for {
				if (modifier == "Tan") {
					target_modified_num = math.Atan(target_num)
					terms_value_array[i] = target_num
					terms_array[i] = "math.Tan(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + ")"
				} else if (modifier == "Frexp") {
					target_modified_num, exponent = math.Frexp(target_num)
					terms_value_array[i] = target_num
					terms_array[i] = "(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + "*math.Pow(2, float64(" + strconv.Itoa(exponent) + ")))"
				} else if (modifier == "Cbrt") {
					target_modified_num = math.Pow(target_num, 3)
					terms_value_array[i] = target_num
					terms_array[i] = "math.Cbrt(" + strconv.FormatFloat(target_modified_num, 'f', -1, 64) + ")"
				} 	

				// Checking for infinity overflows
				if strings.Contains(strconv.FormatFloat(target_modified_num, 'f', -1, 64), "Inf") {
					modifier = "Tan"
					continue
				} else {
					break
				}
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
		result_string = "((math.Round((" + result_string + "*" +  strconv.Itoa(divider) + ")))/" + strconv.Itoa(divider) + ")"
	} else {
		result_string = "(int(math.Round(" + result_string + ")))"

	}


	//fmt.Println(result_string)
	return result_string
}

func obfuscateVariableName(real_value string) string {
	if (*ignore_vars_bool) {
		return real_value
	}

	if _, exists := names_dictionary[real_value]; !exists {
		var result []rune
		for i := 0; i < 20; i++ {
			result = append(result, unicode_chars[rand.Intn(len(unicode_chars))])
		}
		if valueExists(names_dictionary, string(result)) {
			debug("again")
			return obfuscateFunctionName(real_value)
		}
		names_dictionary[real_value] = string(result)
	}
	return names_dictionary[real_value]
}

func obfuscateFunctionName(real_value string) string {
	if (*ignore_functions_bool) {
		return real_value
	}

	if _, exists := names_dictionary[real_value]; !exists {
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
	if (*ignore_strings_obfuscation_bool) {
		return `"` + real_value + `"`
	}

	real_value, _ = strconv.Unquote(`"` + real_value + `"`)
	byte_array := []byte(real_value)
	result_string := "" 
	i := 0

	if (real_value == "") {
		return `""`
	}

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
	if (*ignore_bools_bool) {
		return real_value
	}

	int1 := rand.Intn(10000)

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

func obfuscateHex(real_value string) string {
	if (*ignore_hexes_bool) {
		return real_value
	}

	if strings.HasPrefix(real_value, "0x") {
		real_value = real_value[2:]
	}

    byteArray, err := hex.DecodeString(real_value)
    if err != nil {
        fmt.Println("Error decoding hex:", err)
		return real_value
    }

	result_string := "("
	for i := 0; i < len(byteArray); i++ {
		if (i+1 < len(byteArray)) {
			result_string = result_string + ","
		}
        byte_int := int(byteArray[i])
		result_string = result_string + "byte(" + obfuscateIntFloat(float64(byte_int)) + ")"
    }
	result_string = result_string + ")"
	
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

func addImport(file *ast.File, fset *token.FileSet, import_str string) (*ast.File, *token.FileSet) {
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

	return file, fset
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

func addFunction(file *ast.File, fset *token.FileSet, function_name string, function_content string, inputs []string, input_types []string, outputs []string, output_types []string) (*ast.File, *token.FileSet) {
	
	inputs_parsed := parseFieldList(inputs, input_types)
	outputs_parsed := parseFieldList(outputs, output_types)

	// The function body as a string.
	funcBody := `
	package main

	import (
			"fmt"
	)

	func myfunc() {
	` + function_content + `
	}
	`

	// Parse the function body string into an AST.
	body, err := parser.ParseFile(fset, "", funcBody, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing function body:", err)
	}

	// Extract the body from the parsed AST.
	var funcBodyStmts []ast.Stmt
	for _, decl := range body.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil {
				funcBodyStmts = fn.Body.List
			}
		}
	}

	// Create the AST nodes representing the new function.
	newFunc := &ast.FuncDecl{
		Name: ast.NewIdent(function_name),
		Type: &ast.FuncType{
			Params:  inputs_parsed,
			Results: outputs_parsed,
		},
		Body: &ast.BlockStmt{
			List: funcBodyStmts,
		},
	}

	// Add the new function declaration to the end of the file.
	file.Decls = append(file.Decls, newFunc)
	return file, fset
}

func addGlobalVar(file *ast.File, var_name string, var_type string, var_type_token token.Token, var_content string) *ast.File {
	globalVar := &ast.GenDecl{
        Tok: token.VAR,
        Specs: []ast.Spec{
            &ast.ValueSpec{
                Names: []*ast.Ident{
                    ast.NewIdent(var_name),
                },
                Type: ast.NewIdent(var_type), // Type of the variable
                Values: []ast.Expr{
                    &ast.BasicLit{
                        Kind:  var_type_token,
                        Value: var_content, // Initial value of the variable
                    },
                },
            },
        },
    }

    // Add the new global variable declaration to the AST
    file.Decls = append(file.Decls, globalVar)

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

func addAESFunctions(file *ast.File, fset *token.FileSet) (*ast.File, *token.FileSet) {

	file = addGlobalVar(file, "aes_key_obf", "string", token.STRING, "\"" + string(aes_key_obf) + "\"")
	file = addGlobalVar(file, "iv_obf", "string", token.STRING, "\"" + string(iv_obf) + "\"")

	funcBody := ""
	funcBody = `
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
	`
	file, fset = addFunction(file, fset, "PKCS5UnPadding", funcBody, strings.Split("src", " "), strings.Split("[]byte", " "), strings.Split("", " "), strings.Split("[]byte", " "))

	funcBody = `
	ciphertext, _ := base64.StdEncoding.DecodeString(encrypted)
	block, _ := aes.NewCipher([]byte(aes_key_obf))
	mode := cipher.NewCBCDecrypter(block, []byte(iv_obf))
	mode.CryptBlocks(ciphertext, ciphertext)
	ciphertext = PKCS5UnPadding(ciphertext)
	return string(ciphertext)
	`
	file, fset = addFunction(file, fset, "aesDecrypt", funcBody, strings.Split("encrypted", " "), strings.Split("string", " "), strings.Split("", " "), strings.Split("string", " "))
	


	return file, fset
}

func aesEncrypt(plaintext string) (string) {
	if (plaintext == "") {
		return `""`
	}

	plaintext, _ = strconv.Unquote(`"` + plaintext + `"`)

	var plainTextBlock []byte
	length := len(plaintext)

	if length%16 != 0 {
		extendBlock := 16 - (length % 16)
		plainTextBlock = make([]byte, length+extendBlock)
		copy(plainTextBlock[length:], bytes.Repeat([]byte{uint8(extendBlock)}, extendBlock))
	} else {
		plainTextBlock = make([]byte, length)
	}

	copy(plainTextBlock, plaintext)
	block, err := aes.NewCipher([]byte(aes_key_obf))

	if err != nil {
		debug(err)
		return ""
	}

	ciphertext := make([]byte, len(plainTextBlock))
	mode := cipher.NewCBCEncrypter(block, []byte(iv_obf))
	mode.CryptBlocks(ciphertext, plainTextBlock)

	str := base64.StdEncoding.EncodeToString(ciphertext)

	return str
}

func shuffle(arr []interface{}) []interface{} {
    // Make a copy of the slice to avoid modifying the original
    shuffled := make([]interface{}, len(arr))
    copy(shuffled, arr)

    // Shuffle the elements
    for i := len(shuffled) - 1; i > 0; i-- {
        j := rand.Intn(i + 1)
        shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
    }

    return shuffled
}

func trimFirstLastChars(input string) string {
    if len(input) <= 2 {
        return ""
    }

    trimmedString := input[1 : len(input)-1]

    return trimmedString
}

func containsNonNumericChars(str string) bool {
    for _, char := range str {
        if !unicode.IsDigit(char) {
            return true
        }
    }
    return false
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}