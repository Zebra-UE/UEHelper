package referencechainsearch

import "regexp"

func DumpChain() {

}

func FindPreventingReference(line string) {
	//(root) (NeverGCed)  GCObjectReferencer /Engine/Transient.GCObjectReferencer_2147482645
	re := regexp.MustCompile(`\((\w+)\)\s+\((\w+)\)\s+\w+\s+(/.+)$`)

	re.FindStringSubmatch()

}
