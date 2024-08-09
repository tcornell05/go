package chat

import "fmt"

// ANSI color codes map
var ColorMap = map[string]string{
	// Regular Colors
	"black":  "\033[0;30m",
	"red":    "\033[0;31m",
	"green":  "\033[0;32m",
	"yellow": "\033[0;33m",
	"blue":   "\033[0;34m",
	"purple": "\033[0;35m",
	"cyan":   "\033[0;36m",
	"white":  "\033[0;37m",

	// Bold
	"bold_black":  "\033[1;30m",
	"bold_red":    "\033[1;31m",
	"bold_green":  "\033[1;32m",
	"bold_yellow": "\033[1;33m",
	"bold_blue":   "\033[1;34m",
	"bold_purple": "\033[1;35m",
	"bold_cyan":   "\033[1;36m",
	"bold_white":  "\033[1;37m",

	// Underline
	"underline_black":  "\033[4;30m",
	"underline_red":    "\033[4;31m",
	"underline_green":  "\033[4;32m",
	"underline_yellow": "\033[4;33m",
	"underline_blue":   "\033[4;34m",
	"underline_purple": "\033[4;35m",
	"underline_cyan":   "\033[4;36m",
	"underline_white":  "\033[4;37m",

	// Background
	"bg_black":  "\033[40m",
	"bg_red":    "\033[41m",
	"bg_green":  "\033[42m",
	"bg_yellow": "\033[43m",
	"bg_blue":   "\033[44m",
	"bg_purple": "\033[45m",
	"bg_cyan":   "\033[46m",
	"bg_white":  "\033[47m",

	// High Intensity
	"hi_black":  "\033[0;90m",
	"hi_red":    "\033[0;91m",
	"hi_green":  "\033[0;92m",
	"hi_yellow": "\033[0;93m",
	"hi_blue":   "\033[0;94m",
	"hi_purple": "\033[0;95m",
	"hi_cyan":   "\033[0;96m",
	"hi_white":  "\033[0;97m",

	// Bold High Intensity
	"bold_hi_black":  "\033[1;90m",
	"bold_hi_red":    "\033[1;91m",
	"bold_hi_green":  "\033[1;92m",
	"bold_hi_yellow": "\033[1;93m",
	"bold_hi_blue":   "\033[1;94m",
	"bold_hi_purple": "\033[1;95m",
	"bold_hi_cyan":   "\033[1;96m",
	"bold_hi_white":  "\033[1;97m",

	// High Intensity backgrounds
	"hi_bg_black":  "\033[0;100m",
	"hi_bg_red":    "\033[0;101m",
	"hi_bg_green":  "\033[0;102m",
	"hi_bg_yellow": "\033[0;103m",
	"hi_bg_blue":   "\033[0;104m",
	"hi_bg_purple": "\033[0;105m",
	"hi_bg_cyan":   "\033[0;106m",
	"hi_bg_white":  "\033[0;107m",

	// Reset
	"reset": "\033[0m",
}

func Colorize(text, color string) string {
	return fmt.Sprintf("%s%s%s", ColorMap[color], text, ColorMap["reset"])
}

func ColorizeTest() {
	fmt.Println(Colorize("This is a black message", "black"))
	fmt.Println(Colorize("This is a red message", "red"))
	fmt.Println(Colorize("This is a green message", "green"))
	fmt.Println(Colorize("This is a yellow message", "yellow"))
	fmt.Println(Colorize("This is a blue message", "blue"))
	fmt.Println(Colorize("This is a purple message", "purple"))
	fmt.Println(Colorize("This is a cyan message", "cyan"))
	fmt.Println(Colorize("This is a white message", "white"))

	fmt.Println(Colorize("This is a bold black message", "bold_black"))
	fmt.Println(Colorize("This is a bold red message", "bold_red"))
	fmt.Println(Colorize("This is a bold green message", "bold_green"))
	fmt.Println(Colorize("This is a bold yellow message", "bold_yellow"))
	fmt.Println(Colorize("This is a bold blue message", "bold_blue"))
	fmt.Println(Colorize("This is a bold purple message", "bold_purple"))
	fmt.Println(Colorize("This is a bold cyan message", "bold_cyan"))
	fmt.Println(Colorize("This is a bold white message", "bold_white"))

	fmt.Println(Colorize("This is an underlined black message", "underline_black"))
	fmt.Println(Colorize("This is an underlined red message", "underline_red"))
	fmt.Println(Colorize("This is an underlined green message", "underline_green"))
	fmt.Println(Colorize("This is an underlined yellow message", "underline_yellow"))
	fmt.Println(Colorize("This is an underlined blue message", "underline_blue"))
	fmt.Println(Colorize("This is an underlined purple message", "underline_purple"))
	fmt.Println(Colorize("This is an underlined cyan message", "underline_cyan"))
	fmt.Println(Colorize("This is an underlined white message", "underline_white"))

	fmt.Println(Colorize("This is a background black message", "bg_black"))
	fmt.Println(Colorize("This is a background red message", "bg_red"))
	fmt.Println(Colorize("This is a background green message", "bg_green"))
	fmt.Println(Colorize("This is a background yellow message", "bg_yellow"))
	fmt.Println(Colorize("This is a background blue message", "bg_blue"))
	fmt.Println(Colorize("This is a background purple message", "bg_purple"))
	fmt.Println(Colorize("This is a background cyan message", "bg_cyan"))
	fmt.Println(Colorize("This is a background white message", "bg_white"))

	fmt.Println(Colorize("This is a high intensity black message", "hi_black"))
	fmt.Println(Colorize("This is a high intensity red message", "hi_red"))
	fmt.Println(Colorize("This is a high intensity green message", "hi_green"))
	fmt.Println(Colorize("This is a high intensity yellow message", "hi_yellow"))
	fmt.Println(Colorize("This is a high intensity blue message", "hi_blue"))
	fmt.Println(Colorize("This is a high intensity purple message", "hi_purple"))
	fmt.Println(Colorize("This is a high intensity cyan message", "hi_cyan"))
	fmt.Println(Colorize("This is a high intensity white message", "hi_white"))

	fmt.Println(Colorize("This is a bold high intensity black message", "bold_hi_black"))
	fmt.Println(Colorize("This is a bold high intensity red message", "bold_hi_red"))
	fmt.Println(Colorize("This is a bold high intensity green message", "bold_hi_green"))
	fmt.Println(Colorize("This is a bold high intensity yellow message", "bold_hi_yellow"))
	fmt.Println(Colorize("This is a bold high intensity blue message", "bold_hi_blue"))
	fmt.Println(Colorize("This is a bold high intensity purple message", "bold_hi_purple"))
	fmt.Println(Colorize("This is a bold high intensity cyan message", "bold_hi_cyan"))
	fmt.Println(Colorize("This is a bold high intensity white message", "bold_hi_white"))

	fmt.Println(Colorize("This is a high intensity background black message", "hi_bg_black"))
	fmt.Println(Colorize("This is a high intensity background red message", "hi_bg_red"))
	fmt.Println(Colorize("This is a high intensity background green message", "hi_bg_green"))
	fmt.Println(Colorize("This is a high intensity background yellow message", "hi_bg_yellow"))
	fmt.Println(Colorize("This is a high intensity background blue message", "hi_bg_blue"))
	fmt.Println(Colorize("This is a high intensity background purple message", "hi_bg_purple"))
	fmt.Println(Colorize("This is a high intensity background cyan message", "hi_bg_cyan"))
	fmt.Println(Colorize("This is a high intensity background white message", "hi_bg_white"))
}
