package entities

type Channel struct {
	UserNameSpace string
	ChatRoomName  string
} //@name Channel

func (x *Channel) MarshalJSON() ([]byte, error) {
	out := make([]byte, 0, len(x.UserNameSpace)+len(x.ChatRoomName)+1)
	out = append(out, []byte(x.UserNameSpace)...)
	out = append(out, ':')
	out = append(out, []byte(x.ChatRoomName)...)
	return out, nil
}

func (x *Channel) UnmarshalJSON(b []byte) error {
	isUserNameSpace := true
	sl := make([]byte, 0, 32)
	for i := range b {
		if isUserNameSpace {
			if b[i] == ':' {
				isUserNameSpace = false
				x.UserNameSpace = string(sl)
				sl = sl[:0]
				continue
			}
			sl = append(sl, b[i])
			if i == len(b)-1 {
				x.UserNameSpace = string(sl)
				return nil
			}
			continue
		}
		sl = append(sl, b[i])
	}

	x.ChatRoomName = string(sl)
	return nil
}
