package profiles

import (
	"embed"

	"github.com/goccy/go-yaml"
)

var (
	//go:embed included
	includedEmbed embed.FS

	includedProfiles []Profile
)

func GetIncluded() []Profile {
	if len(includedProfiles) > 0 {
		return includedProfiles
	}

	files, _ := includedEmbed.ReadDir("included")

	for i := range files {
		file, err := includedEmbed.Open("included/" + files[i].Name())
		if err != nil {
			panic(err)
		}
		p := Profile{}
		err = yaml.NewDecoder(file).Decode(&p)
		if err != nil {
			panic(err)
		}
		includedProfiles = append(includedProfiles, p)
	}

	return includedProfiles
}
