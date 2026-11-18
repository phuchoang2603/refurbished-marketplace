package images

import rego.v1

third_party_images := {"quay.io/opstree/redis", "getmeili/meilisearch", "busybox"}

docs contains doc if {
	some entry in input
	doc := entry.contents
	not is_test_hook(doc)
}

is_test_hook(doc) if contains(doc.metadata.annotations["helm.sh/hook"], "test")

images contains image if {
	some doc in docs
	walk(doc, [path, image])
	path[count(path) - 1] == "image"
	is_string(image)
}

repository(image) := split(split(image, "@")[0], ":")[0]

allowed_registry(image) if startswith(image, "ghcr.io/")

allowed_registry(image) if repository(image) in third_party_images

deny contains msg if {
	some image in images
	not allowed_registry(image)
	msg := sprintf("image %s must come from ghcr.io", [image])
}

deny contains msg if {
	some image in images
	not regex.match(`(:[^/:@]+|@sha256:[0-9a-f]+)$`, image)
	msg := sprintf("image %s needs an explicit tag", [image])
}

deny contains msg if {
	some image in images
	endswith(image, ":latest")
	msg := sprintf("image %s must not use the latest tag", [image])
}
