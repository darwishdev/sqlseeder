build_push:
	git add .  && git commit -m '$(msg)' && git push && git tag $(tag) && git push origin $(tag)
