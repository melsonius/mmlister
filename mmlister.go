package main

import (
    "fmt"
    "flag"
    "os"
    "io/ioutil"
    "strings"
    "time"
    "encoding/json"
    "gopkg.in/yaml.v2"
    "path/filepath"
)

type FileList struct {
    ModifiedTime time.Time
    IsLink       bool
    IsDir        bool
    LinksTo      string
    Size         int64
    Name         string
    Children     []FileList
}

// map output formats to the methods that print them
var displayFunctionByFormat = map[string]func(string, bool){
    "json" : printJson,
    "yaml" : printYaml,
    "text" : printText,
}

func validateInput(path, outputFormat string) (bool, []string) {
    var errors []string

    if (path == "") {
        errors = append(errors, "Path must be specified.")
    } else {
        // determine if path points to a valid directory
        info, err := os.Stat(path);
        if (err != nil) {
            errors = append(errors, "An error occurred while trying to read path: " + err.Error())
        } else if (!info.IsDir()) {
            errors = append(errors, "Path must be a directory.")
        }
    }

    // the output flag is only acceptable if we've defined a display function for it.
    var _, displayFunctionExists = displayFunctionByFormat[outputFormat]
    if (!displayFunctionExists) {
        errors = append(errors, "Unknown output type \"" + outputFormat + "\"")
    }

    return len(errors) == 0, errors;
}

func main() {
    // set up the three explicit flags and parse the command line parameters
    pathFlag := flag.String("path", "", "REQUIRED: path to folder");
    recursiveFlag := flag.Bool("recursive", false, "list files recursively.  (default \"false\")");
    outputFlag := flag.String("output", "text", "json|yaml|text");
    flag.Parse();

    // invalid input prints any errors plus the default help
    var validInput, errors = validateInput(*pathFlag, *outputFlag);
    if (!validInput) {
        for _, e := range errors {
            fmt.Fprintln(os.Stderr, e, "\n");
        }

        flag.PrintDefaults();
        return;
    }

    // print using one of the supported methods
    var displayFunction = displayFunctionByFormat[*outputFlag]
    displayFunction(*pathFlag, *recursiveFlag);
}

func readDirectoryStructure(root string, recursive bool) [] FileList {
    var fileList []FileList
    files, _ := ioutil.ReadDir(root)
    for _, f := range files {
        var children []FileList;
        if (recursive) {
            children = readDirectoryStructure(root + "/" + f.Name(), recursive)
        }

        // determine whether it's a symbolic link, and if so, where it links to
        isLink := f.Mode() & os.ModeSymlink != 0
        linkedTo := ""
        if (isLink) {
            linkedTo, _ = filepath.EvalSymlinks(root + "/" + f.Name())
        }

        var file = FileList{f.ModTime(), isLink, f.IsDir(), linkedTo, f.Size(), f.Name(), children};
        fileList = append(fileList, file);
    }

    return fileList
}

func printText(path string, recursive bool) {
    var files = readDirectoryStructure(path, recursive)
    fmt.Println(path);
    printAsText(files, 1)
}

func printAsText(files []FileList, depth int) {
    for _, f := range files {
        var name = f.Name
        if (f.IsDir) {
            name += "/"
        }

        if (f.IsLink) {
            name += "* (" + f.LinksTo + ")"
        }

        fmt.Printf("%s%s\n", strings.Repeat("  ", depth), name)
        printAsText(f.Children, depth + 1)
    }
}

func printJson(path string, recursive bool) {
    var files = readDirectoryStructure(path, recursive)
    var jsonOutput, _ = json.MarshalIndent(files, "", "  ")
    fmt.Println(string(jsonOutput))
}

func printYaml(path string, recursive bool) {
    var files = readDirectoryStructure(path, recursive)
    var yamlOutput, _ = yaml.Marshal(files)
    fmt.Println(string(yamlOutput))
}