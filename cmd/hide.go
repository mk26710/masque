package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mk26710/masque/helpers"
	"github.com/mk26710/masque/models"
	"github.com/spf13/cobra"
)

// hideCmd represents the hide command
var hideCmd = &cobra.Command{
	Use:   "hide [path to diretory]",
	Short: "Hides files in specified direcatory (not recursive) and creates a map.json in it",
	Args:  cobra.MinimumNArgs(1),
	Run:   hideRunner,
}

func CreateHideEntry(targetAbs string, file fs.DirEntry) (models.MasqueEntry, error) {
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

func hideRunner(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.PrintErrln("You must provide a path to target directory")
		return
	}

	targetDir := args[0]

	if exists := helpers.DirExists(targetDir); !exists {
		cmd.PrintErrln("You have provided an invalid directory path!")
		cmd.PrintErrf("Provided path: %s\n", targetDir)
		return
	}

	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	files, err := os.ReadDir(targetAbs)
	if err != nil {
		fmt.Println(err)
		return
	}

	if masqued := helpers.IsMasqued(targetAbs); masqued {
		cmd.Printf("%s is already masqued!\n", targetAbs)
		return
	}

	var entries []models.MasqueEntry

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".masque.json") {
			continue
		}

		if entry, err := CreateHideEntry(targetAbs, file); err == nil {
			entries = append(entries, entry)
		}
	}

	j, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		cmd.PrintErrln(err)
		return
	}

	mapPath := path.Join(targetAbs, "map.masque.json")
	os.WriteFile(mapPath, j, 0644)

	for _, entry := range entries {
		oldpath := path.Join(targetAbs, entry.OldName)
		newpath := path.Join(targetAbs, entry.NewName)

		if err := os.Rename(oldpath, newpath); err != nil {
			cmd.Println(err)
		}
	}

	cmd.Println("Done!")
}

func init() {
	rootCmd.AddCommand(hideCmd)
}
