package myslave

func (m *MySlave) setupPredicate() {
	if len(m.dbAllowed)+len(m.dbExcluded) == 0 {
		m.Predicate = m.predicateAllowAll
	} else if len(m.dbAllowed) > 0 {
		m.Predicate = m.predicateUseAllow
	} else {
		m.Predicate = m.predicateUseExclude
	}
}

func (m *MySlave) predicateAllowAll(schema, table string) bool {
	return true
}

func (m *MySlave) predicateUseAllow(schema, table string) bool {
	if _, present := m.dbAllowed[schema]; present {
		return true
	}

	return false
}

func (m *MySlave) predicateUseExclude(schema, table string) bool {
	if _, excluded := m.dbExcluded[schema]; excluded {
		return false
	}

	return true
}
