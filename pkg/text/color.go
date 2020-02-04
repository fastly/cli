package text

import "github.com/fatih/color"

// Bold is a Sprint-class function that makes the arguments bold.
var Bold = color.New(color.Bold).SprintFunc()

// BoldRed is a Sprint-class function that makes the arguments bold and red.
var BoldRed = color.New(color.Bold, color.FgRed).SprintFunc()

// BoldYellow is a Sprint-class function that makes the arguments bold and yellow.
var BoldYellow = color.New(color.Bold, color.FgYellow).SprintFunc()

// BoldGreen is a Sprint-class function that makes the arguments bold and green.
var BoldGreen = color.New(color.Bold, color.FgGreen).SprintFunc()

// Reset is a Sprint-class function that resets the color for the arguments.
var Reset = color.New(color.Reset).SprintFunc()
