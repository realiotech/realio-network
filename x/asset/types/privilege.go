package types

type PrivilegeImplMsg interface {
	NeedPrivilege() string
}
