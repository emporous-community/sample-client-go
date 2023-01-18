package commands

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	managerapi "github.com/emporous/emporous-go/api/services/collectionmanager/v1alpha1"
	"github.com/emporous/emporous-go/config"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/square/go-jose.v2/json"

	"github.com/emporous-community/sample-client-go/cmd/sample-client/commands/internal"
)

// PullOptions describe configuration options that can
// be set using the pull subcommand.
type PullOptions struct {
	*RootOptions
	Source         string
	Output         string
	AttributeQuery string
}

// NewPullCmd creates a new cobra.Command for the pull subcommand.
func NewPullCmd(rootOpts *RootOptions) *cobra.Command {
	o := PullOptions{RootOptions: rootOpts}

	cmd := &cobra.Command{
		Use:           "pull SRC",
		Short:         "Pull an Emporous collection based on content or attribute address",
		SilenceErrors: false,
		SilenceUsage:  false,
		Args:          cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cobra.CheckErr(o.Complete(args))
			cobra.CheckErr(o.Run(cmd.Context()))
		},
	}

	cmd.Flags().StringVarP(&o.Output, "output", "o", o.Output, "output location for artifacts")
	cmd.Flags().StringVar(&o.AttributeQuery, "attributes", o.AttributeQuery, "attribute query config path")

	return cmd
}

func (o *PullOptions) Complete(args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}
	o.Source = args[0]

	absPath, err := filepath.Abs(o.Output)
	if err != nil {
		return err
	}

	o.Output = absPath
	return nil
}

func (o *PullOptions) Run(ctx context.Context) error {
	client, cleanup, err := clientSetup(ctx, o.ServerAddress)
	if err != nil {
		return err
	}
	defer cleanup()

	req := managerapi.Retrieve_Request{
		Source:      o.Source,
		Destination: o.Output,
	}

	if o.AttributeQuery != "" {
		query, err := config.ReadAttributeQuery(o.AttributeQuery)
		if err != nil {
			return err
		}
		var attributes map[string]interface{}
		err = json.Unmarshal(query.Attributes, &attributes)
		if err != nil {
			return err
		}
		filter, err := structpb.NewStruct(attributes)
		if err != nil {
			return err
		}
		req.Filter = filter
	}

	cred, err := internal.GetCredentials(o.Source)
	if err != nil {
		return err
	}
	req.Auth = cred

	resp, err := client.RetrieveContent(ctx, &req)
	if err != nil {
		return err
	}

	if len(resp.Digests) == 0 {
		fmt.Fprintln(o.IOStreams.Out, "No matching collections")
		return nil
	}

	for _, digest := range resp.Digests {
		fmt.Fprintln(o.IOStreams.Out, digest)
	}

	return nil
}
