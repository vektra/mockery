package internal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/chigopher/pathlib"
	"github.com/rs/zerolog"
	"github.com/xeipuuv/gojsonschema"
)

func httpsGet(ctx context.Context, url string) (string, error) {
	log := zerolog.Ctx(ctx)
	//nolint: gosec
	response, err := http.Get(url)
	if err != nil {
		log.Err(err).Msg("failed to download file")
		return "", fmt.Errorf("downloading file: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		log.Err(err).Msg("failed to read response body")
		return "", fmt.Errorf("reading response body: %w", err)
	}
	if response.StatusCode != 200 {
		log.Debug().Int("status-code", response.StatusCode).Msg("got non-200 response when downloading file. Logging response body:")
		log.Debug().Msg(string(body))
		return "", fmt.Errorf("got http code %d: %w", response.StatusCode, errBadHTTPStatus)
	}
	return string(body), nil
}

func download(ctx context.Context, url string) (string, error) {
	if strings.HasPrefix(url, "file://") {
		templatePath := pathlib.NewPath(strings.SplitAfterN(url, "file://", 2)[1])
		b, err := templatePath.ReadFile()
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		templateString, err := httpsGet(ctx, url)
		if err != nil {
			return "", fmt.Errorf("downloading url: %w", err)
		}
		return templateString, nil
	}
	return "", fmt.Errorf("unsupported protocol specifier in %s", url)
}

type RemoteTemplateCache struct {
	cache map[string]*RemoteTemplate
	mu    sync.Mutex
}

func NewRemoteTemplateCache() *RemoteTemplateCache {
	return &RemoteTemplateCache{
		cache: make(map[string]*RemoteTemplate),
	}
}

func (c *RemoteTemplateCache) Get(templateURL, schemaURL string) *RemoteTemplate {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, ok := c.cache[templateURL]; !ok {
		template := NewRemoteTemplate(templateURL, schemaURL)
		c.cache[templateURL] = template
		return template
	} else {
		return cached
	}
}

type RemoteTemplate struct {
	templateURL    string
	templateString string
	templateOnce   sync.Once
	templateErr    error

	schemaURL  string
	schema     *gojsonschema.Schema
	schemaOnce sync.Once
	schemaErr  error

	requireSchemaExists bool
}

func NewRemoteTemplate(templateURL, schemaURL string) *RemoteTemplate {
	return &RemoteTemplate{
		templateURL: templateURL,
		schemaURL:   schemaURL,
	}
}

// Template will return the template string. It downloads the remote template once
// and caches the result for future calls.
func (r *RemoteTemplate) Template(ctx context.Context) (string, error) {
	r.templateOnce.Do(func() {
		var err error
		r.templateString, err = download(ctx, r.templateURL)
		if err != nil {
			r.templateErr = fmt.Errorf("downloading template: %w", err)
		}
	})

	return r.templateString, r.templateErr
}

// Schema returns the JSON Schema as a string. It downloads the remote schema once
// and caches the result for future calls.
func (r *RemoteTemplate) Schema(ctx context.Context) (*gojsonschema.Schema, error) {
	log := zerolog.Ctx(ctx)
	log.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("remote-template", r.templateURL)
	})

	r.schemaOnce.Do(func() {
		log.Debug().Msg("schema not downloaded before")
		schemaString, err := download(ctx, r.schemaURL)
		if err != nil {
			log.Debug().Err(err).Msg("schema download encountered error")
			r.schemaErr = fmt.Errorf("downloading schema: %w", err)
			return
		}
		r.schema, err = gojsonschema.NewSchema(gojsonschema.NewStringLoader(schemaString))
		if err != nil {
			r.schemaErr = fmt.Errorf("creating JSON schema: %w", err)
		}
	})

	return r.schema, r.schemaErr
}
