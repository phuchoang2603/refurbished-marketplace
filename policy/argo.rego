package argo

import rego.v1

deny contains "Namespace ecommerce must be enrolled in Istio ambient" if {
	not ambient_ecommerce
}

ambient_ecommerce if {
	some entry in input
	doc := entry.contents
	doc.kind == "Namespace"
	doc.metadata.name == "ecommerce"
	doc.metadata.labels["istio.io/dataplane-mode"] == "ambient"
}
