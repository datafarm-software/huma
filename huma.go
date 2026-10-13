package huma

import (
	"context"
	"net/http"
	"os/exec"
	"reflect"
	"regexp"

	"github.com/danielgtaylor/huma/v2"
)

type Mode int

const (
	Production Mode = iota
	Development
)

const MajorApiVersionRoutePrefix = "/v1"

type ApiError struct {
	Schema string `json:"$schema" doc:"Link to Object Model."`
	Title  string `json:"title" doc:"Name associated with error code."`
	Status int    `json:"status" doc:"Http Status Code."`
	Detail string `json:"detail" doc:"Human Readable explanation of what went wrong."`
}

var ApiErrorRegistry huma.Registry = huma.NewMapRegistry("#/errors", huma.DefaultSchemaNamer)

func (h *ApiError) MediaType() *huma.MediaType {
	return &huma.MediaType{
		Schema:  huma.SchemaFromType(ApiErrorRegistry, reflect.TypeFor[ApiError]()),
		Example: h,
	}
}

func FiveHundredExample() *ApiError {
	return &ApiError{
		Schema: "http://localhost:3030/schemas/ErrorModel.json",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
	}
}

func FiveHundredResponse() *huma.Response {
	return &huma.Response{
		Description: "Internal Server Error",
		Content: map[string]*huma.MediaType{
			"application/json": FiveHundredExample().MediaType(),
		},
	}
}

func FourHundredExample() *ApiError {
	return &ApiError{
		Schema: "http://localhost:3030/schemas/ErrorModel.json",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: "No Authorization Header Provided.",
	}
}

func FourHundredResponse() *huma.Response {
	return &huma.Response{
		Description: "Bad Request",
		Content: map[string]*huma.MediaType{
			"application/json": FourHundredExample().MediaType(),
		},
	}
}

func FourOhOneExample() *ApiError {
	return &ApiError{
		Schema: "http://localhost:3030/schemas/ErrorModel.json",
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: "Unknown user.",
	}
}

func FourOhOneResponse() *huma.Response {
	return &huma.Response{
		Description: "Unauthorized",
		Content: map[string]*huma.MediaType{
			"application/json": FourOhOneExample().MediaType(),
		},
	}
}

func FourOhFourExample() *ApiError {
	return &ApiError{
		Schema: "http://localhost:3030/schemas/ErrorModel.json",
		Title:  "Not Found",
		Status: http.StatusUnauthorized,
		Detail: "Device Not Found.",
	}
}

func FourOhFourResponse() *huma.Response {
	return &huma.Response{
		Description: "Not Found",
		Content: map[string]*huma.MediaType{
			"application/json": FourOhFourExample().MediaType(),
		},
	}
}

var (
	Basic  = []map[string][]string{{"basic": {}}}
	Bearer = []map[string][]string{{"bearer": {}}}
)

func BaseOperation(method string, middlewares *huma.Middlewares) huma.Operation {
	op := huma.Operation{
		RequestBody: &huma.RequestBody{},
		Responses: map[string]*huma.Response{
			"500": FiveHundredResponse(),
			"400": FourHundredResponse(),
			"401": FourOhOneResponse(),
			"404": FourOhFourResponse(),
		},
		Method:   method,
		Tags:     []string{method},
		Security: Bearer,
	}
	if middlewares != nil {
		op.Middlewares = *middlewares
	}
	return op
}

func DeviceIdParam() *huma.Param {
	return &huma.Param{
		Name:     "deviceId",
		In:       "path",
		Required: true,
		Schema: &huma.Schema{
			Type:    "string",
			Pattern: `^\w{1,30}$`,
		},
	}
}

type HumaHandler[I, O any] func(context.Context, *I) (*O, error)

type OperationHandler struct {
	HumaHandler[any, any]
	Operation huma.Operation
}

func RegisterHumaOperations(api huma.API, mode Mode, handlers ...OperationHandler) {
	if mode == Production {
		api.OpenAPI().Servers = append(api.OpenAPI().Servers, &huma.Server{
			URL: MajorApiVersionRoutePrefix,
		})
	}
	for _, h := range handlers {
		huma.Register(api, h.Operation, h.HumaHandler)
	}
}

var versionRegex = regexp.MustCompile(`[\d]+.[\d]+.[\d]+`)

func FindVersion() string {
	var cmd *exec.Cmd
	cmd = exec.Command("git", "describe", "--tags", "--abbrev=0")
	stdOut, _ := cmd.Output()
	verStr := string(stdOut)
	if verStr == "" {
		cmd = exec.Command("cat", "/app/.git-version")
		stdOut, _ = cmd.Output()
		verStr = string(stdOut)
	}
	verStr = versionRegex.FindString(verStr)
	if verStr == "" {
		// NOTE: if all else fails, fallback to v1
		verStr = "1"
	}
	return verStr
}
