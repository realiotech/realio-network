package types

type PrivilegeI interface {
	NeedPrivilege() string
}
