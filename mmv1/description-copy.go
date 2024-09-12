package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CopyAllDescriptions() {
	// identifiers := []string{
	// 	"description:",
	// 	"note:",
	// 	"set_hash_func:",
	// 	"warning:",
	// 	"required_properties:",
	// 	"optional_properties:",
	// 	"attributes:",
	// }

	// for i, id := range identifiers {
	// 	CopyText(id, len(identifiers)-1 == i)
	// }

	copyComments()
}

// Used to copy/paste comments from Ruby -> Go YAML files
func copyComments() {
	renamedFields := map[string]string{
		"skip_sweeper": "exclude_sweeper",
		"skip_delete":  "exclude_delete",
		"values":       "enum_values",
	}
	var allProductFiles []string = make([]string, 0)
	files, err := filepath.Glob("products/**/go_product.yaml")
	if err != nil {
		return
	}
	for _, filePath := range files {
		dir := filepath.Dir(filePath)
		allProductFiles = append(allProductFiles, fmt.Sprintf("products/%s", filepath.Base(dir)))
	}

	for _, productPath := range allProductFiles {
		// Gather go and ruby file pairs
		yamlMap := make(map[string][]string)
		yamlPaths, err := filepath.Glob(fmt.Sprintf("%s/*", productPath))
		if err != nil {
			log.Fatalf("Cannot get yaml files: %v", err)
		}
		for _, yamlPath := range yamlPaths {
			if strings.HasSuffix(yamlPath, "_new") {
				continue
			}
			fileName := filepath.Base(yamlPath)
			baseName, found := strings.CutPrefix(fileName, "go_")
			if yamlMap[baseName] == nil {
				yamlMap[baseName] = make([]string, 2)
			}
			if found {
				yamlMap[baseName][1] = yamlPath
			} else {
				yamlMap[baseName][0] = yamlPath
			}
		}

		for _, files := range yamlMap {
			rubyPath := files[0]
			goPath := files[1]
			// TODO rewrite: ServicePerimeters.yaml is an exeption and needs manually copy the comments over after switchover
			if strings.Contains(rubyPath, "products/accesscontextmanager/ServicePerimeters.yaml") {
				continue
			}

			if !strings.Contains(rubyPath, "bigqueryconnection/Connection.yaml") {
				continue
			}

			recordingComments := false
			comments := ""
			commentsMap := make(map[string]string, 0)
			nestedNameLine := ""
			nestedParentLine := ""
			trimmedPreviousLine := ""

			// Ready Ruby yaml
			wholeLineComment, err := regexp.Compile(`^\s*#.*?`)
			if err != nil {
				log.Fatalf("Cannot compile the regular expression: %v", err)
			}

			if err != nil {
				log.Fatalf("Cannot compile the regular expression: %v", err)
			}

			file, _ := os.Open(rubyPath)
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				if wholeLineComment.MatchString(line) {
					if !recordingComments {
						recordingComments = true
						comments = line
					} else {
						comments = fmt.Sprintf("%s\n%s", comments, line)
					}
				} else {
					normalizedLine := line

					indexOfComment := strings.Index(normalizedLine, " # ")
					if indexOfComment > 0 { // The comments are in the same line with the code
						comments = normalizedLine[indexOfComment:]
						recordingComments = true
						normalizedLine = strings.TrimRight(normalizedLine[:indexOfComment], " ")
					}

					normalizedLine = strings.ReplaceAll(normalizedLine, "'", "")
					normalizedLine = strings.ReplaceAll(normalizedLine, `"`, "")
					normalizedLine = strings.ReplaceAll(normalizedLine, `\`, "")
					normalizedLine = strings.ReplaceAll(normalizedLine, ": :", ": ")
					normalizedLine = strings.ReplaceAll(normalizedLine, "- :", "- ")
					trimmed := strings.TrimSpace(normalizedLine)
					index := strings.Index(normalizedLine, trimmed)

					if index == 0 {
						nestedParentLine = ""
						nestedNameLine = ""
					} else if index >= 2 && (strings.HasPrefix(trimmedPreviousLine, "- !ruby/object") || strings.HasPrefix(trimmedPreviousLine, "--- !ruby/object")) {
						normalizedLine = fmt.Sprintf("%s- %s", normalizedLine[:index-2], normalizedLine[index:])

						if strings.HasPrefix(trimmed, "name:") {
							if nestedNameLine != "" {
								nestedParentLine = nestedNameLine
							}
							nestedNameLine = normalizedLine
						}
					}

					trimmedPreviousLine = trimmed

					if recordingComments {
						if !strings.HasPrefix(comments, "# Copyright") {
							// The line is a type, for example - !ruby/object:Api::Type::Array.
							// The lines of types are not present in Go yaml files
							if strings.HasPrefix(trimmed, "- !ruby/object") || strings.HasPrefix(trimmed, "--- !ruby/object") {
								continue
							}

							// Remove suffix " !ruby/object" as the types are not present in Go yaml files
							indexOfRuby := strings.Index(normalizedLine, ": !ruby/object")
							if indexOfRuby >= 0 {
								normalizedLine = normalizedLine[:indexOfRuby+1]
							}
							// Remove suffix Api::Type::
							indexOfRuby = strings.Index(normalizedLine, " Api::Type::")
							if indexOfRuby >= 0 {
								normalizedLine = normalizedLine[:indexOfRuby]
							}

							// Some fields are renamed during yaml file conversion
							field := strings.Split(normalizedLine, ":")[0]
							if shouldUseFieldName(normalizedLine) {
								normalizedLine = field
							}

							field = strings.TrimSpace(field)
							if goName, ok := renamedFields[field]; ok {
								normalizedLine = strings.Replace(normalizedLine, field, goName, 1)
							}

							key := fmt.Sprintf("%s$%s$%s", nestedParentLine, nestedNameLine, normalizedLine)
							commentsMap[key] = comments
						}
						recordingComments = false
						comments = ""
					}
				}
			}

			if len(commentsMap) == 0 {
				continue
			}

			// Read Go yaml while writing to a temp file
			nestedNameLine = ""
			nestedParentLine = ""
			newFilePath := fmt.Sprintf("%s_new", goPath)
			fo, _ := os.Create(newFilePath)
			w := bufio.NewWriter(fo)
			file, _ = os.Open(goPath)
			defer file.Close()
			scanner = bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					continue
				}
				if !wholeLineComment.MatchString(line) { // This line is not a comment
					// Replace ' with whitespace
					normalizedLine := strings.ReplaceAll(line, "'", "")
					normalizedLine = strings.ReplaceAll(normalizedLine, `"`, "")
					trimmed := strings.TrimSpace(normalizedLine)
					index := strings.Index(normalizedLine, trimmed)

					if index == 0 {
						nestedParentLine = ""
						nestedNameLine = ""
					} else if index >= 2 && strings.HasPrefix(trimmed, "- name:") {
						if nestedNameLine != "" {
							nestedParentLine = nestedNameLine
						}
						nestedNameLine = normalizedLine
					}

					field := strings.Split(normalizedLine, ":")[0]
					if shouldUseFieldName(normalizedLine) {
						normalizedLine = field
					}

					key := fmt.Sprintf("%s$%s$%s", nestedParentLine, nestedNameLine, normalizedLine)
					if comments, ok := commentsMap[key]; ok {
						delete(commentsMap, key)
						line = fmt.Sprintf("%s\n%s", comments, line)
					}
				}
				_, err := w.WriteString(fmt.Sprintf("%s\n", line))
				if err != nil {
					log.Fatalf("Error when writing the line %s: %#v", line, err)
				}
			}

			// Flush writes any buffered data to the underlying io.Writer.
			if err = w.Flush(); err != nil {
				panic(err)
			}

			if len(commentsMap) > 0 {
				log.Printf("Some comments in rubyPath %s are not copied over: %#v", rubyPath, commentsMap)
			}
			// Overwrite original file with temp
			os.Rename(newFilePath, goPath)
		}

	}

}

// custom template files in Go yaml files have different names
// The format of primary_resource_name for enum is different in Go yaml files
func shouldUseFieldName(line string) bool {
	filedNames := []string{"templates/terraform/", "primary_resource_name:", "default_value:", "deprecation_message:"}
	for _, fieldName := range filedNames {
		if strings.Contains(line, fieldName) {
			return true
		}
	}
	return false
}

// Used to copy/paste text from Ruby -> Go YAML files
func CopyText(identifier string, last bool) {
	var allProductFiles []string = make([]string, 0)
	files, err := filepath.Glob("products/**/go_product.yaml")
	if err != nil {
		return
	}
	for _, filePath := range files {
		dir := filepath.Dir(filePath)
		allProductFiles = append(allProductFiles, fmt.Sprintf("products/%s", filepath.Base(dir)))
	}

	for _, productPath := range allProductFiles {
		if strings.Contains(productPath, "healthcare") || strings.Contains(productPath, "memorystore") {
			continue
		}
		// Gather go and ruby file pairs
		yamlMap := make(map[string][]string)
		yamlPaths, err := filepath.Glob(fmt.Sprintf("%s/*", productPath))
		if err != nil {
			log.Fatalf("Cannot get yaml files: %v", err)
		}
		for _, yamlPath := range yamlPaths {
			if strings.HasSuffix(yamlPath, "_new") {
				continue
			}
			fileName := filepath.Base(yamlPath)
			baseName, found := strings.CutPrefix(fileName, "go_")
			if yamlMap[baseName] == nil {
				yamlMap[baseName] = make([]string, 2)
			}
			if found {
				yamlMap[baseName][1] = yamlPath
			} else {
				yamlMap[baseName][0] = yamlPath
			}
		}

		for _, files := range yamlMap {
			rubyPath := files[0]
			goPath := files[1]
			var text []string
			currText := ""
			recording := false

			if strings.Contains(rubyPath, "product.yaml") {
				// log.Printf("skipping %s", rubyPath)
				continue
			}

			// Ready Ruby yaml
			file, _ := os.Open(rubyPath)
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, identifier) && !strings.HasPrefix(strings.TrimSpace(line), "#") {
					currText = strings.SplitAfter(line, identifier)[1]
					recording = true
				} else if recording {
					if terminateText(line) {
						text = append(text, currText)
						currText = ""
						recording = false
					} else {
						currText = fmt.Sprintf("%s\n%s", currText, line)
					}
				}
			}
			if recording {
				text = append(text, currText)
			}

			// Read Go yaml while writing to a temp file
			index := 0
			firstLine := true
			newFilePath := fmt.Sprintf("%s_new", goPath)
			fo, _ := os.Create(newFilePath)
			w := bufio.NewWriter(fo)
			file, _ = os.Open(goPath)
			defer file.Close()
			scanner = bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if firstLine {
					if strings.Contains(line, "NOT CONVERTED - RUN YAML MODE") {
						firstLine = false
						if !last {
							w.WriteString(fmt.Sprintf("NOT CONVERTED - RUN YAML MODE\n"))
						}
						continue
					} else {
						break
					}
				}

				if strings.Contains(line, identifier) {
					if index >= len(text) {
						log.Printf("did not replace %s correctly! Is the file named correctly?", goPath)
						w.Flush()
						break
					}
					line = fmt.Sprintf("%s%s", line, text[index])
					index += 1
				}
				w.WriteString(fmt.Sprintf("%s\n", line))
			}

			if !firstLine {
				if index != len(text) {
					log.Printf("potential issue with %s, only completed %d index out of %d replacements", goPath, index, len(text))
				}
				if err = w.Flush(); err != nil {
					panic(err)
				}

				// Overwrite original file with temp
				os.Rename(newFilePath, goPath)
			} else {
				os.Remove(newFilePath)
			}
		}

	}

}

// quick and dirty logic to determine if a description/note is terminated
func terminateText(line string) bool {
	terminalStrings := []string{
		"!ruby/",
	}

	for _, t := range terminalStrings {
		if strings.Contains(line, t) {
			return true
		}
	}

	if regexp.MustCompile(`^\s*https:[\s$]*`).MatchString(line) {
		return false
	}

	return regexp.MustCompile(`^\s*[a-z_]+:[\s$]*`).MatchString(line)
}
