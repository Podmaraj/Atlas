package transform

import (
	"compress/gzip"
	"net/http"
	"strings"

	"edgecore/internal/gateway/pipeline"
	"edgecore/internal/models"
)

type CompressionPlugin struct {
	minBytes int
}

func NewCompressionPlugin() *CompressionPlugin {
	return &CompressionPlugin{
		minBytes: 1024, // 1KB threshold
	}
}

func (p *CompressionPlugin) Name() string {
	return "compression"
}

func (p *CompressionPlugin) Init(config models.JSONMap) error {
	if min, ok := config["min_bytes"].(float64); ok && min > 0 {
		p.minBytes = int(min)
	}
	return nil
}

func (p *CompressionPlugin) ExecuteRequest(ctx *pipeline.PipelineContext) error {
	acceptEnc := ctx.Request.Header.Get("Accept-Encoding")
	if strings.Contains(acceptEnc, "gzip") {
		ctx.SetMetadata("enable_gzip", true)
		gzipWriter := gzip.NewWriter(ctx.Writer)
		ctx.Writer = &gzipResponseWriter{
			ResponseWriter: ctx.Writer,
			writer:         gzipWriter,
		}
	}
	return nil
}

func (p *CompressionPlugin) ExecuteResponse(ctx *pipeline.PipelineContext) error {
	if enabled, ok := ctx.GetMetadata("enable_gzip"); ok && enabled.(bool) {
		if grw, ok := ctx.Writer.(*gzipResponseWriter); ok {
			grw.writer.Close()
		}
	}
	return nil
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Del("Content-Length")
	return g.writer.Write(b)
}

func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Del("Content-Length")
	g.ResponseWriter.WriteHeader(statusCode)
}
