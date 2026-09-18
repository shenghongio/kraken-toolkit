package executor

const (
	DefaultSSHPort = 22
	DefaultSSHUser = "root"
)

//var groupRegex = regexp.MustCompile(`^\[([^\[\]]+)\]$`)

// Host 表示一个SSH目标主机
type Host struct {
	Address string
	User    string
	Port    int
	Passwd  string
}
