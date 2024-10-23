build_push:
	git add . && git commit -m '$(msg)' && gith psuh && git tag $(tag) && git push origin $(tag)
