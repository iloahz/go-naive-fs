package naivefs

type FSType int

const (
	FSTypeDefault = 0
	FSTypeLocal   = 1
	FSTypeSamba   = 2
	FSTypeMinio   = 3
)

func (ft FSType) String() string {
	switch ft {
	case FSTypeDefault:
		return "default"
	case FSTypeLocal:
		return "local"
	case FSTypeSamba:
		return "samba"
	case FSTypeMinio:
		return "minio"
	default:
		return "unknown"
	}
}
