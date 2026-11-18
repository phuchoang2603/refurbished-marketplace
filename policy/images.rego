package images

import rego.v1

deny contains msg if {
	some entry in input
	walk(entry.contents, [path, image])
	path[count(path) - 1] == "image"
	is_string(image)
	endswith(image, ":latest")
	msg := sprintf("image %s must not use the latest tag", [image])
}
