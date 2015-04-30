package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyFromBehavior(t *testing.T) {
	target := hooks{
		RawPreStart:  "from target",
		RawPostStart: "from target",
	}
	source := hooks{
		RawPreStart: "from source",
	}
	target.CopyFrom(source)
	assert.Equal(t, "from source", target.RawPreStart, "Source hook should have precedence")
	assert.Equal(t, "from target", target.RawPostStart, "Undefined hooks in target should not affect existing hooks")
}

func TestCopyFromReturnValue(t *testing.T) {
	target := hooks{
		RawPreStart: "foo",
	}
	source := hooks{
		RawPostStart: "bar",
	}
	assert.False(t, target.CopyFrom(source), "Copying unrelated hooks should not trigger an override")
	target = hooks{
		RawPreStart: "foo",
	}
	source = hooks{
		RawPreStart:  "bar",
		RawPostStart: "bar",
	}
	assert.True(t, target.CopyFrom(source), "Copying related hooks should trigger an override")
}
