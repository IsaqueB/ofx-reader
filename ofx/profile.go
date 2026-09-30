package ofx

// BankProfile is an extension point for bank-specific quirks.
// The generic parser intentionally works without any profiles. A future
// profile can match a statement/document and normalize fields without changing
// the parser or the public API.
type BankProfile interface {
	Name() string
	Match(header Header, statement BankStatement) bool
	Normalize(statement *BankStatement) error
}

// ProfileRegistry stores bank profiles in matching order.
type ProfileRegistry struct {
	profiles []BankProfile
}

func NewProfileRegistry(profiles ...BankProfile) *ProfileRegistry {
	r := &ProfileRegistry{}
	r.profiles = append(r.profiles, profiles...)
	return r
}

func (r *ProfileRegistry) Register(profile BankProfile) {
	if profile != nil {
		r.profiles = append(r.profiles, profile)
	}
}

func (r *ProfileRegistry) apply(header Header, statement *BankStatement) (string, error) {
	if r == nil || statement == nil {
		return "", nil
	}
	for _, p := range r.profiles {
		if p.Match(header, *statement) {
			if err := p.Normalize(statement); err != nil {
				return p.Name(), err
			}
			return p.Name(), nil
		}
	}
	return "", nil
}
