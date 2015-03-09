package crane

// Dependencies contains 5 fields:
// all: contains all dependencies
// required: contains all non-optional dependencies
// link: containers linked to
// volumesFrom: containers that provide volumes
// net: container the net stack is shared with
type Dependencies struct {
	All         []string
	Required    []string
	Link        []string
	VolumesFrom []string
	Net         string
}

// includes checks whether the given needle is
// included in the dependency list
func (d *Dependencies) includes(needle string) bool {
	return d.includesAsKind(needle, "all")
}

// includesAsKind checks whether the given needle is
// included in the dependency list as the given kind
func (d *Dependencies) includesAsKind(needle string, kind string) bool {
	for _, name := range d.forKind(kind) {
		if name == needle {
			return true
		}
	}
	return false
}

// forKind returns the list of dependencies for
// a certain kind of dependency
func (d *Dependencies) forKind(kind string) []string {
	switch kind {
	case "all":
		return d.All
	case "required":
		return d.Required
	case "link":
		return d.Link
	case "volumesFrom":
		return d.VolumesFrom
	case "net":
		return []string{d.Net}
	default:
		return []string{}
	}
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

// satisfied is true when there are no required
// dependencies left.
func (d *Dependencies) satisfied() bool {
	return len(d.Required) == 0
}

// register a dependency
func (d *Dependencies) add(name string, kind string, required bool) {
	if !d.includesAsKind(name, "all") {
		d.All = append(d.All, name)
	}
	if required && !d.includesAsKind(name, "required") {
		d.Required = append(d.Required, name)
	}
	if !d.includesAsKind(name, kind) {
		switch kind {
		case "link":
			d.Link = append(d.Link, name)
		case "volumesFrom":
			d.VolumesFrom = append(d.VolumesFrom, name)
		case "net":
			d.Net = name
		}
	}
}

// remove removes the given name from Required
func (d *Dependencies) remove(resolved string) {
	for i, name := range d.Required {
		if name == resolved {
			d.Required = append(d.Required[:i], d.Required[i+1:]...)
		}
	}
}
