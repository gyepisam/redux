# build html from markdown
redo-ifchange $2.md
blackfriday-tool -css github.css $2.md $3
