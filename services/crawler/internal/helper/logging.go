package helper

import (
	"fmt"

	"golang.org/x/net/html"
)

func PrintHtmlNodeInfo(
	node *html.Node,
) {
	fmt.Println()
	fmt.Println("node.Attr", node.Attr)
	fmt.Println("node.Data", node.Data)
	fmt.Println("node.DataAtom", node.DataAtom)
	// fmt.Println("node.FirstChild", node.FirstChild)
	// fmt.Println("node.LastChild", node.LastChild)
	fmt.Println("node.Namespace", node.Namespace)
	// fmt.Println("node.NextSibling", node.NextSibling)
	// fmt.Println("node.Parent", node.Parent)
	// fmt.Println("node.PrevSibling", node.PrevSibling)
	fmt.Println("node.Type", node.Type)
}

func LogPageInfo() {

}

func LogPageError() {

}
