package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/xerrors"
	git "gopkg.in/src-d/go-git.v4"
)

func main() {
	log.Println("Go pj builder")

	if len(os.Args) != 2 {
		log.Fatal("Usage: go-pj-builder [project name]")
	}

	if err := createNewProject(os.Args[1]); err != nil {
		log.Fatalf("Error: %+v", err)
	}
}

func createNewProject(name string) error {
	// check target directory
	if len(name) == 0 {
		return xerrors.New("app name is empty")
	}
	log.Printf("App name: %s\n", name)

	targetDir := filepath.Join(".", name)
	_, err := os.Stat(targetDir)
	if err == nil {
		return xerrors.Errorf("%s already exists", targetDir)
	}

	// Create working directory
	workDir, err := ioutil.TempDir("", "pj")
	if err != nil {
		return xerrors.Errorf("failed to create temp dir: %w", err)
	}
	log.Printf("Working directory: %+v\n", workDir)

	// Clone golang-standards/project-layout
	log.Println("Cloning golang-standards/project-layout")
	if _, err = git.PlainClone(workDir, false, &git.CloneOptions{
		URL:      "https://github.com/golang-standards/project-layout",
		Progress: os.Stdout,
	}); err != nil {
		return xerrors.Errorf("failed to clone golang-standards/project-layout.")
	}

	// find directories
	dirsWithAppName := []string{}
	readmeFiles := []string{}
	if err := filepath.Walk(workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if strings.HasSuffix(path, "_your_app_") {
				dirsWithAppName = append(dirsWithAppName, path)
			}
		} else {
			_, filename := filepath.Split(path)
			if filename == "README.md" {
				readmeFiles = append(readmeFiles, path)
			}
		}

		return nil
	}); err != nil {
		return xerrors.Errorf("failed in listing directories: %w", err)
	}

	for _, path := range dirsWithAppName {
		if err := os.Rename(path, strings.ReplaceAll(path, "_your_app_", name)); err != nil {
			return xerrors.Errorf("failed to rename directory with appname: %w", err)
		}
	}

	for _, path := range readmeFiles {
		if err := os.Remove(path); err != nil {
			return xerrors.Errorf("failed to remove README.md: %w", err)
		}
	}

	// Remove git directory
	if err := os.RemoveAll(filepath.Join(workDir, ".git")); err != nil {
		return xerrors.Errorf("failed to remove .git directory: %w", err)
	}

	// Move directory
	if err := os.Rename(workDir, targetDir); err != nil {
		return xerrors.Errorf("failed to move directory: %w", err)
	}

	log.Printf("Project created: %s\n", targetDir)

	return nil
}
