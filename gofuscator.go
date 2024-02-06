package main

import (
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

func main() {


	filePath := "/var/sample_go_project/main.go"

	// Parse the file
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		os.Exit(1)
	}

	// Inspect the AST to find variable names
	ast.Inspect(file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			// Check if it is a variable (not a type, function, etc.)
			if ident.Obj != nil && ident.Obj.Kind == ast.Var {
				//fmt.Println(ident.Name)
				ident.Name = "amogus_var"
			}
		}
		return true
	})

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
	//fmt.Println(importPaths)

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			// Check if it is the old function name and declared in the current file
			if node.Name.Name != "main" && node.Recv == nil && node.Name.Obj != nil && node.Name.Obj.Pos().IsValid() && fset.Position(node.Name.Obj.Pos()).Filename == filePath {
				node.Name.Name = "amogus_func"
			}

		case *ast.CallExpr:
			// Check if it is a function call
			if ident, ok := node.Fun.(*ast.Ident); ok {
				// Check if it is the old function name and declared in the current file
				if ident.Name != "main" && ident.Obj != nil && ident.Obj.Pos().IsValid() && fset.Position(ident.Obj.Pos()).Filename == filePath {
					ident.Name = "amogus_func"
				}
			}
		case *ast.BasicLit:
			// Check if it is a string literal
			if node.Kind == token.STRING && !isInArray(strings.Trim(node.Value, "\""), importPaths){
				node.Value = "\"" + "amogus_str" + "\""
			}
			if node.Kind == token.INT {
				integer_value, _ := strconv.Atoi(node.Value)
				node.Value = obfuscateInteger(float64(integer_value))
			}
			if node.Kind == token.FLOAT {
				float_value, _ := strconv.ParseFloat(node.Value, 64)
				node.Value = obfuscateInteger(float64(float_value))
			}
			
		}
		return true
	})

	os.Remove("outfile.txt")
	outputFile, err := os.Create("outfile.txt")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	err = printer.Fprint(outputFile, fset, file)
	if err != nil {
		fmt.Println("Error writing to output file:", err)
		os.Exit(1)
	}
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

func obfuscateInteger(real_value float64) string {
	var terms_array []string
	var terms_value_array []float64
	var operations_array []string
	
	terms_amount := rand.Intn(3) + 3
	debug(terms_amount)
	
	possible_operations_array := []string{"*", "/"}
	possible_modifiers_array := []string{"Sqrt", "Sin", "Cos", "Log", "Tan"}

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
			}

		} else {
			x := 0
			total := terms_value_array[0]
			for {
				if (x+2 >= terms_amount) {
					break
				}

				if (operations_array[x] == "*") {
					fmt.Println(total)
					fmt.Println("*")
					fmt.Println(terms_value_array[x+1])
					total = total * terms_value_array[x+1]
				} else if (operations_array[x] == "/") {
					fmt.Println(total)
					fmt.Println("/")
					fmt.Println(terms_value_array[x+1])
					total = total / terms_value_array[x+1]
				}

				fmt.Println(total)
				fmt.Println()

				x = x + 1
			}

			target_num := float64(real_value) - total
			target_log := math.Atan(target_num)

			fmt.Println("total:")
			fmt.Println(total)
			fmt.Println("real:")
			fmt.Println(real_value)
			fmt.Println("target:")
			fmt.Println(target_num)

			operations_array[len(operations_array)-1] = "+"
			terms_value_array[len(terms_value_array)-1] = target_num
			terms_array[len(terms_array)-1] = "math.Tan(" + strconv.FormatFloat(target_log, 'f', -1, 64) + ")"
			
		}

		i = i + 1
	}

	fmt.Println(terms_array)
	fmt.Println(terms_value_array)
	fmt.Println(operations_array)
	
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
		result_string = "(math.Round((" + result_string + ")*" +  strconv.Itoa(divider) + ")/" + strconv.Itoa(divider) + ")"
	} else {
		result_string = "(math.Round(" + result_string + "))"

	}


	fmt.Println(result_string)
	return result_string
}
