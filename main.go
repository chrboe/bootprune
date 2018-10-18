package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func isLinuxImage(filename string) (bool, string) {
	if strings.HasPrefix(filename, "vmlinuz-") {
		return true, filename[len("vmlinuz-"):]
	}

	return false, ""
}

func createTempFile(versions []string) *os.File {
	tmpfile, err := ioutil.TempFile("", "bootprune.tmp.*")
	if err != nil {
		log.Fatal(err)
	}

	return tmpfile
}

func makePromptString(versions []string) string {
	var b strings.Builder

	for _, v := range versions {
		fmt.Fprintf(&b, "keep %s\n", v)
	}

	b.WriteString("\n")
	b.WriteString("# Commands:\n")
	b.WriteString("# k, keep <version> = keep this version\n")
	b.WriteString("# d, drop <version> = delete all files associated with this version\n")

	return b.String()
}

func containsVersion(versions []string, version string) bool {
	for _, v := range versions {
		if v == version {
			return true
		}
	}

	return false
}

func parseReadback(lines []string, versions []string) []string {
	deleteVersions := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "drop ") || strings.HasPrefix(line, "d ") {
			spaceIndex := strings.Index(line, " ")
			version := strings.Trim(line[spaceIndex:], " ")
			if containsVersion(versions, version) {
				deleteVersions = append(deleteVersions, version)
			}
		}
	}

	return deleteVersions
}

func getEditor() string {
	var e string

	if e = os.Getenv("VISUAL"); e != "" {
		return e
	}

	if e = os.Getenv("EDITOR"); e != "" {
		return e
	}

	return "vim"
}

func promptEditor(versions []string) ([]string, error) {
	tmpfile := createTempFile(versions)
	defer os.Remove(tmpfile.Name())

	_, err := tmpfile.WriteString(makePromptString(versions))
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(getEditor(), tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	tmpfile.Seek(0, 0)

	readbackLines := make([]string, 0)

	scanner := bufio.NewScanner(tmpfile)
	for scanner.Scan() {
		readbackLines = append(readbackLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	deleteVersions := parseReadback(readbackLines, versions)
	return deleteVersions, nil
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [y/N]: ", s)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	response = strings.ToLower(strings.TrimSpace(response))

	if response == "y" || response == "yes" {
		return true
	} else if response == "n" || response == "no" {
		return false
	} else {
		return false
	}
}

func main() {
	files, err := ioutil.ReadDir("/boot")
	if err != nil {
		log.Fatal(err)
	}

	versions := make([]string, 0)
	for _, file := range files {
		isImage, version := isLinuxImage(file.Name())
		if isImage {
			versions = append(versions, version)
		}
	}

	deleteVersions, err := promptEditor(versions)
	if err != nil {
		log.Fatal(err)
	}

	matches := make([]string, 0)
	for _, dv := range deleteVersions {
		newMatches, err := filepath.Glob(fmt.Sprintf("/boot/*%s*", dv))
		if err != nil {
			log.Fatal(err)
		}

		matches = append(matches, newMatches...)
	}

	if len(matches) == 0 {
		fmt.Println("Did not delete any files.")
		return
	}

	fmt.Println("Deleting the following files:")
	for _, m := range matches {
		fmt.Println("\t" + m)
	}

	confirmed := askForConfirmation("Is this okay?")
	if confirmed {
		deleted := 0
		for _, m := range matches {
			err = os.Remove(m)
			if err != nil {
				fmt.Println(err)
			} else {
				deleted++
			}
		}

		fmt.Println("Deleted", deleted, "files.")
	} else {
		fmt.Println("Did not delete any files.")
	}
}
