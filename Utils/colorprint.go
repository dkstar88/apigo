package Utils

import "github.com/fatih/color"

func ColorPrintln(fgColor color.Attribute, line string) {
	color.New(fgColor).Println(line)
	color.Unset()
}

func ColorPrint(fgColor color.Attribute, str string) {
	color.New(fgColor).Print(str)
	color.Unset()
}

func ColorPrintPair(fgColorLeft color.Attribute, left string, fgColorRight color.Attribute, right string) {
	ColorPrint(fgColorLeft, left)
	print(": ")
	ColorPrint(fgColorRight, right)
}

func ColorPrintSummary(left string, fgColorRight color.Attribute, right string) {
	ColorPrint(color.FgWhite, left)
	print(": ")
	ColorPrint(fgColorRight, right)
	println("")
}
