package ui

import "fmt"

const (
	Bold    = "\033[1m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	Reset   = "\033[0m"
)

func PrintBanner() {
	print(`
    _         _                        _         _     _  __      
   / \  _   _| |_ ___  _ __ ___   __ _| |_ ___  | |   (_)/ _| ___ 
  / _ \| | | | __/ _ \| '_ ' _ \ / _' | __/ _ \ | |   | | |_ / _ \
 / ___ \ |_| | || (_) | | | | | | (_| | ||  __/ | |___| |  _|  __/
/_/   \_\__,_|\__\___/|_| |_| |_|\__,_|\__\___| |_____|_|_|  \___|
                                                                    
`)
}

func PrintWelcome() {
	Printf("Welcome to %s%sAutomate Life%s, your gateway to automation\n\n", Bold, Green, Reset)
	Printf("Run %s%sautomateLife%s then one of the following commands to start:\n\n", Bold, Blue, Reset)
	println("init: creates a config file in your current directory")
	println("start: clones repository and optionally runs tests or builds")
	println("verify: verifies that the current directory has the necessary parameters for automation")
	println("test: runs the tests in your project")
	println("build: builds the project and creates artifacts (IPA for iOS, binaries for other languages)")
}

func Printf(format string, args ...interface{}) {
	print(fmt.Sprintf(format, args...))
}

func Println(args ...interface{}) {
	println(args)
}

func Success(message string) {
	Printf("%s%s %s%s\n", Bold, Green, message, Reset)
}

func Error(message string) {
	Printf("%s%sError:%s %s\n", Bold, Red, Reset, message)
}

func Info(message string) {
	Printf("%s%sInfo:%s %s\n", Bold, Blue, Reset, message)
}

func Warning(message string) {
	Printf("%s%sWarning:%s %s\n", Bold, Yellow, Reset, message)
}
