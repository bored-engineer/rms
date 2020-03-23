package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/richardlehane/mscfb"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"
)

// compoundCmd represents the compound command
var compoundCmd = &cobra.Command{
	Use:   "compound",
	Short: "Commmands to interact with compound (binary) files",
}

// compoundUnpackCmd represents the unpack command on compound
var compoundUnpackOutput string
var compoundUnpackCmd = &cobra.Command{
	Use:   "unpack [file.compound]",
	Args:  cobra.ExactArgs(1),
	Short: "Unpack a compound file into a folder",
	RunE: func(cmd *cobra.Command, args []string) error {

		input, err := os.Open(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to open input file")
		}
		defer input.Close()

		doc, err := mscfb.New(input)
		if err != nil {
			return errors.Wrap(err, "failed to start compound reader")
		}

		for {
			entry, err := doc.Next()
			if entry == nil {
				break
			} else if err != nil {
				return errors.Wrap(err, "failed to read next compound file")
			}

			entryName := filepath.Join(filepath.Join(entry.Path...), entry.Name)

			if entry.Size == 0 {
				fmt.Printf("Skipping empty entry: %s\n", entryName)
				continue
			}

			destPath := filepath.Join(compoundUnpackOutput, filepath.Clean(entryName))
			destDir := filepath.Dir(destPath)

			if err := os.MkdirAll(destDir, 0755); err != nil {
				return errors.Wrapf(err, "failed to create directory for file %s", destPath)
			}

			output, err := os.OpenFile(destPath, os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				return errors.Wrapf(err, "failed to open output file %s", destPath)
			}

			n, err := io.Copy(output, entry)
			if err != nil {
				output.Close()
				return errors.Wrapf(err, "failed to copy file %s", destPath)
			}

			if err := output.Close(); err != nil {
				return errors.Wrapf(err, "failed to close file %s", destPath)
			}
			fmt.Printf("Wrote %d bytes from entry: %s\n", n, entryName)
		}

		outputPath, err := filepath.Abs(compoundUnpackOutput)
		if err != nil {
			outputPath = compoundUnpackOutput
		}
		fmt.Printf("Unpacked %s to %s\n", args[0], outputPath)

		return nil
	},
}

func init() {
	compoundUnpackCmd.Flags().StringVarP(&compoundUnpackOutput, "output", "o", "unpacked", "Output directory for the unpacked file")
	compoundCmd.AddCommand(compoundUnpackCmd)
	rootCmd.AddCommand(compoundCmd)
}
