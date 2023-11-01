package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "/exit" {
			fmt.Println("Bye!")
			break
		}
		if line == "/help" {
			fmt.Println("The program calculates the sum of numbers")
			continue
		}
		if line == "" {
			continue
		}
		if strings.TrimSpace(line)[0] == '/' {
			fmt.Println("Unknown command")
			continue
		}
		if IsAssignmentInput(line) {
			err := ReadAssignmentInput(line)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
		} else {
			parsedInput, err := parseVariablesInString(strings.TrimSpace(line))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			output, err := ReadCalculationInput(parsedInput)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			fmt.Println(output)
		}
	}
}

var VariableDictionary = map[string]float32{}

// Assignment

func ReadAssignmentInput(line string) error {
	words := strings.Split(strings.TrimSpace(line), "=")
	varName := strings.TrimSpace(words[0])
	isValidVarName := IsWord(varName)
	if !isValidVarName {
		return errors.New("Invalid identifier")
	}
	operation := strings.TrimSpace(words[1])
	replacedOperation, err := parseVariablesInString(operation)
	if err != nil {
		return err
	}
	result, err := ReadCalculationInput(replacedOperation)
	if err != nil {
		return err
	}
	VariableDictionary[varName] = result
	return nil
}

func parseVariablesInString(s string) (string, error) {
	splitString := strings.Fields(s)
	newString := s
	for _, word := range splitString {
		_, numberError := strconv.ParseFloat(word, 32)
		isWord := IsWord(word)
		if numberError == nil {
			continue
		}
		if isWord {
			if value, ok := VariableDictionary[word]; ok {
				newString = strings.Replace(
					newString,
					word,
					strconv.FormatFloat(float64(value), 'f', 3, 32),
					1)
			} else {
				return "", errors.New("Unknown variable: " + word)
			}
		} else if !isSign(word) && !isParenthesis(word) {
			return "", errors.New("Invalid Identifier: " + word)
		}
	}
	return newString, nil
}

func isParenthesis(word string) bool {
	return strings.HasSuffix(word, "(") || strings.HasSuffix(word, ")") ||
		strings.HasPrefix(word, "(") || strings.HasPrefix(word, ")")
}

func IsAssignmentInput(line string) bool {
	amountOfEquals := strings.Count(line, "=")
	return amountOfEquals == 1
}

func IsWord(word string) bool {
	for _, letter := range word {
		if !unicode.In(letter, unicode.Latin) {
			return false
		}
	}
	return true
}

// Calculation

func ReadCalculationInput(line string) (float32, error) {
	re := regexp.MustCompile(`([\(\)])`)
	spacedParenthesis := re.ReplaceAllString(line, " $1 ")
	if !isValidInput(spacedParenthesis) {
		return 0, errors.New("Invalid expression")
	}
	posfix, err := infixToPostfix(spacedParenthesis)
	if err != nil {
		return -1, err
	}
	output, err := readPosfix(posfix)
	if err != nil {
		return -1, err
	}
	result, err := strconv.ParseFloat(output, 32)
	if err != nil {
		return -1, err
	}
	return float32(result), nil
}

func isValidInput(line string) bool {
	wordSlice := strings.Split(line, " ")
	var lastValue *string
	for i, word := range wordSlice {
		if word == " " || word == "" {
			continue
		}
		if lastValue == nil {
			lastValue = &wordSlice[i]
			continue
		}
		if *lastValue == "(" || word == "(" || *lastValue == ")" || word == ")" {
			lastValue = &wordSlice[i]
			continue
		}
		_, errWord := strconv.ParseFloat(word, 32)
		_, errLast := strconv.ParseFloat(*lastValue, 32)
		isWordNumber := errWord == nil
		isLastNumber := errLast == nil
		isWordSign := isSign(word)
		isLastSign := isSign(*lastValue)
		areBothNumbers := isWordNumber && isLastNumber
		areBothSigns := isWordSign && isLastSign

		if areBothNumbers || areBothSigns {
			return false
		} else if !isWordNumber && !isWordSign {
			return false
		}
		lastValue = &wordSlice[i]
	}
	return true
}

func readSignRuneList(signList []rune) (string, error) {
	if len(signList) == 1 {
		return string(signList[0]), nil
	}
	sign, _ := signAddition(signList[0])
	resultSign, err := sign(signList[1])
	if err != nil {
		return "", err
	}
	for _, currentSign := range signList[2:] {
		s, _ := signAddition([]rune(resultSign)[0])
		resultSign, err = s(currentSign)
		if err != nil {
			return "", err
		}
	}
	return resultSign, nil
}

func signAddition(first rune) (signAdder func(second rune) (s string, err error), neutralRune rune) {
	return func(second rune) (string, error) {
		if (first == '*' || second == '/') || (first == '/' || second == '*') {
			return "", errors.New("Invalid expression")
		} else if first == '-' && second == '-' {
			return "+", nil
		} else if first == '-' || second == '-' {
			return "-", nil
		} else {
			return "+", nil
		}
	}, '+'
}

func getOperationOfSign(sign string) (func(a float32) func(b float32) float32, float32) {
	switch sign {
	case "+":
		return addition, 0
	case "-":
		return substraction, 0
	case "*":
		return multiplication, 0
	case "/":
		return division, 0
	default:
		return nil, 0
	}
}

func addition(first float32) func(float32) float32 {
	return func(second float32) float32 {
		return first + second
	}
}

func substraction(first float32) func(float32) float32 {
	return func(second float32) float32 {
		return first - second
	}
}

func multiplication(first float32) func(float32) float32 {
	return func(second float32) float32 {
		return first * second
	}
}

func division(first float32) func(float32) float32 {
	return func(second float32) float32 {
		return first / second
	}
}

func isSign(s string) bool {
	var isSign bool
	for _, r := range []rune(s) {
		isAddition := r == '+'
		isSubstraction := r == '-'
		isDivision := r == '/'
		isMultiplication := r == '*'
		isParenthesis := r == '(' || r == ')'
		if isAddition || isSubstraction || isDivision || isMultiplication || isParenthesis {
			isSign = true
		} else {
			isSign = false
			break
		}
	}
	return isSign
}

func infixToPostfix(operation string) (string, error) {
	operands := strings.Fields(operation)
	operandStack := Stack{storage: make([]string, 0)}
	result := ""
	for _, char := range operands {
		// if scanned character is operand, add it to output string
		if !isSign(char) {
			result = result + " " + char
		} else if char == "(" {
			operandStack.Push(char)
		} else if char == ")" {
			top, err := operandStack.Peek()
			for top, err = operandStack.Peek(); top != "(" && err == nil; top, err = operandStack.Peek() {
				result += " " + top
				_, _ = operandStack.Pop()
			}
			if err != nil {
				return "", errors.New("Invalid expression")
			}
			_, _ = operandStack.Pop()
		} else {
			for top, _ := operandStack.Peek(); !operandStack.isEmpty() && getOperatorPrecedence(char) <= getOperatorPrecedence(top); top, _ = operandStack.Peek() {
				result = result + " " + top
				_, _ = operandStack.Pop()
			}
			operandStack.Push(char)
		}
	}

	for _, operand := range operandStack.storage {
		if operand == "(" || operand == ")" {
			return "", errors.New("Invalid expression")
		}
	}

	for !operandStack.isEmpty() {
		top, _ := operandStack.Pop()
		result = result + " " + top
	}
	return result, nil
}

func readPosfix(posfixOp string) (string, error) {
	posFixStack := Stack{storage: make([]string, 0)}
	fields := strings.Fields(posfixOp)

	for _, word := range fields {
		if isSign(word) {
			singleSign, err := readSignRuneList([]rune(word))
			if err != nil {
				return "", err
			}
			secondOperand, err := posFixStack.Pop()
			if err != nil {
				return "", errors.New("invalid postfix expression")
			}
			firstOperand, err := posFixStack.Pop()
			if err != nil {
				return "", errors.New("invalid postfix expression")
			}

			operation, _ := getOperationOfSign(singleSign)

			firstFloat, err := strconv.ParseFloat(firstOperand, 64)
			if err != nil {
				return "", errors.New("invalid operand")
			}

			secondFloat, err := strconv.ParseFloat(secondOperand, 64)
			if err != nil {
				return "", errors.New("invalid operand")
			}

			result := operation(float32(firstFloat))(float32(secondFloat))
			stringResult := strconv.FormatFloat(float64(result), 'f', -1, 64)
			posFixStack.Push(stringResult)
		} else {
			posFixStack.Push(word)
		}
	}

	finalResult, err := posFixStack.Pop()
	if err != nil {
		return "", errors.New("invalid postfix expression")
	}
	return finalResult, nil
}
func getOperatorPrecedence(operator string) int {
	switch operator {
	case "*":
		return 1
	case "/":
		return 1
	case "+":
		return 0
	case "-":
		return 0
	default:
		return -1
	}
}

type Stack struct {
	storage []string
}

func (s *Stack) Push(value string) {
	s.storage = append(s.storage, value)
}

func (s *Stack) Pop() (string, error) {
	last := len(s.storage) - 1
	if last <= -1 { // check the size
		return "", errors.New("Stack is empty") // and return error
	}

	value := s.storage[last]     // save the value
	s.storage = s.storage[:last] // remove the last element

	return value, nil // return saved value and nil error
}

func (s *Stack) Peek() (string, error) {
	last := len(s.storage) - 1
	if last <= -1 { // check the size
		return "", errors.New("Stack is empty") // and return error
	}
	return s.storage[last], nil
}

func (s *Stack) HasParenthesis() bool {
	joinedString := strings.Join(s.storage, "")
	return strings.ContainsAny(joinedString, "()")
}

func (s *Stack) isEmpty() bool {
	_, err := s.Peek()
	if len(s.storage) > 0 && err == nil {
		return false
	} else {
		return true
	}
}
