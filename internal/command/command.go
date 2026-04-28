package command

type Command struct {
	Name  string
	Usage string
	Run   func(args []string) error
}
