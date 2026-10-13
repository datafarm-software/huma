package huma

import (
	"fmt"
	"io"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/datafarm-software/datafarm-api/api/sensor/data"
)

type CsvMarshaller interface {
	Csv() (string, error)
}

func (h ApiError) Csv() (csvStr string, err error) {
	return fmt.Sprintf("Title, Status, Detail\n%s,%d,%s\n",
		h.Title, h.Status, h.Detail), nil
}

var CsvMediaType = &huma.MediaType{
	Schema: &huma.Schema{
		Type: "string",
	},
}

func EnableCsv(config *huma.Config) {
	config.DefaultFormat = "application/json"
	config.Formats["text/csv"] = huma.Format{
		Marshal: func(w io.Writer, v any) error {
			cm, ok := v.(data.CsvMarshaller)
			if !ok {
				return fmt.Errorf("csv marshal did not receive marshaller, got: %T", v)
			}
			csv, _ := cm.Csv()
			_, err := w.Write([]byte(csv))
			return err
		},
		Unmarshal: func(data []byte, v any) error {
			return fmt.Errorf("text/csv request bodies are not supported")
		},
	}
	defaultTransformer := huma.NewSchemaLinkTransformer("#/components/schemas/", "/schemas")
	config.Transformers = []huma.Transformer{
		func(ctx huma.Context, status string, v any) (any, error) {
			negotiatedContentType := ctx.Header("Accept")
			if !strings.Contains(negotiatedContentType, "text/csv") {
				return defaultTransformer.Transform(ctx, status, v)
			}
			if errM, ok := v.(*huma.ErrorModel); ok {
				return ApiError{
					Title:  errM.Title,
					Status: errM.Status,
					Detail: errM.Detail,
				}, nil
			}
			return v, nil
		},
	}
	config.CreateHooks = []func(huma.Config) huma.Config{
		func(c huma.Config) huma.Config {
			c.OnAddOperation = append(c.OnAddOperation, defaultTransformer.OnAddOperation)
			return c
		},
	}
}
