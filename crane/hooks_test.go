package crane

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyFromBehavior(t *testing.T) {
	target := Hooks{
		RawPreBuild:  "from target",
		RawPostBuild: "from target",
		RawPreStart:  "from target",
		RawPostStart: "from target",
	}
	source := Hooks{
		RawPreBuild: "from source",
		RawPreStart: "from source",
	}
	target.CopyFrom(source)
	assert.Equal(t, "from source", target.RawPreBuild, "Source hook should have precedence")
	assert.Equal(t, "from target", target.RawPostBuild, "Undefined hooks in target should not affect existing hooks")
	assert.Equal(t, "from source", target.RawPreStart, "Source hook should have precedence")
	assert.Equal(t, "from target", target.RawPostStart, "Undefined hooks in target should not affect existing hooks")
}

func TestCopyFromReturnValue(t *testing.T) {
	target := Hooks{
		RawPreStart: "foo",
	}
	source := Hooks{
		RawPostStart: "bar",
	}
	assert.False(t, target.CopyFrom(source), "Copying unrelated hooks should not trigger an override")
	target = Hooks{
		RawPreStart: "foo",
	}
	source = Hooks{
		RawPreStart:  "bar",
		RawPostStart: "bar",
	}
	assert.True(t, target.CopyFrom(source), "Copying related hooks should trigger an override")
}
