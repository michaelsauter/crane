package crane

// Dependencies contains 4 fields:
// all: contains all dependencies
// link: containers linked to
// volumesFrom: containers that provide volumes
// net: container the net stack is shared with
type Dependencies struct {
	All         []string
	Link        []string
	VolumesFrom []string
	Net         string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	for _, name := range d.All {
		if name == needle {
			return true
		}
	}
	return false
}

// mustRun checks whether the given needle needs
// to be running in order to be satisfied.
func (d *Dependencies) mustRun(needle string) bool {
	if needle == d.Net {
		return true
	}
	for _, name := range d.Link {
		if name == needle {
			return true
		}
	}
	return false
}

// satisfied is true when there are no
// dependencies left.
func (d *Dependencies) satisfied() bool {
	return len(d.All) == 0
}

// remove removes the given name from All
func (d *Dependencies) remove(resolved string) {
	for i, name := range d.All {
		if name == resolved {
			d.All = append(d.All[:i], d.All[i+1:]...)
		}
	}
}
