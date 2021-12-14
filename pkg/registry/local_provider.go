package registry

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"

	"git.backbone/corpix/unregistry/pkg/errors"
	"git.backbone/corpix/unregistry/pkg/server"
)

type (
	LocalProviderConfig struct {
		Containers map[string]string `yaml:"containers"`
	}
	LocalProviderManifestDescriptor struct {
		Hash    []byte // sha256 of payload
		Payload []byte
	}
	LocalProvider struct {
		config    LocalProviderConfig
		tags      map[string]string
		manifests map[string]*LocalProviderManifestDescriptor
		blobs     map[string]interface{}
	}
)

//

var _ Provider = new(LocalProvider)

//

func (p *LocalProvider) init() error {
	var err error

	for k, v := range p.config.Containers {
		err = p.preloadManifest(k, v)
		if err != nil {
			return errors.Wrapf(
				err, "failed to preload manifest %q from %q",
				k, v,
			)
		}
	}

	return nil
}

func (p *LocalProvider) preloadManifest(tag string, dir string) error {
	var (
		manifests    ImageManifests
		manifestPath = path.Join(dir, manifestFile)
		configPath   string
	)

	stream, err := os.Open(manifestPath)
	if err != nil {
		return errors.Wrap(err, "unable to open image manifest %q for reading")
	}
	defer stream.Close()

	//

	err = json.NewDecoder(stream).Decode(&manifests)
	if err != nil {
		return errors.Wrap(err, "failed to decode image manifest")
	}

	if len(manifests) < 1 {
		return errors.New("empty image manifest")
	}

	for index, manifest := range manifests {
		configPath = path.Join(dir, manifest.Config)

		configStream, err := os.Open(configPath)
		if err != nil {
			return errors.Wrapf(
				err, "unable to open image manifest config #%d %q for reading",
				index, configPath,
			)
		}

		config := ImageManifestConfig{}
		err = json.NewDecoder(configStream).Decode(&config)
		if err != nil {
			return errors.Wrapf(
				err, "unable to decode image manifest config #%d %q",
				index, configPath,
			)
		}

		//

		configBytes, err := json.Marshal(config)
		if err != nil {
			return errors.Wrapf(
				err, "unable to encode image manifest config for image manifest #%d %q",
				index, configPath,
			)
		}
		configBytesDigestHex := DigestHexString(configBytes)
		p.blobs[configBytesDigestHex] = configBytes

		layers := make([]ManifestRecord, len(manifest.Layers))
		for layerIndex, layer := range manifest.Layers {
			layerPath := path.Join(dir, layer)
			size, buf, err := DigestFile(layerPath)
			if err != nil {
				return errors.Wrapf(
					err, "unable to calculate digest of image layer listed in #%d %q, layer path %q",
					index, configPath, layerPath,
				)
			}

			layerDigestHex := hex.EncodeToString(buf)
			layers[layerIndex] = ManifestRecord{
				MediaType: LayerMediaType,
				Size:      size,
				Digest:    digestPrefix + layerDigestHex,
			}

			_, blobExists := p.blobs[layerDigestHex]
			if !blobExists {
				p.blobs[layerDigestHex] = layerPath
			}
		}

		registryManifest := Manifest{
			SchemaVersion: SchemaVersion,
			MediaType:     ManifestMediaType,
			Config: ManifestRecord{
				MediaType: ImageConfigMediaType,
				Size:      int64(len(configBytes)),
				Digest:    digestPrefix + configBytesDigestHex,
			},
			Layers: layers,
		}
		registryManifestBytes, err := json.Marshal(registryManifest)
		if err != nil {
			return errors.Wrapf(
				err, "unable to encode registry manifest for image manifest #%d %q",
				index, configPath,
			)
		}

		// save digest to in-memory store

		digest := Digest(registryManifestBytes)
		digestHex := hex.EncodeToString(digest)

		p.manifests[digestHex] = &LocalProviderManifestDescriptor{
			Hash:    digest,
			Payload: registryManifestBytes,
		}

		// alias tag to digest

		alias, tagExists := p.tags[tag]
		if tagExists {
			return errors.Errorf(
				"unable to map tag %q to image manifest config #%d %q: conflict, already mapped to %q",
				tag, index, configPath, alias,
			)
		}

		p.tags[tag] = digestHex
	}

	return nil
}

func (p *LocalProvider) tag(name string, reference string) string {
	return name + ":" + reference
}

//

func (p *LocalProvider) GetManifest(name string, reference string) (Stream, error) {
	switch {
	case strings.HasPrefix(reference, digestPrefix):
		manifest, ok := p.manifests[strings.TrimPrefix(reference, digestPrefix)]
		if !ok {
			return nil, NewErrNotFound(reference)
		}

		return io.NopCloser(bytes.NewBuffer(manifest.Payload)), nil
	default:
		tag := p.tag(name, reference)
		alias, ok := p.tags[tag]
		if !ok {
			return nil, NewErrNotFound(tag)
		}

		return p.GetManifest(name, digestPrefix+alias)
	}
}

func (p *LocalProvider) GetBlob(name string, reference string) (Stream, error) {
	switch {
	case strings.HasPrefix(reference, digestPrefix):
		blob, ok := p.blobs[strings.TrimPrefix(reference, digestPrefix)]
		if !ok {
			return nil, NewErrNotFound(reference)
		}

		switch v := blob.(type) {
		case string:
			stream, err := os.Open(v)
			if err != nil {
				return nil, server.NewError(
					server.StatusInternalServerError,
					"could not open layer blob",
					errors.Wrapf(
						err, "failed to open blob tar %q",
						blob,
					),
					nil,
				)
			}

			return stream, nil
		case []byte:
			return io.NopCloser(bytes.NewBuffer(v)), nil
		default:
			return nil, errors.Errorf("unsupported blob type %T", blob)
		}
	default:
		return nil, NewErrNotFound(reference)
	}

}

//

func NewLocalProvider(c LocalProviderConfig) (*LocalProvider, error) {
	p := &LocalProvider{
		config:    c,
		tags:      map[string]string{},
		manifests: map[string]*LocalProviderManifestDescriptor{},
		blobs:     map[string]interface{}{},
	}

	err := p.init()
	if err != nil {
		return nil, err
	}

	return p, nil
}
