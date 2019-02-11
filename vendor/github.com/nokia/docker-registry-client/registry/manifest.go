package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	digest "github.com/opencontainers/go-digest"
)

// fetchManifest sends a HTTP request to the URL of the manifest addressed by repository/reference.
// httpVerb is the HTTP verb to send: GET or HEAD. acceptedMediaTypes is the list of acceptable/expected media types.
// Returns with the HTTP response.
func (registry *Registry) fetchManifest(repository, reference string, httpVerb string, acceptedMediaTypes []string) (resp *http.Response, err error) {

	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.%s url=%s repository=%s reference=%s", strings.ToLower(httpVerb), url, repository, reference)

	req, err := http.NewRequest(httpVerb, url, nil)
	if err != nil {
		return nil, err
	}

	for i := range acceptedMediaTypes {
		if i == 0 {
			req.Header.Set("Accept", acceptedMediaTypes[i])
		} else {
			req.Header.Add("Accept", acceptedMediaTypes[i])
		}
	}
	resp, err = registry.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// manifest downloads any type of manifest addressed by repository/reference.
// acceptedMediaTypes is the list of acceptable/expected mediatypes (shouldn't be empty).
func (registry *Registry) manifest(repository, reference string, acceptedMediaTypes []string) (distribution.Manifest, error) {
	resp, err := registry.fetchManifest(repository, reference, "GET", acceptedMediaTypes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	actualMediaType := resp.Header.Get("Content-Type")
	acceptable := false
	for i := range acceptedMediaTypes {
		if actualMediaType == acceptedMediaTypes[i] {
			acceptable = true
			break
		}
	}
	if !acceptable && actualMediaType != manifestlist.MediaTypeManifestList {
		// NOTE: a manifest list may be legally returned even it wasn't asked for
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: %s (registry may not support the manifest type(s) you want: %v)", actualMediaType, acceptedMediaTypes)
	}
	decoder := json.NewDecoder(resp.Body)
	switch actualMediaType {

	case schema1.MediaTypeSignedManifest:
		deserialized := &schema1.SignedManifest{}
		err = decoder.Decode(&deserialized)
		if err != nil {
			return nil, err
		}
		return deserialized, nil

	case schema2.MediaTypeManifest:
		deserialized := &schema2.DeserializedManifest{}
		err = decoder.Decode(&deserialized)
		if err != nil {
			return nil, err
		}
		if deserialized.MediaType != actualMediaType {
			return nil, fmt.Errorf("mediaType in manifest should be '%s' not '%s'", actualMediaType, deserialized.MediaType)
		}
		return deserialized, nil

	case manifestlist.MediaTypeManifestList:
		deserialized := &manifestlist.DeserializedManifestList{}
		err = decoder.Decode(&deserialized)
		if err != nil {
			return nil, err
		}
		if deserialized.MediaType != actualMediaType {
			return nil, fmt.Errorf("mediaType in manifest should be '%s' not '%s'", actualMediaType, deserialized.MediaType)
		}
		if acceptable {
			return deserialized, nil
		}

		// if `reference` is an image digest, a manifest list may be received even if a schema2 Manifest was requested.
		// (since the Docker Image Digest is often the hash of a manifest list (a.k.a. fat manifest), not a schema2 manifest)
		// if that's the case: select and unwrap the default referred manifest in this case
		if len(deserialized.Manifests) == 0 {
			return nil, fmt.Errorf("empty manifest list: repository=%s reference=%s", repository, reference)
		}

		// use linux/amd64 manifest by default
		// TODO: query current platform architecture, OS and Variant and use those as selection criteria
		for _, m := range deserialized.Manifests {
			if strings.ToLower(m.Platform.Architecture) == "amd64" && strings.ToLower(m.Platform.OS) == "linux" {
				// address the manifest explicitly with its digest
				return registry.manifest(repository, m.Digest.String(), acceptedMediaTypes)
			}
		}
		// fallback: use the first manifest in the list
		// NOTE: emptiness of the list was checked above
		return registry.manifest(repository, deserialized.Manifests[0].Digest.String(), acceptedMediaTypes)

	default:
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: %s", actualMediaType)
	}

}

// Manifest returns a schema1 or schema2 manifest addressed by repository/reference.
// The schema of the result depends on the answer of the registry, it is usually the
// native schema of the manifest.
func (registry *Registry) Manifest(repository, reference string) (distribution.Manifest, error) {
	mediaTypes := []string{schema2.MediaTypeManifest, schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest}
	return registry.manifest(repository, reference, mediaTypes)
}

// ManifestV1 returns with the schema1 manifest addressed by repository/reference.
// If the registry answers with any other manifest schema it returns an error.
func (registry *Registry) ManifestV1(repository, reference string) (*schema1.SignedManifest, error) {
	mediaTypes := []string{schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest}
	m, err := registry.manifest(repository, reference, mediaTypes)
	if err != nil {
		return nil, err
	}
	signedManifest, ok := m.(*schema1.SignedManifest)
	if !ok {
		mediaType, _, _ := m.Payload()
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: %s", mediaType)
	}
	return signedManifest, nil
}

// ManifestList returns with the ManifestList (a.k.a. fat manifest) addressed by repository/reference
// If the registry answers with any other manifest schema it returns an error.
func (registry *Registry) ManifestList(repository, reference string) (*manifestlist.DeserializedManifestList, error) {
	mediaTypes := []string{manifestlist.MediaTypeManifestList}
	m, err := registry.manifest(repository, reference, mediaTypes)
	if err != nil {
		return nil, err
	}
	deserializedManifest, ok := m.(*manifestlist.DeserializedManifestList)
	if !ok {
		mediaType, _, _ := m.Payload()
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: %s", mediaType)
	}
	return deserializedManifest, nil
}

// ManifestV2 returns with the schema2 manifest addressed by repository/reference
// If the registry answers with any other manifest schema it usually returns an error,
// except the following case.
// If reference is an image digest (sha256:...) and it is the hash of a manifestlist,
// then the method will return the first manifest with 'linux' OS and 'amd64' architecture,
// or in the absence thereof, the first manifest in the list. (Rationale: the Docker
// Image Digests returned by `docker images --digests` sometimes refer to manifestlists)
func (registry *Registry) ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error) {
	mediaTypes := []string{schema2.MediaTypeManifest}
	m, err := registry.manifest(repository, reference, mediaTypes)
	if err != nil {
		return nil, err
	}
	deserializedManifest, ok := m.(*schema2.DeserializedManifest)
	if !ok {
		mediaType, _, _ := m.Payload()
		return nil, fmt.Errorf("unexpected manifest schema was received from registry: %s", mediaType)
	}
	return deserializedManifest, nil
}

// ManifestDescriptor fills the returned Descriptor according to the Content-Type, Content-Length and
// Docker-Content-Digest headers of the HTTP Response to a Manifest query.
func (registry *Registry) manifestDescriptor(repository, reference string, acceptedMediaTypes []string) (distribution.Descriptor, error) {
	resp, err := registry.fetchManifest(repository, reference, "HEAD", acceptedMediaTypes)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	defer resp.Body.Close()

	var desc distribution.Descriptor
	desc.MediaType = resp.Header.Get("Content-Type")
	desc.Digest = digest.Digest(resp.Header.Get("Docker-Content-Digest"))
	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err == nil {
		// make sure desc.Size is zero in case of an error
		desc.Size = size
	}
	return desc, nil
}

// ManifestDescriptor fills the returned Descriptor according to the Content-Type, Content-Length and
// Docker-Content-Digest headers of the HTTP Response to a Manifest query.
func (registry *Registry) ManifestDescriptor(repository, reference string) (distribution.Descriptor, error) {
	mediaTypes := []string{manifestlist.MediaTypeManifestList, schema2.MediaTypeManifest, schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest}
	return registry.manifestDescriptor(repository, reference, mediaTypes)
}

// ManifestDigest returns the 'Docker-Content-Digest' field of the header when making a
// request to the schema1 manifest.
func (registry *Registry) ManifestDigest(repository, reference string) (digest.Digest, error) {
	mediaTypes := []string{}
	desc, err := registry.manifestDescriptor(repository, reference, mediaTypes)
	if err != nil {
		return "", err
	}
	return desc.Digest, nil
}

// ManifestV2Digest returns the 'Docker-Content-Digest' field of the header when making a
// request to the schema2 manifest.
func (registry *Registry) ManifestV2Digest(repository, reference string) (digest.Digest, error) {
	mediaTypes := []string{schema2.MediaTypeManifest}
	desc, err := registry.manifestDescriptor(repository, reference, mediaTypes)
	if err != nil {
		return "", err
	}
	return desc.Digest, nil
}

// DeleteManifest deletes the manifest given by its digest from the repository
func (registry *Registry) DeleteManifest(repository string, digest digest.Digest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, digest)
	registry.Logf("registry.manifest.delete url=%s repository=%s reference=%s", url, repository, digest)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}
	return nil
}

// PutManifest uploads manifest to the given repository/reference.
// Manifest is typically either of type schema2.DeserializedManifest or schema1.SignedManifest
func (registry *Registry) PutManifest(repository, reference string, manifest distribution.Manifest) error {
	url := registry.url("/v2/%s/manifests/%s", repository, reference)
	registry.Logf("registry.manifest.put url=%s repository=%s reference=%s", url, repository, reference)

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(payload)
	req, err := http.NewRequest("PUT", url, buffer)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", mediaType)
	resp, err := registry.Client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	return err
}

// PutManifestV2 uploads the given schama2.Manifest to the given repository/reference and returns with its digest.
// If you want to upload a schema2.DeserializedManifest, please use the generic PutManifest().
func (registry *Registry) PutManifestV2(repository, reference string, manifest *schema2.Manifest) (digest.Digest, error) {
	deserializedManifest, err := schema2.FromStruct(*manifest)
	if err != nil {
		return "", err
	}
	_, canonical, err := deserializedManifest.Payload()
	if err != nil {
		return "", err
	}
	digest := digest.FromBytes(canonical)
	err = registry.PutManifest(repository, reference, deserializedManifest)
	return digest, err
}
