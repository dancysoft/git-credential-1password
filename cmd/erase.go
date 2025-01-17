package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/develerik/git-credential-1password/git"
	"github.com/develerik/git-credential-1password/onepassword"
)

// eraseCmd represents the store operation.
var eraseCmd = &cobra.Command{
	Use:              "erase",
	TraverseChildren: true,
	Args:             cobra.NoArgs,
	Short:            "Remove a matching credential, if any, from the helper’s storage",
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteCredentials(os.Stdin); err != nil {
			if _, err = fmt.Fprintf(os.Stderr, "%s", err); err != nil {
				panic(err)
			}
			os.Exit(1)
		}
	},
}

func deleteCredentials(r io.Reader) error {
	data, err := git.ReadInput(r)

	if err != nil {
		return err
	}

	c := onepassword.Client{
		Account:  account,
		Vault:    vault,
		NoSignin: noSignin,
	}

	if err := c.Login(cache); err != nil {
		return err
	}

	return c.DeleteCredentials(data["protocol"], data["host"], archive)
}
