package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/dtan4/k8sec/k8s"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api/v1"
)

var listOpts = struct {
	base64encode bool
}{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List secrets",
	Long: `List secrets

$ k8sec list rails
NAME    TYPE    KEY             VALUE
rails   Opaque  database-url    "postgres://example.com:5432/dbname"

Show values as base64-encoded string:

$ k8sec list --base64 rails
NAME    TYPE    KEY             VALUE
rails   Opaque  database-url    cG9zdGdyZXM6Ly9leGFtcGxlLmNvbTo1NDMyL2RibmFtZQ==
`,
	RunE: doList,
}

func doList(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		return errors.New("Too many arguments.")
	}

	clientset, err := k8s.NewKubeClient(rootOpts.kubeconfig)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize Kubernetes API client.")
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, strings.Join([]string{"NAME", "TYPE", "KEY", "VALUE"}, "\t"))

	var v string

	if len(args) == 1 {
		secret, err := clientset.Core().Secrets(rootOpts.namespace).Get(args[0])
		if err != nil {
			return errors.Wrap(err, "Failed to retrieve secrets.")
		}

		for key, value := range secret.Data {
			if listOpts.base64encode {
				v = base64.StdEncoding.EncodeToString(value)
			} else {
				v = strconv.Quote(string(value))
			}

			fmt.Fprintln(w, strings.Join([]string{secret.Name, string(secret.Type), key, v}, "\t"))
		}
	} else {
		secrets, err := clientset.Core().Secrets(rootOpts.namespace).List(v1.ListOptions{})
		if err != nil {
			return errors.Wrap(err, "Failed to retrieve secrets.")
		}

		for _, secret := range secrets.Items {
			for key, value := range secret.Data {
				if listOpts.base64encode {
					v = base64.StdEncoding.EncodeToString(value)
				} else {
					v = strconv.Quote(string(value))
				}

				fmt.Fprintln(w, strings.Join([]string{secret.Name, string(secret.Type), key, v}, "\t"))
			}
		}
	}

	w.Flush()

	return nil
}

func init() {
	RootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listOpts.base64encode, "base64", false, "Show values as base64-encoded string")
}
