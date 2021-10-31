package models

import "encoding/json"

type UserList struct {
	User []*UserInfo `redis:"user"`
}

func (ul UserList) MarshalBinary() ([]byte, error) {
	return json.Marshal(ul)
}

func (ul *UserList) UnmarshalBinary(info []byte) error {
	return json.Unmarshal(info, &ul)
}

type UserInfo struct {
	Id       int    `redis:"id"`
	Password string `redis:"password"`
	Login    string `redis:"login"`
	Name     string `redis:"name"`
	Phone    string `redis:"phone"`
}

func (ui UserInfo) MarshalBinary() ([]byte, error) {
	return json.Marshal(ui)
}
func (ui *UserInfo) UnmarshalBinary(info []byte) error {
	return json.Unmarshal(info, &ui)
}
