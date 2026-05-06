package command

// Command represents a CLI subcommand with a name, usage string,
// and a Run function that receives the remaining arguments.
type Command struct {
	Name  string
	Usage string
	Run   func(args []string) error
}
