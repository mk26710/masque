package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mk26710/masque/helpers"
	"github.com/mk26710/masque/models"
	"github.com/spf13/cobra"
)

// unhideCmd represents the unhide command
var unhideCmd = &cobra.Command{
	Use:   "unhide [path to masqued directory]",
	Short: "Restores filename in the provided directory according to map.masque.json",
	Args:  cobra.MinimumNArgs(1),
	RunE:  unhideRunner,
}

func unhideRunner(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("you must provide a path to target directory as first argument")
	}

	targetDir := args[0]

	if exists := helpers.DirExists(targetDir); !exists {
		return fmt.Errorf("provided path must be a path to a directory")
	}

	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return err
	}

	if masqued := helpers.IsMasqued(targetAbs); !masqued {
		return fmt.Errorf("%s does not contain %s", targetAbs, helpers.MAP_FILE_NAME)
	}

	file, err := os.Open(path.Join(targetAbs, helpers.MAP_FILE_NAME))
	if err != nil {
		return err
	}
	defer file.Close()

	var entries []models.MasqueEntry

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return err
	}

	for _, entry := range entries {
		oldpath := path.Join(targetAbs, entry.NewName)
		newpath := path.Join(targetAbs, entry.OldName)

		os.Rename(oldpath, newpath)
	}

	if err := os.Remove(path.Join(targetAbs, helpers.MAP_FILE_NAME)); err != nil {
		return err
	}

	cmd.Println("Unmasqued: ", targetAbs)

	return nil
}

func init() {
	rootCmd.AddCommand(unhideCmd)
}
