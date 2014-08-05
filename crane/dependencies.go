package crane

// Dependencies contains 4 fields:
// all: contains all dependencies
// link: containers linked to
// volumesFrom: containers that provide volumes
// net: container the net stack is shared with
type Dependencies struct {
	all         []string
	link        []string
	volumesFrom []string
	net         string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	return d.includesAsKind(needle, "all")
}

// includes checks whether the given needle is
// included in the dependency list as the given kind
func (d *Dependencies) includesAsKind(needle string, kind string) bool {
	for _, name := range d.forKind(kind) {
		if name == needle {
			return true
		}
	}
	return false
}

// returns the list of dependencies for a certain
// kind of dependency
func (d *Dependencies) forKind(kind string) []string {
	switch kind {
	case "all":
		return d.all
	case "link":
		return d.link
	case "volumesFrom":
		return d.volumesFrom
	case "net":
		return []string{d.net}
	default:
		return []string{}
	}
}

// mustRun checks whether the given needle needs
// to be running
func (d *Dependencies) mustRun(needle string) bool {
	if needle == d.net {
		return true
	}
	for _, name := range d.link {
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
