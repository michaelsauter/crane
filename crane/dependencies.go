package crane

// Dependencies contain 3 fields:
// list: contains all dependencies
// linked: contains dependencies that
// are being linked to.
// net: container the net stack is shared with
type Dependencies struct {
	list   []string
	linked []string
	net    string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	for _, name := range d.list {
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

// satisfied is true when the list is empty
func (d *Dependencies) satisfied() bool {
	return len(d.list) == 0
}

// remove removes the given name from the list
func (d *Dependencies) remove(resolved string) {
	for i, name := range d.list {
		if name == resolved {
			d.list = append(d.list[:i], d.list[i+1:]...)
		}
	}
}
