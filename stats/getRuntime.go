package stats

import "time"

// GetCoSchedCATRuntimes returns the runtime of application when running in parallel to cosched with CAT
func GetCoSchedCATRuntimes(application string, cosched string) *map[int]RuntimeT {
	temp, exists := runtimeStats.Runtimes[application]
	if !exists {
		return nil
	}

	if (*temp).CoSchedCATRuntimes == nil {
		return nil
	}

	ret, exists := (*(*temp).CoSchedCATRuntimes)[cosched]
	if exists {
		return &ret
	}

	return nil
}

// GetCoSchedCATRuntimes returns the runtime of application when running in parallel to cosched with CAT
func GetCoSchedCATRuntimesNormalized(application string, cosched string) *map[int]RuntimeT {
	ref := GetReferenceRuntime(application)
	cat := GetCoSchedCATRuntimes(application, cosched)
	if ref == nil || cat == nil {
		return nil
	}

	meanInNanoseconds := int64(ref.Mean * 1e+9)
	ret := make(map[int]RuntimeT)
	for catKey, catR := range *cat {
		ret[catKey] = normalizeRuntimeT(catR, meanInNanoseconds)
	}

	return &ret
}

// GetCoSchedRuntimes returns the runtime of application when running in parallel to cosched without CAT
func GetCoSchedRuntimes(application string, cosched string) *RuntimeT {
	temp, exists := runtimeStats.Runtimes[application]
	if !exists {
		return nil
	}

	if (*temp).CoSchedRuntimes == nil {
		return nil
	}

	ret, exists := (*temp.CoSchedRuntimes)[cosched]
	if !exists {
		return nil
	}

	return &ret
}

// GetCoSchedRuntimesNormalized returns the normalized runtime of application when running in parallel to cosched without CAT
func GetCoSchedRuntimesNormalized(application string, cosched string) *RuntimeT {
	ref := GetReferenceRuntime(application)
	co := GetCoSchedRuntimes(application, cosched)
	if ref == nil || co == nil {
		return nil
	}

	meanInNanoseconds := int64(ref.Mean * 1e+9)
	ret := normalizeRuntimeT(*co, meanInNanoseconds)
	return &ret
}

// GetCATRuntimes returns all cat individual runtimes with CAT
func GetCATRuntimes(application string) *map[int]RuntimeT {
	_, exists := runtimeStats.Runtimes[application]
	if exists {
		return runtimeStats.Runtimes[application].CATRuntimes
	}
	return nil
}

// GetCATRuntimesNormalized returns all cat individual runtimes with CAT normalized
func GetCATRuntimesNormalized(application string) *map[int]RuntimeT {
	ref := GetReferenceRuntime(application)
	cat := GetCATRuntimes(application)
	if ref == nil || cat == nil {
		return nil
	}

	meanInNanoseconds := int64(ref.Mean * 1e+9)
	ret := make(map[int]RuntimeT)
	for catKey, catR := range *cat {
		ret[catKey] = normalizeRuntimeT(catR, meanInNanoseconds)
	}

	return &ret
}

// GetReferenceRuntime returns the individual runtime without CAT
func GetReferenceRuntime(application string) *RuntimeT {
	_, exists := runtimeStats.Runtimes[application]
	if exists {
		return &runtimeStats.Runtimes[application].ReferenceRuntimes
	}
	return nil
}

// GetReferenceRuntimeNormalized returns the individual runtime without CAT normalized
func GetReferenceRuntimeNormalized(application string) *RuntimeT {
	ref := GetReferenceRuntime(application)
	if ref == nil {
		return nil
	}
	meanInNanoseconds := int64(ref.Mean * 1e+9)
	ret := normalizeRuntimeT(*ref, meanInNanoseconds)
	return &ret
}

func normalizeRuntimeT(ref RuntimeT, meanInNanoseconds int64) RuntimeT {
	var rs []DataPerRun
	for _, v := range *ref.RawRuntimesByMask {
		for _, t := range v {
			var r DataPerRun
			r.Output = t.Output
			r.Runtime = time.Duration(int64(t.Runtime) / meanInNanoseconds)
			rs = append(rs, r)
		}
	}

	ret := newRuntimeT(NoCATMask, rs)
	return ret
}
