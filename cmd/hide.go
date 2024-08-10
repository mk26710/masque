package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mk26710/masque/helpers"
	"github.com/mk26710/masque/models"
	"github.com/spf13/cobra"
)

// hideCmd represents the hide command
var hideCmd = &cobra.Command{
	Use:   "hide [path to diretory]",
	Short: "Hides files in specified direcatory (not recursive) and creates a map.json in it",
	Args:  cobra.MinimumNArgs(1),
	RunE:  hideRunner,
}

func CreateMasqueEntry(targetAbs string, file fs.DirEntry) (models.MasqueEntry, error) {
	if file.IsDir() {
		return models.MasqueEntry{}, fmt.Errorf("%s is not a directory", file)
	}

	fp := filepath.Join(targetAbs, file.Name())

	sha, err := helpers.GetSha256(fp)
	if err != nil {
		return models.MasqueEntry{}, fmt.Errorf("can not obtain SHA256 for %s", fp)
	}

	result := models.MasqueEntry{
		NewName: sha + filepath.Ext(fp),
		OldName: file.Name(),
	}

	return result, nil
}

func CreateAllMasqueEntries(targetAbs string) ([]models.MasqueEntry, error) {
	files, err := os.ReadDir(targetAbs)
	if err != nil {
		return []models.MasqueEntry{}, err
	}

	if masqued := helpers.IsMasqued(targetAbs); masqued {
		return []models.MasqueEntry{}, fmt.Errorf("%s is already masqued", targetAbs)
	}

	wg := sync.WaitGroup{}
	out := make(chan models.MasqueEntry, 1000)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".masque.json") {
			continue
		}

		wg.Add(1)
		go func(targetAbs string, file fs.DirEntry) {
			defer wg.Done()
			if entry, err := CreateMasqueEntry(targetAbs, file); err == nil {
				out <- entry
			}
		}(targetAbs, file)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	var entries []models.MasqueEntry
	for entry := range out {
		entries = append(entries, entry)
	}

	return entries, nil
}

func HideMasqueEntris(targetAbs string, entries []models.MasqueEntry) []error {
	var errors []error

	for _, entry := range entries {
		oldpath := path.Join(targetAbs, entry.OldName)
		newpath := path.Join(targetAbs, entry.NewName)

		if err := os.Rename(oldpath, newpath); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func hideRunner(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a path to target directory")
	}

	targetDir := args[0]

	if exists := helpers.DirExists(targetDir); !exists {
		return fmt.Errorf("tou have provided an invalid directory path")
	}

	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	entries, err := CreateAllMasqueEntries(targetAbs)
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	mapPath := path.Join(targetAbs, "map.masque.json")
	os.WriteFile(mapPath, j, 0644)

	errors := HideMasqueEntris(targetAbs, entries)
	for _, err := range errors {
		cmd.Println(err)
	}

	cmd.Println("Done!")

	return nil
}

func init() {
	rootCmd.AddCommand(hideCmd)
}
