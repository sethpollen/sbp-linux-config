// Specifies the information my prompt has to import from a corp codebase.

package sbpgo

// Any method can return nil if the system doesn't support it.
type CorpContext interface {
	// Directory which contains all p4 repositories as children.
	P4Root() *string

	// Command for getting the status of a p4 repository.
	P4StatusCommand() *string

	// Extracts a WorkspaceStatus from the output of the P4StatusCommand.
	P4Status(output []byte) (*WorkspaceStatus, error)

	// Command for getting the commit log of a Mercurial repository.
	HgLogCommand() *string

	// Extracts a WorkspaceStatus from the output of the HgLogCommand.
	HgLog(output []byte) (*WorkspaceStatus, error)
}
