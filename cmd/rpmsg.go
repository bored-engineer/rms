package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/bored-engineer/rms/rpmsg"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"
)

// rpmsgCmd represents the rpmsg command
var rpmsgCmd = &cobra.Command{
	Use:   "rpmsg",
	Short: "Commmands to interact with restricted-permissions messages (rmpsg)",
}

// rpmsgDecodeCmd represents the decode command on rpmsg
var rpmsgDecodeOutput string
var rpmsgDecodeCmd = &cobra.Command{
	Use:   "decode [message.rpmsg]",
	Args:  cobra.ExactArgs(1),
	Short: "Decode a rpmsg file into a raw compound file",
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := os.Open(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to open input file")
		}
		defer input.Close()
		r, err := rpmsg.NewReader(input)
		if err != nil {
			return errors.Wrap(err, "failed to start rpmsg reader")
		}
		output, err := os.OpenFile(rpmsgDecodeOutput, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return errors.Wrap(err, "failed to open output file")
		}
		defer output.Close()
		n, err := io.Copy(output, r)
		if err != nil {
			return errors.Wrap(err, "failed to decode")
		}
		fmt.Printf("Decoded %d bytes to compound file: %s\n", n, rpmsgDecodeOutput)
		return nil
	},
}

func init() {
	rpmsgDecodeCmd.Flags().StringVarP(&rpmsgDecodeOutput, "output", "o", "rpmsg.compound", "Output file for the decoded file")
	rpmsgCmd.AddCommand(rpmsgDecodeCmd)
	rootCmd.AddCommand(rpmsgCmd)
}
