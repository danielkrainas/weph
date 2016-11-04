package cmd

import (
	"github.com/spf13/cobra"

	"github.com/danielkrainas/wiph/context"
)

type ExecutorFunc func(ctx context.Context, args []string) error

type Info struct {
	Use   string
	Short string
	Long  string
	Run   ExecutorFunc
	Flags []*Flag
}

type FlagType string

var (
	FlagString FlagType = "string"
)

type Flag struct {
	Short       string
	Long        string
	Description string
	Type        FlagType
}

var registry map[string]*Info = make(map[string]*Info)

func Register(name string, info *Info) {
	registry[name] = info
}

func CreateDispatcher(ctx context.Context, info *Info) func() error {
	root := makeCobraCommand(ctx, info)
	for _, info := range registry {
		cmd := makeCobraCommand(ctx, info)
		root.AddCommand(cmd)
	}

	return func() error {
		return root.Execute()
	}
}

func makeCobraRunner(ctx context.Context, innerFunc ExecutorFunc, flags []*Flag) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx = contextWithFlags(ctx, cmd, flags)
		return innerFunc(ctx, args)
	}
}

func makeCobraCommand(ctx context.Context, info *Info) *cobra.Command {
	cmd := &cobra.Command{
		Use:   info.Use,
		Short: info.Short,
		Long:  info.Long,
	}

	for _, f := range info.Flags {
		switch f.Type {
		case FlagString:
			cmd.PersistentFlags().StringP(f.Long, f.Short, "", "")
		}

	}

	if info.Run != nil {
		cmd.RunE = makeCobraRunner(ctx, info.Run, info.Flags)
	}

	return cmd
}

func contextWithFlags(ctx context.Context, cmd *cobra.Command, flags []*Flag) context.Context {
	if len(flags) < 1 {
		return ctx
	}

	values := make(map[string]interface{})
	for _, f := range flags {
		v := cmd.PersistentFlags().Lookup(f.Long).Value.String()
		values["flags."+f.Long] = v
	}

	return context.WithValues(ctx, values)
}
