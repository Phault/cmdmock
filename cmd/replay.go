package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Phault/cmdmock/recorder"
	"github.com/spf13/cobra"
)

func resolvePath(p string) (absPath string, err error) {
	if !path.IsAbs(p) {
		workingDir, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("Unable to retrieve working directory: %v\n", err)
		}

		return path.Join(workingDir, p), nil
	}

	return p, nil
}

func loadRecording(directory string, name string) (rec *recorder.Recording, err error) {
	directory, err = resolvePath(directory)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve path: %v", err)
	}

	filePath := addExtIfMissing(path.Join(directory, name), ".json")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v\n", err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&rec)

	if err != nil {
		return nil, fmt.Errorf("unable to decode recording: %v\n", err)
	}

	return
}

var replayCmd = &cobra.Command{
	Use:   "replay name",
	Short: "Replays a recorded command.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		rec, err := loadRecording(OutputDir, name)
		if err != nil {
			return fmt.Errorf("unable to load recording: %v", err)
		}

		for i, entry := range rec.Timeline {
			prevTimeOffset := time.Duration(0)

			if i > 0 {
				prevEntry := rec.Timeline[i-1]
				prevTimeOffset = prevEntry.TimeOffset
			}

			time.Sleep(entry.TimeOffset - prevTimeOffset)

			err := entry.Event.Execute()
			if err != nil {
				return fmt.Errorf("unable to load recording: %v", err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(replayCmd)
}
