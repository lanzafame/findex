package main

import (
	"bufio"
	"compress/zlib"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
)

func index(indexfile, root string) {
	fout, ferr := os.Create(indexfile)
	if ferr != nil {
		fmt.Println("error: ", ferr)
		os.Exit(1)
	}
	defer fout.Close()

	writer := zlib.NewWriter(fout)
	defer writer.Close()

	filepath.Walk(root,
		func(path string, f os.FileInfo, err error) error {
			fmt.Fprintln(writer, path)
			return nil
		})
}

func search(indexfile, pattern string) {

	fin, ferr := os.Open(indexfile)
	if ferr != nil {
		fmt.Println("error: ", ferr)
		os.Exit(1)
	}
	defer fin.Close()

	decompressor, derr := zlib.NewReader(fin)
	if derr != nil {
		fmt.Println("error: ", derr)
		os.Exit(1)
	}
	defer decompressor.Close()

	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Printf("error: invalid regex - '%s'\n", pattern)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(decompressor)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		path := scanner.Text()
		if re.MatchString(path) {
			fmt.Println(path)
		}
	}
}

func main() {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Println("error: cannot determine current user")
		os.Exit(1)
	}
	defaultIndex := path.Join(currentUser.HomeDir, ".findex")
	indexfile := flag.String("i", defaultIndex, "Index file")
	flag.Parse()

	if len(flag.Args()) != 2 {
		fmt.Println("Usage:")
		fmt.Printf("\t%s [-i index file] search <pattern>\n", path.Base(os.Args[0]))
		fmt.Printf("\t%s [-i index file] index <path>\n", path.Base(os.Args[0]))
		os.Exit(1)
	}

	switch flag.Args()[0] {
	case "search":
		if _, err := os.Stat(*indexfile); os.IsNotExist(err) {
			fmt.Println("error: index file does not exist.")
			os.Exit(1)
		}
		search(*indexfile, flag.Args()[1])
	case "index":
		index(*indexfile, flag.Args()[1])
	default:
		fmt.Printf("error: unknown command %s\n", flag.Args()[0])
	}
}
