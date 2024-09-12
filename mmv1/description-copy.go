package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func CopyAllDescriptions() {
	identifiers := []string{
		"description:",
		"note:",
		"set_hash_func:",
		"warning:",
		"required_properties:",
		"optional_properties:",
		"attributes:",
	}

	for i, id := range identifiers {
		CopyText(id, len(identifiers)-1 == i)
	}

	// copyComments()
}

// Used to copy/paste text from Ruby -> Go YAML files
func copyComments() {
	renamedFields := map[string]string{
		"skip_sweeper": "exclude_sweeper",
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
		if !strings.Contains(productPath, "accesscontextmanager") {
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
			// log.Printf("files %#v", files)
			rubyPath := files[0]
			// rubyPath := "products/accesscontextmanager/AccessLevel.yaml"
			// goPath := "products/accesscontextmanager/go_AccessLevel.yaml"
			goPath := files[1]

			// if !strings.Contains(rubyPath, "products/accesscontextmanager/AccessLevel.yaml") {
			// 	continue
			// }

			recordingComments := false
			comments := ""
			commentsMap := make(map[string]string, 0)
			commentsAreTerminated := false
			nestedNameLine := ""

			// if strings.Contains(rubyPath, "product.yaml") {
			// 	// log.Printf("skipping %s", rubyPath)
			// 	continue
			// }

			// Ready Ruby yaml
			r, err := regexp.Compile(`^\s*#.*?`)
			if err != nil {
				log.Fatalf("Cannot compile the regular expression: %v", err)
			}

			file, _ := os.Open(rubyPath)
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()

				if r.MatchString(line) {
					// log.Printf("line %s", line)
					if !recordingComments {
						recordingComments = true
						comments = line
					} else {
						comments = fmt.Sprintf("%s\n%s", comments, line)
					}
				} else {
					// Replace ' with whitespace
					normalizedLine := strings.ReplaceAll(line, "'", "")
					trimmed := strings.TrimSpace(normalizedLine)
					index := strings.Index(normalizedLine, trimmed)

					if index == 0 {
						nestedNameLine = ""
					} else if index >= 2 && strings.HasPrefix(trimmed, "name:") {
						nestedNameLine = fmt.Sprintf("%s- %s", normalizedLine[:index-2], normalizedLine[index:])
						nestedNameLine = strings.ReplaceAll(nestedNameLine, "'", "")
					}

					if recordingComments {
						if !strings.HasPrefix(comments, "# Copyright") {
							// The line is a type, for example - !ruby/object:Api::Type::Array
							if strings.Contains(normalizedLine, "!ruby/object") {
								commentsAreTerminated = true
								continue
							}

							if commentsAreTerminated {
								log.Printf("line will be trimmed %s", normalizedLine)
								if index >= 2 {
									normalizedLine = fmt.Sprintf("%s- %s", normalizedLine[:index-2], normalizedLine[index:])
								}
								commentsAreTerminated = false
							}

							// Some fields are renamed during yaml file conversion
							field := strings.Split(normalizedLine, ":")[0]
							if goName, ok := renamedFields[field]; ok {
								normalizedLine = strings.Replace(normalizedLine, field, goName, 1)
							}

							key := fmt.Sprintf("%s$%s", nestedNameLine, normalizedLine)
							commentsMap[key] = comments
						}
						recordingComments = false
						comments = ""
					}
				}
			}

			if len(commentsMap) > 0 {
				j, _ := json.Marshal(commentsMap)
				log.Printf("test rubyPath %s commentsmap %#v", rubyPath, string(j))
			}

			// Read Go yaml while writing to a temp file
			nestedNameLine = ""
			newFilePath := fmt.Sprintf("%s_new", goPath)
			fo, _ := os.Create(newFilePath)
			w := bufio.NewWriter(fo)
			file, _ = os.Open(goPath)
			defer file.Close()
			scanner = bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				// log.Printf("line1 %s", line)

				if !r.MatchString(line) { // This line is not a comment
					// log.Printf("line2 %s", line)

					// log.Printf("line3 %s", line)

					// log.Printf("line %s", line)

					// Replace ' with whitespace
					normalizedLine := strings.ReplaceAll(line, "'", "")
					trimmed := strings.TrimSpace(normalizedLine)
					if strings.HasPrefix(trimmed, "- name:") {
						nestedNameLine = normalizedLine
					}
					key := fmt.Sprintf("%s$%s", nestedNameLine, normalizedLine)

					// if strings.Contains(normalizedLine, "- name: parent") {
					// 	log.Printf("line comments nestedNameLine %s: %d", nestedNameLine, len(nestedNameLine))
					// 	log.Printf("line comments normalizedLine %s: %d", normalizedLine, len(normalizedLine))
					// }
					if comments, ok := commentsMap[key]; ok {
						log.Printf("line has comments normalizedLine %s", normalizedLine)
						delete(commentsMap, key)
						line = fmt.Sprintf("%s\n%s", comments, line)
					}
				}
				// log.Printf("line4 %s", line)
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
				for key := range commentsMap {
					arr := strings.Split(key, "$")
					log.Printf("left comments name  %s: %d", arr[0], len(arr[0]))
					log.Printf("left comments line  %s: %d", arr[1], len(arr[1]))
				}

				j, _ := json.Marshal(commentsMap)
				log.Printf("some comments left: test rubyPath %s commentsmap %#v", rubyPath, string(j))

				// os.Remove(newFilePath)

			} else {
				// Overwrite original file with temp
				os.Rename(newFilePath, goPath)
			}
		}

	}

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
