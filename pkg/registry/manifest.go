package registry

import (
	"time"
)

// see https://github.com/moby/moby/blob/8955d8da8951695a98eb7e15bead19d402c6eb27/image/tarexport/tarexport.go#L18
// https://docs.docker.com/registry/spec/manifest-v2-2/

type (
	Manifest struct {
		SchemaVersion int              `json:"schemaVersion"`
		MediaType     string           `json:"mediaType"`
		Config        ManifestRecord   `json:"config"`
		Layers        []ManifestRecord `json:"layers"`
	}

	ManifestRecord struct {
		MediaType string `json:"mediaType"`
		Size      int64  `json:"size"`
		Digest    string `json:"digest"`
	}

	ImageManifest struct {
		Config   string
		RepoTags []string
		Layers   []string
	}

	ImageManifests []ImageManifest

	ImageManifestConfig struct {
		Architecture string                    `json:"architecture"`
		OS           string                    `json:"os"`
		RootFS       ImageManifestConfigRootFS `json:"rootfs"`
		Config       map[string]interface{}    `json:"config"`
		Created      time.Time                 `json:"created"`
	}

	ImageManifestConfigRootFS struct {
		FSType  string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	}
)

const (
	manifestFile = "manifest.json"

	SchemaVersion        = 2
	ManifestMediaType    = "application/vnd.docker.distribution.manifest.v2+json"
	ImageConfigMediaType = "application/vnd.docker.container.image.v1+json"
	LayerMediaType       = "application/vnd.docker.image.rootfs.diff.tar.gzip"
)
