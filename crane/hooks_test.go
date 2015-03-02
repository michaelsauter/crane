package crane

import (
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
	if target.RawPreStart != "from source" {
		t.Errorf("Source hook should have precedence but got %v", target.RawPreStart)
	}
	if target.RawPostStart != "from target" {
		t.Errorf("Undefined hooks in target should not affect existing hooks, got %v", target.RawPostStart)
	}
}

func TestCopyFromReturnValue(t *testing.T) {
	target := hooks{
		RawPreStart: "foo",
	}
	source := hooks{
		RawPostStart: "bar",
	}
	if overriden := target.CopyFrom(source); overriden {
		t.Errorf("Copying unrelated hooks should not trigger an override")
	}
	target = hooks{
		RawPreStart: "foo",
	}
	source = hooks{
		RawPreStart:  "bar",
		RawPostStart: "bar",
	}
	if overriden := target.CopyFrom(source); !overriden {
		t.Errorf("Copying related hooks should trigger an override")
	}
}
