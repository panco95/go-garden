package user

type ExistsArgs struct {
	Username string
}

type ExistsReply struct {
	Exists bool
}
