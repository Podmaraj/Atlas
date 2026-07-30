package transform

import (
	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

type TransformPlugin struct {
	addHeaders      map[string]string
	removeHeaders   []string
	addQueryParams  map[string]string
	removeQueryParams []string
}

func NewTransformPlugin() *TransformPlugin {
	return &TransformPlugin{
		addHeaders:        make(map[string]string),
		removeHeaders:     make([]string, 0),
		addQueryParams:    make(map[string]string),
		removeQueryParams: make([]string, 0),
	}
}

func (p *TransformPlugin) Name() string {
	return "request-transformer"
}

func (p *TransformPlugin) Init(config models.JSONMap) error {
	if addH, ok := config["add_headers"].(map[string]interface{}); ok {
		for k, v := range addH {
			if strVal, isStr := v.(string); isStr {
				p.addHeaders[k] = strVal
			}
		}
	}
	if remH, ok := config["remove_headers"].([]interface{}); ok {
		for _, v := range remH {
			if strVal, isStr := v.(string); isStr {
				p.removeHeaders = append(p.removeHeaders, strVal)
			}
		}
	}
	if addQ, ok := config["add_query_params"].(map[string]interface{}); ok {
		for k, v := range addQ {
			if strVal, isStr := v.(string); isStr {
				p.addQueryParams[k] = strVal
			}
		}
	}
	return nil
}

func (p *TransformPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	req := ctx.Request

	// Add/Overwrite headers
	for k, v := range p.addHeaders {
		req.Header.Set(k, v)
	}

	// Remove specified headers
	for _, k := range p.removeHeaders {
		req.Header.Del(k)
	}

	// Modify Query parameters
	if len(p.addQueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range p.addQueryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	return nil
}

func (p *TransformPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	return nil
}
