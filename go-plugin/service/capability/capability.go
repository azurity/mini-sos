package capability

import (
	"errors"

	"github.com/google/uuid"
)

var ErrCapExist = errors.New("capability exist")
var ErrCapNotExist = errors.New("capability not exist")

var CapCall = uuid.MustParse("0191757a-b9f0-7212-8be4-d89f541b774f")
var CapListDir = uuid.MustParse("0191757d-5e8a-78e3-9163-545d699b8366")
var CapSetDir = uuid.MustParse("01917589-1f8f-7b80-a5a3-202bd380a063")

var KnownCapability = map[uuid.UUID]bool{}

func init() {
	KnownCapability[CapCall] = true
	KnownCapability[CapListDir] = true
}
