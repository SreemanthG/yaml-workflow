package yaml

import (
	"bytes"
	"io/ioutil"
	"unicode"

	"github.com/hashicorp/go-hclog"
	"github.com/lyraproj/issue/issue"
	"github.com/lyraproj/pcore/pcore"
	"github.com/lyraproj/pcore/px"
	"github.com/lyraproj/pcore/pximpl"
	"github.com/lyraproj/servicesdk/grpc"
	"github.com/lyraproj/servicesdk/service"
	"github.com/lyraproj/servicesdk/serviceapi"
)

const ManifestLoaderID = `Yaml::ManifestLoader`

type manifestLoader struct {
	ctx         px.Context
	serviceName string
}

type manifestService struct {
	ctx     px.Context
	service serviceapi.Service
}

func (m *manifestService) Invoke(identifier, name string, arguments ...px.Value) px.Value {
	return m.service.Invoke(m.ctx.Fork(), identifier, name, arguments...)
}

func (m *manifestService) Metadata() (px.TypeSet, []serviceapi.Definition) {
	return m.service.Metadata(m.ctx.Fork())
}

func (m *manifestService) State(name string, parameters px.OrderedMap) px.PuppetObject {
	return m.service.State(m.ctx.Fork(), name, parameters)
}

func WithService(serviceName string, sf func(c px.Context, s serviceapi.Service)) {
	rt := pximpl.InitializeRuntime()
	rt.SetLogger(grpc.NewHclogLogger(hclog.Default()))

	pcore.Do(func(c px.Context) {
		c.DoWithLoader(service.FederatedLoader(c.Loader()), func() {
			sb := service.NewServiceBuilder(c, serviceName)
			sb.RegisterApiType(`Yaml::Service`, &manifestService{})
			sb.RegisterAPI(ManifestLoaderID, &manifestLoader{c, serviceName})
			s := sb.Server()
			c.Set(`Yaml::ServiceLoader`, s)
			sf(c, s)
		})
	})
}

func Start(serviceName string) {
	WithService(serviceName, func(c px.Context, s serviceapi.Service) {
		grpc.Serve(c, s)
	})
}

func (m *manifestLoader) LoadManifest(moduleDir string, fileName string) serviceapi.Definition {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(px.Error(px.UnableToReadFile, issue.H{`path`: fileName, `detail`: err.Error()}))
	}

	c := m.ctx
	mf := munged(fileName)
	sb := service.NewServiceBuilder(c, mf)

	sb.RegisterStateConverter(ResolveState)
	sb.RegisterStep(CreateStep(c, fileName, content))
	s, _ := m.ctx.Get(`Yaml::ServiceLoader`)
	return s.(*service.Server).AddApi(mf, &manifestService{c, sb.Server()})
}

func munged(path string) string {
	b := bytes.NewBufferString(``)
	pu := true
	ps := true
	for _, c := range path {
		if c == '/' {
			if !ps {
				b.WriteString(`::`)
				ps = true
			}
		} else if c == '_' || c >= '0' && c <= '9' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' {
			if ps || pu {
				// First character of the name must be an upper case letter
				if ps && (c == '_' || c >= '0' && c <= '9') {
					// Must insert extra character
					b.WriteRune('X')
				} else {
					c = unicode.ToUpper(c)
				}
			}
			b.WriteRune(c)
			ps = false
			pu = false
		} else {
			pu = true
		}
	}
	if ps {
		b.Truncate(b.Len() - 2)
	}
	return b.String()
}
