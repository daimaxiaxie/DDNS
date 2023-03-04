package common

type CloudProvider interface {
	Init(config []byte) error
	Update() error
	Stop()

	Info() CloudInfo
}
