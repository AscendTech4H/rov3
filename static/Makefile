# This Makefile downloads the web UI static files

# Versions of web resources
MATERIALIZE_VERSION = 0.100.2
MATERIAL_ICONS_VERSION = 3.0.1

# Files used by process
MATERIALIZE_FILES = materialize.min.js materialize.min.css
MATERIAL_ICON_EXTENSIONS = woff2 woff ttf
MATERIAL_ICON_FILES = material-icons.css $(foreach ext,$(MATERIAL_ICON_EXTENSIONS),MaterialIcons-Regular.$(ext))
MATERIAL_FONT_EXTENSIONS = woff2 woff
MATERIAL_FONT_VARIANTS = Bold Light Medium Regular Thin
MATERIAL_FONT_FILES = $(foreach variant,$(MATERIAL_FONT_VARIANTS),$(foreach ext,$(MATERIAL_FONT_EXTENSIONS),fonts/roboto/Roboto-$(variant).$(ext)))
FILES = $(MATERIALIZE_FILES) $(MATERIAL_ICON_FILES) $(MATERIAL_FONT_FILES)

# All target (default)
all: $(FILES)

# Materialize
materialize.min.css:
	wget https://cdnjs.cloudflare.com/ajax/libs/materialize/$(MATERIALIZE_VERSION)/css/$@
materialize.min.js:
	wget https://cdnjs.cloudflare.com/ajax/libs/materialize/$(MATERIALIZE_VERSION)/js/$@

# Material icons (license @ https://github.com/google/material-design-icons/blob/master/LICENSE)
MaterialIcons-Regular.%:
	wget https://github.com/google/material-design-icons/raw/$(MATERIAL_ICONS_VERSION)/iconfont/$@
material-icons.css:
	wget https://github.com/google/material-design-icons/raw/$(MATERIAL_ICONS_VERSION)/iconfont/$@

# Materialize roboto font
fonts:
	mkdir fonts
fonts/roboto: fonts
	mkdir fonts/roboto
fonts/roboto/%: fonts/roboto
	wget https://cdnjs.cloudflare.com/ajax/libs/materialize/0.100.2/$@ -O $@

# Clean target
clean:
	rm -r $(FILES) fonts
