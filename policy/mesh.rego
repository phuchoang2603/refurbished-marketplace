package mesh

import rego.v1

namespace_wide contains doc if {
	some entry in input
	doc := entry.contents
	doc.kind == "PeerAuthentication"
	doc.metadata.namespace == "ecommerce"
	not doc.spec.selector
}

deny contains "ecommerce needs a namespace-wide PeerAuthentication with mTLS mode STRICT" if {
	not any_strict
}

any_strict if {
	some doc in namespace_wide
	doc.spec.mtls.mode == "STRICT"
}

deny contains msg if {
	some doc in namespace_wide
	doc.spec.mtls.mode != "STRICT"
	msg := sprintf("PeerAuthentication %s must be STRICT, not %s", [doc.metadata.name, doc.spec.mtls.mode])
}
