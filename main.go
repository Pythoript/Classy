package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
)

var htmlClassRegex = regexp2.MustCompile(`class\s*=\s*["']?([\w\s-]+)["']?`, regexp2.None)
var cssClassRegex = regexp2.MustCompile(`\.[a-zA-Z_][\w-]*`, regexp2.None)
var jsQuerySelectorRegex = regexp2.MustCompile(`querySelector(All)?\(\s*["']\.([\w\s.-]+)["']\s*\)`, regexp2.None)
var jsClassListRegex = regexp2.MustCompile(`classList\.(add|remove|toggle)\(\s*["']([\w\s-]+)["'](?:,\s*["']([\w\s-]+)["'])*\s*\)`, regexp2.None)
var jsClassNameRegex = regexp2.MustCompile(`className\s*=\s*["']([^"'\s]+(?:\s+[^"'\s]+)*)["']`, regexp2.None)

func regexp2FindAllString(re *regexp2.Regexp, s string) []string {
	var matches []string
	m, _ := re.FindStringMatch(s)
	for m != nil {
		matches = append(matches, m.String())
		m, _ = re.FindNextMatch(m)
	}
	return matches
}

type ClassUsage struct {
	name  string
	count int
}

func main() {
	dir := flag.String("dir", ".", "Directory to recursively scan for HTML, CSS, JS, and PHP files")
	preview := flag.Bool("preview", false, "Only show class renaming without modifying files")
	allowDuplicates := flag.Bool("allow-duplicates", false, "Allow duplicate classes in HTML attributes")
	flag.Parse()

	classCount := make(map[string]int)
	err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && isSupportedFile(path) {
			processFile(path, classCount)
		}
		return nil
	})

	if err != nil {
		return
	}

	classes := rankClasses(classCount)
	classMap := generateClassMap(classes)
	if *preview {
		printClassMap(classMap)
	} else {
		err := filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && isSupportedFile(path) {
				renameClassesInFile(path, classMap, *allowDuplicates)
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error during file processing: %v\n", err)
		}
	}
}

func isSupportedFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".html" || ext == ".css" || ext == ".js" || ext == ".php"
}

func processFile(path string, classCount map[string]int) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		ext := filepath.Ext(path)
		if ext == ".css" {
			extractClassesFromCSS(line, classCount)
		} else if ext == ".js" {
			extractClassesFromJS(line, classCount)
		} else {
			extractClassesFromHTML(line, classCount)
		}
	}
}

func extractClassesFromHTML(line string, classCount map[string]int) {
	matches := regexp2FindAllString(htmlClassRegex, line)
	for _, match := range matches {
		classes := strings.Fields(match)
		for _, className := range classes {
			classCount[className]++
		}
	}
}

func extractClassesFromCSS(line string, classCount map[string]int) {
	matches := regexp2FindAllString(cssClassRegex, line)
	for _, match := range matches {
		classCount[match[1:]]++
	}
}

func extractClassesFromJS(line string, classCount map[string]int) {
	queryMatches := regexp2FindAllString(jsQuerySelectorRegex, line)
	for _, match := range queryMatches {
		classes := strings.Split(match, ".")
		for _, className := range classes[1:] {
			classCount[className]++
		}
	}

	classListMatches := regexp2FindAllString(jsClassListRegex, line)
	for _, match := range classListMatches {
		classes := extractClassesFromClassList(match)
		for _, className := range classes {
			classCount[className]++
		}
	}

	classNameMatches := regexp2FindAllString(jsClassNameRegex, line)
	for _, match := range classNameMatches {
		classes := strings.Fields(match)
		for _, className := range classes {
			classCount[className]++
		}
	}
}

func extractClassesFromClassList(match string) []string {
	insideParens := match[strings.Index(match, "(")+1 : strings.LastIndex(match, ")")]
	classNames := strings.Split(insideParens, ",")
	classes := []string{}
	for _, class := range classNames {
		trimmedClass := strings.TrimSpace(class)
		trimmedClass = strings.Trim(trimmedClass, `"'`)
		classes = append(classes, trimmedClass)
	}
	return classes
}

func rankClasses(classCount map[string]int) []ClassUsage {
	classes := make([]ClassUsage, 0, len(classCount))
	for className, count := range classCount {
		classes = append(classes, ClassUsage{name: className, count: count})
	}
	sort.Slice(classes, func(i, j int) bool {
		return classes[i].count > classes[j].count
	})
	return classes
}

func generateClassMap(classes []ClassUsage) map[string]string {
	classMap := make(map[string]string)
	for i, class := range classes {
		classMap[class.name] = generateShortClassName(i)
	}
	return classMap
}

func generateShortClassName(index int) string {
	firstChar := 'a' + rune(index%26)
	second := index / 26
	if second == 0 {
		return string(firstChar)
	}
	secondChar := ""
	if second > 9 {
		secondChar = fmt.Sprintf("a%d", second-10)
	} else {
		secondChar = fmt.Sprintf("%d", second)
	}
	return fmt.Sprintf("%s%s", string(firstChar), secondChar)
}

func renameClassesInFile(path string, classMap map[string]string, allowDuplicates bool) {
	tempPath := path + ".tmp"
	inputFile, err := os.Open(path)
	if err != nil {
		return
	}
	defer inputFile.Close()

	outputFile, err := os.Create(tempPath)
	if err != nil {
		return
	}
	defer outputFile.Close()

	scanner := bufio.NewScanner(inputFile)
	writer := bufio.NewWriter(outputFile)

	for scanner.Scan() {
		line := scanner.Text()
		var updatedLine string
		if filepath.Ext(path) == ".css" {
			updatedLine = updateCSSClasses(line, classMap)
		} else if filepath.Ext(path) == ".js" {
			updatedLine = updateJSClasses(line, classMap)
		} else {
			updatedLine = updateHTMLClasses(line, classMap, allowDuplicates)
		}
		writer.WriteString(updatedLine + "\n")
	}

	writer.Flush()
	os.Rename(tempPath, path)
}

func updateCSSClasses(line string, classMap map[string]string) string {
	updatedLine := line
	matches := regexp2FindAllString(cssClassRegex, updatedLine)
	for _, match := range matches {
		if newClass, found := classMap[match[1:]]; found {
			updatedLine = strings.Replace(updatedLine, match, fmt.Sprintf(".%s", newClass), 1)
		}
	}
	return updatedLine
}

func updateHTMLClasses(line string, classMap map[string]string, allowDuplicates bool) string {
	updatedLine := line
	matches := regexp2FindAllString(htmlClassRegex, updatedLine)
	for _, match := range matches {
		originalClasses := strings.Fields(match)
		newClasses := []string{}
		for _, className := range originalClasses {
			if newClass, found := classMap[className]; found {
				newClasses = append(newClasses, newClass)
			} else {
				newClasses = append(newClasses, className)
			}
		}
		if !allowDuplicates {
			newClasses = uniqueClasses(newClasses)
		}
		updatedLine = strings.Replace(updatedLine, match, "class=\""+strings.Join(newClasses, " ")+"\"", 1)
	}
	return updatedLine
}

func updateJSClasses(line string, classMap map[string]string) string {
	updatedLine := line

	queryMatches := regexp2FindAllString(jsQuerySelectorRegex, updatedLine)
	for _, match := range queryMatches {
		parts := strings.SplitN(match, ".", 2)
		if len(parts) == 2 && strings.HasSuffix(parts[1], `")`) {
			classes := strings.Split(parts[1][:len(parts[1])-2], ".")
			for i, className := range classes {
				if newClass, found := classMap[className]; found {
					classes[i] = newClass
				}
			}
			updatedLine = strings.Replace(updatedLine, match, parts[0]+"."+strings.Join(classes, ".")+`")`, 1)
		}
	}

	classListMatches := regexp2FindAllString(jsClassListRegex, updatedLine)
	for _, match := range classListMatches {
		start := strings.Index(match, "(") + 1
		end := strings.LastIndex(match, ")")
		classNames := strings.Split(match[start:end], ",")
		for i, className := range classNames {
			className = strings.Trim(className, `"' `)
			if newClass, found := classMap[className]; found {
				classNames[i] = `"` + newClass + `"`
			} else {
				classNames[i] = `"` + className + `"`
			}
		}
		updatedLine = strings.Replace(updatedLine, match, match[:start]+strings.Join(classNames, ", ")+match[end:], 1)
	}

	classNameMatches := regexp2FindAllString(jsClassNameRegex, updatedLine)
	for _, match := range classNameMatches {
		start := strings.Index(match, "=") + 1
		classes := strings.Fields(strings.Trim(match[start:], " \"'`"))
		for i, className := range classes {
			if newClass, found := classMap[className]; found {
				classes[i] = newClass
			}
		}
		updatedLine = strings.Replace(updatedLine, match, fmt.Sprintf("className = '%s'", strings.Join(classes, " ")), 1)
	}

	return updatedLine
}

func uniqueClasses(classes []string) []string {
	seen := make(map[string]struct{})
	unique := []string{}
	for _, class := range classes {
		if _, found := seen[class]; !found {
			seen[class] = struct{}{}
			unique = append(unique, class)
		}
	}
	return unique
}

func printClassMap(classMap map[string]string) {
	for original, newClass := range classMap {
		fmt.Printf("%s -> %s\n", original, newClass)
	}
}
