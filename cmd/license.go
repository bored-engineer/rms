package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/bored-engineer/rms/aadrm"

	"github.com/spf13/cobra"

	"github.com/pkg/errors"
)

// licenseCmd represents the license command
var licenseUserAgent string
var licensePlatformID string
var licenseInsecure bool
var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Commmands to interact with licenses and aadrm",
}

// licenseShowCmd represents the fetch command on license
var licenseShowCmd = &cobra.Command{
	Use:   "show [user.license]",
	Args:  cobra.ExactArgs(1),
	Short: "Print the contents of a user license",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return errors.Wrapf(err, "failed to open license file %s", args[0])
		}
		defer f.Close()

		userLicense, err := aadrm.DecodeEndUserLicense(f)
		if err != nil {
			return errors.Wrap(err, "failed to decode license")
		}

		fmt.Println(userLicense.String())
		return nil
	},
}

// licenseFetchCmd represents the fetch command on license
var licenseFetchOutput string
var licenseFetchCmd = &cobra.Command{
	Use:   "fetch [access_token] [content.license]",
	Args:  cobra.ExactArgs(2),
	Short: "Fetch a user license using access_token",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Create the client
		client := aadrm.NewClient(&http.Client{
			Transport: &oauth2.Transport{
				Source: oauth2.StaticTokenSource(&oauth2.Token{
					AccessToken: strings.TrimSpace(args[0]),
				}),
				Base: &http.Transport{
					Proxy: http.ProxyFromEnvironment,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: licenseInsecure,
						RootCAs: aadrm.NewCertPool(),
					},
				},
			},
		})
		client.RMSPlatformID = licensePlatformID
		client.UserAgent = licenseUserAgent

		// Read in the file and find the start (sometimes there's a random prefix)
		license, err := ioutil.ReadFile(args[1])
		if err != nil {
			return errors.Wrapf(err, "failed to read license file %s", args[1])
		}
		if idx := bytes.Index(license, []byte("<?xml")); idx == -1 {
			return errors.New("license does not have xml prefix")
		} else if idx > 0 {
			license = license[idx:]
		}

		// Make the request
		userLicense, rawLicense, _, err := client.GetEndUserLicense(ctx, license)
		if err != nil {
			return errors.Wrap(err, "failed to request EndUserLicense")
		}

		// Print the license and write it to a file
		fmt.Println(userLicense.String())
		if err := ioutil.WriteFile(licenseFetchOutput, rawLicense, 0644); err != nil {
			return errors.Wrap(err, "failed to write license")
		}

		return nil
	},
}

// licenseDecryptCmd represents the fetch command on license
var licenseDecryptOutput string
var licenseDecryptCmd = &cobra.Command{
	Use:   "decrypt [user.license] [content.encrypted]",
	Args:  cobra.ExactArgs(2),
	Short: "Decrypt the contents of a file using the user license",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return errors.Wrapf(err, "failed to open license file %s", args[0])
		}
		defer f.Close()

		userLicense, err := aadrm.DecodeEndUserLicense(f)
		if err != nil {
			return errors.Wrap(err, "failed to decode license")
		}

		ciphertext, err := ioutil.ReadFile(args[1])
		if err != nil {
			return errors.Wrapf(err, "failed to read license file %s", args[1])
		}

		plaintext, err := userLicense.Key.Decrypt(ciphertext)
		if err != nil {
			return errors.Wrap(err, "failed to decrypt")
		}

		if err := ioutil.WriteFile(licenseDecryptOutput, plaintext, 0644); err != nil {
			return errors.Wrapf(err, "failed to write decrypted file %s", licenseDecryptOutput)
		}

		fmt.Printf("Decrypted %d bytes from %s\n", len(plaintext), args[1])
		return nil
	},
}

func init() {
	licenseCmd.AddCommand(licenseShowCmd)
	licenseFetchCmd.Flags().StringVarP(&licenseFetchOutput, "output", "o", "user.license", "Output file for the user license")
	licenseCmd.AddCommand(licenseFetchCmd)
	licenseCmd.AddCommand(licenseDecryptCmd)
	licenseDecryptCmd.Flags().StringVarP(&licenseDecryptOutput, "output", "o", "decrypted.compound", "Output file for the decryption")
	licenseCmd.PersistentFlags().BoolVar(&licenseInsecure, "insecure", false, "Disable all x509/TLS verification")
	licenseCmd.PersistentFlags().StringVarP(&licenseUserAgent, "user-agent", "u", "Outlook/16.35.20030802 CFNetwork/1121.1.2 Darwin/19.3.0 (x86_64)", "User Agent to present to aadrm")
	licenseCmd.PersistentFlags().StringVarP(&licensePlatformID, "platform-id", "p", "AppName=com.microsoft.Outlook;AppVersion=16.35;DevicePlatform=Mac;OSVersion=10.15.3;SDKVersion=4.2.21;ClientID=00000000-0000-0000-0000-000000000000", "X-MS-RMS-Platform-Id to present to aadrm")
	rootCmd.AddCommand(licenseCmd)
}
