package doppler

import rego.v1

deny contains msg if {
	some entry in input
	doc := entry.contents
	regex.match(`DOPPLER_(CONFIG|PROJECT|TOKEN)`, json.marshal(doc))
	msg := sprintf("%s %s must not set Doppler config in Helm or Argo values", [doc.kind, doc.metadata.name])
}
