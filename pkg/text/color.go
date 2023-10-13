package text

import "github.com/fatih/color"

// Bold is a Sprint-class function that makes the arguments bold.
var Bold = color.New(color.Bold).SprintFunc()

// BoldCyan is a Sprint-class function that makes the arguments bold and cyan.
var BoldCyan = color.New(color.Bold, color.FgCyan).SprintFunc()

// BoldRed is a Sprint-class function that makes the arguments bold and red.
var BoldRed = color.New(color.Bold, color.FgRed).SprintFunc()

// BoldYellow is a Sprint-class function that makes the arguments bold and yellow.
var BoldYellow = color.New(color.Bold, color.FgYellow).SprintFunc()

// BoldGreen is a Sprint-class function that makes the arguments bold and green.
var BoldGreen = color.New(color.Bold, color.FgGreen).SprintFunc()

// Reset is a Sprint-class function that resets the color for the arguments.
var Reset = color.New(color.Reset).SprintFunc()

// Prompt is a Sprint-class function that makes the arguments bold and grey.
var Prompt = color.New(color.Bold, color.FgHiBlack).SprintFunc()

// ColorFn is a function returned from a color.SprintFunc() call.
type ColorFn func(a ...any) string
