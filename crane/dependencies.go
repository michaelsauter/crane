package crane

// Dependencies contains 4 fields:
// all: contains all dependencies
// linked: containers linked to
// volumesFrom: containers that provide volumes
// net: container the net stack is shared with
type Dependencies struct {
	all         []string
	linked      []string
	volumesFrom []string
	net         string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	for _, name := range d.all {
		if name == needle {
			return true
		}
	}
	return false
}

// mustRun checks whether the given needle needs
// to be running
func (d *Dependencies) mustRun(needle string) bool {
	if needle == d.net {
		return true
	}
	for _, name := range d.linked {
		if name == needle {
			return true
		}
	}
	return false
}

// satisfied is true when there are no
// dependencies
func (d *Dependencies) satisfied() bool {
	return len(d.all) == 0
}

// remove removes the given name from all
func (d *Dependencies) remove(resolved string) {
	for i, name := range d.all {
		if name == resolved {
			d.all = append(d.all[:i], d.all[i+1:]...)
		}
	}
}
