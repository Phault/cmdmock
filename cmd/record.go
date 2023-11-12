package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/Phault/cmdmock/recorder"
	"github.com/spf13/cobra"
)

func addExtIfMissing(p string, ext string) string {
	if path.Ext(p) != ext {
		return fmt.Sprint(p, ext)
	}

	return p
}

func saveRecording(recording *recorder.Recording, directory string, name string, force bool) (outputFile string, err error) {
	directory, err = resolvePath(directory)
	if err != nil {
		return "", fmt.Errorf("unable to resolve path: %v", err)
	}

	filePath := addExtIfMissing(path.Join(directory, name), ".json")

	if _, err := os.Stat(filePath); err == nil {
		if !force {
			return "", fmt.Errorf("a recording already exists at %s", filePath)
		}
	} else if !force && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("unable to assess whether an existing recording exists: %v", err)
	}

	recordingJson, err := json.MarshalIndent(recording, "", " ")
	if err != nil {
		return "", fmt.Errorf("unable to convert to json: %v\n", err)
	}

	os.MkdirAll(path.Dir(filePath), os.ModePerm)
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to create file: %v\n", err)
	}

	_, err = file.Write(recordingJson)

	if err != nil {
		return "", fmt.Errorf("unable to write recording: %v\n", err)
	}

	return filePath, nil
}

var (
	quiet bool
	force bool
)

var recordCmd = &cobra.Command{
	Use:                   "record name [flags] -- command...",
	DisableFlagsInUseLine: true,
	Short:                 "Records a command's output for later replay.",
	Args:                  cobra.MinimumNArgs(2),
	RunE: func(cobraCmd *cobra.Command, args []string) error {
		name := args[0]
		commandArgs := args[1:]

		var cmd *recorder.RecordedCmd
		if len(commandArgs) == 1 {
			cmd = recorder.Command(commandArgs[0])
		} else {
			cmd = recorder.Command(commandArgs[0], commandArgs[1:]...)
		}

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		recording, err := cmd.Run()
		if err != nil {
			_, isExitError := err.(*exec.ExitError)

			if !isExitError {
				return fmt.Errorf("failed to run the command: %v", err)
			}
		}

		outputFile, err := saveRecording(recording, OutputDir, name, force)
		if err != nil {
			return fmt.Errorf("failed to save recording: %v", err)
		}

		cobraCmd.Println("Saved recording to", outputFile)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)

	recordCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Do not passthrough output from the recorded command.")
	recordCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite any existing recording.")
}
