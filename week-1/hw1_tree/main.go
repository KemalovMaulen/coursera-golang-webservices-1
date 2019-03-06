package main

import (
	"io"
	"os"
	"strconv"
	"strings"
	"io/ioutil"
)


type Tree struct {
	name     string
	size     int
	child    []Tree
}

func buildTree(path string, fileNode *Tree) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return  err
	}
	for _, file := range files {
		child := &Tree{name: file.Name() }
		if file.IsDir() {
			child.size = -1
			if err:= buildTree(path+"/"+child.name, child); err != nil {
				return  err
			}
		} else {
			child.size = int(file.Size())
		}
		fileNode.child = append(fileNode.child, *child)
	}

	return nil
}

func nullify(pref string) string {
	pref = strings.Replace(pref,"├───", "│\t", -1 )
	pref = strings.Replace(pref,"└───", "\t", -1 )
	return pref
}

func sortChild(node *Tree) {
	newChilds := []Tree{}
	for _, child := range node.child {
		if child.size == -1 {
			sortChild(&child)
			newChilds = append(newChilds, child)
		}
	}
	for _, child := range node.child {
		if child.size != -1 {
			newChilds = append(newChilds, child)
		}
	}
	node.child = newChilds
}

func printTree(out io.Writer, node *Tree, printFiles bool, prefix string) string {
	result := ""
	if node.size == -1 {
		//fmt.Println(prefix + node.name)
		result += prefix + node.name + "\n"
	} else if node.name != "" && printFiles {
		outRes := prefix + node.name + " (" +  strconv.Itoa(node.size) + "b)"
		if node.size == 0 {
			outRes = prefix + node.name + " (empty)"
		}
		//fmt.Println(outRes)
		result += outRes +"\n"
	}
	prefix = nullify(prefix)
	for ind, child := range node.child {
		if len(node.child)-1 == ind || (!printFiles && len(node.child)-1 != ind && node.child[ind+1].size != -1) {
			result += printTree(out, &child, printFiles, prefix + "└───" )
		} else {
			result += printTree(out, &child, printFiles, prefix + "├───")
		}
	}
	return result
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	root := &Tree{}
	err := buildTree(path , root)
	if err != nil {
		return  err
	}
	if !printFiles {
		sortChild(root)
	}
	result := printTree(out, root, printFiles, "")
	out.Write([]byte(result))
	return nil
}


func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"

	//path := "testdata"
	//printFiles := true

	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

